---
title: 一文记录 k8s 与 docker
date: 2023-12-20
categories: 
- 计算机基础
---

# k8s 网络
## linux 虚拟网络 veth pair 和 bridge
- Network namespace 实现网络隔离
- Veth pair提供了一种连接两个network namespace的方法
- Bridge 实现同一网络中多个namespace的连接
- 添加路由信息，查看路由信息
- iptabels 和 NAT
-  实战练习

``` sh
sudo ip netns add ns1
sudo ip netns add ns2
sudo ip netns add ns3

sudo brctl addbr virtual-bridge

sudo ip link add veth-ns1 type veth peer name veth-ns1-br
sudo ip link set veth-ns1 netns ns1
sudo brctl addif virtual-bridge veth-ns1-br

sudo ip link add veth-ns2 type veth peer name veth-ns2-br
sudo ip link set veth-ns2 netns ns2
sudo brctl addif virtual-bridge veth-ns2-br

sudo ip link add veth-ns3 type veth peer name veth-ns3-br
sudo ip link set veth-ns3 netns ns3
sudo brctl addif virtual-bridge veth-ns3-br


sudo ip -n ns1 addr add local 192.168.1.1/24 dev veth-ns1
sudo ip -n ns2 addr add local 192.168.1.2/24 dev veth-ns2
sudo ip -n ns3 addr add local 192.168.1.3/24 dev veth-ns3

sudo ip link set virtual-bridge up
sudo ip link set veth-ns1-br up
sudo ip link set veth-ns2-br up
sudo ip link set veth-ns3-br up
sudo ip -n ns1 link set veth-ns1 up
sudo ip -n ns2 link set veth-ns2 up
sudo ip -n ns3 link set veth-ns3 up

sudo ip netns delete ns1
sudo ip netns delete ns2
sudo ip netns delete ns3
sudo ip link set virtual-bridge down
sudo brctl delbr virtual-bridge

$ sudo ip netns exec ns1 ping 192.168.1.2
PING 192.168.1.2 (192.168.1.2): 56 data bytes
64 bytes from 192.168.1.2: seq=0 ttl=64 time=0.068 ms
--- 192.168.1.2 ping statistics ---
3 packets transmitted, 3 packets received, 0% packet loss
round-trip min/avg/max = 0.060/0.064/0.068 ms
$ sudo ip netns exec ns1 ping 192.168.1.3
PING 192.168.1.3 (192.168.1.3): 56 data bytes
64 bytes from 192.168.1.3: seq=0 ttl=64 time=0.055 ms
--- 192.168.1.3 ping statistics ---
3 packets transmitted, 3 packets received, 0% packet loss
round-trip min/avg/max = 0.055/0.378/1.016 ms
```

## docker 网络 和 docker0
- docker0网桥和缺省路由
- docker0
- route
- iptables 和 nat

``` sh
# 查看网桥
$ brctl show
bridge name	bridge id		STP enabled	interfaces
docker0		8000.02421557ce52	no		veth91e1730
							            vethc858a6a
# 查看docker 网络
docker network inspect bridge

# 查看container route信息
# 目的地址为172.17的网络不走route，其它走默认的172.17.0.1 route
$ docker exec busybox1 route -n
Kernel IP routing table
Destination     Gateway         Genmask         Flags Metric Ref    Use Iface
0.0.0.0         172.17.0.1      0.0.0.0         UG    0      0        0 eth0
172.17.0.0      0.0.0.0         255.255.0.0     U     0      0        0 eth0

# 查看iptables
# 出口不为0docker的流量都使用SNAT
$ sudo iptables -t nat -S | grep docker
-A POSTROUTING -s 172.17.0.0/16 ! -o docker0 -j MASQUERADE
-A DOCKER -i docker0 -j RETURN
```


## pod 网络
### pause 
- pause容器实现1个pod中多个container的网络共享
- Pause 用于实现容器之间共享网络，如果其中部分容器挂掉，其余容器网路正常工作
- https://github.com/kubernetes/kubernetes/blob/master/build/pause/linux/pause.c

``` sh
$ docker ps | grep etcd
8fd1337b0bf2   73deb9a3f702                "etcd --advertise-cl…"   3 hours ago      Up 3 hours                k8s_etcd_etcd-minikube_kube-system_94aa022caf543792dfcddf4a2ca05a30_0
1202ef34af2b   registry.k8s.io/pause:3.9   "/pause"                 3 hours ago      Up 3 hours                k8s_POD_etcd-minikube_kube-system_94aa022caf543792dfcddf4a2ca05a30_0

$ docker inspect 8fd1337b0bf2 | grep -i networkMode
$ docker inspect 8fd1337b0bf2 | grep -i networkMode
            "NetworkMode": "container:1202ef34af2b155e938cbe770870ba6c8edd3a57c88545a697816c340a6ce320",

```
### CNI 标准和插件
- CNI标准: https://github.com/containernetworking/cni
- CNI 插件:https://github.com/containernetworking/plugins


``` sh
    $ ls -l /opt/cni/bin/
    -rwxr-xr-x 1 root root 2660408 Nov  7  2023 bandwidth
    -rwxr-xr-x 1 root root 3018552 Nov  7  2023 bridge
    -rwxr-xr-x 1 root root 1984728 Nov  7  2023 cnitool
    -rwxr-xr-x 1 root root 7432152 Nov  7  2023 dhcp
    -rwxr-xr-x 1 root root 3096120 Nov  7  2023 firewall
    -rwxr-xr-x 1 root root 2250104 Nov  7  2023 host-local
    -rwxr-xr-x 1 root root 2775128 Nov  7  2023 ipvlan
    -rwxr-xr-x 1 root root 2305848 Nov  7  2023 loopback
    -rwxr-xr-x 1 root root 2799704 Nov  7  2023 macvlan
    -rwxr-xr-x 1 root root 2615256 Nov  7  2023 portmap
    -rwxr-xr-x 1 root root 2891096 Nov  7  2023 ptp
    -rwxr-xr-x 1 root root 2367288 Nov  7  2023 tuning
    -rwxr-xr-x 1 root root 2771032 Nov  7  2023 vlan
```

## service 网络
### 背景
- Zookeeper提供名字服务，pod自身实现负载均衡，RPC框架实现负载均衡
- Service 为 Pods 提供的固定 IP，其他服务可以通过 Service IP 找到提供服务的Endpoints。
- Service提供负载均衡。Service 由多个 Endpoints 组成，kubernetes 对组成 Service 的 Pods 提供的负载均衡方案，例如随机访问、robin 轮询等。
- 暂时将Pod等同于Endpoint

<p align="center">
  <img src="/images/k8s_services_background.png" width=600 height=350>
  <br/>
</p>

### 实现原理
- Service IP IP 由API server分配，写入etcd
- Etcd 中存储service和endpoints
- Controllermanager watch etcd的变换生成endpoints
- node 中的kube-proxy watch service 和 endpoints的变化

<p align="center">
  <img src="/images/k8s_services.png" width=600 height=350>
  <br/>
</p>


### kube-proxy 服务发现和负载均衡
- Order -> item 的流程
- 服务发现：[环境变量和DNS](https://kubernetes.io/zh-cn/docs/concepts/services-networking/service/#environment-variables)
- servicename.namespace.svc.cluster.local
- kub-proxy 通过watch etcd中service和endpoint的变更，维护本地的iptables/ipvs
- kub-proxy 通过转发规则实现service ip 到 pod ip的转发，通过规则实现负载均衡


<p align="center">
  <img src="/images/k8s_services_name_space_load_balacing.png" width=600 height=350>
  <br/>
</p>

### [service 类型](https://kubernetes.io/zh-cn/docs/concepts/services-networking/service/#loadbalancer)
- ClusterIP
- NodePort
- LoadBalancer


## ingress 网络
### 背景
- 集群外部访问集群内部资源？nodeport,loadbalancer。一个服务一个port或者一个外网IP，一个域名
- Ingress 是 Kubernetes 中的一种 API 对象，用于管理入站网络流量，基于域名和URL路径把用户的请求转发到对应的service
- ingress相当于七层负载均衡器，是k8s对反向代理的抽象
- ingress负载均衡，将请求自动负载到后端的pod

<p align="center">
  <img src="/images/k8s_ingress_background.png" width=600 height=600>
  <br/>
</p>

### 实现原理
- ingress 资源对象用于编写资源配置规则
- Ingress-controller 监听apiserver感知集群中service和pod的变化动态更新配置规则，并重载proxy反向代理的配置
- proxy反向代理负载均衡器，例如ngnix，接收并按照ingress定义的规则进行转发，常用的是ingress-nginx等，直接转发到pod中
<p align="center">
  <img src="/images/k8s_ingress.png" width=600 height=350>
  <br/>
</p>

支持的路由方式
- 通过使用路径规则。例如： /app1 路径映射到一个服务，将 /app2 路径映射到另一个服务。路径匹配支持精确匹配和前缀匹配两种方式。
- 基于主机的路由匹配。例如，可以将 app1.example.com 主机名映射到一个服务，将 app2.example.com 主机名映射到另一个服务。主机匹配也可以与路径匹配结合使用，实现更细粒度的路由控制。
- 其他条件的路由匹配：：请求方法（如 GET、POST）、请求头（如 Content-Type）、查询参数等。


# docker k8s 常用命令
``` shell
    # minikube
    minikube start
    minikube status
    minikube ssh

    # docker
    docker ps  # 查看所有正在运行的容器
    docker ps -a # 查看所有的容器，包括正在运行的和停止的

    # 用交互式的方式启动容器
    docker start -ai <容器名或容器ID>

    # 打开容器进行交互式终端对话框
    docker exec -it <容器名或容器ID> bash

    # 容器中执行命令
    docker exec <容器名或容器ID> ls

```

参考资料
- [分享PPT](https://github.com/wxquare/effective-resourses/blob/master/share/k8s%20%E7%BD%91%E7%BB%9C%E5%85%A5%E9%97%A8.pdf)
- [Minikube 环境安装](https://github.com/caicloud/kube-ladder/blob/master/tutorials/lab1-installation.md)
- [Kubectl 命令和集群体验](https://github.com/caicloud/kube-ladder/blob/master/tutorials/lab2-application-and-service.md)
- [Linux network namespace,veth,bridge 和 路由](https://www.zhaohuabing.com/post/2020-03-12-linux-network-virtualization/)
- [从0到1搭建linux虚拟网络](https://zhuanlan.zhihu.com/p/199298498)
- [Docker 网络：模拟docker网络](https://morningspace.github.io/tech/k8s-net-mimic-docker/)
- [Docker 网络：从docker0开始](https://morningspace.github.io/tech/k8s-net-docker0/)
- [Pod网络和pause容器](https://morningspace.github.io/tech/k8s-net-pod-1/)
- [认识CNI插件](https://morningspace.github.io/tech/k8s-net-cni/)
- [深度解读CNI：容器网络接口](https://mp.weixin.qq.com/s/_nzbZYpKlpw4jKd5MFpuzw)
- [官方文档：服务service](https://kubernetes.io/zh-cn/docs/concepts/services-networking/service/1)
- [创建service之后，k8s会发生什么](https://zhuanlan.zhihu.com/p/677236869)
- [探究k8s service iptables 路由规则](https://luckymrwang.github.io/2021/02/20/%E6%8E%A2%E7%A9%B6K8S-Service%E5%86%85%E9%83%A8iptables%E8%B7%AF%E7%94%B1%E8%A7%84%E5%88%99/)
- [官方文档：在minikube中使用nginx ingress 控制配置ingress](https://kubernetes.io/zh-cn/docs/tasks/access-application-cluster/ingress-minikube/)
- [官方文档：ingress](https://kubernetes.io/zh-cn/docs/concepts/services-networking/ingress/)
