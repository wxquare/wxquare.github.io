---
title: 一文记录Python实践
date: 2023-10-20
categories:
- 计算机基础
---

Python程序为什么慢？
----
　　不同的场景下，代码是有不同的要求，大体有三个等级，“管用、更好、更快”。相比C/C++，Python具有较好的开发系效率，但是程序的性能运行速度会差一些。究其原因是Python为了灵活性，牺牲了效率。
1. **动态类型**。对于C/C++等静态类型语言，由于变量的类型固定，变量之间的运算很容易指定特定的函数。而动态类型在运行的时间需要大量if else判断处理，直到找到符合条件的函数。**动态类型增加语言的易用性，但是牺牲了程序的运行效率**。

2. GIL(Global Interpreter Lock)全局解释锁，CPython在解释执行任何Python代码的时候，首先都需要**they acquire GIL when running，release GIL when blocking for I/O**。如果没有涉及I/O操作，只有CPU密集型操作时，解释器每隔100一段时间（100ticks）就会释放GIL。GIL是实现Python解释器的（Cython)时所引入的一个概念，不是Python的特性。
由于GIL的存在，使得Python对于计算密集型任务，多线程threading模块形同虚设，因为线程在实际运行的时候必须获得GIL，而GIL只有一个，因此无法发挥多核的优势。为了绕过GIL的限制，只能使用multiprocessing等多进程模块，每个进程各有一个CPython解释器，也就各有一个GIL。

3. CPython不支持JIL（Just-In-Time Compiler),JIL 能充分利用程序运行时信息，进行类型推导等优化，对于重复执行的代码段来说加速效果明显。对于CPython如果想使用JIT对计算密集型任务进行优化，可以尝试使用JIT包numba，它能使得相应的函数变成JIT编译。



Python程序优化的思路？
----
　　最近在做一些算法优化方面的工作,简单总结一下思路:
1. 熟悉算法的整体流程，对于算法代码，最开始尽可能不要使用多线程和多进程方法，
2. 在1的基础上跑出算法的CPU profile，整体了解算法耗时分布和瓶颈。Python提供的cProfile模块灵活的针对特定函数或者文件产生profile文件，根据profile数据进行代码性能优化。
 - 可以直接将生成profile代码写在Python脚本
 - 使用命名行方式生成profile
 - 分析工具pstats
 - 图形化工具[Gprof2Dot](https://github.com/jrfonseca/gprof2dot)。python gprof2dot.py -f pstats result.out | dot -Tpng -o result.png
 [https://blog.csdn.net/asukasmallriver/article/details/74356771](https://blog.csdn.net/asukasmallriver/article/details/74356771)
3. 程序（算法）本身的剪枝。比如视频追踪中，考虑是否让每个像素点都参与计算？优化后选择梯度变化最大的1w个像素点参与计算，能提高分辨率大的视频追踪效率。
4. 使用矩阵操作代替循环操作。(get_values())
5. 任务分解，在理解算法的基础上寻找并行机会，利用多线程或者多进程充分利用机器资源。生产者消费者模型，专门的线程负责图像获取和图形变换，专门的线程负责特征提取和追踪。
6. 使用C/C++重写效率低的瓶颈部分
7. 使用GPU计算




---
title: Python通过swig调用复杂C++库
categories:
- Python
---
　　
　　最近项目中需要用到visp库中的模板追踪算法，visp库用C++编写的，代码多，功能丰富。但是，对于项目来说直接调visp库并不方便，因此我们摘取visp库中的所需代码，提供python调用的接口，并根据项目需求进行优化和扩展。开源项目越来越多，以后工作也可能会遇到提取复杂库中部分功能，然后提供python调用的接口，因此这里总结一下，过程并不复杂，但是也遇到一些坑。主要注意以下几点：
1. **依赖库采用静态方式编译**。最开始的时候采用默认的动态编译，导致项目依赖复杂，部署起来非常不方便。要注意的是，visp库依赖opencv库，两个库都要采用静态方式编译。
2. **提取所需代码，封装成类**。提取visp库中的模板追踪算法，封装成类。为了便于后续的优化工作，接口扩展性尽可能好。
3. **swig实现python调用C++**。有很多方法实现python调用C++，我这里采用swig，适合懒人。

### 静态编译opencv库
　　opencv采用cmake项目管理，通过ccmake可以很方便的设置静态编译选项。BUILD_SHARED_LIBS设置为OFF即为静态编译。另外,为了保持系统整洁，避免安装到系统路径,设置了安装路径，CMAKE_INSTALL_PREFIX=/home/terse/code/terse-visp/opencv-3.4.6/build
1. git clone https://github.com/opencv/opencv.git
2. cd opencv-3.4.6 
3. mkdir build && cd build
4. ccmake ..(关闭动态编译选项，设置安装路径），cmake ..
5. make -j4
6. make install

### 静态编译visp库
　　visp库https://github.com/lagadic/visp.git 和opencv库一样都采用cmake管理，编译过程和opencv一样，这里只需要设置静态编译和设置安装路径：
关闭动态编译选项：BUILD_SHARED_LIBS=OFF
设置安装路径： CMAKE_INSTALL_PREFIX=/home/terse/code/terse-visp/visp/build

1. git clone https://github.com/lagadic/visp.git
2. cd visp
3. mkdir build && cd build
4. ccmake ..(关闭动态编译选项，设置安装路径），cmake ..
5. make -j4
6. make install

### 提取模板追踪算法,封装成C++类
　　visp库中提供了模板追踪算法，但是它不能解决遮挡的情况，参考区域很大的时候，追踪速度也很慢，因此在项目中针对这些问题做了一些优化，这个不是本文的重点就不赘述了。下面从visp中摘取的代码，封装成C++类，say_hello成员函数，没有实际用途，只是为了后续的验证python代码的正确性。
```
#ifndef VISP_H_
#define VISP_H_

#include <visp3/io/vpImageIo.h>
#include <visp3/tt/vpTemplateTrackerSSDInverseCompositional.h>
#include <visp3/tt/vpTemplateTrackerWarpHomography.h>
#include <visp3/core/vpException.h>  
#include <opencv2/opencv.hpp>
#include <fstream>

class TemplateTracker
{
public:
    TemplateTracker();

    void SetSampling(unsigned int sample_i, unsigned int sample_j);
    void SetLambda(double lamda);
    void SetIterationMax(unsigned int n);
    void SetPyramidal(unsigned int nlevels, unsigned int level_to_stop);
    void SetUseTemplateSelect(bool bselect);
    void SetThresholdGradient(float threshold);

    int Init(unsigned char* imgData, unsigned int h, unsigned int w, int* ref, unsigned int points_num, bool bshow);
    int InitWithMask(unsigned char* imgData, unsigned int h, unsigned int w, int* ref, unsigned int points_num, bool bshow, unsigned char* mask_data,int h2,int w2);

    int ComputeH(unsigned char* imgData, unsigned int h,unsigned int w,float* H_matrix,int num);
    int ComputeHWithMask(unsigned char* imgData, unsigned int h,unsigned int w,float* H_matrix,int num,unsigned char* mask_data,int h2,int w2);
    
    void Reset();
    ~TemplateTracker();
    void say_hello();


private:
    vpTemplateTrackerWarpHomography warp_;
    vpTemplateTrackerSSDInverseCompositional tracker_;
    int height_, width_;
    vpImage<unsigned char> I_;
    bool bshow_;
};

#endif /* VISP_H_ */
```

```
#include "visp.h"  


TemplateTracker::TemplateTracker() :
    tracker_(&warp_),
    bshow_(false)
{};

void TemplateTracker::SetSampling(unsigned int sample_i, unsigned int sample_j)
{
    tracker_.setSampling(sample_i, sample_j);
}

void TemplateTracker::SetLambda(double lamda)
{
    tracker_.setLambda(lamda);
}

void TemplateTracker::SetIterationMax(unsigned int n)
{
    tracker_.setIterationMax(n);
}

void TemplateTracker::SetPyramidal(unsigned int nlevels, unsigned int level_to_stop)
{
    tracker_.setPyramidal(nlevels, level_to_stop);
}

void TemplateTracker::SetUseTemplateSelect(bool bselect)
{
    tracker_.setUseTemplateSelect(bselect);
}

void TemplateTracker::SetThresholdGradient(float threshold)
{
    tracker_.setThresholdGradient(threshold);
}


int TemplateTracker::Init(unsigned char* imgData, unsigned int h, unsigned int w, int* ref, unsigned int points_num, bool bshow)
{
    cv::Mat img_gray(h, w, CV_8UC1, (unsigned char*)(imgData));  //浅拷贝
    I_.init(imgData, h, w, true);
    height_ = h;
    width_ = w;
    bshow_ = bshow;
    std::vector<vpImagePoint> v_ip;
    for (int i = 0; i < points_num/2; i++)
    {
        vpImagePoint ip(ref[i * 2], ref[i * 2 + 1]);
        v_ip.push_back(ip);
    }

    try{
        tracker_.initFromPoints(I_, v_ip);
    }catch(vpException &e){
        return e.getCode();
    }

    return 0;
}


int TemplateTracker::InitWithMask(unsigned char* imgData, unsigned int h, unsigned int w, int* ref, unsigned int points_num, bool bshow, unsigned char* mask_data,int h2,int w2)
{
    cv::Mat img_gray(h, w, CV_8UC1, (unsigned char*)(imgData));  //浅拷贝
    I_.init(imgData, h, w, true);
    height_ = h;
    width_ = w;
    bshow_ = bshow;
    if (NULL != mask_data)
    {
        cv::Mat mask_gray(h, w, CV_8UC1, (unsigned char*)(mask_data));  //浅拷贝
        I_.SetMask(mask_gray);
    }

    std::vector<vpImagePoint> v_ip;
    for (int i = 0; i < points_num/2; i++)
    {
        vpImagePoint ip(ref[i * 2], ref[i * 2 + 1]);
        v_ip.push_back(ip);
    }

    try{
        tracker_.initFromPoints(I_, v_ip);
    }catch(vpException &e){
        return e.getCode();
    }

    return 0;
}



int TemplateTracker::ComputeH(unsigned char* imgData, unsigned int h,unsigned int w,float* H_matrix,int num)
{
    I_.init(imgData, height_, width_, true);
    try{
        tracker_.track(I_);
    }catch(vpTrackingException &e){
        std::cout << e.getMessage() << std::endl;
        return e.getCode();
    }
    vpColVector p = tracker_.getp();
    vpHomography H = warp_.getHomography(p);
    for (int m = 0; m < 3; m++)
    {
        for (int n = 0; n < 3; n++)
        {
            H_matrix[m * 3 + n] = H[m][n];
        }
    }
    return 0;
}


int TemplateTracker::ComputeHWithMask(unsigned char* imgData, unsigned int h,unsigned int w,float* H_matrix,int num,unsigned char* mask_data,int h2,int w2)
{
    I_.init(imgData, height_, width_, true);
    if (NULL != mask_data)
    {
        cv::Mat mask_gray(height_, width_, CV_8UC1, (unsigned char*)(mask_data));  //浅拷贝
        I_.SetMask(mask_gray);
    }

    try{
        tracker_.track(I_);
    }catch(vpTrackingException &e){
        std::cout << e.getMessage() << std::endl;
        return e.getCode();
    }
    vpColVector p = tracker_.getp();
    vpHomography H = warp_.getHomography(p);
    for (int m = 0; m < 3; m++)
    {
        for (int n = 0; n < 3; n++)
        {
            H_matrix[m * 3 + n] = H[m][n];
        }
    }
    return 0;
}

void TemplateTracker::Reset()
{
    tracker_.resetTracker();
}

TemplateTracker::~TemplateTracker(){};

void TemplateTracker::say_hello(){
    std::cout << "hello" << std::endl;
}
```


### 采用swig实现python调用C++
　　python调用C++的方法有很多，例如ctypes、PyObject、Boost.python,采用了swig方法，使用之后感觉确挺方便的。为了给追踪功能提供numpy参数的输入和输出，这里需要引入numpy.i文件。
参考：http://www.swig.org/Doc1.3/Python.html#Python

#### 1. 定义接口文件：visp.i
```
/* File: visp.i */
%module visp

%{
  #define SWIG_FILE_WITH_INIT
  #include "visp.h"
%}

%include "numpy.i"

%init %{
    import_array();
%}

%apply (unsigned char* IN_ARRAY2, int DIM1, int DIM2) {(unsigned char* imgData, unsigned int h, unsigned int w)}
%apply (unsigned char* IN_ARRAY2, int DIM1, int DIM2) {(unsigned char* mask_data, int h2, int w2)}
%apply (int* IN_ARRAY1, int DIM1) {(int* ref, unsigned int points_num)}
%apply (float* INPLACE_ARRAY1, int DIM1) {(float* H_matrix,int num)}

%include "visp.h"
```
#### 2. swig 编译visp.i 文件生成C++和py代码，生成visp_wrap.cxx,visp.py
```
swig -c++ -python -py3 visp.i //python3
```
#### 3. 分别编译visp.cc和visp_wrap.cxx代码
```
g++  -O2 -fPIC  -c visp.cc -I/home/terse/code/terse-visp/VispSource/build/include
g++ -O2 -fPIC -c visp_wrap.cxx -I/home/terse/anaconda3/include/python3.6m -I/home/terse/code/terse-visp/VispSource/build/include -I//home/terse/anaconda3/lib/python3.6/site-packages/numpy/core/include/
```
#### 4. 链接生成_visp.so文件
```
g++ -shared visp_wrap.o visp.o -L/home/terse/code/terse-visp/VispSource/build/lib -lvisp_ar -lvisp_blob -lvisp_core -lvisp_detection -lvisp_core -lvisp_gui -lvisp_imgproc -lvisp_io -lvisp_klt -lvisp_mbt -lvisp_me -lvisp_robot -lvisp_sensor -lvisp_tt -lvisp_tt_mi -lvisp_vision -lvisp_visual_features -lvisp_vs -lvisp_tt  -lvisp_ar -lvisp_blob -lvisp_core -lvisp_detection -lvisp_core -lvisp_gui -lvisp_imgproc -lvisp_io -lvisp_klt -lvisp_mbt -lvisp_me -lvisp_robot -lvisp_sensor -lvisp_tt -lvisp_tt_mi -lvisp_vision -lvisp_visual_features -lvisp_vs -Wl,-Bstatic -L/home/terse/code/terse-visp/opencv-3.4.6/build/lib -lopencv_dnn -lopencv_ml -lopencv_objdetect -lopencv_shape -lopencv_stitching -lopencv_superres -lopencv_videostab -lopencv_calib3d -lopencv_features2d -lopencv_highgui -lopencv_videoio -lopencv_imgcodecs -lopencv_video -lopencv_photo -lopencv_imgproc -lopencv_flann -lopencv_core -Wl,-Bstatic -L/home/terse/code/terse-visp/opencv-3.4.6/build/share/OpenCV/3rdparty/lib -littnotify -llibprotobuf -llibjasper -lquirc -lippiw -lippicv -Wl,-Bdynamic -lpython3.7m -Wl,-Bdynamic  -llapack  -fopenmp -ldl  -lz -lrt -ltiff -o _visp.so


```
　　这里有个坑纠结了挺久了，最开始生成的_visp.so文件中通过ldd -r 查看一直有几个未定义的符号。排查后发现来自opencv_core库中，通过nm查看，发现libopencv_core.a中是未定义的，而libopencv_sore.so是正常的。最后发现那几个未定义的符号在/home/terse/code/terse-visp/opencv-3.4.6/build/3rdparty/lib/libittnotify.a库中，链接时加入这个库就将未定义的符号解决了。这个链接文件中有许多依赖的库是不需要的，这里就没有仔细排查了，只要没少就能链接成功。

### 简单测试
　　通过ldd -r 检查_visp.so文件没有问题，理论上就没什么问题里，这里通过代码中故意遗留的函数测试一下。
``` 
import visp
import cv2
import numpy as np

def get_frame(cap, frame_index):
    pos = cap.get(cv2.CAP_PROP_POS_FRAMES)
    if pos != frame_index:
        cap.set(cv2.CAP_PROP_POS_FRAMES, frame_index)
    ret, frame = cap.read()
    gray = cv2.cvtColor(frame, cv2.COLOR_BGR2GRAY)
    return gray

if __name__ == '__main__':
    video_name = "./test_data/1166/input.mp4"

    ref_area = [105, 73, 479, 62, 126, 309, 120, 297, 471, 57, 457, 291]
    key_frame = 5520

    cap = cv2.VideoCapture(video_name) #video name
    tracker = visp.TemplateTracker()
    tracker.SetLambda(0.001)
    tracker.SetPyramidal(3,0)
    tracker.SetIterationMax(200)
    tracker.SetSampling(1,1)  #x和y方向的降采样率
    tracker.say_hello()
    img = get_frame(cap, key_frame)
    ret_code = tracker.Init(img,ref_area,True)

    H_array = np.empty(9,dtype=np.float32)
    img = get_frame(cap, key_frame+1)
    ret_code = tracker.ComputeH(img,H_array)
    print(ret_code,H_array)

```



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