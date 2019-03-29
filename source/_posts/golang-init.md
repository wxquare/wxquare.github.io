---
title: golang程序启动与init函数
---

　　在golang中，可执行文件的入口函数并不是我们写的main函数，编译器在编译go代码时会插入一段起引导作用的汇编代码，它引导程序进行命令行参数、运行时的初始化，例如内存分配器初始化、垃圾回收器初始化、协程调度器的初始化。golang引导初始化之后就会进入用户逻辑，因为存在特殊的init函数，main函数也不是程序最开始执行的函数。

## 一、golang程序启动流程
　　golang可执行程序由于运行时runtime的存在，其启动过程还是非常复杂的，这里通过gdb调试工具简单查看其启动流程：  
1. 找一个golang编译的可执行程序test，info file查看其入口地址：gdb test，info files
(gdb) info files
Symbols from "/home/terse/code/go/src/learn_golang/test_init/main".
Local exec file:
	/home/terse/code/go/src/learn_golang/test_init/main', 
    file type elf64-x86-64.
	**Entry point: 0x452110**
	.....
2. 利用断点信息找到目标文件信息：
(gdb) b *0x452110
Breakpoint 1 at 0x452110: file /usr/local/go/src/runtime/rt0_linux_amd64.s, line 8.
3. 依次找到对应的文件对应的行数，设置断点，调到指定的行，查看具体的内容：
(gdb) b _rt0_amd64  
(gdb) b b runtime.rt0_go  
至此，由汇编代码针对特定平台实现的引导过程就全部完成了，后续的代码都是用Go实现的。分别实现命令行参数初始化，内存分配器初始化、垃圾回收器初始化、协程调度器的初始化等功能。
```
	CALL	runtime·args(SB)
	CALL	runtime·osinit(SB)
	CALL	runtime·schedinit(SB)

	CALL	runtime·newproc(SB)

	CALL	runtime·mstart(SB)
```

## 二、特殊的init函数
1. init函数先于main函数自动执行，不能被其他函数调用
2. init函数没有输入参数、没有返回值
3. 每个包可以含有多个同名的init函数，每个源文件也可以有多个同名的init函数
4. **执行顺序** 变量初始化 > init函数 > main函数。在复杂项目中，初始化的顺序如下：
	- 先初始化import包的变量，然后先初始化import的包中的init函数，，再初始化main包变量，最后执行main包的init函数
	- 从上到下初始化导入的包（执行init函数），遇到依赖关系，先初始化没有依赖的包
	- 从上到下初始化导入包中的变量，遇到依赖，先执行没有依赖的变量初始化
	- main包本身变量的初始化，main包本身的init函数
	- 同一个包中不同源文件的初始化是按照源文件名称的字典序

util.go
```
package util

import (
	"fmt"
)

var c int = func() int {
	fmt.Println("util variable init")
	return 3
}()

func init() {
	fmt.Println("call util.init")
}
```

main.go
```
package main

import (
	"fmt"
	_ "util"
)

var a int = func() int {
	fmt.Println("main variable init")
	return 3
}()

func init() {
	fmt.Println("call main.init")
}

func main() {
	fmt.Println("call main.main")
}
```
执行结果：  
　　　util variable init
　　　call util.init
　　　main variable init
　　　call main.init
　　　call main.main


参考：《Go语言学习笔记13、14、15章》
