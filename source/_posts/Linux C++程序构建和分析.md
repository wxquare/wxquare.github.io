---
title: Linux C++ 程序构建和分析
categories:
- other
---

## 1、C/C++项目构建
- 配置 configure
- 编译 make、cmake、ccmake
- 安装 make install

[如何编写makefile？](http://scc.qibebt.cas.cn/docs/linux/base/%B8%FA%CE%D2%D2%BB%C6%F0%D0%B4Makefile-%B3%C2%F0%A9.pdf)  
如何使用cmake构建C++项目？  


## 2、程序调试
[gdb调试必知必会](https://linuxtools-rst.readthedocs.io/zh_CN/latest/tool/gdb.html#gdb)


## 3、目标文件分析
1. nm
2. readelf/objdump
3. size 查看程序各部分大小
4. ldd 查看程序的依赖信息


## 4、程序性能分析
- top
- time

## 5、程序效率优化
- pstack 跟踪栈空间，可显示每个进程的栈跟踪
- strace 分析程序的系统调用
- prof
- gprof 
	- 用gcc、g++、xlC编译程序时，使用-pg参数，如：g++ -pg -o test.exe test.cpp编译器会自动在目标代码中插入用于性能测试的代码片断，这些代码在程序运行时采集并记录函数的调用关系和调用次数，并记录函数自身执行时间和被调用函数的执行时间。
	- 执行编译后的可执行程序，如：./test.exe。该步骤运行程序的时间会稍慢于正常编译的可执行程序的运行时间。程序运行结束后，会在程序所在路径下生成一个缺省文件名为gmon.out的文件，这个文件就是记录程序运行的性能、调用关系、调用次数等信息的数据文件。
	- 使用gprof命令来分析记录程序运行信息的gmon.out文件，如：gprof test.exe gmon.out则可以在显示器上看到函数调用相关的统计、分析信息。上述信息也可以采用gprof test.exe gmon.out> gprofresult.txt重定向到文本文件以便于后续分析。    

参考：http://www.cnblogs.com/me115/archive/2013/06/05/3117967.html


## 6、程序内存泄露分析
- 调试内存泄漏的工具valgrind



## 7、cmake 通过find_package使用第三方库
https://www.jianshu.com/p/46e9b8a6cb6a
find_package原理
首先明确一点，cmake本身不提供任何搜索库的便捷方法，所有搜索库并给变量赋值的操作必须由cmake代码完成，比如下面将要提到的FindXXX.cmake和XXXConfig.cmake。只不过，库的作者通常会提供这两个文件，以方便使用者调用。
find_package采用两种模式搜索库：

Module模式：搜索CMAKE_MODULE_PATH指定路径下的FindXXX.cmake文件，执行该文件从而找到XXX库。其中，具体查找库并给XXX_INCLUDE_DIRS和XXX_LIBRARIES两个变量赋值的操作由FindXXX.cmake模块完成。

Config模式：搜索XXX_DIR指定路径下的XXXConfig.cmake文件，执行该文件从而找到XXX库。其中具体查找库并给XXX_INCLUDE_DIRS和XXX_LIBRARIES两个变量赋值的操作由XXXConfig.cmake模块完成。

两种模式看起来似乎差不多，不过cmake默认采取Module模式，如果Module模式未找到库，才会采取Config模式。如果XXX_DIR路径下找不到XXXConfig.cmake文件，则会找/usr/local/lib/cmake/XXX/中的XXXConfig.cmake文件。总之，Config模式是一个备选策略。通常，库安装时会拷贝一份XXXConfig.cmake到系统目录中，因此在没有显式指定搜索路径时也可以顺利找到。




----------------------
## 常见问题解决：
### 1、pkg-config：解决编译和链接时候出现的undefined symbol问题
pkg-config能方便使用第三方库和头文件和库文件，其运行原理 
 
- 它首先根据PKG_CONFIG_PATH环境变量下寻找库对应的pc文件  
- 然后从pc文件中获取该库对应的头文件和库文件的位置信息
  
例如在项目中需要使用opencv库，该库包含的头文件和库文件比较多  

- 首先查看是否有对应的opencv.pc find /usr -name opencv.pc  
- 查看该路径是否包含在PKG_CONFIG_PATH  
- 使用pkg-config --cflags --libs opencv 查看库对应的头文件和库文件信息  
- pkg-config --modversion opencv 查看版本信息

参考链接：[https://blog.csdn.net/luotuo44/article/details/24836901](https://blog.csdn.net/luotuo44/article/details/24836901)

### 2、ldd：帮助解决运行时出现的问题
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

实战1：ffmpage的编译
a. 下载地址：http://www.ffmpeg.org/releases/ffmpeg-4.1.4.tar.bz2
b. 查看ffmpeg配置参数：./configure --help
c. 编译机器 taf.YYBOP.VideoImplant.TSZ19
d. sudo yum install yasm  或者 -disable-x86asm 选项
e. 创建build_debug目录，编译debug版本
./configure --prefix=/usr/local/app/terse/ffmpeg-4.1.4/build_debug --enable-shared  --shlibdir=/usr/local/app/terse/ffmpeg-4.1.4/build_debug/lib
make -j16
make install
d. 创建build_release目录,编译release版本
./configure --prefix=/home/terse/code/C++/KCF2/third_party/FFmpeg-release-4.1/build --enable-shared  --shlibdir=/home/terse/code/C++/KCF2/third_party/FFmpeg-release-4.1/build/lib --disable-debug
make -j16
make install

实战2：opencv的编译
a. opencv下载地址：https://github.com/opencv/opencv/archive/3.4.6.zip
b. opencv_contrib下载地址： https://github.com/opencv/opencv_contrib/archive/3.4.6.zip
c. 编译机器 taf.YYBOP.VideoImplant.TSZ19
d. 通过ccmake配置参数
release版本：
OPENCV_EXTRA_MODULES_PATH=/usr/local/app/terse/opencv_contrib-3.4.6/modules
BUILD_SHARED_LIBS=ON
CMAKE_BUILD_TYPE=Release
CMAKE_INSTALL_PREFIX=/usr/local/app/terse/opencv-3.4.6/release 
Debug版本：
OPENCV_EXTRA_MODULES_PATH=/usr/local/app/terse/opencv_contrib-3.4.6/modules
BUILD_SHARED_LIBS=ON
CMAKE_BUILD_TYPE=Debug
CMAKE_INSTALL_PREFIX=/usr/local/app/terse/opencv-3.4.6/Debug 