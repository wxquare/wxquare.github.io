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