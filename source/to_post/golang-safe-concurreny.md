

golang 安全并发
	当一个函数或者类型既能在单一执行体（线程、协程）下正常运行时，也能在多执行体下正常运行是，我们称之为并发安全。当一个程序中有非并发安全的类型下，我们可以借助外在手段保证程序的并发安全。

# 一、竞争条件
　　竞争条件指的是程序在多个goroutine交叉执行操作时，没有给出正确的结果。竞争条件是很恶劣的一种场景，因为这种问题会一直潜伏在你的程序里，然后在非常少见的时候蹦出来，或许只是会在很大的负载时才会发生，又或许是会在使用了某一个编译器、某一种平台或者某一种架构的时候才会出现。这些使得竞争条件带来的问题非常难以复现而且难以分析诊断。
　　这个程序包含了一个特定的竞争条件，叫作数据竞争。无论任何时候，只要有两个goroutine并发访问同一变量，且至少其中的一个是写操作的时候就会发生数据竞争。我们来重复一下数据竞争的定义，因为实在太重要了：数据竞争会在两个以上的goroutine并发访问相同的变量且至少其中一个为写操作时发生。根据上述定义，有三种方式可以避免数据竞争：
```
package main

import (
	"fmt"
	"time"
)

var balance int

func Deposit(amount int) { balance = balance + amount }
func Balance() int       { return balance }

func main() {
	go func() {
		Deposit(200)                // A1
		fmt.Println("=", Balance()) // A2
	}()

	go Deposit(100) // B

	time.Sleep(10 * time.Second)
}
```

1. 第一种避免数据竞争的方法是，避免从多个goroutine访问变量goroutine是唯一一个能够访问clients map的goroutine。这些变量都被限定在了一个单独的goroutine中。由于其它的goroutine不能够直接访问变量，它们只能使用一个channel来发送请求给指定的goroutine来查询更新变量。这也就是Go的口头禅“不要使用共享数据来通信；使用通信来共享数据”。一个提供对一个指定的变量通过channel来请求的goroutine叫做这个变量的monitor（监控）goroutine。
```
package main

import (
	"fmt"
)

var deposits = make(chan int) // send amount to deposit
var balances = make(chan int) // receive balance

func Deposit(amount int) { deposits <- amount }
func Balance() int       { return <-balances }

func teller() {
	var balance int // balance is confined to teller goroutine
	for {
		select {
		case amount := <-deposits:
			balance += amount
		case balances <- balance:
		}
	}
}

func init() {
	go teller() // start the monitor goroutine
}

func main() {
	Deposit(100)
	Deposit(200)
	fmt.Println(Balance())
}

```

第二种避免数据竞争的方法是允许很多goroutine去访问变量，但是在同一个时刻最多只有一个goroutine在访问。这种方式被称为“互斥”。　
1. 二元信号量sema
2. sync.mutex
```
package main

import (
	"sync"
)

func Withdraw(amount int) bool {
	mu.Lock()
	defer mu.Unlock()
	deposit(-amount)
	if balance < 0 {
		deposit(amount)
		return false // insufficient funds
	}
	return true
}

func Deposit(amount int) {
	mu.Lock()
	defer mu.Unlock()
	deposit(amount)
}

func Balance() int {
	mu.Lock()
	defer mu.Unlock()
	return balance
}

// This function requires that the lock be held.
func deposit(amount int) { balance += amount }
```

由于Balance函数只需要读取变量的状态，所以我们同时让多个Balance调用并发运行事实上是安全的，只要在运行的时候没有存款或者取款操作就行。在这种场景下我们需要一种特殊类型的锁，其允许多个只读操作并行执行，但写操作会完全互斥。这种锁叫作“多读单写”锁(multiple readers, single writer lock)，Go语言提供的这样的锁是sync.RWMutex.
RWMutex只有当获得锁的大部分goroutine都是读操作，而锁在竞争条件下，也就是说，goroutine们必须等待才能获取到锁的时候，RWMutex才是最能带来好处的。RWMutex需要更复杂的内部记录，所以会让它比一般的无竞争锁的mutex慢一些。



### 三、内存同步
在现代计算机中可能会有一堆处理器，每一个都会有其本地缓存(local cache)。为了效率，对内存的写入一般会在每一个处理器中缓冲，并在必要时一起flush到主存。这种情况下这些数据可能会以与当初goroutine写入顺序不同的顺序被提交到主存。像channel通信或者互斥量操作这样的原语会使处理器将其聚集的写入flush并commit，这样goroutine在某个时间点上的执行结果才能被其它处理器上运行的goroutine得到。
在一个独立的goroutine中，每一个语句的执行顺序是可以被保证的，也就是说goroutine内顺序是连贯的。但是在不使用channel且不使用mutex这样的显式同步操作时，我们就没法保证事件在不同的goroutine中看到的执行顺序是一致的了。尽管goroutine A中一定需要观察到x=1执行成功之后才会去读取y，但它没法确保自己观察得到goroutine B中对y的写入，所以A还可能会打印出y的一个旧版的值。



### 四、一次性初始化
sync包为我们提供了一个专门的方案来解决这种一次性初始化的问题：sync.Once。概念上来讲，一次性的初始化需要一个互斥量mutex和一个boolean变量来记录初始化是不是已经完成了；互斥量用来保护boolean变量和客户端数据结构。Do这个唯一的方法需要接收初始化函数作为其参数。让我们用sync.Once来简化前面的Icon函数吧：
```
var loadIconsOnce sync.Once
var icons map[string]image.Image
// Concurrency-safe.
func Icon(name string) image.Image {
    loadIconsOnce.Do(loadIcons)
    return icons[name]
}
```
每一次对Do(loadIcons)的调用都会锁定mutex，并会检查boolean变量(译注：Go1.9中会先判断boolean变量是否为1(true)，只有不为1才锁定mutex，不再需要每次都锁定mutex)。在第一次调用时，boolean变量的值是false，Do会调用loadIcons并会将boolean变量设置为true。随后的调用什么都不会做，但是mutex同步会保证loadIcons对内存(这里其实就是指icons变量啦)产生的效果能够对所有goroutine可见。用这种方式来使用sync.Once的话，我们能够避免在变量被构建完成之前和其它goroutine共享该变量。


### 五、竞争条件检测
即使我们小心到不能再小心，但在并发程序中犯错还是太容易了。幸运的是，Go的runtime和工具链为我们装备了一个复杂但好用的动态分析工具，竞争检查器(the race detector)。
只要在go build，go run或者go test命令后面加上-race的flag，就会使编译器创建一个你的应用的“修改”版或者一个附带了能够记录所有运行期对共享变量访问工具的test，并且会记录下每一个读或者写共享变量的goroutine的身份信息。竞争检查器会报告所有的已经发生的数据竞争。然而，它只能检测到运行时的竞争条件；并不能证明之后不会发生数据竞争。所以为了使结果尽量正确，请保证你的测试并发地覆盖到了你的包。由于需要额外的记录，因此构建时加了竞争检测的程序跑起来会慢一些，且需要更大的内存，即使是这样，这些代价对于很多生产环境的工作来说还是可以接受的。对于一些偶发的竞争条件来说，让竞争检查器来干活可以节省无数日夜的debugging。



### 六、协程与线程的区别
1. 协程是动态栈，大小可从2KB到2GB，因此协程数量可达数千上万个
2. 协程的调度有golang运行时runtime自行调度
3. 通过GOMAXPROCS决定多少个操作系统系统同时执行Go的代码
4. Goroutine没有ID号



1. 函数的并发安全性
2. 类型的并发安全性
