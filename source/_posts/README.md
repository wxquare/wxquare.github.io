功能描述
--------------
基于KCF2.0开发的命令行追踪程序，指定视频、追踪区域（矩形）、追踪的开始帧、结束帧的索引，将追踪结果以csv格式保存在文件中。  
**用法：**  
　　**usage: ./kcf2.0 video_name left_top_x left_top_y width height start_index end_index result_file debug**

**参数说明：**  
- vdieo_name: 指定视频  
- left_top_x: 追踪区域左上角的x坐标  
- left_top_y: 追踪区域左上角的y坐标  
- width： 追踪区域的宽度  
- height： 追踪区域的高度  
- start_index: 从哪一帧开始追踪  
- end_index: 追踪哪一帧结束  
- result_file: 追踪结果保存的文件，每一行表示一帧的追踪结果，例如工程目录下的data/result.csv  
- debug: 对于有UI支持的系统，debug为0可以显示追踪过程，通常设置为1

例如：
    ./bin/kcf_20190620 ./data/bag.mp4 312 146 106 98 1 196 ./data/result.csv 0

目录结构
--------------
    ├── bin
    │   ├── kcf_20190620                    公司开发机上编译完成
    │   └── kcf_terse                       个人ubuntu
    ├── build                               编译目录
    ├── CMakeLists.txt
    ├── data                                准备的测试数据
    │   ├── bag.mp4                         追踪的视频
    │   ├── result.csv                      追踪的结果保存结果
    │   └── track_area                      追踪区域
    ├── main_KCF2.0.cpp    
    ├── main_terse.cpp                      主函数 
    ├── README.md
    ├── src                                 kcf核心源码
    │   ├── CMakeLists.txt
    │   ├── cn
    │   ├── complexmat.hpp
    │   ├── kcf.cpp
    │   ├── kcf.h
    │   └── piotr_fhog
    └── third_party                         第三方依赖
        ├── FFmpeg-release-4.1              最新稳定版本的FFmpeg
        └── opencv-3.4.6                    版本要求大于2.4,3.4.6版本验证可行



源码编译
------------------
项目所有依赖采用静态编译，正常情况下，可使用bin目录下面的kcf可执行程勋运行追踪功能
如遇到兼容性问题，可重新编译生成适合特定环境的可执行程序。

**1. 编译ffmpeg**
- ffmpeg默认采用的是静态编译，如果需要使用动态编译可通过--enable-shared指定
- 进入FFmpeg-release-4.1目录，执行./configure --prefix=/usr --disable-x86asm
- make -j8
- sudo make install

**2. 编译opencv**
- opencv使用cmake管理项目，设置了许多的开关选项，ccmake可方便设选项
- cd opencv-3.4.6
- mkdir build
- cd build
- ccmake ..(关闭动态编译选项，设置BUILD_SHARED_LIBS为OFF)
- cmake ..
- make
- sudo make install

**3. 编译kcf2.0**
- 进入build
- 执行ccmake ..，可知道依赖Opencv，指定OpenCV目录
- make，在build目录生成kcf2.0可执行文件


性能测试：
-------
测试容器：CONTAINER_NAME=taf.YYBOP.VideoImplant.PSZ21
	[mqq@10-57-20-60 ~/terse/KCF2.0/build]$ time ./kcf_20190620 bag.mp4 312 146 106 98 1 196 result.csv 1  
	real    0m1.593s
	user    0m21.380s
	sys     0m6.140s


更新日志
------
v20190620:
- 项目构建
- 全部采用静态编译
- opencv读视频帧依赖于ffpmeg，在编译opencv时需要加上ffmpeg依赖，否则会出现：VIDIOC_REQBUFS: Inappropriate ioctl for device
