---
title: Linux程序常见内存问题和Valgrind
categories: 
- 计算机基础
---


## Linux程序的内存问题
1. 使用未初始化的变量
对于位于程序中不同段的变量，其初始值是不同的，全局变量和静态变量初始值为0，而局部变量和动态申请的变量，其初始值为随机值。如果程序使用了为随机值的变量，那么程序的行为就变得不可预期。

2. 内存访问越界
比如访问数组时越界；对动态内存访问时超出了申请的内存大小范围。


3. 内存覆盖
C 语言的强大和可怕之处在于其可以直接操作内存，C 标准库中提供了大量这样的函数，比如 strcpy, strncpy, memcpy, strcat 等，这些函数有一个共同的特点就是需要设置源地址 (src)，和目标地址(dst)，src 和 dst 指向的地址不能发生重叠，否则结果将不可预期。


4. 动态内存管理错误
常见的内存分配方式分三种：静态存储，栈上分配，堆上分配。全局变量属于静态存储，它们是在编译时就被分配了存储空间，函数内的局部变量属于栈上分配，而最灵活的内存使用方式当属堆上分配，也叫做内存动态分配了。常用的内存动态分配函数包括：malloc, alloc, realloc, new等，动态释放函数包括free, delete。一旦成功申请了动态内存，我们就需要自己对其进行内存管理，而这又是最容易犯错误的。下面的一段程序，就包括了内存动态管理中常见的错误。
a. 使用完后未释放
b. 释放后仍然读写
c. 释放了再释放

5. 内存泄露
内存泄露（Memory leak）指的是，在程序中动态申请的内存，在使用完后既没有释放，又无法被程序的其他部分访问。内存泄露是在开发大型程序中最令人头疼的问题，以至于有人说，内存泄露是无法避免的。其实不然，防止内存泄露要从良好的编程习惯做起，另外重要的一点就是要加强单元测试（Unit Test），而memcheck就是这样一款优秀的工具

## valgrind内存检测
```
#include <iostream>
using namespace std;


int main(int argc, char const *argv[])
{
    int a[5];
    a[0] = a[1] = a[3] = a[4] = 0;

    int s=0;
    for(int i=0;i<5;i++){
        s+=a[i];
    }
    if(s == 0){
        std::cout << s << std::endl;
    }
    a[5] = 10;
    std::cout << a[5] << std::endl;


    int *invalid_write = new int[10];
    delete [] invalid_write;
    invalid_write[0] = 3;

    int *undelete = new int[10];
    
    return 0;
}
```
```
==102507== Memcheck, a memory error detector
==102507== Copyright (C) 2002-2017, and GNU GPL'd, by Julian Seward et al.
==102507== Using Valgrind-3.14.0 and LibVEX; rerun with -h for copyright info
==102507== Command: ./a.out
==102507== 
==102507== Conditional jump or move depends on uninitialised value(s)
==102507==    at 0x1091F6: main (learn_valgrind.cpp:14)
==102507== 
10
==102507== Invalid write of size 4
==102507==    at 0x109270: main (learn_valgrind.cpp:23)
==102507==  Address 0x4dc30c0 is 0 bytes inside a block of size 40 free'd
==102507==    at 0x483A55B: operator delete[](void*) (in /usr/lib/x86_64-linux-gnu/valgrind/vgpreload_memcheck-amd64-linux.so)
==102507==    by 0x10926B: main (learn_valgrind.cpp:22)
==102507==  Block was alloc'd at
==102507==    at 0x48394DF: operator new[](unsigned long) (in /usr/lib/x86_64-linux-gnu/valgrind/vgpreload_memcheck-amd64-linux.so)
==102507==    by 0x109254: main (learn_valgrind.cpp:21)
==102507== 
==102507== 
==102507== HEAP SUMMARY:
==102507==     in use at exit: 40 bytes in 1 blocks
==102507==   total heap usage: 4 allocs, 3 frees, 73,808 bytes allocated
==102507== 
==102507== LEAK SUMMARY:
==102507==    definitely lost: 40 bytes in 1 blocks
==102507==    indirectly lost: 0 bytes in 0 blocks
==102507==      possibly lost: 0 bytes in 0 blocks
==102507==    still reachable: 0 bytes in 0 blocks
==102507==         suppressed: 0 bytes in 0 blocks
==102507== Rerun with --leak-check=full to see details of leaked memory
==102507== 
==102507== For counts of detected and suppressed errors, rerun with: -v
==102507== Use --track-origins=yes to see where uninitialised values come from
==102507== ERROR SUMMARY: 2 errors from 2 contexts (suppressed: 0 from 0)

```



1. https://www.ibm.com/developerworks/cn/linux/l-cn-valgrind/index.html
2. http://senlinzhan.github.io/2017/12/31/valgrind/
3. https://www.ibm.com/developerworks/cn/aix/library/au-memorytechniques.html