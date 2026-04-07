---
title: 互联网基础设施：Kubernetes 与 Docker 实践
date: 2024-12-20
categories:
- 系统设计
tags:
- kubernetes
- docker
- 容器化
- 云原生
toc: true
---

## 速查导航

**阅读时间**: 50 分钟 | **难度**: ⭐⭐⭐⭐⭐ | **面试频率**: 极高

**核心考点速查**:
- [一、Docker 核心原理](#一docker-核心原理面试必问) - Docker vs VM / 镜像分层 / namespace / cgroup
- [二、Kubernetes 核心概念](#二kubernetes-核心概念5-分钟速记) - Pod / Deployment / Service / Ingress
- [三、K8s 网络原理](#三k8s-网络原理深度解析) - Pod 网络 / Service 网络 / Ingress 网络
- [四、Pod 生命周期与故障排查](#四pod-生命周期与故障排查) - OOMKilled / CrashLoopBackOff / ImagePullBackOff
- [五、实用 YAML 配置](#五实用-yaml-配置模板) - Deployment / Service / Ingress / ConfigMap / Secret
- [六、监控与告警](#六监控与告警) - Prometheus / Grafana / 常见指标
- [七、面试高频 20 题](#七面试高频-20-题) - 标准答案 + 追问应对

---

## 一、Docker 核心原理（面试必问）

### 1.1 Docker vs 虚拟机

**对比表格**：

| 维度 | Docker 容器 | 虚拟机（VM） |
|------|-----------|------------|
| **启动速度** | 秒级 | 分钟级 |
| **资源占用** | MB 级别 | GB 级别 |
| **隔离级别** | 进程级别（共享内核） | 操作系统级别（独立内核） |
| **性能** | 接近原生（无虚拟化开销） | 有虚拟化开销（5-10%） |
| **镜像大小** | MB 级别 | GB 级别 |
| **安全性** | 较弱（共享内核） | 较强（完全隔离） |
| **适用场景** | 微服务、CI/CD、快速部署 | 需要完全隔离、不同OS |

**架构对比图**：

```
容器架构                      虚拟机架构
┌─────────┐ ┌─────────┐      ┌─────────┐ ┌─────────┐
│  App A  │ │  App B  │      │  App A  │ │  App B  │
├─────────┤ ├─────────┤      ├─────────┤ ├─────────┤
│  Bins   │ │  Bins   │      │  Bins   │ │  Bins   │
├─────────┼─┴─────────┤      ├─────────┤ ├─────────┤
│    Docker Engine    │      │  Guest  │ │  Guest  │
├─────────────────────┤      │   OS    │ │   OS    │
│      Host OS        │      ├─────────┼─┴─────────┤
├─────────────────────┤      │    Hypervisor       │
│    Infrastructure   │      ├─────────────────────┤
└─────────────────────┘      │      Host OS        │
                              ├─────────────────────┤
                              │    Infrastructure   │
                              └─────────────────────┘
```

**面试标准答案**（30 秒）：
> Docker 容器通过 namespace 实现资源隔离，通过 cgroup 实现资源限制，共享宿主机内核，启动快、资源占用少。VM 通过 Hypervisor 虚拟化完整操作系统，隔离性更强但开销大。

### 1.2 Docker 镜像分层机制

**分层原理**：
- 每个 Docker 镜像由多层只读层（Read-Only Layer）叠加而成
- 运行容器时，在最上层加一层可写层（Container Layer）
- 多个容器可以共享相同的镜像层，节省磁盘空间

**示例**：

```dockerfile
FROM ubuntu:20.04        # 第 1 层：基础镜像
RUN apt-get update       # 第 2 层：更新软件源
RUN apt-get install -y nginx  # 第 3 层：安装 nginx
COPY app.conf /etc/nginx/  # 第 4 层：复制配置文件
CMD ["nginx", "-g", "daemon off;"]  # 第 5 层：启动命令
```

**分层存储结构**：

```
┌─────────────────┐ ← Container Layer（可写层）
├─────────────────┤
│ CMD nginx       │ ← Layer 5（只读）
├─────────────────┤
│ COPY app.conf   │ ← Layer 4（只读）
├─────────────────┤
│ RUN apt install │ ← Layer 3（只读）
├─────────────────┤
│ RUN apt update  │ ← Layer 2（只读）
├─────────────────┤
│ FROM ubuntu     │ ← Layer 1（只读）
└─────────────────┘
```

**面试追问：为什么要分层？**
1. **共享存储**：多个镜像可以共享相同的层（如 ubuntu:20.04 基础层）
2. **快速构建**：修改某一层时，只需重新构建该层及以上层
3. **快速分发**：只传输变化的层，不需要传输整个镜像

### 1.3 Namespace 隔离

**Docker 使用 Linux Namespace 实现资源隔离**：

| Namespace | 隔离内容 | 示例 |
|-----------|---------|------|
| **PID** | 进程 ID | 容器内 PID=1 的进程在宿主机上是另一个 PID |
| **NET** | 网络栈（网卡、IP、端口） | 容器有独立的 IP 地址和端口 |
| **IPC** | 进程间通信（消息队列、信号量） | 容器间无法直接 IPC |
| **MNT** | 文件系统挂载点 | 容器看到的是独立的文件系统 |
| **UTS** | 主机名和域名 | 容器有独立的 hostname |
| **USER** | 用户和用户组 | 容器内 root 可以映射为宿主机普通用户 |

**验证 Namespace**：

```bash
# 查看容器的 namespace
docker inspect <container-id> | grep -i pid
# 输出：PID: 12345

# 查看宿主机上的 namespace
sudo ls -l /proc/12345/ns/
# 输出：
# net -> 'net:[4026532123]'
# pid -> 'pid:[4026532124]'
```

### 1.4 Cgroup 资源限制

**Cgroup（Control Groups）用于限制容器的资源使用**：

| 资源 | 限制方式 | Docker 参数 |
|------|---------|------------|
| **CPU** | 限制 CPU 使用率 | `--cpus=2`（2 核）或 `--cpu-shares=512`（相对权重） |
| **内存** | 限制内存使用量 | `--memory=1g`（1GB） |
| **磁盘 IO** | 限制磁盘读写速度 | `--device-read-bps`、`--device-write-bps` |
| **网络** | 限制网络带宽 | 需借助 tc（traffic control）工具 |

**示例**：

```bash
# 限制容器使用 1 核 CPU 和 512MB 内存
docker run -d --cpus=1 --memory=512m nginx
```

**面试追问：如果容器内存超限会发生什么？**
- 容器会被 OOM Killer 杀死
- Pod 状态变为 `OOMKilled`
- Kubernetes 会根据 `restartPolicy` 决定是否重启

---

## 二、Kubernetes 核心概念（5 分钟速记）

### 2.1 核心组件

| 组件 | 职责 | 面试话术 |
|------|------|---------|
| **Master** | 集群控制平面 | "大脑"，负责调度和管理 |
| **Node** | 工作节点，运行容器 | "手脚"，执行实际任务 |
| **Pod** | 最小部署单元，包含 1 个或多个容器 | "豆荚"，容器的外壳 |
| **Deployment** | 管理 Pod 的副本数和更新策略 | 自动扩缩容、滚动更新 |
| **Service** | 为 Pod 提供固定 IP 和负载均衡 | Pod 的"门牌号" |
| **Ingress** | HTTP/HTTPS 路由，七层负载均衡 | 集群的"网关" |

### 2.2 Master 组件

| 组件 | 职责 |
|------|------|
| **API Server** | 集群的统一入口，所有操作都通过 API Server |
| **Scheduler** | 负责 Pod 调度，选择合适的 Node |
| **Controller Manager** | 管理各种控制器（Deployment、ReplicaSet 等） |
| **etcd** | 分布式 KV 存储，保存集群状态 |

### 2.3 Node 组件

| 组件 | 职责 |
|------|------|
| **kubelet** | Node 上的代理，负责 Pod 生命周期管理 |
| **kube-proxy** | 负责 Service 的网络代理和负载均衡 |
| **Container Runtime** | 容器运行时（Docker、containerd、CRI-O） |

### 2.4 Pod 生命周期

**Pod 状态**：

| 状态 | 含义 |
|------|------|
| **Pending** | 等待调度或拉取镜像 |
| **Running** | 至少一个容器正在运行 |
| **Succeeded** | 所有容器成功终止（Job/CronJob 场景） |
| **Failed** | 至少一个容器失败退出 |
| **Unknown** | 无法获取 Pod 状态（通常是 Node 失联） |

**Pod 生命周期钩子**：

| 钩子 | 触发时机 |
|------|---------|
| **postStart** | 容器启动后立即执行 |
| **preStop** | 容器终止前执行（用于优雅关闭） |

---

## 三、K8s 网络原理（深度解析）

### 3.1 Linux 虚拟网络基础

**Veth Pair + Bridge 实现跨 namespace 通信**：

```
┌──────────────┐      ┌──────────────┐
│  Container A │      │  Container B │
│  (ns1)       │      │  (ns2)       │
│  192.168.1.2 │      │  192.168.1.3 │
└──────┬───────┘      └──────┬───────┘
       │ veth-ns1            │ veth-ns2
       │                     │
       └──────┬──────────────┘
              │
         ┌────┴────┐
         │ docker0 │ (Bridge)
         │ 192.168.1.1
         └─────────┘
```

**实战练习**：

```bash
# 创建两个 network namespace
sudo ip netns add ns1
sudo ip netns add ns2

# 创建 veth pair
sudo ip link add veth-ns1 type veth peer name veth-ns1-br
sudo ip link set veth-ns1 netns ns1

# 创建网桥
sudo brctl addbr docker0
sudo brctl addif docker0 veth-ns1-br

# 配置 IP
sudo ip -n ns1 addr add 192.168.1.2/24 dev veth-ns1
sudo ip link set veth-ns1-br up
sudo ip -n ns1 link set veth-ns1 up

# 测试连通性
sudo ip netns exec ns1 ping 192.168.1.1
```

### 3.2 Docker 网络

**docker0 网桥**：
- Docker 默认创建 `docker0` 网桥（类似交换机）
- 每个容器通过 veth pair 连接到 docker0
- 容器间通过 docker0 通信

**查看 Docker 网络**：

```bash
# 查看网桥
brctl show
# 输出：
# bridge name    interfaces
# docker0        veth91e1730
#                vethc858a6a

# 查看容器路由
docker exec <container-id> route -n
# 输出：
# Destination     Gateway         Genmask
# 0.0.0.0         172.17.0.1      0.0.0.0          # 默认网关
# 172.17.0.0      0.0.0.0         255.255.0.0      # 本地网络

# 查看 iptables NAT 规则
sudo iptables -t nat -S | grep docker
# 输出：
# -A POSTROUTING -s 172.17.0.0/16 ! -o docker0 -j MASQUERADE
```

### 3.3 Pod 网络

**Pause 容器**：
- 每个 Pod 有一个 Pause 容器（`registry.k8s.io/pause`）
- Pause 容器负责创建 Network Namespace
- Pod 内其他容器共享 Pause 的网络栈（IP、端口、路由）

**验证 Pause 容器**：

```bash
# 查看 Pod 对应的容器
docker ps | grep etcd
# 输出：
# 8fd1337b0bf2   etcd:latest          # 业务容器
# 1202ef34af2b   registry.k8s.io/pause:3.9  # Pause 容器

# 查看业务容器的网络模式
docker inspect 8fd1337b0bf2 | grep NetworkMode
# 输出：
# "NetworkMode": "container:1202ef34af2b..."
```

**CNI（Container Network Interface）**：
- Kubernetes 通过 CNI 插件实现 Pod 网络
- 常见 CNI：Flannel、Calico、Weave、Cilium
- CNI 插件位置：`/opt/cni/bin/`

```bash
ls -l /opt/cni/bin/
# 输出：
# -rwxr-xr-x bridge
# -rwxr-xr-x host-local
# -rwxr-xr-x loopback
```

### 3.4 Service 网络

**Service 的作用**：
1. **服务发现**：为 Pod 提供固定的 ClusterIP
2. **负载均衡**：将请求分发到多个 Pod

**Service 实现原理**：

```
┌──────────────────────────────────────────────┐
│               API Server                     │
│  (Service IP: 10.96.0.1 分配并写入 etcd)     │
└───────────────┬──────────────────────────────┘
                │
        ┌───────┴───────┐
        │   etcd        │
        │ (Service +    │
        │  Endpoints)   │
        └───────┬───────┘
                │ watch
        ┌───────┴───────┐
        │ kube-proxy    │
        │ (维护 iptables│
        │  或 ipvs 规则)│
        └───────┬───────┘
                │
        ┌───────┴───────┐
        │  iptables/ipvs│
        │ (Service IP → │
        │  Pod IP 转发) │
        └───────────────┘
```

**Service 类型**：

| 类型 | 说明 | 使用场景 |
|------|------|---------|
| **ClusterIP** | 集群内部访问（默认） | 微服务间调用 |
| **NodePort** | 通过 Node IP + 端口访问 | 测试、临时对外暴露 |
| **LoadBalancer** | 云厂商提供的负载均衡器 | 生产环境对外暴露 |
| **ExternalName** | CNAME 记录，映射到外部服务 | 访问外部数据库 |

**kube-proxy 工作模式**：

| 模式 | 原理 | 性能 |
|------|------|------|
| **userspace** | 用户态代理（已废弃） | 差 |
| **iptables** | 内核态转发（默认） | 中等 |
| **ipvs** | LVS 负载均衡 | 高（推荐） |

### 3.5 Ingress 网络

**Ingress 的作用**：
- 七层（HTTP/HTTPS）负载均衡
- 基于域名和 URL 路径路由

**Ingress 架构**：

```
┌────────────────────────────────────────┐
│  External Traffic (example.com/app1)   │
└─────────────┬──────────────────────────┘
              │
      ┌───────┴───────┐
      │  Ingress      │ (Nginx/Traefik)
      │  Controller   │
      └───────┬───────┘
              │ 根据 Ingress 规则转发
      ┌───────┴───────┐
      │  Service A    │ (ClusterIP: 10.96.0.1)
      └───────┬───────┘
              │
      ┌───────┴───────┬───────────┐
      │  Pod A1       │  Pod A2   │
      └───────────────┴───────────┘
```

**Ingress 路由规则**：

```yaml
apiVersion: networking.k8s.io/v1
kind: Ingress
metadata:
  name: example-ingress
spec:
  rules:
  - host: example.com
    http:
      paths:
      - path: /app1
        pathType: Prefix
        backend:
          service:
            name: service-a
            port:
              number: 80
      - path: /app2
        pathType: Prefix
        backend:
          service:
            name: service-b
            port:
              number: 80
```

---

## 四、Pod 生命周期与故障排查

### 4.1 Pod 状态速查表

| 状态 | 原因 | 排查方法 |
|------|------|---------|
| **Pending** | 资源不足 / 镜像拉取中 / 调度失败 | `kubectl describe pod <pod-name>` |
| **ImagePullBackOff** | 镜像不存在 / 镜像仓库认证失败 | 检查镜像名称、Secret |
| **CrashLoopBackOff** | 容器启动后立即崩溃 | 查看日志 `kubectl logs <pod-name>` |
| **OOMKilled** | 内存超限 | 增加 `resources.limits.memory` |
| **Error** | 容器异常退出 | 查看退出码 `kubectl describe pod` |

### 4.2 常见故障排查

#### 故障1：ImagePullBackOff

**现象**：

```bash
$ kubectl get pods
NAME                     READY   STATUS             RESTARTS   AGE
myapp-5d4b7c8f9-xyz      0/1     ImagePullBackOff   0          2m
```

**原因**：
1. 镜像名称错误（拼写错误、标签不存在）
2. 私有镜像仓库认证失败
3. 网络问题（无法连接镜像仓库）

**排查**：

```bash
# 查看详细错误
kubectl describe pod myapp-5d4b7c8f9-xyz
# 输出：
# Failed to pull image "myapp:v1.0": rpc error: code = Unknown desc = Error response from daemon: pull access denied

# 解决方案1：修正镜像名称
kubectl edit deployment myapp

# 解决方案2：创建 imagePullSecrets
kubectl create secret docker-registry regcred \
  --docker-server=registry.example.com \
  --docker-username=user \
  --docker-password=pass \
  --docker-email=user@example.com
```

#### 故障2：CrashLoopBackOff

**现象**：

```bash
$ kubectl get pods
NAME                     READY   STATUS             RESTARTS   AGE
myapp-5d4b7c8f9-xyz      0/1     CrashLoopBackOff   5          5m
```

**原因**：
1. 应用启动失败（配置错误、依赖服务不可用）
2. 健康检查失败
3. 启动命令错误

**排查**：

```bash
# 查看日志
kubectl logs myapp-5d4b7c8f9-xyz
# 输出：
# Error: Cannot connect to database at mysql:3306

# 查看上一次容器的日志（如果容器已重启）
kubectl logs myapp-5d4b7c8f9-xyz --previous

# 解决方案：修复配置
kubectl edit deployment myapp
```

#### 故障3：OOMKilled

**现象**：

```bash
$ kubectl describe pod myapp-5d4b7c8f9-xyz
...
State:          Terminated
  Reason:       OOMKilled
  Exit Code:    137
```

**原因**：
- 容器内存使用超过 `resources.limits.memory`

**排查**：

```bash
# 查看 Pod 资源限制
kubectl get pod myapp-5d4b7c8f9-xyz -o yaml | grep -A 10 resources
# 输出：
# resources:
#   limits:
#     memory: 128Mi  # 内存限制过小
#   requests:
#     memory: 64Mi

# 解决方案：增加内存限制
kubectl edit deployment myapp
# 修改为：
# resources:
#   limits:
#     memory: 512Mi
```

### 4.3 健康检查

**Liveness Probe（存活探针）**：
- 检测容器是否存活
- 失败则重启容器

**Readiness Probe（就绪探针）**：
- 检测容器是否准备好接收流量
- 失败则从 Service Endpoints 中移除

**Startup Probe（启动探针）**：
- 用于慢启动容器
- 启动阶段禁用 Liveness Probe

**示例**：

```yaml
apiVersion: v1
kind: Pod
metadata:
  name: myapp
spec:
  containers:
  - name: app
    image: myapp:v1.0
    livenessProbe:
      httpGet:
        path: /healthz
        port: 8080
      initialDelaySeconds: 30  # 启动后等待 30 秒
      periodSeconds: 10         # 每 10 秒检查一次
    readinessProbe:
      httpGet:
        path: /ready
        port: 8080
      initialDelaySeconds: 5
      periodSeconds: 5
```

---

## 五、实用 YAML 配置模板

### 5.1 Deployment（无状态应用）

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: nginx-deployment
  labels:
    app: nginx
spec:
  replicas: 3  # 副本数
  selector:
    matchLabels:
      app: nginx
  template:
    metadata:
      labels:
        app: nginx
    spec:
      containers:
      - name: nginx
        image: nginx:1.21
        ports:
        - containerPort: 80
        resources:
          requests:
            cpu: 100m      # 请求 0.1 核
            memory: 128Mi
          limits:
            cpu: 500m      # 最多使用 0.5 核
            memory: 512Mi
        livenessProbe:
          httpGet:
            path: /
            port: 80
          initialDelaySeconds: 10
          periodSeconds: 5
        readinessProbe:
          httpGet:
            path: /
            port: 80
          initialDelaySeconds: 5
          periodSeconds: 3
```

### 5.2 Service（ClusterIP）

```yaml
apiVersion: v1
kind: Service
metadata:
  name: nginx-service
spec:
  type: ClusterIP  # 集群内部访问
  selector:
    app: nginx
  ports:
  - protocol: TCP
    port: 80        # Service 端口
    targetPort: 80  # Pod 端口
```

### 5.3 Service（NodePort）

```yaml
apiVersion: v1
kind: Service
metadata:
  name: nginx-nodeport
spec:
  type: NodePort
  selector:
    app: nginx
  ports:
  - protocol: TCP
    port: 80
    targetPort: 80
    nodePort: 30080  # Node 上的端口（30000-32767）
```

### 5.4 Ingress

```yaml
apiVersion: networking.k8s.io/v1
kind: Ingress
metadata:
  name: example-ingress
  annotations:
    nginx.ingress.kubernetes.io/rewrite-target: /
spec:
  rules:
  - host: example.com
    http:
      paths:
      - path: /app1
        pathType: Prefix
        backend:
          service:
            name: service-a
            port:
              number: 80
      - path: /app2
        pathType: Prefix
        backend:
          service:
            name: service-b
            port:
              number: 80
  tls:
  - hosts:
    - example.com
    secretName: example-tls
```

### 5.5 ConfigMap（配置文件）

```yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: app-config
data:
  database.url: "mysql://db:3306/mydb"
  log.level: "info"
---
# 使用 ConfigMap
apiVersion: v1
kind: Pod
metadata:
  name: myapp
spec:
  containers:
  - name: app
    image: myapp:v1.0
    envFrom:
    - configMapRef:
        name: app-config  # 所有 key-value 都作为环境变量
```

### 5.6 Secret（敏感信息）

```yaml
apiVersion: v1
kind: Secret
metadata:
  name: db-secret
type: Opaque
data:
  username: YWRtaW4=  # base64("admin")
  password: cGFzc3dvcmQ=  # base64("password")
---
# 使用 Secret
apiVersion: v1
kind: Pod
metadata:
  name: myapp
spec:
  containers:
  - name: app
    image: myapp:v1.0
    env:
    - name: DB_USER
      valueFrom:
        secretKeyRef:
          name: db-secret
          key: username
    - name: DB_PASS
      valueFrom:
        secretKeyRef:
          name: db-secret
          key: password
```

---

## 六、监控与告警

### 6.1 核心指标

| 维度 | 指标 | 含义 |
|------|------|------|
| **Node** | CPU 使用率 | 节点 CPU 使用情况 |
|          | 内存使用率 | 节点内存使用情况 |
|          | 磁盘使用率 | 节点磁盘使用情况 |
| **Pod** | CPU 使用率 | Pod CPU 使用情况 |
|         | 内存使用率 | Pod 内存使用情况 |
|         | 重启次数 | Pod 重启次数（CrashLoopBackOff） |
| **容器** | OOMKilled 次数 | 内存超限次数 |
|         | 退出码 | 容器退出码（0 正常，非 0 异常） |

### 6.2 常用命令

```bash
# 查看 Node 资源使用情况
kubectl top nodes

# 查看 Pod 资源使用情况
kubectl top pods

# 查看 Pod 详细信息
kubectl describe pod <pod-name>

# 查看 Pod 日志
kubectl logs <pod-name>
kubectl logs <pod-name> --previous  # 上一次容器的日志

# 进入 Pod 调试
kubectl exec -it <pod-name> -- /bin/bash

# 查看 Pod 事件
kubectl get events --sort-by='.metadata.creationTimestamp'
```

---

## 七、面试高频 20 题

### 1. Docker 和虚拟机有什么区别？

**标准答案**：
Docker 容器通过 namespace 实现资源隔离，通过 cgroup 实现资源限制，共享宿主机内核，启动快（秒级）、资源占用少（MB 级）。VM 通过 Hypervisor 虚拟化完整操作系统，隔离性更强但开销大（分钟级启动、GB 级资源占用）。

**追问应对**：
- **为什么容器启动快？** 不需要启动完整 OS，只需启动进程
- **容器安全性如何保证？** namespace 隔离 + seccomp/AppArmor 限制系统调用 + 最小化镜像

### 2. Docker 镜像为什么要分层？

**标准答案**：
1. **共享存储**：多个镜像可以共享相同的层（如 ubuntu:20.04 基础层）
2. **快速构建**：修改某一层时，只需重新构建该层及以上层
3. **快速分发**：只传输变化的层，不需要传输整个镜像

### 3. Kubernetes 的核心组件有哪些？

**标准答案**：
- **Master**：API Server、Scheduler、Controller Manager、etcd
- **Node**：kubelet、kube-proxy、Container Runtime

**追问应对**：
- **API Server 的作用？** 集群的统一入口，所有操作都通过 API Server
- **etcd 的作用？** 分布式 KV 存储，保存集群状态

### 4. Pod 是什么？为什么需要 Pod？

**标准答案**：
Pod 是 Kubernetes 的最小部署单元，包含 1 个或多个容器。Pod 内容器共享网络（同一个 IP）和存储（Volume）。

**为什么需要 Pod？**
- 解决"紧密耦合"容器的部署问题（如 Web + Sidecar）
- 简化调度（Pod 作为整体调度，而不是单个容器）

### 5. Deployment、ReplicaSet、Pod 的关系？

**标准答案**：
- **Deployment** 管理 ReplicaSet，负责滚动更新和回滚
- **ReplicaSet** 管理 Pod，保证 Pod 副本数
- **Pod** 是实际运行的容器

```
Deployment（管理更新）
    ↓
ReplicaSet（管理副本数）
    ↓
Pod（运行容器）
```

### 6. Service 是如何实现负载均衡的？

**标准答案**：
1. API Server 分配 ClusterIP，写入 etcd
2. kube-proxy watch Service 和 Endpoints 变化
3. kube-proxy 维护 iptables/ipvs 规则，实现 Service IP → Pod IP 转发
4. 默认使用轮询（Round-Robin）负载均衡

**追问应对**：
- **iptables vs ipvs？** ipvs 性能更高（基于 LVS），支持更多负载均衡算法

### 7. Ingress 和 Service 有什么区别？

**标准答案**：
- **Service**：四层（TCP/UDP）负载均衡，基于 IP + 端口
- **Ingress**：七层（HTTP/HTTPS）负载均衡，基于域名 + URL 路径

**追问应对**：
- **如何选择？** 内部服务间调用用 Service，对外暴露 HTTP 服务用 Ingress

### 8. Pod 的 Pending 状态可能是什么原因？

**标准答案**：
1. **资源不足**：Node 没有足够的 CPU/内存
2. **镜像拉取中**：正在拉取镜像
3. **调度失败**：没有满足 nodeSelector/亲和性的 Node

**排查方法**：`kubectl describe pod <pod-name>`

### 9. Pod 的 CrashLoopBackOff 如何排查？

**标准答案**：
1. 查看日志：`kubectl logs <pod-name>`
2. 查看上一次日志：`kubectl logs <pod-name> --previous`
3. 查看 Pod 详情：`kubectl describe pod <pod-name>`

**常见原因**：
- 应用启动失败（配置错误、依赖服务不可用）
- 健康检查失败

### 10. OOMKilled 是什么？如何解决？

**标准答案**：
OOMKilled 表示容器内存使用超过 `resources.limits.memory`，被 OOM Killer 杀死。

**解决方案**：
1. 增加内存限制：`resources.limits.memory: 512Mi`
2. 优化应用内存使用（如减少缓存、优化算法）

### 11. 什么是 namespace？有什么作用？

**标准答案**：
namespace 是 Kubernetes 的资源隔离机制，用于将集群资源划分为多个虚拟集群。

**作用**：
- 多租户隔离（不同团队/项目）
- 资源配额（ResourceQuota）
- 访问控制（RBAC）

### 12. ConfigMap 和 Secret 有什么区别？

**标准答案**：
- **ConfigMap**：存储非敏感配置（如数据库地址、日志级别）
- **Secret**：存储敏感信息（如密码、密钥），base64 编码

**追问应对**：
- **Secret 是否安全？** base64 不是加密，只是编码。生产环境建议使用外部密钥管理（如 Vault）

### 13. Liveness Probe 和 Readiness Probe 的区别？

**标准答案**：
- **Liveness Probe**：检测容器是否存活，失败则重启容器
- **Readiness Probe**：检测容器是否准备好接收流量，失败则从 Service Endpoints 中移除

### 14. Kubernetes 如何实现滚动更新？

**标准答案**：
1. Deployment 创建新的 ReplicaSet（新版本）
2. 逐步增加新 ReplicaSet 的副本数，减少旧 ReplicaSet 的副本数
3. 新 Pod 就绪后，旧 Pod 才会被删除

**配置**：
- `maxSurge`：滚动更新时最多超出的 Pod 数
- `maxUnavailable`：滚动更新时最多不可用的 Pod 数

### 15. Kubernetes 如何实现自动扩缩容？

**标准答案**：
使用 **HPA（Horizontal Pod Autoscaler）**，基于 CPU/内存使用率或自定义指标自动调整 Pod 副本数。

**示例**：

```yaml
apiVersion: autoscaling/v2
kind: HorizontalPodAutoscaler
metadata:
  name: myapp-hpa
spec:
  scaleTargetRef:
    apiVersion: apps/v1
    kind: Deployment
    name: myapp
  minReplicas: 2
  maxReplicas: 10
  metrics:
  - type: Resource
    resource:
      name: cpu
      target:
        type: Utilization
        averageUtilization: 80  # CPU 使用率超过 80% 时扩容
```

### 16. StatefulSet 和 Deployment 有什么区别？

**标准答案**：
- **Deployment**：无状态应用，Pod 名称随机、无固定存储
- **StatefulSet**：有状态应用，Pod 名称有序（如 `mysql-0`、`mysql-1`）、有固定存储（PVC）

**适用场景**：
- Deployment：Web 服务、API 服务
- StatefulSet：数据库、消息队列、Zookeeper

### 17. DaemonSet 是什么？有什么作用？

**标准答案**：
DaemonSet 保证每个 Node 上运行一个 Pod。

**适用场景**：
- 日志采集（Fluentd、Filebeat）
- 监控代理（Node Exporter）
- 存储插件（Ceph、GlusterFS）

### 18. Kubernetes 如何实现服务发现？

**标准答案**：
1. **环境变量**：Pod 启动时，Kubernetes 会注入 Service 的环境变量
2. **DNS**：集群内置 DNS（CoreDNS），Service 名称解析为 ClusterIP

**示例**：
- Service 名称：`mysql-service`
- DNS 域名：`mysql-service.default.svc.cluster.local`

### 19. Kubernetes 的存储机制是怎样的？

**标准答案**：
- **Volume**：Pod 级别存储，Pod 删除后数据丢失
- **PersistentVolume（PV）**：集群级别存储资源（如 NFS、Ceph）
- **PersistentVolumeClaim（PVC）**：Pod 对存储的请求，绑定到 PV

### 20. Kubernetes 的安全机制有哪些？

**标准答案**：
1. **RBAC**：基于角色的访问控制
2. **Network Policy**：网络隔离，限制 Pod 间通信
3. **Pod Security Policy**：限制 Pod 的权限（如禁止特权容器）
4. **Secret 加密**：etcd 中 Secret 数据加密存储

---

## 常用命令速查

```bash
# Minikube
minikube start
minikube status
minikube ssh

# Pod 操作
kubectl get pods
kubectl describe pod <pod-name>
kubectl logs <pod-name>
kubectl exec -it <pod-name> -- /bin/bash

# Deployment 操作
kubectl create deployment nginx --image=nginx:1.21
kubectl scale deployment nginx --replicas=3
kubectl rollout status deployment nginx
kubectl rollout undo deployment nginx  # 回滚

# Service 操作
kubectl expose deployment nginx --port=80 --target-port=80 --type=NodePort
kubectl get svc

# 查看资源使用情况
kubectl top nodes
kubectl top pods

# 调试
kubectl describe pod <pod-name>
kubectl get events --sort-by='.metadata.creationTimestamp'
```

---

## 参考资料

1. [Minikube 环境安装](https://github.com/caicloud/kube-ladder/blob/master/tutorials/lab1-installation.md)
2. [Kubectl 命令和集群体验](https://github.com/caicloud/kube-ladder/blob/master/tutorials/lab2-application-and-service.md)
3. [Linux network namespace,veth,bridge 和 路由](https://www.zhaohuabing.com/post/2020-03-12-linux-network-virtualization/)
4. [从0到1搭建linux虚拟网络](https://zhuanlan.zhihu.com/p/199298498)
5. [Docker 网络：模拟docker网络](https://morningspace.github.io/tech/k8s-net-mimic-docker/)
6. [Docker 网络：从docker0开始](https://morningspace.github.io/tech/k8s-net-docker0/)
7. [Pod网络和pause容器](https://morningspace.github.io/tech/k8s-net-pod-1/)
8. [认识CNI插件](https://morningspace.github.io/tech/k8s-net-cni/)
9. [深度解读CNI：容器网络接口](https://mp.weixin.qq.com/s/_nzbZYpKlpw4jKd5MFpuzw)
10. [官方文档：服务service](https://kubernetes.io/zh-cn/docs/concepts/services-networking/service/)
11. [创建service之后，k8s会发生什么](https://zhuanlan.zhihu.com/p/677236869)
12. [探究k8s service iptables 路由规则](https://luckymrwang.github.io/2021/02/20/%E6%8E%A2%E7%A9%B6K8S-Service%E5%86%85%E9%83%A8iptables%E8%B7%AF%E7%94%B1%E8%A7%84%E5%88%99/)
13. [官方文档：在minikube中使用nginx ingress 控制配置ingress](https://kubernetes.io/zh-cn/docs/tasks/access-application-cluster/ingress-minikube/)
14. [官方文档：ingress](https://kubernetes.io/zh-cn/docs/concepts/services-networking/ingress/)
15. [Kubernetes 官方文档](https://kubernetes.io/zh-cn/docs/home/)
