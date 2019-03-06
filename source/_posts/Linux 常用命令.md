### 1、linux基础命令
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



###2、系统信息查看工具
- 查看操作系统发行版：lsb_release -a
- 查看内核版本信息：uname -a
- 查看cpu信息：cat /proc/cpuinfo
- 查看cpu核数：cat /proc/cpuinfo | grep processor | wc -l
- 查看内存信息：cat /proc/meminfo
- 显示架构：arch
- 查看进程间ipc资源情况：ipcs
- 显示当前所有的系统资源limit信息： ulimit -a
- 对生成的core文件的大小不进行限制：ulimit -c unlimited


###3、系统资源管理和监控
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
启动这个工具来收集系统性能数据： /etc/init.d/sysstat start

###4、网络工具
- netstat命令用于显示各种网络相关信息
- 查询某端口port被某个进程占用：netstat -antp | grep port，然后使用ps pid查询进程名称
- 也可以使用lsof -i:port 直接查询该端口的进程
- ping 测试网络连通情况
- traceroute IP 探测前往ip的路由信息
- 直接下载文件或者网页:wget
- 网络远程复制：scp -r localpath ID@host:path
- 使用ssh协议下载： scp -r ID@host:path localpath



###5、环境变量
- 全局/etc/profile->/etc/profile.d;
- 读取当前用户下面的：~/.bash_profile->~/.bash_login->~/.profile
- 读取当前用户目录下面的：~/.bashrc
- export环境变量，退出失效


###6、查看GPU信息
1. 查看gpu信息 nvidia-smi
2. 查看gpu驱动版本信息 cat /proc/driver/nvidia/version
3. [pkgconfig?](https://blog.csdn.net/luotuo44/article/details/24836901) PKG_CONFIG_PATH环境变量


参考：https://linuxtools-rst.readthedocs.io/zh_CN/latest/tool/index.html



