###1、C/C++项目构建
- 配置 configure
- 编译 make、cmake、ccmake
- 安装 make install

[如何编写makefile？](http://scc.qibebt.cas.cn/docs/linux/base/%B8%FA%CE%D2%D2%BB%C6%F0%D0%B4Makefile-%B3%C2%F0%A9.pdf)  
如何使用cmake构建C++项目？  


##2、程序调试
[gdb调试必知必会](https://linuxtools-rst.readthedocs.io/zh_CN/latest/tool/gdb.html#gdb)


##3、目标文件分析
1. nm
2. readelf
3. objdump
4. ldd 查看陈程序的依赖信息


##4、程序性能分析
- top
- time

##5、程序效率优化
- pstack 跟踪栈空间，可显示每个进程的栈跟踪
- strace 分析程序的系统调用
- prof
- gprof 
	- 用gcc、g++、xlC编译程序时，使用-pg参数，如：g++ -pg -o test.exe test.cpp编译器会自动在目标代码中插入用于性能测试的代码片断，这些代码在程序运行时采集并记录函数的调用关系和调用次数，并记录函数自身执行时间和被调用函数的执行时间。
	- 执行编译后的可执行程序，如：./test.exe。该步骤运行程序的时间会稍慢于正常编译的可执行程序的运行时间。程序运行结束后，会在程序所在路径下生成一个缺省文件名为gmon.out的文件，这个文件就是记录程序运行的性能、调用关系、调用次数等信息的数据文件。
	- 使用gprof命令来分析记录程序运行信息的gmon.out文件，如：gprof test.exe gmon.out则可以在显示器上看到函数调用相关的统计、分析信息。上述信息也可以采用gprof test.exe gmon.out> gprofresult.txt重定向到文本文件以便于后续分析。    

参考：http://www.cnblogs.com/me115/archive/2013/06/05/3117967.html


##6、程序内存泄露分析
- 调试内存泄漏的工具valgrind

----------------------
##常见问题解决：
### 1、pkg-config：解决编译和链接时候出现的undefined symbol问题
pkg-config能方便使用第三方库和头文件和库文件，其运行原理 
 
- 它首先根据PKG_CONFIG_PATH环境变量下寻找库对应的pc文件  
- 然后从pc文件中获取该库对应的头文件和库文件的位置信息
  
例如在项目中需要使用opencv库，该库包含的头文件和库文件比较多  

- 首先查看是否有对应的opencv.pc find /usr -name opencv.pc  
- 查看该路径是否包含在PKG_CONFIG_PATH  
- 使用pkg-config --cflgs --libs opencv 查看库对应的头文件和库文件信息  
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
