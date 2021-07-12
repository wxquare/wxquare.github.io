---
title: time/top/perf/gprof性能分析
categories:
- other
---

	会使用time、top、perf和gprof工具分析程序的性能。

## 一、time获取程序的运行时间
### shell time
　　Linux系统中有两个time，一个是默认的shelltime，它能帮助获取程序运行的时间，包括程序在用户态时间user、内核态时间sys，运行时间real。通过(user+sys)/real计算出CPU暂用率，判断该程序时CPU密集型还是IO密集型程序。
例如：
```
$time ffmpeg -y -i in.mp4 -vf "crop=w=100:h=100:x=12:y=34" -acodec copy out.mp4
	real	0m16.987s
	user	0m32.255s
	sys	0m6.129s
```

### /usr/bin/time -v
　　除了使用shell time，还可以使用/usr/bin/time，它能帮助获取更多的程序运行的信息，通常会加上-v参数，使time 输出足够详细的信息。
例如：
```
$ /usr/bin/time -v ffmpeg -y -i in.mp4 -vf "crop=w=100:h=100:x=12:y=34" -acodec copy out.mp4
	Command being timed: "ffmpeg -y -i in.mp4 -vf crop=w=100:h=100:x=12:y=34 -acodec copy out.mp4"
	User time (seconds): 0.01                                #用户态时间
	System time (seconds): 0.10                              #内核态时间
	Percent of CPU this job got: 116%                        #CPU占用率
	Elapsed (wall clock) time (h:mm:ss or m:ss): 0:00.10     #实际运行时间
	Average shared text size (kbytes): 0
	Average unshared data size (kbytes): 0
	Average stack size (kbytes): 0
	Average total size (kbytes): 0
	Maximum resident set size (kbytes): 20064
	Average resident set size (kbytes): 0
	Major (requiring I/O) page faults: 0                     #缺页异常
	Minor (reclaiming a frame) page faults: 2953
	Voluntary context switches: 2737						 #上下文切换
	Involuntary context switches: 693
	Swaps: 0
	File system inputs: 0
	File system outputs: 368
	Socket messages sent: 0
	Socket messages received: 0
	Signals delivered: 0
	Page size (bytes): 4096
	Exit status: 0
```
参考：https://my.oschina.net/yumm007/blog/920412

## 二、top Linux下的任务管理器
top命令是Linux下常用的性能分析工具，能够实时显示系统中各个进程的资源占用状况，类似于Windows的任务管理器。top是一个动态显示过程,即可以通过用户按键来不断刷新当前状态.如果在前台执行该命令,它将独占前台,直到用户终止该程序为止.比较准确的说,top命令提供了实时的对系统处理器的状态监视.它将显示系统中CPU最“敏感”的任务列表.该命令可以按CPU使用.内存使用和执行时间对任务进行排序；而且该命令的很多特性都可以通过交互式命令或者在个人定制文件中进行设定。
1. 实时查看当前机器任务数量：
   Tasks:  87 total,   1 running,  86 sleeping,   0 stopped,   0 zombie
2. 实时查看机器CPU信息：
   Cpu(s):  0.0%us,  0.2%sy,  0.0%ni, 99.7%id,  0.0%wa,  0.0%hi,  0.0%si,  0.2%st
3. 内存信息：
   Mem:    377672k total,   322332k used,    55340k free,    32592k buffers
4. 查看具体某个进程资源使用信息：   	  
	- PID：进程的ID
	- USER：进程所有者
	- PR：进程的优先级别，越小越优先被执行
	- NInice：值
	- VIRT：进程占用的虚拟内存
	- RES：进程占用的物理内存
	- SHR：进程使用的共享内存
	- S：进程的状态。S表示休眠，R表示正在运行，Z表示僵死状态，N表示该进程优先值为负数
	- %CPU：进程占用CPU的使用率
	- %MEM：进程使用的物理内存和总内存的百分比
	- TIME+：该进程启动后占用的总的CPU时间，即占用CPU使用时间的累加值。
	- COMMAND：进程启动命令名称
5. 常用的交互命令：
	q：退出
	P：按照CPU使用率排序
	M：按照内存使用排序
	H：显示线程信息
	1：监控每个逻辑CPU的信息
参考：https://linuxtools-rst.readthedocs.io/zh_CN/latest/tool/top.html



## 三、perf工具
参考： 
https://www.ibm.com/developerworks/cn/linux/l-cn-perf1/index.html
https://www.cnblogs.com/arnoldlu/p/6241297.html
https://zhuanlan.zhihu.com/p/22194920


## 四、gprof工具
https://blog.csdn.net/stanjiang2010/article/details/5655143

