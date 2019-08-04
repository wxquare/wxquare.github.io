---
title: Linux 常用命令和shell编程
categories:
- C/C++
---

## 1、linux基础命令
- 帮助命令：man、info
- 查找命令路径：which、whereis
- 查看文件文件个数：find ./ | wc -l
- 以时间顺序显示目录项：ls -lrt
- 查看文件时同时显示行数：cat -n xxx
- 查看两个文件的差别：diff file1 file2
- 动态显示文本最新信息，常用于查看日志： tail -f xxx.log
- 软连接/硬链接： ln cc ccAgain 和 ln -s cc ccAgain
- command1 && command2
- comamand1 || command2
- <font color=red >查找txt和pdf文件：find . \( -name "*.txt" -o -name "*.pdf" \) -print </font>
- find查找文件时指定深度：find . -maxdepth 1 -type f
- find只查找目录：find . -type d -print 
- [文本处理](https://linuxtools-rst.readthedocs.io/zh_CN/latest/base/03_text_processing.html)
- 打包：taf -cvf xxx.tar .  解包： tar -xvf xxx.tar 
- 压缩与解压：-z 解压gz文件；-j解压bz2；-J解压xz文件



## 2、系统信息查看工具
- 查看操作系统发行版：lsb_release -a
- 查看内核版本信息：uname -a
- 查看cpu信息：cat /proc/cpuinfo
- 查看cpu核数：cat /proc/cpuinfo | grep processor | wc -l
- 查看内存信息：cat /proc/meminfo
- 显示架构：arch
- 查看进程间ipc资源情况：ipcs
- 显示当前所有的系统资源limit信息： ulimit -a
- 对生成的core文件的大小不进行限制：ulimit -c unlimited


## 3、系统资源管理和监控
- 查询正在运行的进程信息：ps -ef 或者 ps -ajx
- 查询某用户的进程： ps -ef | grep username 或者 ps -lu username
- 实时显示进程信息： top linux下的任务管理器
- 查看用户打开的文件： lsof -u username
- 查看某进程打开的文件： lsof -p pid
- 杀死某进程：kill -9 pid
- pmap输出进程内存你的状况，用来分析线程堆栈
- 查看**内存**使用量：free -m 或者 vmstat n m
- 查看**磁盘**使用情况：df -h
- iostat 监视I/O子系统，ubuntu安装systat。通过iostat方便查看CPU、网卡、tty设备、磁盘、CD-ROM 等等设备的活动情况, 负载信息
-  sar 找出系统瓶颈的利器
*ubuntu系统下，默认可能没有安装这个包，使用apt-get install sysstat 来安装；
安装完毕，将性能收集工具的开关打开： vi /etc/default/sysstat
设置 ENABLED=”true”
启动这个工具来收集系统性能数据： /etc/init.d/sysstat start. 
 

## 4、网络工具
- netstat命令用于显示各种网络相关信息
- 查询某端口port被某个进程占用：netstat -antp | grep port，然后使用ps pid查询进程名称
- 也可以使用lsof -i:port 直接查询该端口的进程
- ping 测试网络连通情况
- traceroute IP 探测前往ip的路由信息
- 直接下载文件或者网页:wget
- 网络远程复制：scp -r localpath ID@host:path
- 使用ssh协议下载： scp -r ID@host:path localpath
- nc服务器编程常用，既可以作为客户端又可以指定端口作为服务端。



## 5、环境变量
- 全局/etc/profile->/etc/profile.d;
- 读取当前用户下面的：~/.bash_profile->~/.bash_login->~/.profile
- 读取当前用户目录下面的：~/.bashrc
- export环境变量，退出失效


## 6、查看GPU信息
- 查看gpu信息 nvidia-smi
- 查看gpu驱动版本信息 cat /proc/driver/nvidia/version
- [pkgconfig?](https://blog.csdn.net/luotuo44/article/details/24836901) PKG_CONFIG_PATH环境变量


## 7、测试系统磁盘的性能
dd是Linux/UNIX 下的一个非常有用的命令，作用是用指定大小的块拷贝一个文件，并在拷贝的同时进行指定的转换。另外在linux中，有两个特殊的设备：/dev/null：回收站、无底洞，经常作为写端，不会产生IO，/dev/zero产生字符，经常作为读端，也不会产生IO。
（1）测试磁盘写能力
    dd if=/dev/zero of=/test1.img bs=4k count=10000
    因为/dev//zero是一个伪设备，它只产生空字符流，对它不会产生IO，所以，IO都会集中在of文件中，of文件只用于写，所以这个命令相当于测试磁盘的写能力。命令结尾添加oflag=direct将跳过内存缓存，添加oflag=sync将跳过hdd缓存。
（2）测试磁盘读能力
    dd if=/dev/sda of=/dev/null bs=4k  count=10000
    因为/dev/sdb是一个物理分区，对它的读取会产生IO，/dev/null是伪设备，相当于黑洞，of到该设备不会产生IO，所以，这个命令的IO只发生在/dev/sdb上，也相当于测试磁盘的读能力。
（3）测试同时读写能力
    time dd if=/dev/sda of=/test1.img  bs=4k count=10000
    在这个命令下，一个是物理分区，一个是实际的文件，对它们的读写都会产生IO（对/dev/sda是读，对/test.img是写），假设它们都在一个磁盘中，这个命令就相当于测试磁盘的同时读写能力。


## 8、使用dd和nc命令测试网络性能
nc是netcat的简写，有着网络界的瑞士军刀美誉。因为它短小精悍、功能实用，被设计为一个简单、可靠的网络工具
（1）实现任意TCP/UDP端口的侦听，nc可以作为server以TCP或UDP方式侦听指定端口
（2）端口的扫描，nc可以作为client发起TCP或UDP连接
（3）机器之间传输文件
（4）机器之间网络测速   
nc命令有个-l参数可以用来监听指定端口，因此我们要完成上面的功能，就只需要简单的从/dev/zero或者其他虚拟设备读入数据：

time nc -l -p 5001 < /test.img

然后另外一台电脑使用nc来连接到这个端口并读入数据：
time nc 192.168.0.11 5001 > /dev/null
上面的测试的结果中，是从磁盘读数据通过网络获取，通过time命令或缺时间参数，可以计算出网络的性能。更准备的测试应该从/dev/zero中多数据会更好一些


## 9、新学习的命令
netstat tcpdump ipcs ipcrm
cpu 内存 硬盘 等等与系统性能调试相关的命令必须熟练掌握，设置修改权限 tcp网络状态查看 各进程状态 抓包相关等相关命令
awk sed需掌握
系统命令工具

进程间通信设施状态 ipcs
虚拟内存统计工具 vmstat
Linux系统运行时长 uptime
CPU平均负载和磁盘活动 iostat
监控，收集和汇报系统活动 sar
监控多处理器使用情况 mpstat
监控进程的内存使用情况 pmap
进程监控工具 atop
控制台的流量监控工具 vnstat
系统进程状态 ps
详细网络流量分析 tcpdump
内存使用状态 free
系统管理员调优和基准测量工具 nmon
密切关注Linux系统 glances
查看系统调用 strace
网络参数工具

Linux网络统计监控工具 netstat
显示和修改网络接口控制器 ethtool
网络数据包分析利刃 tcpdump
监控并帮助你调试网络 bmon
获取实时网络统计信息 iptraf
显示主机上网络接口带宽使用情况 iftop
网络流量和 ip 流量分布 ntopng


日志监控工具

实时网络日志分析器 GoAccess

多窗口之下日志监控 MultiTail
日志分析系统 LogWatch/Swatch
参数监控工具

监控apache网络服务器整体性能 apachetop
ftp 服务器基本信息 ftptop
IO监控 iotop
电量消耗和电源管理 powertop
监控 mysql 的线程和性能 mytop
系统运行参数分析 htop/top/atop
工作基本上离不开 Linux，也是面试常见问题之一


参考：https://linuxtools-rst.readthedocs.io/zh_CN/latest/tool/index.html



