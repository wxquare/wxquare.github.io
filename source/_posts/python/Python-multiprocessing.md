---
title: Python 多线程和多进程
categories:
- Python
---

## 一、multiprocessing
　　Python多线程和多进程的使用非常方面，因为multiprocessing提供了非常好的封装。为了方便设置线程和进程的数量，通常都会使用池pool技术。
```
from multiprocessing.dummy import Pool as DummyPool   # thread pool
from multiprocessing import Pool                      # process pool
```
multilprocessing包的使用可参考：
- https://docs.python.org/3/library/multiprocessing.html#module-multiprocessing
- https://thief.one/2016/11/24/Multiprocessing-Pool/

## 二、Python多线程无法利用多核CPU
　　GIL的存在导致Python多线程无法利用多核CPU。Python中每个进程会有一个GIL，该进程的线程只有在获得该GIL的情况下才能运行。线程在遇到I/O 操作时会释放这把锁，如果是纯计算的没有I\O操作的程序，解释器会每隔100次操作释放该锁，让别的线程有机会执行。因此同一时刻，在python进程中只会有一个线程在运行，其它线程都处于等待状态。    
**python多线程注意事项：**
1. 对于CPU密集型代码（循环、计算等），由于计算工作量多和大，计算很快就会达到100，然后触发GIL的释放与在竞争，多个线程来回切换损耗资源，所以在多线程遇到CPU密集型代码时，单线程会比较快。例如下面代码中的python_cpu_100_100ms()函数。
2. 对于I/O密集型代码（文件处理、网络爬虫、sleep等待),开启多线程实际上是**并发(不是并行)**，线程A在IO等待时，会切换到线程B，从而提升效率。例如：io_100_100ms()函数。
3. 通常来说python多线程是不适合用于CPU密集型代码的，但是有一种类外，就是当函数底层实现为C/C++代码时,例如以numpy和cv2实现的C_cpu_100_300ms()函数。
4. 很多程序既包括CPU操作包括IO操作，对于这种程序我们可以用time先简单测试一下cpu利用率，然后考虑是否使用多线程，以及设置合理的线程数量。例如下面代码中的test()函数。　　
```
from multiprocessing.dummy import Pool as DummyPool
from multiprocessing import Pool
import time
import os

# avoid numpy thread
os.environ["MKL_NUM_THREADS"] = "1" 
os.environ["NUMEXPR_NUM_THREADS"] = "1" 
os.environ["OMP_NUM_THREADS"] = "1"

import numpy as np
import cv2

# avoid cv2 threads
cv2.setNumThreads(1)

# cpu intensive
def python_cpu_100_100ms():
    x = 0
    for i in range(1,2000000):
        x += i

# IO-intensive
def io_100_100ms():
   time.sleep(0.1)

# cpu intensive
def C_cpu_100_300ms():
    arr2 = np.ndarray((10000,10000,3), dtype=np.uint8)
    hist = cv2.calcHist([arr2, arr2],[0],None,histSize=[256],ranges=[0,256])

# CPU and io
def test():
    python_cpu_100_100ms()
    time.sleep(0.1)

if __name__=="__main__":
    # pool = Pool(processes=4)
    pool = DummyPool(processes=4)
    for i in range(500):
        ret = pool.apply_async(test, args=())   #维持执行的进程总数为10，当一个进程执行完后启动一个新进程.
    pool.close()
    pool.join()

```

## 三、关于Python多进程
1. 通常来说Python多进程能利用机器的多核资源，是程序并行，提高性能
2. 进程的代价始终比线程的代价要高，当程序耗时较低的不建议使用多进程
3. 虽然Python多进程支持变量共享，但是不建议使用。建议使用进程返回值，进程同步之后，再整理各个进程执行的结果。 
