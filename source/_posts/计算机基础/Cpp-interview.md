---
title: 一文记录 C/C++基础知识和编码规范
date: 2023-10-13
categories: 
- 计算机基础
---

[了解google C++编码规范](https://zh-google-styleguide.readthedocs.io/en/latest/)

## C/C++基础
### 1. 关键字的作用const用途
	- 定义常量、指针、引用、对象，const进行修饰的变量的值在程序的任意位置将不能再被修改，就如同常数一样使用
	- 修饰参数const int x；
	- 修饰成员变量；const成员变量必须通过初始化列表进行初始化。
	- 修饰成员函数；
	- C++ mutable变量突破const修饰成员函数的限制
	
### 2. 关键字static的用途
	- 静态全局变量
	- 修饰函数内部的局部变量，作用相当于全局变量
	- 类的静态成员变量
	- 类的静态成员函数
	参考：https://www.cnblogs.com/wxquare/p/6692924.html

### 3. 宏定义和内敛函数的区别
	- 内联函数和宏定义减少函数调用所带来的时间和空间的开销，以空间换时间的策略
	- 宏是在预编译阶段简单文本替代，inline在编译阶段实现展开，宏定义是预编译期、内敛是编译器优化
	- 宏肯定会被替代，而复杂的inline函数不会被展开
	- 宏容易出错（运算顺序），且难以被调试,inline不会
	- 宏不是类型安全，而inline是类型安全的，会提供参数与返回值的类型检查
	- 当函数size太大,inline虚函数,函数中存在循环或递归，内敛可能失效
	- 当函数被声明为内联函数之后, 编译器会将其内联展开, 而不是按通常的函数调用机制进行调用
	- 使用宏时要非常谨慎, 尽量以内联函数, 枚举和常量代替之.
	参考：https://www.cnblogs.com/wxquare/p/6800488.html

### 4. extern 与 static
　　extern是C/C++语言中表明函数和全局变量作用范围（可见性）的关键字，该关键字告诉编译器，其声明的函数和变量可以在本模块或其它模块中使用。通常，在模块的头文件中对本模块提供给其它模块引用的函数和全局变量以关键字extern声明。例如，如果模块B欲引用该模块A中定义的全局变量和函数时只需包含模块A的头文件即可。这样，模块B中调用模块A中的函数时，在编译阶段，模块B虽然找不到该函数，但是并不会报错；它会在连接阶段中从模块A编译生成的目标代码中找到此函数。与extern对应的关键字是static，被它修饰的全局变量和函数只能在本模块使用。

### 5. extern "C"
　　extern "C"是为了实现C和C++的混合编程，而C和C++的编译和链接是不完全相同的，extern "C"表明它按照类C的编译和连接规约来编译和连接，而不是C++的编译和链接。C++是一个面向对象语言，它为了支持函数的重载，在编译的时候会带上参数的类型来唯一标识每个函数，而C语言并不需要这么做。C语言中并没有重载和类这些特性，故并不像C++那样print(int i)，会被编译为_print_int，而是直接编译为_print等。因此如果直接在C++中调用C的函数会失败，因为连接是调用C中的print(3)时，它会去找_print_int(3)。因此extern"C"的作用就体现出来了。假设一个C函数print(int i)，为了在C++中能够调用它，必须要加上extern关键字。
　　参考：https://www.cnblogs.com/skynet/archive/2010/07/10/1774964.html

### 6. struct和class的区别
　　在C++中struct和class中struct和class的区别比较小，主要**类成员的默认访问权限**和**继承权限**。
默认继承权限：如果不明确指定，来自class的继承按照private继承处理，来自struct的继承按照public继承处理。成员的默认访问权限：class的成员默认是private权限，struct默认是public权限。
仅当只有数据成员时使用 struct, 其它一概使用 class.（google 编码规范）


### 7. 什么是类型安全、内存安全，C/C++不是类型安全
静态类型 vs 动态类型: 静态类型（C/C++，java，golang），动态类型（python）
弱类型 vs 强类型：弱类型（C、C++），强类型（python、golang）
类型安全： 一般来说弱类型存在隐含的类型转换都不是类型安全的，而强类型是类型安全的
https://www.zhihu.com/question/19918532/answer/21647195

### 8. C++ 中的四种类型转换
C风格的强制类型转换(Type Cast)很简单，不管什么类型的转换统统是：TYPE b = (TYPE)a。
C++风格的类型转换提供了4种类型转换操作符来应对不同场合的应用。
const_cast：字面上理解就是去const属性。
**static_cast**：命名上理解是静态类型转换。如int转换成char。类似于C风格的强制转换。无条件转换，静态类型转换。基本类型转换用static_cast。
**dynamic_cast**：命名上理解是动态类型转换。如子类和父类之间的多态类型转换。有条件转换，动态类型转换，运行时类型安全检查(转换失败返回NULL)。多态类之间的类型转换用daynamic_cast。
reinterpret_cast：仅仅重新解释类型，但没有进行二进制的转换。
4种类型转换的格式，如：TYPE B = static_cast<TYPE>(a)
参考：https://www.cnblogs.com/goodhacker/archive/2011/07/20/2111996.html

### 9. 关键字volatile的作用
　　volatile int i = 10
　　volatile 指出 i 是随时可能发生变化的，每次使用它的时候必须从 i的地址中读取，因而编译器生成的汇编代码会重新从i的地址读取数据放在 b 中。而优化做法是，由于编译器发现两次从 i读数据的代码之间的代码没有对 i 进行过操作，它会自动把上次读的数据放在 b 中。而不是重新从 i 里面读。这样以来，如果 i是一个寄存器变量或者表示一个端口数据就容易出错，所以说volatile直接存取原始内存地址，禁止执行期寄存器的优化。

### 10. 指针和引用
　　指针指向一块内存，它的内容是所指内存的地址；而引用则是某块内存的别名，引用初始化后不能改变指向。使用时，引用更加安全，指针更加灵活。
- 初始化。引用必须初始化，且初始化之后不能呢改变；指针可以不必初始化，且指针可以改变所指的对象
- 空值。指针可以指向空值，不存在指向空值的引用。当引用或者指针作为参数传递的时候，拿到一个引用的时候，是不需要判断引用是否为空的，而拿到一个指针的时候，我们则需要判断它是否为空。这点经常在判断函数参数是否有效的时候使用。
- 引用和指针指向一个对象时，引用的创建和销毁不会调用类的拷贝构造函数和析构函数。delete一个指针会调用该对象的析构函数，注意防止二次析构。
- 引用和指针与const。存在常量指针和常量引用指针，表示指向的对象是常量，不能通过指针或者常量修改常量；存在指针常量，不存在引用常量，因为引用本身不能修改指向的特性和与指针常量的特性相同，不需要引用常量。
- 函数参数传递时使用指针或者引用的效果是相同的，都是简洁操作主调函数中的相关变量，当时引用会更加的安全，因为指针一些修改指向，将不能影响主调函数中的相关变量。所以参数传递时尽可能使用引用。
- sizeof引用的时候是对象的大小，sizeof指针是指针本身的大小
- 引用和指针的实现是相同的，“引用只是一个别名，不会占内存空间”的说法是错误的，实际上引用也会再用内存空间。

### 11.C++的空类八个默认函数
　　C++空类会默认产生的8个类成员函数，需要牢记函数的具体形式，尽可能少用默认函数，自己重新定义。参考：
- 缺省构造函数
- 拷贝构造函数
- 赋值构造函数
- 析构函数
- 取值操作符函数
- const 取值操作符
- 移动构造函数C++11
- 移动赋值构造函数C++11
- 如果你的类型需要, 就让它们支持拷贝 / 移动. 否则, 就把隐式产生的拷贝和移动函数禁用.（google编码规范）

### 12.C++11中delete和default的作用
**=default**显式缺省，告知编译器生成函数默认的缺省版本
**=delete**显式删除，告知编译器不生成函数默认的缺省版本
C++11中引进这两种新特性的目的是为了增强对“类默认函数的控制”，从而让程序员更加精准地去控制默认版本的函数

### 13.C++11禁用隐式类型转换explicit
　　explicit关键字用来修饰类的构造函数，被修饰的构造函数的类，不能发生相应的隐式类型转换，只能以显示的方式进行类型转换。

### 14.new/delete和malloc/free的使用
(1) new/delete是C++的运算符，malloc/free是C/C++的库函数
(2) new/delete和malloc/free必须配套使用
(3) mallocl/free仅仅是在堆中分配内存，需要自己指定分配内存大小以及指针类型的转换
(4) new/delete会根据对象的类型调用对应的构造函数和析构函数，因此在C++中使用更加多
(5) new是类型安全的，而malloc不是，比如：
```
	int* p = new float[2]; // 编译时指出错误
	int* p = malloc(2*sizeof(float)); // 编译时无法指出错误
```
### 15 new/operator new和placement new
参考：https://www.cnblogs.com/luxiaoxun/archive/2012/08/10/2631812.html
(1) new：新建对象时用，是C++操作符。本质上是调用operator new函数分配内存，然后调用构造函数生成类的对象，返回对应类型的指针。
(2) operator new就像operator + 一样，是可以重载的。如果类中没有重载operator new，那么调用的就是全局的::operator new来完成堆的分配。要实现不同的内存分配行为，应该重载operator new。
(3) placement new：只是operator new重载的一个版本。它并不分配内存，只是返回指向已经分配好的某段内存的一个指针。因此不能删除它，但需要调用对象的析构函数。如果你想在已经分配的内存中创建一个对象，使用new时行不通的。也就是说placement new允许你在一个已经分配好的内存中（栈或者堆中）构造一个新的对象。原型中void* p实际上就是指向一个已经分配好的内存缓冲区的的首地址。
```
placement new函数形式：void *operator new( size_t, void * p ) throw() { return p; }
```
### 16.sizeof
(1)空对象的大小为1个字节
(2)编译器内存对齐
(3)继承
(4)虚函数的影响
参考：https://www.cnblogs.com/wxquare/p/6675523.html 学习C++对象模型

### 17.编译器内存对齐
　　现代计算机中内存空间都是按照 byte 划分的，从理论上讲似乎对任何类型的变量的访问可以从任何地址开始，但是实际的计算机系统对基本类型数据在内存中存放的位置有限制，它们会要求这些数据的首地址的值是某个数k（通常它为4或8）的倍数，这就是所谓的内存对齐。假如没有内存对齐机制，数据可以任意存放，现在一个int变量存放在从地址1开始的联系四个字节地址中，该处理器去取数据时，要先从0地址开始读取第一个4字节块,剔除不想要的字节（0地址）,然后从地址4开始读取下一个4字节块,同样剔除不要的数据（5，6，7地址）,最后留下的两块数据合并放入寄存器.这需要做很多工作。**因此编译器内存对齐是为了提高数据读写的效率。**每个特定平台上的编译器都有自己的默认“对齐系数”（也叫对齐模数）。gcc中默认#pragma pack(4)，可以通过预编译命令#pragma pack(n)，n = 1,2,4,8,16来改变这一系数。

### 18.浅拷贝和深拷贝
　　对于含有堆内存的对象，浅拷贝只是对指针的拷贝，拷贝后两个指针指向同一个内存空间，深拷贝对指针所指向的内容进行拷贝。默认拷贝构造函数为浅拷贝。

### 19.friend友元函数和友元类
　　C++中使用类对数据进行了隐藏和封装，类的数据成员一般都定义为私有成员，成员函数一般都定义为公有的，以此提供类与外界的通讯接口。但是，有时需要定义一些函数，这些函数不是类的一部分，但又需要频繁地访问类的数据成员，这时可以将这些函数定义为该函数的友元函数。除了友元函数外，还有友元类，两者统称为友元。**友元的作用是提高了程序的运行效率（即减少了类型检查和安全性检查等都需要时间开销），但它破坏了类的封装性和隐藏性，使得非成员函数可以访问类的私有成员。**实际中这一特性很少使用。
参考：https://www.cnblogs.com/wxquare/p/5015440.html


### 20.类的初始化列表
(1)初始化列表是C++11中新增的类成员初始化方式。 
(2)没有默认构造函数的类自定义类型成员必须使用初始化列表
(3)const成员、引用类型成员必须使用初始化列表。
(4)初始化列表中初始化初始化的顺序与成员定义的顺序相同，与初始化列表的顺序无关
**初始化列表的优点**：主要是对于自定义类型，初始化列表是作用在函数体之前，他调用构造函数对对象进行初始化。然而在函数体内，需要先调用构造函数，然后进行赋值，这样效率就不如初始化列表

### 21. 重载、覆盖和重写
(1) **重载(overload)**:同类中同名函数，参数的类型、个数或者返回类型不同。与virtual无关。（同类中）
(2) **覆盖(override)**:基类函数virtual函数，派生类中重写该函数，函数名称和参数完全相同。（基类和派生类中，基类函数virtual函数）
(3) **重写(overwrite)**：派生类的函数屏蔽了与其同名的基类函数，与virtual无关，是一种派生类和基类之间**同名覆盖**。

### 22. 继承/多继承/虚继承
关于继承这个问题，不同语言有自己的设计思路，有的支持继承，单继承。有的支持组合，C++支持多继承。
(1) 单继承、多继承，继承时构造函数和析构函数的调用顺序。C++支持多继承。
(2) 继承方式，public/protected/private,默认为private继承，通常为public继承
(3) 友元函数不能被继承,那么基类的友元函数是不能被派生类继承
(4) **静态成员和静态成员函数是可以继承的**
(5) 虚继承的概念
(6) C++ 对象内存模型：https://www.cnblogs.com/wxquare/p/6675523.html
　　虚拟继承是多重继承中特有的概念。虚拟基类是为解决多重继承而出现的。如:类D继承自类B1、B2，而类B1、B2都继承自类A，因此在类D中两次出现类A中的变量和函数。为了节省内存空间，可以将B1、B2对A的继承定义为虚拟继承，而A就成了虚拟基类。实现的代码如下：虚拟继承在一般的应用中很少用到，所以也往往被忽视，这也主要是因为在C++中，多重继承是不推荐的，也并不常用，而一旦离开了多重继承，虚拟继承就完全失去了存在的必要因为这样只会降低效率和占用更多的空间。
```
class A
class B1:public virtual A;
class B2:public virtual A;
class D:public B1,public B2;
```
谷歌编码规范：
1. 使用组合常常比使用继承更合理. 如果使用继承的话, 定义为 public 继承.
2. C++ 实践中, 继承主要用于两种场合: 实现继承, 子类继承父类的实现代码; 接口继承, 子类仅继承父类的方法名称.
3. 必要的话, 析构函数声明为 virtual. 如果你的类有虚函数, 则析构函数也应该为虚函数
4. 真正需要用到多重实现继承的情况少之又少. 只在以下情况我们才允许多重继承: 最多只有一个基类是非抽象类; 其它基类都是以 Interface 为后缀的 纯接口类.

### 23. virtual虚
C++中的virtual虚问题是一大难点，需要掌握以下几点：
(1) **虚基类成员函数**，派生类override这个虚函数，虚成员函数可以实现多态性（多态）
(2) **虚析构函数**，在继承关系中，基类的构造函数经常为虚函数，这是因为当用一个基类的指针删除一个派生类的对象时，如果基类的析构函数不是虚函数，派生类的析构函数不会被调用，造成内存泄露。因此对于含有虚函数的类的析构函数一般为虚函数。
(3) 虚函数的实现，虚函数表和虚函数指针，参考C++内存模型。
(4) 虚函数的动态绑定机制与运行期多态。参考：https://www.cnblogs.com/wxquare/p/5017326.html
(5) 虚继承，在多重继承关系中，为了避免菱形继承导致的资源浪费，会使用虚继承。
(6) 虚继承的实现，虚基表，参考C++内存模型。
(7) 内敛函数不能为虚函数，因为内敛函数时静态的
(8) 静态函数不能为虚函数，因为静态函数属于类，不属于对象
(9) 构造函数不能为虚函数，需要构造函数初始化虚函数表
(10) 纯虚函数，类似于的接口的作用。


### 24.C++对象模型
https://www.cnblogs.com/wxquare/p/5017326.html

### 25.C++11中的移动语义
C++中的拷贝语义和移动语义，右值引用和移动语义？C++11右值引用和移动语义 对含堆内存类的临时对象的拷贝和赋值函数的优化，使的深拷贝转化为浅拷贝。拷贝语义和移动语义
参考：https://www.cnblogs.com/wxquare/p/6836271.html

### 26.C++11智能指针
参考：https://www.cnblogs.com/wxquare/p/4759020.html

### 27.模板编程
C++模板编程问题？模板，函数模板，类模板，C++ 类模板碰到static，每个类型一个static值，C++ 类中不能包含虚函数模板，类模板可以包含虚函数？模板的声明和实现为何要放在头文件中？
C++中模板与泛型编程：https://www.cnblogs.com/wxquare/p/4743180.html
模板的声明和实现为什么要放在头文件中？
https://www.cnblogs.com/wanyao/archive/2011/06/29/2093588.html
什么是模板元编程？



### 28. 访函数/函数指针/lamda表达式
(1)函数对象(function object)又叫仿函数(functor)，就是重载了调用运算符()的类，所生成的对象，就叫做函数对象/仿函数。因为重载了()之后，我们就能像函数一样去使用这个类，同时类里面又可以储存一些信息，所以要比普通的函数更加灵活
(2)函数指针也是一个函数对象，因为指针在C++中都是对象。
(3)lamda表达式
访函数和函数指针的区别，哪个效率更高？


### 29. C++异常处理
为什么C++很少使用异常处理？
https://www.zhihu.com/question/22889420
1. 返回错误码
2. 断言
3. 异常处理Exception。代价是产生的二进制文件大小的增加，因为异常产生的位置决定了需要如何做栈展开（stack unwinding），这些数据需要存储在表里。典型情况，使用异常和不使用异常比，二进制文件大小会有约百分之十到二十的上升。C++ 由于本身是强调实时性、高性能、低开销的语言，异常在某些使用场景下会被人诟病。然而，异常对表达性的改进是巨大的。因而，除非项目有特别严苛的实时性、空间之类的限制，使用异常应当是缺省选择。
构造函数可以抛异常，析构函数不能抛异常？
https://www.cnblogs.com/fly1988happy/archive/2012/04/11/2442765.html
C++标准中假定了析构函数中不应该，也不永许抛出异常的。通常异常发生时，c++的机制会调用已经构造对象的析构函数来释放资源，此时若析构函数本身也抛出异常，则前一个异常尚未处理，又有新的异常，会造成程序崩溃的问题


### 30. RTTI
RTTI是”Runtime Type Information”的缩写，意思是运行时类型信息，它提供了运行时确定对象类型的方法。
实现机制是虚函数和虚函数表
谷歌禁用使用 RTTI.
在运行时判断类型通常意味着设计问题. 如果你需要在运行期间确定一个对象的类型, 这通常说明你需要考虑重新设计你的类.
随意地使用 RTTI 会使你的代码难以维护. 它使得基于类型的判断树或者 switch 语句散布在代码各处. 如果以后要进行修改, 你就必须检查它们


### 31. 前置声明(forward declaration)
所谓「前置声明」（forward declaration）是类、函数和模板的纯粹声明，没伴随着其定义
尽可能地避免使用前置声明。使用 #include 包含需要的头文件即可。

### 32. #include头文件的顺序
1. C 系统文件
2. C++ 标准库文件
3. 其它库的.h文件
4. 本项目内的的. 文件
```
#include <sys/types.h>
#include <unistd.h>

#include <hash_map>
#include <vector>

#include "base/basictypes.h"
#include "base/commandlineflags.h"
#include "foo/public/bar.h"
```

### 33. 命名空间namespace
1. 命名空间将全局作用域细分为独立的, 具名的作用域, 可有效防止全局作用域的命名冲突
2. 不应该使用 using 指示 引入整个命名空间的标识符号 
3. 禁止用内联命名空间

```
// 禁止 —— 污染命名空间
using namespace foo;
```

```
// .h 文件
namespace mynamespace {

// 所有声明都置于命名空间中
// 注意不要使用缩进
class MyClass {
    public:
    ...
    void Foo();
};

} // namespace mynamespace
```
```
// .h 文件
namespace mynamespace {

// 所有声明都置于命名空间中
// 注意不要使用缩进
class MyClass {
    public:
    ...
    void Foo();
};

} // namespace mynamespace
```

### 34. 接口类
接口是指满足特定条件的类, 这些类以 Interface 为后缀 (不强制).

### 35. 函数使用引用参数
1. 在 C 语言中, 如果函数需要修改变量的值, 参数必须为指针, 如 int foo(int *pval). 在 C++ 中, 函数还可以声明为引用参数: int foo(int &val).
2. 定义引用参数可以防止出现 (*pval)++ 这样丑陋的代码. 引用参数对于拷贝构造函数这样的应用也是必需的. 同时也更明确地不接受空指针.
3. 函数参数列表中, 所有引用参数都必须是 const

### 36. 函数返回值后置语法
只有在常规写法 (返回类型前置) 不便于书写或不便于阅读时使用返回类型后置语法.
C++11引入了后置返回值的语法：
auto foo(int x) -> int;
后置返回类型是显式地指定 Lambda 表达式 的返回值的唯一方式. 某些情况下, 编译器可以自动推导出 Lambda 表达式的返回类型, 但并不是在所有的情况下都能实现. 即使编译器能够自动推导, 显式地指定返回类型也能让读者更明了.有时在已经出现了的函数参数列表之后指定返回类型, 能够让书写更简单, 也更易读, 尤其是在返回类型依赖于模板参数时. 例如:
```
template <class T, class U> auto add(T t, U u) -> decltype(t + u);
```

### 37. 流
1. 流用来替代 printf() 和 scanf()
2. 有了流, 在打印时不需要关心对象的类型. 不用担心格式化字符串与参数列表不匹配 (虽然在 gcc 中使用 printf 也不存在这个问题). 流的构造和析构函数会自动打开和关闭对应的文件
3. 不要使用流, 除非是日志接口需要. 使用 printf 之类的代替.

### 38.前置自增和自减
不考虑返回值的话, 前置自增 (++i) 通常要比后置自增 (i++) 效率更高. 因为后置自增 (或自减) 需要对表达式的值 i 进行一次拷贝. 如果 i 是迭代器或其他非数值类型, 拷贝的代价是比较大的. 既然两种自增方式实现的功能一样, 为什么不总是使用前置自增呢?

### 39.constexpr
1. 在 C++11 里，用 constexpr 来定义真正的常量，或实现常量初始化
2. 变量可以被声明成 constexpr 以表示它是真正意义上的常量，即在编译时和运行时都不变。函数或构造函数也可以被声明成 constexpr, 以用来定义 constexpr 变量
3. 靠 constexpr 特性，方才实现了 C++ 在接口上打造真正常量机制的可能。好好用 constexpr 来定义真・常量以及支持常量的函数。避免复杂的函数定义，以使其能够与constexpr一起使用。 千万别痴心妄想地想靠 constexpr 来强制代码
 

### 40. Lambda 表达式
1. 适当使用 lambda 表达式。别用默认 lambda 捕获，所有捕获都要显式写出来。
2. Lambda 表达式是创建匿名函数对象的一种简易途径，常用于把函数当参数传，例如：
```
std::sort(v.begin(), v.end(), [](int x, int y) {
    return Weight(x) < Weight(y);
});
```

### 41 .命名
1. 文件名：http_server_logs.h，http_server_logs.cc
2. 类型性：MyExcitingClass
3. 变量名：string table_name
4. 函数名：AddTableEntry()
5. 命名空间：websearch::index




## C/C++ 常见问题
1. 如何让类对象只在栈（堆）上分配空间？https://blog.csdn.net/hxz_qlh/article/details/13135433
只能在栈上建立对象：
2. C++不可继承类的实现？
https://www.cnblogs.com/wxquare/p/7280025.html

3. 如何定义和实现一个类的成员函数为回调函数？
友元函数/静态成员函数消除this指针的影响
4. C++复制构造函数的参数为什么是引用类型？
- 编译时报错
- 需要首先调用该类的拷贝构造函数来初始化形参(局部对象)，造成无线循环递归

5. C++全局对象如何在main函数之前构造和析构？
https://blog.csdn.net/iyangyoulei/article/details/46925973


## 其它
　STL是C++程序重要的组成部分，这里主要记录工作中遇到的问题，目标是熟悉C++的两个网站和两本书籍：
1. http://www.cplusplus.com/reference/
2. https://en.cppreference.com/w/
3. 《Effective STL》书籍
4. 《STL源码分析》书籍

### 1. 熟悉STL中17种容器及其背后对应的数据结构

### 2. map和unordered_map的区别
1. map背后是红黑树，unordered_map背后是哈希表
2. map是key值有序的，unordered_map是key值无序的
3. 两者内存消耗差不多，但是插入/查找/删除效率unordered_map是map的2到3倍
4. unordered_map是通过链地址法解决冲突的
5. std::map [] operator 和 insert 的区别。如果key已经存在，[] operator会将key对应的value用新值替换，而insert会返回一个pair说这组元素已经存在，如果key不存在，二者效果相同




### 4. priority_queue优先队列的实现
　　priority_queue 优先队列，其底层是用堆来实现的。在优先队列中，队首元素一定是当前队列中优先级最高的那一个。在优先队列中，没有 front() 函数与 back() 函数，而只能通过 top() 函数来访问队首元素（也可称为堆顶元素），也就是优先级最高的元素。基本操作有：
- empty() 如果队列为空返回真
- pop() 删除对顶元素
- push() 加入一个元素
- size() 返回优先队列中拥有的元素个数
- top() 返回优先队列对顶元素
- priority_queue 默认为大顶堆，即堆顶元素为堆中最大元素（比如：在默认的int型中先出队的为较大的数）。

### 5. 迭代器和迭代器失效iterator
　　为了提高C++编程的效率，STL中提供了许多容器，包括vector、list、map、set等。有些容器例如vector可以通过脚标索引的方式访问容器里面的数据，但是大部分的容器不能使用这种方式，例如list、map、set。STL中每种容器在实现的时候设计了一个内嵌的iterator类，不同的容器有自己专属的迭代器，使用迭代器来访问容器中的数据。除此之外，通过迭代器，可以将容器和通用算法结合在一起，只要给予算法不同的迭代器，就可以对不同容器执行相同的操作，例如find查找函数。迭代器对指针的一些基本操作如*、->、++、==、!=、=进行了重载，使其具有了遍历复杂数据结构的能力，其遍历机制取决于所遍历的数据结构，所有迭代的使用和指针的使用非常相似。通过begin，end函数获取容器的头部和尾部迭代器，end 迭代器不包含在容器之内，当begin和end返回的迭代器相同时表示容器为空。
　　容器的插入insert和erase操作可能导致迭代器失效，对于erase操作不要使用操作之前的迭代器，因为erase的那个迭代器一定失效了，正确的做法是返回删除操作时候的那个迭代器。
参考：https://www.cnblogs.com/wxquare/p/4699429.html

### 6. 容器的线程安全性Thread safety
　　STL为了效率，没有给所有操作加锁。不同线程同时读同一容器对象没关系，不同线程同时写不同的容器对象没关系。但不能同时又读又写同一容器对象的。因此，多线程要同时读写时，还是要自己加锁。

### 7. STL 排序
1. sort,快排加插入排序
2. stable_sort，稳定排序
3. sort_heap，堆排序
4. list.sort，链表归并排序
https://www.cnblogs.com/wxquare/p/4922733.html

### 8. STL容器的内存管理方式

### 9. vector和map的内存释放

4. STL容器的内存管理方式？


### 其它常见的问题
1. vector和map的内存释放问题？容器删除数据的时候注意迭代器失效？vector和map正确的内存释放？
2. C++ 的iostream 的局限
根据以上分析，我们可以归纳 iostream 的局限：输入方面，istream 不适合输入带格式的数据，因为“纠错”能力不强，进一步的分析请见孟岩写的《契约思想的一个反面案例》，孟岩说“复杂的设计必然带来复杂的使用规则，而面对复杂的使用规则，用户是可以投票的，那就是你做你的，我不用！”可谓鞭辟入里。如果要用 istream，我推荐的做法是用 getline() 读入一行数据，然后用正则表达式来判断内容正误，并做分组，然后用 strtod/strtol 之类的函数做类型转换。这样似乎更容易写出健壮的程序。输出方面，ostream 的格式化输出非常繁琐，而且写死在代码里，不如 stdio 的小语言那么灵活通用。建议只用作简单的无格式输出。log 方面，由于 ostream 没有办法在多线程程序中保证一行输出的完整性，建议不要直接用它来写 log。如果是简单的单线程程序，输出数据量较少的情况下可以酌情使用。当然，产品代码应该用成熟的 logging 库，而不要用其它东西来凑合。in-memory 格式化方面，由于 ostringstream 会动态分配内存，它不适合性能要求较高的场合。文件 IO 方面，如果用作文本文件的输入或输出，(i|o)fstream 有上述的缺点；如果用作二进制数据输入输出，那么自己简单封装一个 File class 似乎更好用，也不必为用不到的功能付出代价（后文还有具体例子）。ifstream 的一个用处是在程序启动时读入简单的文本配置文件。如果配置文件是其他文本格式（XML 或 JSON），那么用相应的库来读，也用不到 ifstream。性能方面，iostream 没有兑现“高效性”诺言。iostream 在某些场合比 stdio 快，在某些场合比 stdio 慢，对于性能要求较高的场合，我们应该自己实现字符串转换（见后文的代码与测试）。iostream 性能方面的一个注脚：在线 ACM/ICPC 判题网站上，如果一个简单的题目发生超时错误，那么把其中 iostream 的输入输出换成 stdio，有时就能过关。
既然有这么多局限，iostream 在实际项目中的应用就大为受限了，在这上面投入太多的精力实在不值得。说实话，我没有见过哪个 C++ 产品代码使用 iostream 来作为输入输出设施。 

4. STL::list::sort链表归并排序







---
title: C/C++程序的项目构建、编译、调试工具和方法
categories:
- 计算机基础
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





---
title: C/C++程序性能分析的工具
categories: 
- 计算机基础
---


　　C++代码编译测试完成功能之后，有时会遇到一些性能问题，此时需要学会使用一些工具对其进行性能分析，找出程序的性能瓶颈，然后进行优化，基本需要掌握下面几个命令：
1. time分析程序的执行时间
2. top观察程序资源使用情况
3. perf/gprof进一步分析程序的性能
4. 内存问题与valgrind
5. 自己写一个计时器，计算局部函数的时间



## 一、time
### 1.shell time。
　　time非常方便获取程序运行的时间，包括用户态时间user、内核态时间sys和实际运行的时间real。我们可以通过(user+sys)/real计算程序CPU占用率，判断程序时CPU密集型还是IO密集型程序。
	$time ./kcf2.0 ../data/bag.mp4 312 146 106 98 1 196 result.csv 1
	real	0m2.065s
	user	0m4.598s
	sys	    0m0.907s
cpu使用率：(4.598+0.907)/2.065=267%
视频帧数196，196/2.065=95


### 2./usr/bin/time
　　Linux中除了shell time，还有/usr/bin/time，它能获取程序运行更多的信息，通常带有-v参数。
```
$ /usr/bin/time -v  ./kcf2.0 ../data/bag.mp4 312 146 106 98 1 196 result.csv 1
    User time (seconds): 4.28                                  # 用户态时间
	System time (seconds): 1.11                                # 内核态时间
	Percent of CPU this job got: 279%                          # CPU占用率
	Elapsed (wall clock) time (h:mm:ss or m:ss): 0:01.93   
	Average shared text size (kbytes): 0
	Average unshared data size (kbytes): 0
	Average stack size (kbytes): 0
	Average total size (kbytes): 0
	Maximum resident set size (kbytes): 63980                  # 最大内存分配
	Average resident set size (kbytes): 0
	Major (requiring I/O) page faults: 0
	Minor (reclaiming a frame) page faults: 19715              # 缺页异常
	Voluntary context switches: 3613                           # 上下文切换
	Involuntary context switches: 295682
	Swaps: 0
	File system inputs: 0
	File system outputs: 32
	Socket messages sent: 0
	Socket messages received: 0
	Signals delivered: 0
	Page size (bytes): 4096
	Exit status: 0
```


## 二、top
top是linux系统的任务管理器，它既能看系统所有任务信息，也能帮助查看单个进程资源使用情况。
主要有以下几个功能：
1. 查看系统任务信息：
 Tasks:  87 total,   1 running,  86 sleeping,   0 stopped,   0 zombie
2. 查看CPU使用情况
 Cpu(s):  0.0%us,  0.2%sy,  0.0%ni, 99.7%id,  0.0%wa,  0.0%hi,  0.0%si,  0.2%st
3. 查看内存使用情况
 Mem:    377672k total,   322332k used,    55340k free,    32592k buffers
4. 查看单个进程资源使用情况 
	- PID：进程的ID
	- USER：进程所有者
	- PR：进程的优先级别，越小越优先被执行
	- NInice：值
	- VIRT：进程占用的虚拟内存
	- RES：进程占用的物理内存
	- SHR：进程使用的共享内存
	- S：进程的状态。S表示休眠，R表示正在运行，Z表示僵死状态，N表示该进程优先值为负数
	- %CPU：进程占用CPU的使用率
	- %MEM：进程使用的物理内存和总内存的百分比
	- TIME+：该进程启动后占用的总的CPU时间，即占用CPU使用时间的累加值。
	- COMMAND：进程启动命令名称
5. 除此之外top还提供了一些交互命令：
	- q:退出
	- 1:查看每个逻辑核
	- H：查看线程
	- P：按照CPU使用率排序
	- M：按照内存占用排序

参考：https://linuxtools-rst.readthedocs.io/zh_CN/latest/tool/top.html


## 三、perf
参考：https://www.ibm.com/developerworks/cn/linux/l-cn-perf1/index.html
参考：https://zhuanlan.zhihu.com/p/22194920

### 1. perf stat
　　做任何事都最好有条有理。老手往往能够做到不慌不忙，循序渐进，而新手则往往东一下，西一下，不知所措。面对一个问题程序，最好采用自顶向下的策略。先整体看看该程序运行时各种统计事件的大概，再针对某些方向深入细节。而不要一下子扎进琐碎细节，会一叶障目的。有些程序慢是因为计算量太大，其多数时间都应该在使用 CPU 进行计算，这叫做 CPU bound 型；有些程序慢是因为过多的 IO，这种时候其 CPU 利用率应该不高，这叫做 IO bound 型；对于 CPU bound 程序的调优和 IO bound 的调优是不同的。如果您认同这些说法的话，Perf stat 应该是您最先使用的一个工具。它通过概括精简的方式提供被调试程序运行的整体情况和汇总数据。虚拟机上面有些参数不全面，cycles、instructions、branches、branch-misses。下面的测试数据来自服务器。                                          
**$time ./kcf2.0 ../data/bag.mp4 312 146 106 98 1 196 result.csv 1**
```
     25053.120420      task-clock (msec)         #   17.196 CPUs utilized          
         1,509,877      context-switches          #    0.060 M/sec                  
             3,427      cpu-migrations            #    0.137 K/sec                  
            34,025      page-faults               #    0.001 M/sec                  
    65,242,918,152      cycles                    #    2.604 GHz                    
                 0      stalled-cycles-frontend   #    0.00% frontend cycles idle   
                 0      stalled-cycles-backend    #    0.00% backend  cycles idle   
    64,695,693,541      instructions              #    0.99  insns per cycle        
     8,049,836,066      branches                  #  321.311 M/sec                  
        42,734,371      branch-misses             #    0.53% of all branches        

       1.456907056 seconds time elapsed
```
### 2. perf top
　　Perf top 用于实时显示当前系统的性能统计信息。该命令主要用来观察整个系统当前的状态，比如可以通过查看该命令的输出来查看当前系统最耗时的内核函数或某个用户进程。
### 3. perf record/perf report
　　使用 top 和 stat 之后，这时对程序基本性能有了一个大致的了解，为了优化程序，便需要一些粒度更细的信息。比如说您已经断定目标程序计算量较大，也许是因为有些代码写的不够精简。那么面对长长的代码文件，究竟哪几行代码需要进一步修改呢？这便需要使用 perf record 记录单个函数级别的统计信息，并使用 perf report 来显示统计结果。您的调优应该将注意力集中到百分比高的热点代码片段上，假如一段代码只占用整个程序运行时间的 0.1%，即使您将其优化到仅剩一条机器指令，恐怕也只能将整体的程序性能提高 0.1%。俗话说，好钢用在刀刃上，要优化热点函数。

```
perf record – e cpu-clock ./t1 
perf report
```
增加-g参数可以获取调用关系
```
perf record – e cpu-clock – g ./t1 
perf report
```
$perf record -e cpu-clock -g ./kcf2.0 ../data/bag.mp4 312 146 106 98 1 196 result.csv 1
$perf report
![](https://github.com/wxquare/wxquare.github.io/raw/hexo/source/photos/perf_kcf2.0.jpg)

经过perf的分析，我们的目标应该很明确了，cv::DFT和get_feature这两个函数比较耗时，另外还有一个和线程相关的操作也比较耗时，接下来要去分析代码，做代码级别的优化。

## 四、gprof
参考： https://blog.csdn.net/stanjiang2010/article/details/5655143




## 五、内存问题与valgrind
### 5.1常见的内存问题
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

### 5.1 valgrind内存检测

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


## 六、自定义timer计时器
```
class timer {
public:
    clock_t start;
    clock_t end;
    string name;
    timer(string n) {
        start = clock();
        name = n;
    }
    ~timer() {
        end = clock();
        printf("%s time: %f \n", name.c_str(), 
            (end - start) * 1.0 / CLOCKS_PER_SEC * 1000);
    }
};
```