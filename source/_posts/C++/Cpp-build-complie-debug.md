---
title: C/C++程序的项目构建、编译、调试工具和方法
categories:
- C/C++
---

　　在Linux C/C++项目实践中，随和项目越来越复杂，第三方依赖项的增加，有时会遇到一些编译、链接和调试问题，这里总结一下遇到的问题、解决的办法和使用到的工具：
1. 了解gcc/g++编译过程、常见和编译选项解决编译过程遇到的问题
2. 了解链接过程、动态链接、静态链接，解决链接过程中遇到的问题
4. 解决程序运行出现的依赖问题、符号未定义问题
5. 学习会使用gdb调试一些基本问题
5. 学会使用makefile和cmake工具构建项目


## 一、排查编译问题常用工具
### 1. gcc/g++的区别和使用
1. 后缀为.c的，gcc把它当作是C程序，而g++当作是c++程序；后缀为.cpp的，两者都会认为是c++程序，注意，虽然c++是c的超集，但是两者对语法的要求是有区别的
2. 对于C代码，编译和链接都使用gcc
3. 对于C++代码，编译时可以使用gcc/g++，gcc实际也是调用g++；链接时gcc 不能自动和C++使用库链接，因此要使用g++或者gcc -lstdc++  

### 2. 常见gcc编译链接选项
- -c 只编译并生成目标文件
- -g 生成调试信息，gdb可以利用该调试信息
- -o 指定生成的输出文件，可执行程序或者动态链接库文件名
- -I 编译时添加头文件路径
- -L 链接时添加库文件路径
- -D 定义宏，常用于开关控制代码
- -shared 用于生成共享库.so
- -Wall 显示所有警告信息，-w不生成任何警告信息
- -O0选项不进行任何优化，debug会产出和程序预期的结果；O1优化会消耗少多的编译时间，它主要对代码的分支，常量以及表达式等进行优化;O2会尝试更多的寄存器级的优化以及指令级的优化，它会在编译期间占用更多的内存和编译时间。 通常情况下线上代码至少加上O2优化选项。
- -fPIC 位置无关选项，生成动态库时使用，实现真正意义上的多进程共享的.so库。
- -Wl选项告诉编译器将后面的参数传递给链接器
- -Wl,-Bstatic，指明后面是链接今静态库
- -Wl,-Bdynamic,指明后面是链接动态库

### 3. 编译时添加头文件依赖路径
　　-include用来包含头文件，但一般情况下包含头文件都在源码里用#include xxxxxx实现，-include参数很少用。-I参数是用来指定头文件目录，/usr/include目录一般是不用指定的，gcc知道去那里找，但 是如果头文件不在/usr/include里我们就要用-I参数指定了，比如头文件放在/myinclude目录里，那编译命令行就要加上-I /myinclude参数了，如果不加你会得到一个"xxxx.h: No such file or directory"的错误。-I参数可以用相对路径，比如头文件在当前目录，可以用-I.来指定。

## 二、排查链接问题常用工具
1. 查看ld链接器的搜索顺序 ld --verbose | grep SEARCH
2. 链接时指定链接目录 -L/dir
3. -Wl,-Bstatic，指明后面是链接今静态库
4. -Wl,-Bdynamic,指明后面是链接动态库  
5. 运行时找不到动态库so文件，设置LD_LIBRARY_PATH，添加依赖so文件所在路径
6. 链接完成后使用ldd查看动态库依赖关系，如果依赖的某个库找不到，通过这个命令可以迅速定位问题所在
7. ldd -r，帮助检查是否存在未定义的符号undefine symbol,so库链接状态和错误信息



## 三、gdb调试基本使用
### 1. 对C/C++程序的调试，需要在编译前就加上-g选项。
1. $gdb <programe>
2. 设置参数：set args 可指定运行时参数。（如：set args 10 20 30 40 50） 

### 2. 查看源代码
- list ：简记为 l ，其作用就是列出程序的源代码，默认每次显示10行。
- list 行号：将显示当前文件以“行号”为中心的前后10行代码，如：list 12
- list 函数名：将显示“函数名”所在函数的源代码，如：list main
- list ：不带参数，将接着上一次 list 命令的，输出下边的内容

### 3. 设置断点和关闭断点
- break n （简写b n）: 在第n行处设置断点（可以带上代码路径和代码名称： b test.cpp:578）
- break func（简写b func): 在函数func()的入口处设置断点，如：break test_func
- info b （info breakpoints)：显示当前程序的断点设置情况
- delete 断点号n：删除第n个断点
- disable 断点号n：暂停第n个断点
- clear 行号n：清除第n行的断点

### 4. 程序调试运行
- run：简记为 r ，其作用是运行程序，当遇到断点后，程序会在断点处停止运行，等待用户输入下一步的命令。
- continue （简写c ）：继续执行，到下一个断点处（或运行结束）
- next：（简写 n），单步跟踪程序，当遇到函数调用时，也不进入此函数体；此命令同 step 的主要区别是，step 遇到用户自定义的函数，将步进到函数中去运行，而 next 则直接调用函数，不会进入到函数体内。
- step （简写s）：单步调试如果有函数调用，则进入函数；与命令n不同，n是不进入调用的函数的
- until：当你厌倦了在一个循环体内单步跟踪时，这个命令可以运行程序直到退出循环体。
- until+行号： 运行至某行，不仅仅用来跳出循环
- finish： 运行程序，直到当前函数完成返回，并打印函数返回时的堆栈地址和返回值及参数值等信息。
- call 函数(参数)：调用程序中可见的函数，并传递“参数”，如：call gdb_test(55)
- quit：简记为 q ，退出gdb

### 5. 打印程序运行的调试信息
- print 表达式：简记为 p ，其中“表达式”可以是任何当前正在被测试程序的有效表达式，比如当前正在调试C语言的程序，那么“表达式”可以是任何C语言的有效表达式，包括数字，变量甚至是函数调用。
- print a：将显示整数 a 的值
- print name：将显示字符串 name 的值
- print gdb_test(22)：将以整数22作为参数调用 gdb_test() 函数
- print gdb_test(a)：将以变量 a 作为参数调用 gdb_test() 函数
- 扩展info locals： 显示当前堆栈页的所有变量

### 6. 查询运行信息
- where/bt ：当前运行的堆栈列表；
- bt backtrace 显示当前调用堆栈
- up/down 改变堆栈显示的深度
- set args 参数:指定运行时的参数
- show args：查看设置好的参数
- info program： 来查看程序的是否在运行，进程号，被暂停的原因。


## 四、gdb调试coredump问题
 　　Coredump叫做核心转储，它是进程运行时在突然崩溃的那一刻的一个内存快照。操作系统在程序发生异常而异常在进程内部又没有被捕获的情况下，会把进程此刻内存、寄存器状态、运行堆栈等信息转储保存在一个文件里。该文件也是二进制文件，可以使用gdb调试。虽然我们知道进程在coredump的时候会产生core文件，但是有时候却发现进程虽然core了，但是我们却找不到core文件。在ubuntu系统中需要进行设置，ulimit  -c 可以设置core文件的大小，如果这个值为0.则不会产生core文件，这个值太小，则core文件也不会产生，因为core文件一般都比较大。使用**ulimit  -c unlimited**来设置无限大，则任意情况下都会产生core文件。
 　　gdb打开core文件时，有显示没有调试信息，因为之前编译的时候没有带上-g选项，没有调试信息是正常的，实际上它也不影响调试core文件。因为调试core文件时，符号信息都来自符号表，用不到调试信息。如下为加上调试信息的效果。
 调试步骤：
 ＄gdb program core_file 进入
 $ bt或者where # 查看coredump位置
 当程序带有调试信息的情况下，我们实际上是可以看到core的地方和代码行的匹配位置。但往往正常发布环境是不会带上调试信息的，因为调试信息通常会占用比较大的存储空间，一般都会在编译的时候把-g选项去掉。这种情况啊也是可以通过core_dump文件找到错误位置的，但这个过程比较复杂，参考：https://blog.csdn.net/u014403008/article/details/54174109

## 五、gdb调试线上死锁问题
　　如果你的程序是一个服务程序，那么你可以指定这个服务程序运行时的进程ID。gdb会自动attach上去，并调试。对于服务进程，我们除了使用gdb调试之外，还可以使用pstack跟踪进程栈。这个命令在排查进程问题时非常有用，比如我们发现一个服务一直处于work状态（如假死状态，好似死循环），使用这个命令就能轻松定位问题所在；可以在一段时间内，多执行几次pstack，若发现代码栈总是停在同一个位置，那个位置就需要重点关注，很可能就是出问题的地方。gdb比pstack更加强大，gdb可以随意进入进程、线程中改变程序的运行状态和查看程序的运行信息。思考：如何调试死锁？
$gdb <program> <PID>
$pstack pid


## 六、undefined symbol问题解决步骤
1. file 检查so或者可执行文件的架构
```
$ file _visp.so 
_visp.so: ELF 64-bit LSB pie executable, x86-64, version 1 (GNU/Linux), dynamically linked, BuildID[sha1]=6503ba6b7545e38e669ab9ed31f86449d8a5f78b, stripped
```
2. ldd -r _visp.so 命令查看so库链接状态和错误信息
```
undefined symbol: __itt_api_version_ptr__3_0	(./_visp.so)
undefined symbol: __itt_id_create_ptr__3_0	(./_visp.so)
```
3. c++filt symbol 定位错误在那个C++文件中
```
base) terse@ubuntu:~/code/terse-visp$ c++filt __itt_domain_create_ptr__3_0
__itt_domain_create_ptr__3_0
```
4. 还可以使用grep -R __itt_domain_create_ptr__3_0 ./
最终发现这个符号来自XXX/opencv-3.4.6/build/share/OpenCV/3rdparty/libittnotify.a

5. 通过nm命令也能看出该符号确实未定义
```
$ nm _visp.so | grep __itt_domain_create_ptr__3_0
      U __itt_domain_create_ptr__3_0
```


## 七、pkg-config 找第三方库的头文件和库文件
pkg-config能方便使用第三方库和头文件和库文件，其运行原理 
- 它首先根据PKG_CONFIG_PATH环境变量下寻找库对应的pc文件  
- 然后从pc文件中获取该库对应的头文件和库文件的位置信息
  
例如在项目中需要使用opencv库，该库包含的头文件和库文件比较多  
- 首先查看是否有对应的opencv.pc find /usr -name opencv.pc  
- 查看该路径是否包含在PKG_CONFIG_PATH  
- 使用pkg-config --cflags --libs opencv 查看库对应的头文件和库文件信息  
- pkg-config --modversion opencv 查看版本信息
参考链接：[https://blog.csdn.net/luotuo44/article/details/24836901](https://blog.csdn.net/luotuo44/article/details/24836901)


## 八、cmake中的find_package
https://www.jianshu.com/p/46e9b8a6cb6a
find_package原理
首先明确一点，cmake本身不提供任何搜索库的便捷方法，所有搜索库并给变量赋值的操作必须由cmake代码完成，比如下面将要提到的FindXXX.cmake和XXXConfig.cmake。只不过，库的作者通常会提供这两个文件，以方便使用者调用。
find_package采用两种模式搜索库：

Module模式：搜索CMAKE_MODULE_PATH指定路径下的FindXXX.cmake文件，执行该文件从而找到XXX库。其中，具体查找库并给XXX_INCLUDE_DIRS和XXX_LIBRARIES两个变量赋值的操作由FindXXX.cmake模块完成。

Config模式：搜索XXX_DIR指定路径下的XXXConfig.cmake文件，执行该文件从而找到XXX库。其中具体查找库并给XXX_INCLUDE_DIRS和XXX_LIBRARIES两个变量赋值的操作由XXXConfig.cmake模块完成。

两种模式看起来似乎差不多，不过cmake默认采取Module模式，如果Module模式未找到库，才会采取Config模式。如果XXX_DIR路径下找不到XXXConfig.cmake文件，则会找/usr/local/lib/cmake/XXX/中的XXXConfig.cmake文件。总之，Config模式是一个备选策略。通常，库安装时会拷贝一份XXXConfig.cmake到系统目录中，因此在没有显式指定搜索路径时也可以顺利找到。

## 九、ldd解决运行时问题
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

## 十、makefile和cmake的使用
- [跟我学些makefile](https://github.com/wxquare/programming/blob/master/document/%E8%B7%9F%E6%88%91%E4%B8%80%E8%B5%B7%E5%86%99Makefile-%E9%99%88%E7%9A%93.pdf)
- [CMake入门实战](https://www.hahack.com/codes/cmake/)


## 其它问题
1. c++进程内存空间分布
2. ELF是什么？其大小与程序中全局变量的是否初始化有什么关系（注意.bss段）、elf文件格式和运行时内存布局
3. 标准库函数和系统调用的区别
4. 编译器内存对齐和内存对齐的原理
5. 编译器如何区分C和C++？
6. C++动态链接库和静态链接库？如何创建和使用静态链接库和动态链接库？（fPIC, shared）
8. 如何判断计算机的字节序是大端还是小端的？
9. 预编译、编译、汇编、链接
10. GDB的基本工作原理是什么？和断点调试的实现原理：在程序中设置断点，现将该位置原来的指令保存，然后向该位置写入int 3，当执行到int 3的时候，发生软中断。内核会给子进程发出sigtrap信号，当然这个信号首先被gdb捕获，gdb会进行断点命中判定，如果命中的话就会转入等待用户输入进行下一步的处理，否则继续运行，替换int 3，恢复执行
12. gdb调试、coredump、调试运行中的程序？通过ptrace让父进程可以观察和控制其它进程的执行，检查和改变其核心映像以及寄存器，主要通过实现断电调试和系统调用跟踪。
119. 编译器的编译过程？链接的时候做了什么事？在中间层优化时怎么做?编译。词法分析、句法分析、语义分析生成中间的汇编代码。汇编，链接：静态链接库、动态链接库
5.	gcc 和 g++的区别
6.	项目构建工具makefile、cmake
20. 预处理：#include文件、条件预编译指令、注释。保留#pargma编译器指令
21. valgrind(内存、堆栈、函数调用、多线程竞争、缓存，可扩展)，valgrind内存检查的原理、和具体使用！
22. C++内存管理：内存布局、堆栈的区别、内存操作四个原则、内存泄露检查、智能指针、STL内存管理(内存池)
23. gdb调试多进程和多线程命令



参考：
[1]. gdb 调试利器:https://linuxtools-rst.readthedocs.io/zh_CN/latest/tool/gdb.html
[2]. 陈皓专栏gdb调试系列：https://blog.csdn.net/haoel/article/details/2879
[3]. gdb core_dump调试：https://blog.csdn.net/u014403008/article/details/54174109
[4]. 进程调试，死循环和死锁卡死：https://blog.csdn.net/guowenyan001/article/details/46238355




