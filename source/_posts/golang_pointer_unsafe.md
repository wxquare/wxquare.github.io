---
title: golang 指针和unsafe
categories:
- Golang
---

## 一、golang指针和unsafe.pointer
1. 不同类型的指针不能相互转化  
2. 指针变量不能进行运算，不支持c/c++中的++，--运算  
3. 任何类型的指针都可以被转换成unsafe.Pointer类型，反之也是  
4. uintptr值可以被转换成unsafe.Pointer类型，反之也是
5. 对unsafe.Pointer和uintptr两种类型单独解释两句：  
	- unsafe.Pointer是一个指针类型，指向的值不能被解析，类似于C/C++里面的(void *)，只说明这是一个指针，但是指向什么的不知道。
	- uintptr 是一个整数类型，这个整数的宽度足以用来存储一个指针类型数据；那既然是整数类类型，当然就可以对其进行运算了
```      
    package main
    import (
    	"fmt"
    	"unsafe"
    )
    func main() {
    	var ii [4]int = [4]int{11, 22, 33, 44}
    	px := &ii[0]
    	fmt.Println(&ii[0], px, *px)
    	//compile error
    	//pf32 := (*float32)(px)
    
    	//compile error
    	// px = px + 8
    	// px++
    
    	var pointer1 unsafe.Pointer = unsafe.Pointer(px)
    	var pf32 *float32 = (*float32)(pointer1)
    
    	var p2 uintptr = uintptr(pointer1)
    	print(p2)
    	p2 = p2 + 8
    	var pointer2 unsafe.Pointer = unsafe.Pointer(p2)
    	var pi32 *int = (*int)(pointer2)
    
    	fmt.Println(*px, *pf32, *pi32)
    
    }
```
## 二、 nil指针
引用类型声明而没有初始化赋值时，其值为nil。golang需要经常判断nil,防止出现panic错误。  
```
    bool  -> false  
    numbers -> 0 
    string-> ""  
    
    pointers -> nil
    slices -> nil
    maps -> nil
    channels -> nil
    functions -> nil
    interfaces -> nil



    package main
    
    import (
    	"fmt"
    )
    
    type Person struct {
    	AgeYears int
    	Name string
    	Friends  []Person
    }
    
    func main() {
    	var p Person
    	fmt.Printf("%v\n", p)
    
    	var slice1 []int
    	fmt.Println(slice1)
    	if slice1 == nil {
    		fmt.Println("slice1 is nil")
    	}
    	// fmt.Println(slice1[0])  panic
    
    	// var c chan int
    	// close(c)  panic
    }
```
参考：  

- https://studygolang.com/articles/10953  
- https://www.jianshu.com/p/dd80f6be7969  