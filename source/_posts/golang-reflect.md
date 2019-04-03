---
title: golang 反射
categories:
- Golang
---


　　像Java、Python一样，Golang也提供反射机制，简单的来说，反射能在运行期获取接口对象的类型、数据和方法。golang的反射机制依赖于接口，因为接口对象保存了自身类型和实际对象的对象的类型和数据。要想用好反射需要理解：**实际对象、接口对象、反射类型Type和反射值Value类型**。Golang使用反射非常方便，只需导入reflect包，它提供两个入口函数，将任何传入的对象转换为接口类型，从而获取反射类型（Type）和反射值（Value）：

	func TypeOf( i interface{}) Type
	func ValueOf(i interface{}) Value

通过反射，我们可以完成：  
1. 获取对象的类型
2. 获取对象成员的类型，包括可导出成员和不可导出成员
3. 获取对象成员方法，只可以获取可导出成员方法，获取成员方法时，区分基类型和指针类型
4. 获取struct tag
5. 将对象赋值给接口对象是复制行为，而且是unaddressable，因此要修改目标值，必须获取对象指针的反射值
6. 接口对象存储的指针本身是不可以寻址和设置的，需要使用Elem获取目标对象
7. 反射不能对非导出字段进行设置操作，只能改变导出字段的值
8. 可通过Interface方法进行类型断言和转换
9. 反射可动态成员方法 
10. 相比正常的修改变量值和调用方法，反射存在效率低的问题

```

    package main
    
    import (
    	"fmt"
    	"reflect"
    	"unsafe"
    )
    
    type Person struct {
    	name string
    	Age  int `json:"age"`
    	Done chan int
    }
    
    func (p Person) setName(newName string) string {
    	p.name = newName
    	return p.name
    }
    
    func (p *Person) SetAge(newAge int) {
    	p.Age = newAge
    }
    
    func (p Person) PrintNameAge() {
    	fmt.Println(p.name, p.Age)
    }
    
    func main() {
    
    	// struct type
    	p1 := Person{"terse", 12, make(chan int)}
    
    	//1.获取类型信息
    	tp1 := reflect.TypeOf(p1)
    	fmt.Println(tp1.Kind(), tp1.Name())
    
    	//2.获取成员类型信息
    	//能获取可导出和不可导出成员
    	for i := 0; i < tp1.NumField(); i++ {
    		f := tp1.Field(i)
    		fmt.Println(f.Name)
    	}
    
    	//3.获取成员方法信息
    	//只能获取可导出的方法
    	//输出方法集时区分基类型和指针类型
    	for i := 0; i < tp1.NumMethod(); i++ {
    		fmt.Println(tp1.Method(i))
    	}
    	tp2 := reflect.TypeOf(&p1)
    	for i := 0; i < tp2.NumMethod(); i++ {
    		fmt.Println(tp2.Method(i))
    	}
    
    	//4.获取struct tag
    	fmt.Println(tp1.Field(1).Tag.Get("json"))
    
    	//5.接口变量会复制对象，且是unaddressable，所以要修改目标兑现必须使用指针
    	//6.接口存储的指针本身是不可以寻址和进行设置操作，需要使用Elem获取目标对象
    	vp1 := reflect.ValueOf(p1)
    	vp2 := reflect.ValueOf(&p1).Elem()
    	fmt.Println(vp1.CanAddr(), vp2.CanAddr())
    
    	//7.不能对非导出字段进行设置操作
    	name := vp2.FieldByName("name")
    	Age := vp2.FieldByName("Age")
    	fmt.Println(name.CanSet(), name.CanAddr())
    	fmt.Println(Age.CanSet(), Age.CanAddr())
    
    	//8.对可导出字段进行设置操作
    	*(*int)(unsafe.Pointer(Age.UnsafeAddr())) = 100
    	fmt.Println(p1)
    
    	fmt.Println(vp2.CanInterface())
    	//9.Interface方法进行类型断言和转换
    	p, ok := vp2.Interface().(Person)
    	if ok {
    		fmt.Println("vp2 is *Person", reflect.TypeOf(p).Name())
    	}
    
    	//10.动态调用方法
    	vp3 := reflect.ValueOf(&p1)
    	m := vp3.MethodByName("SetAge")
    	in := []reflect.Value{
    		reflect.ValueOf(10),
    	}
    	out := m.Call(in)
    	fmt.Println(out, p1)
    }
```