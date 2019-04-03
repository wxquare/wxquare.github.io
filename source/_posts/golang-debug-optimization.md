---
title: golang 程序测试和优化
---

　　Golang非常注重工程化，提供了非常好用单元测试、性能测试（benchmark）和调优工具（pprof），它们对提高代码的质量和服务的性能非常有帮助。[参考链接](https://tonybai.com/2015/08/25/go-debugging-profiling-optimization)中通过一段http代码非常详细的介绍了golang程序优化的步骤和方便之处。实际工作中，我们很难每次都对代码都有那么高的要求，但是能使用一些工具对程序进行优化程序性能也是golang程序员必备的技能。
- testing 标准库 
- go test 测试工具
- go tool pprof 分析profile数据



## 一、单元测试，测试正确性
1. 为了测试某个文件中的某个函数的性能，在相同目录下定义xxx_test.go文件，使用go build命令编译程序时会忽略测试文件
2. 在测试文件中定义测试某函数的代码，以TestXxxx方式命名，例如TestAdd
3. 在相同目录下运行 go test -v 即可观察代码的测试结果

    	func TestAdd(t *testing.T) {
    		if add(1, 3) != 4 {
    			t.FailNow()
    		}
    	}

	

## 二、性能测试，benchmark
1. 单元测试，测试程序的正确性。benchmark 用户测试代码的效率，执行的时间
2. benchmark测试以BenchMark开头，例如BenchmarkAdd
3. 运行 go test -v -bench=. 程序会运行到一定的测试，直到有比较准备的测试结果

    	func BenchmarkAdd(b *testing.B) {
    		for i := 0; i < b.N; i++ {
    			_ = add(1, 2)
    		}
    	}
    
    	BenchmarkAdd-4  	2000000000	 0.26 ns/op

## 三、pprof性能分析

1. 除了使用使用testing进行单元测试和benchanmark性能测试，golang能非常方便捕获或者监控程序运行状态数据，它包括cpu、内存、和阻塞等，并且非常的直观和易于分析。
2. 有两种捕获方式： a、在测试时输出并保存相关数据；b、在运行阶段，在线采集，通过web接口获得实时数据。
3. Benchamark时输出profile数据：go test -v -bench=. -memprofile=mem.out -cpuprofile=cpu.out
4. 使用go tool pprof xxx.test mem.out 进行交互式查看，例如top5。同理，可以分析其它profile文件。  

(pprof) top5
Showing nodes accounting for 1994.93MB, 63.62% of 3135.71MB total
Dropped 28 nodes (cum <= 15.68MB)
Showing top 5 nodes out of 46
      flat  flat%   sum%        cum   cum%
  475.10MB 15.15% 15.15%   475.10MB 15.15%  regexp/syntax.(*compiler).inst
  455.58MB 14.53% 29.68%   455.58MB 14.53%  regexp.progMachine
  421.55MB 13.44% 43.12%   421.55MB 13.44%  regexp/syntax.(*parser).newRegexp
  328.61MB 10.48% 53.60%   328.61MB 10.48%  regexp.onePassCopy
  314.09MB 10.02% 63.62%   314.09MB 10.02%  net/http/httptest.cloneHeader

- flat：仅当前函数，不包括它调用的其它函数
- cum： 当前函数调用堆栈的累计
- sum： 列表前几行所占百分比的总和

更加详细的golang程序调试和优化请参考：
[1]. https://tonybai.com/2015/08/25/go-debugging-profiling-optimization/
[2]. https://blog.golang.org/profiling-go-programs