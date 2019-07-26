---
title: time/top/perf/gprof程序性能分析
categories:
- 计算机基础
---


　　会使用time、top、perf和gprof工具分析程序的性能。


## 一、time
1. shell time。 time非常方便获取程序运行的时间，包括用户态时间user、内核态时间sys和实际运行的时间real。我们可以通过(user+sys)/real计算程序CPU占用率，判断程序时CPU密集型还是IO密集型程序。
$time ffmpeg -y -i in.mp4 -vf "crop=w=100:h=100:x=12:y=34" -acodec copy out.mp4
real	0m0.144s
user	0m0.048s
sys	    0m0.124s

2. Linux中除了shell time，还有/usr/bin/time，它能获取程序运行更多的信息，通常带有-v参数。
```
$ /usr/bin/time -v ffmpeg -y -i in.mp4 -vf "crop=w=100:h=100:x=12:y=34" -acodec copy out.mp4

User time (seconds): 0.01                            # 用户态时间
System time (seconds): 0.10                          # 内核态时间
Percent of CPU this job got: 124%                    # CPU占用率
Elapsed (wall clock) time (h:mm:ss or m:ss): 0:00.09 # 实际时间
Average shared text size (kbytes): 0
Average unshared data size (kbytes): 0
Average stack size (kbytes): 0
Average total size (kbytes): 0
Maximum resident set size (kbytes): 20096
Average resident set size (kbytes): 0
Major (requiring I/O) page faults: 0                 # 缺页异常 
Minor (reclaiming a frame) page faults: 2953
Voluntary context switches: 2715                     # 上下文切换
Involuntary context switches: 414
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

## 二、top
top是linux系统的任务管理器，它既能看系统所有任务信息，也能帮助查看单个进程资源使用情况。
主要有以下几个功能：
1. 查看系统任务信息：
 Tasks:  87 total,   1 running,  86 sleeping,   0 stopped,   0 zombie
2. 查看CPU使用情况
 Cpu(s):  0.0%us,  0.2%sy,  0.0%ni, 99.7%id,  0.0%wa,  0.0%hi,  0.0%si,  0.2%st
3. 查看内存使用情况
 Mem:    377672k total,   322332k used,    55340k free,    32592k buffers
4. 查看单个进程资源使用情况 
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
5. 除此之外top还提供了一些交互命令：
	- q:退出
	- 1:查看每个逻辑核
	- H：查看线程
	- P：按照CPU使用率排序
	- M：按照内存占用排序

参考：https://linuxtools-rst.readthedocs.io/zh_CN/latest/tool/top.html


## 三、perf
参考：https://www.ibm.com/developerworks/cn/linux/l-cn-perf1/index.html
参考：https://zhuanlan.zhihu.com/p/22194920



## 四、gprof
参考： https://blog.csdn.net/stanjiang2010/article/details/5655143