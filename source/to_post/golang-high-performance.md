---
title: golang 程序性能分析与优化
categories:
- Golang
---


最近参加了dave关于高性能golang的论坛，它通过几个case非常清晰的介绍了golang性能分析与优化的技术，非常值得学习。https://dave.cheney.net/high-performance-go-workshop/dotgo-paris.html。随着计算机技术的发展，硬件资源越来越受限制，我们应该关注程序的性能。
## 一、性能测量
1. time 
　　在Linux中，time命令经常用于统计程序的耗时(real)、用户态cpu耗时(user)、系统态cpu耗时(sys)。在操作系统课程中我们知道程序的运行包括用户态和系统态。对于单核程序，程序有时处于等待状态，因此总是real>user+sys，将(user+sys)/real称为cpu利用率。对于多核程序来说，由于能把多个cpu都利用起来，上面的关系就不成立。

2. benchmarking
　　有时我们有测试某些函数性能的需求，go testing包内置了非常好用的benchmarks。例如有一个产生斐波那契数列的函数，我们可以用testing包测试出它的benchmark。
fib.go:
```
	package benchmarkFib

	func Fib(n int) int {
		switch n {
		case 0:
			return 0
		case 1:
			return 1
		case 2:
			return 2
		default:
			return Fib(n-1) + Fib(n-2)
		}
	}
```
fib_test.go
```
	package benchmarkFib

	import (
		"testing"
	)

	func BenchmarkFib20(b *testing.B) {
		for n := 0; n < b.N; n++ {
			Fib(20)
		}
	}
```
$go test -bench=.
	goos: linux
	goarch: amd64
	pkg: learn_golang/benchmarkFib
	BenchmarkFib20-4   	  100000	     22912 ns/op
	PASS
	ok  	learn_golang/benchmarkFib	2.526s


3. profile
　　benchmark能帮助分析某些函数的性能，但是对于分析整个程序来说还是需要使用profile。golang使用profile是非常方便的，因为很早期的时候就集成到runtime中,它包括两个部分：
- runtime/pprof
- go tool pprof cpu.pprof 分析profile数据
   pprof包括四种类型的profile，其中最常用的是cpu profile和memory profile。
- CPU profile：最常用，运行时每10ms中断并记录当前运行的goroutine的堆栈跟踪，通过cpu profile可以看出函数调用的次数和所占用时间的百分比。
- Memory profile：采样的是分配的堆内存而不是使用的内存
- Block profile
- Mutex contention profile

为了更方便的产生profile文件，dave封装了runtime/pprof。https://github.com/pkg/profile.git
结合dave的例子分析cpu profile：https://github.com/wxquare/learn_golang/tree/master/pprof
% go run main.go moby.txt
2019/05/06 21:26:56 profile: cpu profiling enabled, cpu.pprof
"moby.txt": 181275 words
2019/05/06 21:26:57 profile: cpu profiling disabled, cpu.pprof

a、使用命令分析profile：
% go tool pprof
% top 

b、借助浏览器分析profile： go tool pprof -http=:8080
- 图模式（Graph mode)
- 火焰图模式(Flame Graph mode)
 

## 二、Execution Tracer
   profile是基于采样(sample)的，而Execution Tracer是集成到Go运行时(runtime)中，因此它能知道程序在某个时间点的具体行为。Dave用了一个例子来说明为什么需要tracer，而 go tool pprof执行的效果很差。

1. v1 time ./mandelbrot (原版)
    real    0m1.654s
	user    0m1.630s
	sys     0m0.015s

2. 跑出profile、分析profile
	cd examples/mandelbrot-runtime-pprof
	go run mandelbrot.go > cpu.pprof
    go tool pprof -http=:8080 cpu.pprof

3. 通过profile数据，可以知道fillpixel几乎做了程序所有的工作，但是我们似乎也没有什么可以优化的了？？？这个时候可以考虑引入Execution tracer。运行程序跑出trace数据。
	import "github.com/pkg/profile"

	func main() {
		defer profile.Start(profile.TraceProfile, profile.ProfilePath(".")).Stop()

	然后使用go tool trace trace.out 分析trace数据。

4. 分析trace数据，记住要使用chrome浏览器。
  通过trace数据可以看出只有一个Goroutine在工作，没有利用好机器的资源。

5. 之后的几个优化通过调整使用的gorutine的数量使得程序充分利用CPU计算资源，提高程序的效率。


## 三、编译器优化
1. 逃逸分析（Escape analysis）
	golang在内存分配的时候没有堆(heap)和栈(stack)的区别，由编译器决定是否需要将对象逃逸到堆中。例如：
```
		func Sum() int {
		const count = 100
		numbers := make([]int, count)
		for i := range numbers {
			numbers[i] = i + 1
		}

		var sum int
		for _, i := range numbers {
			sum += i
		}
		return sum
	}

	func main() {
		answer := Sum()
		fmt.Println(answer)
	}
```
$ go build -gcflags=-m test_esc.go 
# command-line-arguments
./test_esc.go:9:17: Sum make([]int, count) does not escape
./test_esc.go:23:13: answer escapes to heap
./test_esc.go:23:13: main ... argument does not escape

2. 内敛（Inlining）
   了解C/C++的应该知道内敛，golang编译器同样支持函数内敛，对于较短且重复调用的函数可以考虑使用内敛

3. Dead code elimination/Branch elimination
	编译器会将代码中一些无用的分支进行优化，分支判断，提高效率。例如下面一段代码由于a和b是常量，编译器也可以推导出Max(a,b)，因此最终F函数为空
```	
	func Max(a, b int) int {
		if a > b {
			return a
		}
		return b
	}

	func F() {
		const a, b = 100, 20
		if Max(a, b) == b {
			panic(b)
		}
	}
```
常用的编译器选项： go build -gcflags="-lN" xxx.go
- "-S",编译时查看汇编代码
- "-l",关闭内敛优化
- "-m",打印编译优化的细节
- "-l -N",关闭所有的优化



## 四、内存和垃圾回收
golang支持垃圾回收，gc能减少编程的负担，但与此同时也可能造成程序的性能问题。那么如何测量golang程序使用的内存，以及如何减少golang gc的负担呢？经历了许多版本的迭代，golang gc 沿着低延迟和高吞吐的目标在进化，相比早起版本，目前有了很大的改善，但仍然有可能是程序的瓶颈。因此要学会分析golang 程序的内存和垃圾回收问题。

如何查看程序的gc信息？
1. 通过设置环境变量？env GODEBUG=gctrace=1
例如： env GODEBUG=gctrace=1 godoc -http=:8080
2. import _ "net/http/pprof"，查看/debug/pprof

tips：
1. 减少内存分配，优先使用第二种APIs
	func (r *Reader) Read() ([]byte, error)
	func (r *Reader) Read(buf []byte) (int, error)
2. 尽量避免string 和 []byte之间的转换
3. 尽量减少两个字符串的合并
4. 对slice预先分配大小
5. 尽量不要使用cgo，因为c和go毕竟是两种语言。cgo是个high overhead的操作，调用cgo相当于阻塞IO，消耗一个线程
6. defer is expensive？在性能要求较高的时候，考虑少用
7. 对IO操作设置超时机制是个好习惯SetDeadline, SetReadDeadline, SetWriteDeadline
8. 当数据量很大的时候，考虑使用流式IO(streaming IO)。io.ReaderFrom / io.WriterTo
