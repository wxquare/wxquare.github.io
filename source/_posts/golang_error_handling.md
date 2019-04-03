---
title: golang错误处理
categories:
- Golang
---


错误处理是任何编程语言都不可避免的话题，golang错误处理的方式虽然备受争议，但总体是符合工程语言的要求的。熟悉golang错误处理的方式，需要掌握以下五点
## 1.根据error接口自定义错误类型
golang中引入了关于错误处理的标准模式error接口，实际中可以通过实现error结构，自定义错误类型。error接口只有一个Error方法，它返回一个string表示错误的内容。  
error接口： 
 
    type error Interface{  
    	Error() string  
    }  
自定义错误类型：
  
    type MyError struct {
    	ErrorInfo string
    }
    
    func (e *MyError) Error() string {
    	return ErrorInfo
    }
    
## 2.通过errors包生成error对象
errors包提供New方法，非常方便生成error对象，例如：

    func foo() error {
    	return errors.New("foo error")
    }

## 3.panic和recover  
- 当一个函数抛出panic错误时，正常的函数流程立即终止
- defer关键字延迟执行的语句将正常执行
- 逐层向上执行panic过程，直到所属的goroutine中所有执行的函数终止
- recover用于终止panic的错误处理流程   

例如：   

    func main() {
    	//defer
    	defer func() {
    		fmt.Println("defer func(){}()")
    		if r := recover(); r != nil {
    			fmt.Println("Runtime error caught!", r)
    		}
    	}()
    	panic("throw a panic")
    	fmt.Println("hello,world")
    }
## 4.defer
defer是golang中非常好用的一个错误处理方式，函数正常退出和出错时，defer中的语句也会被执行，作用相当于C++中的析构函数，对资源泄露非常有帮助。实际使用时需要注意：  
- defer语句的位置  
- defer语句执行的顺序  
defer语句的调用遵循的顺序是先进后出，即最后一个defer语句最先被执行。


## 5.接口对象类型断言
golang中接口对象非常方便，因此提供类型判断，防止出现panic错误。例如：

    type Person struct {
    	Name string
    	age  int
    }
    
    func main() {
    	//Type Assertion
    	var v interface{}
    	v = Person{"bob", 12}
    	if f, ok := v.(Person); ok {
    		fmt.Println(f.Name)
    	}
    }

