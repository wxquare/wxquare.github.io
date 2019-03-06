1. 帮助命令：man、info
2. 查找命令路径：which、whereis
3. 查看文件文件个数：find ./ | wc -l
4. 以时间顺序显示目录项：ls -lrt
5. 查看文件时同时显示行数：cat -n xxx
6. 查看两个文件的差别：diff file1 file2
7. 动态显示文本最新信息，常用于查看日志： tail -f xxx.log
8. 软连接/硬链接： ln cc ccAgain 和 ln -s cc ccAgain
9. command1 && command2
10. comamand1 || command2
11. <font color=red >查找txt和pdf文件：find . \( -name "*.txt" -o -name "*.pdf" \) -print </font>
12. find查找文件时指定深度：find . -maxdepth 1 -type f
13. find只查找目录：find . -type d -print 
14. [文本处理](https://linuxtools-rst.readthedocs.io/zh_CN/latest/base/03_text_processing.html)
15. 打包：taf -cvf xxx.tar .  解包： tar -xvf xxx.tar 
16. 压缩与解压：-z 解压gz文件；-j解压bz2；-J解压xz文件
17. 


###1、系统信息查看工具
- 查看操作系统发行版：lsb_release -a
- 查看内核版本信息：uname -a
- 查看cpu信息：cat /proc/cpuinfo
- 查看cpu核数：cat /proc/cpuinfo | grep processor | wc -l
- 查看内存信息：cat /proc/meminfo
- 显示架构：arch
- 查看ipc资源：ipcs
- 显示当前所有的系统资源limit信息： ulimit -a
- 对生成的core文件的大小不进行限制：ulimit -c unlimited



###2、进程管理和监控工具
- 查询正在运行的进程信息：ps -ef 或者 ps -ajx
- 查询某用户的进程： ps -ef | grep username 或者 ps -lu username
- 实时显示进程信息： top
- 查看用户打开的文件： lsof -u username
- 查看某进程打开的文件： lsof -p pid
- 杀死某进程：kill -9 pid
- pmap输出进程内存你的状况，用来分析线程堆栈
- 查看内存使用量：free -m 或者 vmstat n m
- 查看磁盘使用情况：df -h


###3、网络工具
- netstat命令用于显示各种网络相关信息
- 查询某端口port被某个进程占用：netstat -antp | grep port，然后使用ps pid查询进程名称
- 也可以使用lsof -i:port 直接查询该端口的进程
- ping 测试网络连通情况
- traceroute IP 探测前往ip的路由信息
- 直接下载文件或者网页:wget
- 网络远程复制：scp -r localpath ID@host:path
- 使用ssh协议下载： scp -r ID@host:path localpath



###4、环境变量
- 全局/etc/profile->/etc/profile.d;
- 读取当前用户下面的：~/.bash_profile->~/.bash_login->~/.profile
- 读取当前用户目录下面的：~/.bashrc
- export环境变量，退出失效


###其它
1. 查看gpu信息 nvidia-smi
2. 查看gpu驱动版本信息 cat /proc/driver/nvidia/version
3. [pkgconfig?](https://blog.csdn.net/luotuo44/article/details/24836901) PKG_CONFIG_PATH环境变量
参考：https://linuxtools-rst.readthedocs.io/zh_CN/latest/tool/index.html


### 3、pkg-config：解决编译和链接时候出现的undefined symbol问题
pkg-config能方便使用第三方库和头文件和库文件，其运行原理 
 
- 它首先根据PKG_CONFIG_PATH环境变量下寻找库对应的pc文件  
- 然后从pc文件中获取该库对应的头文件和库文件的位置信息
  
例如在项目中需要使用opencv库，该库包含的头文件和库文件比较多  

- 首先查看是否有对应的opencv.pc find /usr -name opencv.pc  
- 查看该路径是否包含在PKG_CONFIG_PATH  
- 使用pkg-config --cflgs --libs opencv 查看库对应的头文件和库文件信息  
- pkg-config --modversion opencv 查看版本信息

参考链接：[https://blog.csdn.net/luotuo44/article/details/24836901](https://blog.csdn.net/luotuo44/article/details/24836901)

### 4、ldd：帮助解决运行时出现的问题
**现象**：  
- <font color=red >error while loading shared libraries: libopencv_cudabgsegm.so.3.4: cannot open shared object file: No such file or directory </font>  
- ldd ./xxx，发现库文件not found  

      libopencv_cudaobjdetect.so.3.4 => not found  
      libopencv_cudalegacy.so.3.4 => not found

**ld.so 动态共享库搜索顺序**：  
1. ELF可执行文件中动态段DT_RPATH指定；gcc加入链接参数“-Wl,-rpath”指定动态库搜索路径；  
2. 环境变量LD_LIBRARY_PATH指定路径；  
3. /etc/ld.so.cache中缓存的动态库路径。可以通过修改配置文件/etc/ld.so.conf 增删路径（修改后需要运行ldconfig命令）；  
4. 默认的 /lib/;  
5. 默认的 /usr/lib/  

**解决办法**：  
- 确认系统中是包含这个库文件的  
- pkg-config --libs opencv 查看opencv库的路径  
- export LD_LIBRARY_PATH=/usr/local/lib64，增加运行时加载路径  

 参考链接：[https://www.cnblogs.com/amyzhu/p/8871475.html](https://www.cnblogs.com/amyzhu/p/8871475.html)

