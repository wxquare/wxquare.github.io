---
title: Python 多线程和多进程
categories:
- Python
---
　　
　　最近在做一些算法优化的工作，由于对Python认识不够，开始的入坑使用了多线程。发现在一个四核机器，即使使用多线程，CPU使用率始终在100%左右（一个核）。后来发现Python中并行计算要使用多进程，改成多进程模式后，CPU使用率达到340%，也提升了算法的效率。另外multiprocessing对多线程和多进程做了很好的封装，需要掌握。这里总结下面两个问题：
1. Python中的并行计算为什么要使用多进程？
2. Python多线程和多进程简单测试
3. multiprocessing库的使用


## Python中的并行计算为什么要使用多进程？
　　Python在并行计算中必须使用多进程的原因是GIL(Global Interpreter Lock，全局解释器锁)。GIL使得在解释执行Python代码时，会产生互斥锁来限制线程对共享资源的访问，直到解释器遇到I/O操作或者操作次数达到一定数目时才会释放GIL。这使得**Python一个进程内同一时间只能允许一个线程进行运算**”，也就是说多线程无法利用多核CPU。因此：
1. 对于CPU密集型任务（循环、计算等），由于多线程触发GIL的释放与在竞争，多个线程来回切换损耗资源，因此多线程不但不会提高效率，反而会降低效率。所以**计算密集型程序，要使用多进程**。
2. 对于I/O密集型代码（文件处理、网络爬虫、sleep等待),开启多线程实际上是**并发(不是并行)**，线程A在IO等待时，会切换到线程B，从而提升效率。
3. 大多数程序包含CPU和IO操作，但不考虑进程的资源开销，**多进程通常都是优于多线程的**。
4. 由于Python多线程的问题，因此通常情况下都使用多进程，使用多进程需要注意进程间变量的共享。

## Python多线程和多进程简单测试
- job1是一个完成CPU没有任务IO的死循环，观察CPU使用率，无论使用多少线程数量num，CPU使用率始终在100%左右，也就是说只能利用核的资源。而多进程则可以使用多核资源，num为1时CPU使用率为100%，num为2时CPU使用率接近200%。
- job2是一个IO密集型的程序，主要的耗时在print系统调用。num=4时，多线程跑了10.81s，cpu使用率93%；多进程只用了3.23s，CPU使用率130%。
　
```
import multiprocessing
import threading

def job1():
    '''
        full cpu
    '''
    while True:
        continue

NUMS = 100000
def job2():
    '''
        cpu and io
    '''
    for i in range(NUMS):
        print("hello,world")

def multi_threads(num,job):
    threads = []
    for i in range(num):
        t = threading.Thread(target=job,args=())
        threads.append(t)
    for t in threads:
        t.start()
    for t in threads:
        t.join()

def multi_process(num,job):
    process = []
    for i in range(num):
        p = multiprocessing.Process(target=job,args=())
        process.append(p)
    for p in process:
        p.start()
    for p in process:
        p.join()

if __name__ == '__main__':
    # multi_threads(4,job1)
    # multi_process(4,job1)
    # multi_threads(4,job2)
    multi_process(4,job2)

```

## [multiprocessing的使用](https://docs.python.org/3/library/multiprocessing.html#module-multiprocessing)
参考：https://docs.python.org/3/library/multiprocessing.html#module-multiprocessing

1. 单个进程multiprocessing.Process对象，和threading.Thread的API完全一样，start(),join(),参考上文中的测试代码。
2. 进程池
3. 进程间对象共享队列multiprocessing.Queue()
4. 进程同步multiprocessing.Lock()
5. 进程间状态共享multiprocessing.Value,multiprocessing.Array
6. multiprocessing.Manager()
7. 进程池：multiprocessing.Pool()
- pool.map
- pool.imap_unordered
- pool.apply_async
- 


　　Python多线程和多进程的使用非常方面，因为multiprocessing提供了非常好的封装。为了方便设置线程和进程的数量，通常都会使用池pool技术。
```
from multiprocessing.dummy import Pool as DummyPool   # thread pool
from multiprocessing import Pool                      # process pool
```
multilprocessing包的使用可参考：
- https://docs.python.org/3/library/multiprocessing.html#module-multiprocessing
- https://thief.one/2016/11/24/Multiprocessing-Pool/