---
title: C/C++ 基础和常见面试问题
date: 2023-10-13
categories: 
- 系统设计
---

## 基础语法与关键字

### const 关键字
- 定义常量、指针、引用、对象,const进行修饰的变量的值在程序的任意位置将不能再被修改
- 修饰参数: `const int x`
- 修饰成员变量: const成员变量必须通过初始化列表进行初始化
- 修饰成员函数
- C++ mutable变量突破const修饰成员函数的限制

### static 关键字
- 静态全局变量
- 修饰函数内部的局部变量,作用相当于全局变量
- 类的静态成员变量
- 类的静态成员函数

### 宏定义和内联函数
- 内联函数和宏定义减少函数调用所带来的时间和空间的开销,以空间换时间的策略
- 宏是在预编译阶段简单文本替代,inline在编译阶段实现展开
- 宏肯定会被替代,而复杂的inline函数不会被展开
- 宏容易出错(运算顺序),且难以被调试,inline不会
- 宏不是类型安全,而inline是类型安全的
- 当函数size太大,inline虚函数,函数中存在循环或递归,内联可能失效

### extern 与 static
- extern是C/C++语言中表明函数和全局变量作用范围(可见性)的关键字
- 该关键字告诉编译器,其声明的函数和变量可以在本模块或其它模块中使用
- 与extern对应的关键字是static,被它修饰的全局变量和函数只能在本模块使用

### extern "C"
- extern "C"是为了实现C和C++的混合编程
- C和C++的编译和链接是不完全相同的
- extern "C"表明它按照类C的编译和连接规约来编译和连接,而不是C++的编译和链接
- C++是一个面向对象语言,它为了支持函数的重载,在编译的时候会带上参数的类型来唯一标识每个函数
- C语言中并没有重载和类这些特性,故并不像C++那样print(int i)会被编译为_print_int

### struct和class的区别
- 在C++中struct和class的区别比较小,主要类成员的默认访问权限和继承权限
- 默认继承权限:如果不明确指定,来自class的继承按照private继承处理,来自struct的继承按照public继承处理
- 成员的默认访问权限:class的成员默认是private权限,struct默认是public权限
- 仅当只有数据成员时使用struct,其它一概使用class(google编码规范)

### 类型安全
- 静态类型 vs 动态类型: 静态类型(C/C++,java,golang),动态类型(python)
- 弱类型 vs 强类型: 弱类型(C、C++),强类型(python、golang)
- 类型安全: 一般来说弱类型存在隐含的类型转换都不是类型安全的,而强类型是类型安全的

### C++中的四种类型转换
- const_cast: 字面上理解就是去const属性
- static_cast: 命名上理解是静态类型转换。如int转换成char。类似于C风格的强制转换
- dynamic_cast: 命名上理解是动态类型转换。如子类和父类之间的多态类型转换
- reinterpret_cast: 仅仅重新解释类型,但没有进行二进制的转换

### volatile关键字
- volatile指出变量是随时可能发生变化的
- 每次使用它的时候必须从内存中读取
- 编译器生成的汇编代码会重新从内存读取数据
- 防止编译器优化

### 指针和引用
- 指针指向一块内存,它的内容是所指内存的地址
- 引用则是某块内存的别名,引用初始化后不能改变指向
- 使用时,引用更加安全,指针更加灵活
- 初始化:引用必须初始化,且初始化之后不能改变;指针可以不必初始化,且指针可以改变所指的对象
- 空值:指针可以指向空值,不存在指向空值的引用
- 引用和指针指向一个对象时,引用的创建和销毁不会调用类的拷贝构造函数和析构函数
- 引用和指针与const:存在常量指针和常量引用指针,表示指向的对象是常量
- 函数参数传递时使用指针或者引用的效果是相同的,都是简洁操作主调函数中的相关变量
- sizeof引用的时候是对象的大小,sizeof指针是指针本身的大小

## 面向对象与类设计

### C++的空类八个默认函数
- 缺省构造函数
- 拷贝构造函数
- 赋值构造函数
- 析构函数
- 取值操作符函数
- const取值操作符
- 移动构造函数C++11
- 移动赋值构造函数C++11

### C++11中delete和default的作用
- =default显式缺省,告知编译器生成函数默认的缺省版本
- =delete显式删除,告知编译器不生成函数默认的缺省版本
- C++11中引进这两种新特性的目的是为了增强对"类默认函数的控制"

### C++11禁用隐式类型转换explicit
- explicit关键字用来修饰类的构造函数
- 被修饰的构造函数的类,不能发生相应的隐式类型转换
- 只能以显示的方式进行类型转换

### new/delete和malloc/free的使用
- new/delete是C++的运算符,malloc/free是C/C++的库函数
- new/delete和malloc/free必须配套使用
- mallocl/free仅仅是在堆中分配内存,需要自己指定分配内存大小以及指针类型的转换
- new/delete会根据对象的类型调用对应的构造函数和析构函数
- new是类型安全的,而malloc不是

### new/operator new和placement new
- new:新建对象时用,是C++操作符
- operator new就像operator + 一样,是可以重载的
- placement new:只是operator new重载的一个版本
- 它并不分配内存,只是返回指向已经分配好的某段内存的一个指针

### sizeof
- 空对象的大小为1个字节
- 编译器内存对齐
- 继承
- 虚函数的影响

### 编译器内存对齐
- 现代计算机中内存空间都是按照byte划分的
- 实际的计算机系统对基本类型数据在内存中存放的位置有限制
- 它们会要求这些数据的首地址的值是某个数k(通常它为4或8)的倍数
- 这就是所谓的内存对齐
- 编译器内存对齐是为了提高数据读写的效率

### 浅拷贝和深拷贝
- 对于含有堆内存的对象,浅拷贝只是对指针的拷贝
- 拷贝后两个指针指向同一个内存空间
- 深拷贝对指针所指向的内容进行拷贝
- 默认拷贝构造函数为浅拷贝

### friend友元函数和友元类
- 友元的作用是提高了程序的运行效率
- 但它破坏了类的封装性和隐藏性
- 使得非成员函数可以访问类的私有成员
- 实际中这一特性很少使用

### 类的初始化列表
- 初始化列表是C++11中新增的类成员初始化方式
- 没有默认构造函数的类自定义类型成员必须使用初始化列表
- const成员、引用类型成员必须使用初始化列表
- 初始化列表中初始化初始化的顺序与成员定义的顺序相同,与初始化列表的顺序无关
- 初始化列表的优点:主要是对于自定义类型,初始化列表是作用在函数体之前

### 重载、覆盖和重写
- 重载(overload):同类中同名函数,参数的类型、个数或者返回类型不同
- 覆盖(override):基类函数virtual函数,派生类中重写该函数
- 重写(overwrite):派生类的函数屏蔽了与其同名的基类函数

### 继承/多继承/虚继承
- 单继承、多继承,继承时构造函数和析构函数的调用顺序
- 继承方式,public/protected/private,默认为private继承
- 友元函数不能被继承
- 静态成员和静态成员函数是可以继承的
- 虚继承的概念
- C++对象内存模型

### virtual虚
- 虚基类成员函数,派生类override这个虚函数
- 虚析构函数
- 虚函数的实现,虚函数表和虚函数指针
- 虚函数的动态绑定机制与运行期多态
- 虚继承,在多重继承关系中,为了避免菱形继承导致的资源浪费
- 虚继承的实现,虚基表
- 内敛函数不能为虚函数
- 静态函数不能为虚函数
- 构造函数不能为虚函数
- 纯虚函数,类似于的接口的作用

### C++对象模型
- 对象的内存布局
- 虚函数表
- 虚基表
- 继承关系中的内存布局

### C++11中的移动语义
- C++中的拷贝语义和移动语义
- 右值引用和移动语义
- 对含堆内存类的临时对象的拷贝和赋值函数的优化
- 使的深拷贝转化为浅拷贝

### C++11智能指针
- unique_ptr
- shared_ptr
- weak_ptr
- 智能指针的使用场景

### 模板编程
- 模板,函数模板,类模板
- C++类模板碰到static
- 每个类型一个static值
- C++类中不能包含虚函数模板
- 类模板可以包含虚函数
- 模板的声明和实现为何要放在头文件中
- 模板元编程

### 访函数/函数指针/lamda表达式
- 函数对象(function object)又叫仿函数(functor)
- 就是重载了调用运算符()的类,所生成的对象
- 函数指针也是一个函数对象
- lamda表达式

### C++异常处理
- 返回错误码
- 断言
- 异常处理Exception
- 构造函数可以抛异常,析构函数不能抛异常

### RTTI
- RTTI是"Runtime Type Information"的缩写
- 意思是运行时类型信息
- 它提供了运行时确定对象类型的方法
- 实现机制是虚函数和虚函数表
- 谷歌禁用使用RTTI

### 前置声明(forward declaration)
- 所谓「前置声明」(forward declaration)是类、函数和模板的纯粹声明
- 没伴随着其定义
- 尽可能地避免使用前置声明
- 使用#include包含需要的头文件即可

### #include头文件的顺序
- C系统文件
- C++标准库文件
- 其它库的.h文件
- 本项目内的的.h文件

### 命名空间namespace
- 命名空间将全局作用域细分为独立的,具名的作用域
- 可有效防止全局作用域的命名冲突
- 不应该使用using指示引入整个命名空间的标识符号
- 禁止用内联命名空间

### 接口类
- 接口是指满足特定条件的类
- 这些类以Interface为后缀(不强制)

### 函数使用引用参数
- 在C语言中,如果函数需要修改变量的值,参数必须为指针
- 在C++中,函数还可以声明为引用参数
- 定义引用参数可以防止出现(*pval)++这样丑陋的代码
- 引用参数对于拷贝构造函数这样的应用也是必需的
- 同时也更明确地不接受空指针
- 函数参数列表中,所有引用参数都必须是const

### 函数返回值后置语法
- 只有在常规写法(返回类型前置)不便于书写或不便于阅读时使用返回类型后置语法
- C++11引入了后置返回值的语法
- 后置返回类型是显式地指定Lambda表达式的返回值的唯一方式

### 流
- 流用来替代printf()和scanf()
- 有了流,在打印时不需要关心对象的类型
- 不用担心格式化字符串与参数列表不匹配
- 流的构造和析构函数会自动打开和关闭对应的文件
- 不要使用流,除非是日志接口需要
- 使用printf之类的代替

### 前置自增和自减
- 不考虑返回值的话,前置自增(++i)通常要比后置自增(i++)效率更高
- 因为后置自增(或自减)需要对表达式的值i进行一次拷贝
- 如果i是迭代器或其他非数值类型,拷贝的代价是比较大的
- 既然两种自增方式实现的功能一样,为什么不总是使用前置自增呢

### constexpr
- 在C++11里,用constexpr来定义真正的常量
- 或实现常量初始化
- 变量可以被声明成constexpr以表示它是真正意义上的常量
- 即在编译时和运行时都不变
- 函数或构造函数也可以被声明成constexpr
- 以用来定义constexpr变量

### Lambda表达式
- 适当使用lambda表达式
- 别用默认lambda捕获
- 所有捕获都要显式写出来
- Lambda表达式是创建匿名函数对象的一种简易途径
- 常用于把函数当参数传

### 命名
- 文件名:http_server_logs.h,http_server_logs.cc
- 类型性:MyExcitingClass
- 变量名:string table_name
- 函数名:AddTableEntry()
- 命名空间:websearch::index

## STL容器与算法

### 熟悉STL中17种容器及其背后对应的数据结构
- vector
- list
- deque
- stack
- queue
- priority_queue
- set
- multiset
- map
- multimap
- unordered_set
- unordered_multiset
- unordered_map
- unordered_multimap
- array
- forward_list
- string

### map和unordered_map的区别
- map背后是红黑树,unordered_map背后是哈希表
- map是key值有序的,unordered_map是key值无序的
- 两者内存消耗差不多,但是插入/查找/删除效率unordered_map是map的2到3倍
- unordered_map是通过链地址法解决冲突的
- std::map [] operator和insert的区别

### priority_queue优先队列的实现
- priority_queue优先队列,其底层是用堆来实现的
- 在优先队列中,队首元素一定是当前队列中优先级最高的那一个
- 在优先队列中,没有front()函数与back()函数
- 而只能通过top()函数来访问队首元素
- 也就是优先级最高的元素
- priority_queue默认为大顶堆

### 迭代器和迭代器失效iterator
- 为了提高C++编程的效率,STL中提供了许多容器
- 有些容器例如vector可以通过脚标索引的方式访问容器里面的数据
- 但是大部分的容器不能使用这种方式
- STL中每种容器在实现的时候设计了一个内嵌的iterator类
- 不同的容器有自己专属的迭代器
- 使用迭代器来访问容器中的数据
- 通过迭代器,可以将容器和通用算法结合在一起
- 只要给予算法不同的迭代器,就可以对不同容器执行相同的操作
- 迭代器对指针的一些基本操作如*、->、++、==、!=、=进行了重载
- 使其具有了遍历复杂数据结构的能力
- 其遍历机制取决于所遍历的数据结构
- 所有迭代的使用和指针的使用非常相似
- 通过begin,end函数获取容器的头部和尾部迭代器
- end迭代器不包含在容器之内
- 当begin和end返回的迭代器相同时表示容器为空
- 容器的插入insert和erase操作可能导致迭代器失效
- 对于erase操作不要使用操作之前的迭代器
- 因为erase的那个迭代器一定失效了
- 正确的做法是返回删除操作时候的那个迭代器

### 容器的线程安全性Thread safety
- STL为了效率,没有给所有操作加锁
- 不同线程同时读同一容器对象没关系
- 不同线程同时写不同的容器对象没关系
- 但不能同时又读又写同一容器对象的
- 因此,多线程要同时读写时,还是要自己加锁

### STL排序
- sort,快排加插入排序
- stable_sort,稳定排序
- sort_heap,堆排序
- list.sort,链表归并排序

### STL容器的内存管理方式
- 内存分配器
- 内存池
- 内存释放

### vector和map的内存释放
- vector的内存释放
- map的内存释放
- 容器删除数据的时候注意迭代器失效
- vector和map正确的内存释放

## 编译、链接与调试

### 排查编译问题常用工具

#### gcc/g++的区别和使用
- 后缀为.c的,gcc把它当作是C程序,而g++当作是c++程序
- 后缀为.cpp的,两者都会认为是c++程序
- 对于C代码,编译和链接都使用gcc
- 对于C++代码,编译时可以使用gcc/g++,gcc实际也是调用g++
- 链接时gcc不能自动和C++使用库链接,因此要使用g++或者gcc -lstdc++

#### 常见gcc编译链接选项
- -c 只编译并生成目标文件
- -g 生成调试信息,gdb可以利用该调试信息
- -o 指定生成的输出文件,可执行程序或者动态链接库文件名
- -I 编译时添加头文件路径
- -L 链接时添加库文件路径
- -D 定义宏,常用于开关控制代码
- -shared 用于生成共享库.so
- -Wall 显示所有警告信息,-w不生成任何警告信息
- -O0选项不进行任何优化,debug会产出和程序预期的结果
- O1优化会消耗少多的编译时间,它主要对代码的分支,常量以及表达式等进行优化
- O2会尝试更多的寄存器级的优化以及指令级的优化
- 它会在编译期间占用更多的内存和编译时间
- 通常情况下线上代码至少加上O2优化选项
- -fPIC 位置无关选项,生成动态库时使用
- 实现真正意义上的多进程共享的.so库
- -Wl选项告诉编译器将后面的参数传递给链接器
- -Wl,-Bstatic,指明后面是链接今静态库
- -Wl,-Bdynamic,指明后面是链接动态库

#### 编译时添加头文件依赖路径
- -include用来包含头文件
- 但一般情况下包含头文件都在源码里用#include xxxxxx实现
- -include参数很少用
- -I参数是用来指定头文件目录
- /usr/include目录一般是不用指定的,gcc知道去那里找
- 但是如果头文件不在/usr/include里我们就要用-I参数指定了
- 比如头文件放在/myinclude目录里,那编译命令行就要加上-I /myinclude参数了
- 如果不加你会得到一个"xxxx.h: No such file or directory"的错误
- -I参数可以用相对路径,比如头文件在当前目录,可以用-I.来指定

### 排查链接问题常用工具
- 查看ld链接器的搜索顺序 ld --verbose | grep SEARCH
- 链接时指定链接目录 -L/dir
- -Wl,-Bstatic,指明后面是链接今静态库
- -Wl,-Bdynamic,指明后面是链接动态库
- 运行时找不到动态库so文件,设置LD_LIBRARY_PATH,添加依赖so文件所在路径
- 链接完成后使用ldd查看动态库依赖关系
- 如果依赖的某个库找不到,通过这个命令可以迅速定位问题所在
- ldd -r,帮助检查是否存在未定义的符号undefine symbol,so库链接状态和错误信息

### gdb调试基本使用

#### 对C/C++程序的调试
- 需要在编译前就加上-g选项
- $gdb <programe>
- 设置参数:set args 可指定运行时参数
- (如:set args 10 20 30 40 50)

#### 查看源代码
- list:简记为l,其作用就是列出程序的源代码,默认每次显示10行
- list 行号:将显示当前文件以"行号"为中心的前后10行代码
- list 函数名:将显示"函数名"所在函数的源代码
- list:不带参数,将接着上一次list命令的,输出下边的内容

#### 设置断点和关闭断点
- break n(简写b n):在第n行处设置断点
- break func(简写b func):在函数func()的入口处设置断点
- info b(info breakpoints):显示当前程序的断点设置情况
- delete 断点号n:删除第n个断点
- disable 断点号n:暂停第n个断点
- clear 行号n:清除第n行的断点

#### 程序调试运行
- run:简记为r,其作用是运行程序
- 当遇到断点后,程序会在断点处停止运行
- 等待用户输入下一步的命令
- continue(简写c):继续执行,到下一个断点处(或运行结束)
- next:(简写n),单步跟踪程序
- 当遇到函数调用时,也不进入此函数体
- 此命令同step的主要区别是
- step遇到用户自定义的函数,将步进到函数中去运行
- 而next则直接调用函数,不会进入到函数体内
- step(简写s):单步调试如果有函数调用,则进入函数
- 与命令n不同,n是不进入调用的函数的
- until:当你厌倦了在一个循环体内单步跟踪时
- 这个命令可以运行程序直到退出循环体
- until+行号:运行至某行,不仅仅用来跳出循环
- finish:运行程序,直到当前函数完成返回
- 并打印函数返回时的堆栈地址和返回值及参数值等信息
- call 函数(参数):调用程序中可见的函数,并传递"参数"
- quit:简记为q,退出gdb

#### 打印程序运行的调试信息
- print 表达式:简记为p
- 其中"表达式"可以是任何当前正在被测试程序的有效表达式
- 比如当前正在调试C语言的程序
- 那么"表达式"可以是任何C语言的有效表达式
- 包括数字,变量甚至是函数调用
- print a:将显示整数a的值
- print name:将显示字符串name的值
- print gdb_test(22):将以整数22作为参数调用gdb_test()函数
- print gdb_test(a):将以变量a作为参数调用gdb_test()函数
- 扩展info locals:显示当前堆栈页的所有变量

#### 查询运行信息
- where/bt:当前运行的堆栈列表
- bt backtrace 显示当前调用堆栈
- up/down 改变堆栈显示的深度
- set args 参数:指定运行时的参数
- show args:查看设置好的参数
- info program:来查看程序的是否在运行,进程号,被暂停的原因

### gdb调试coredump问题
- Coredump叫做核心转储
- 它是进程运行时在突然崩溃的那一刻的一个内存快照
- 操作系统在程序发生异常而异常在进程内部又没有被捕获的情况下
- 会把进程此刻内存、寄存器状态、运行堆栈等信息转储保存在一个文件里
- 该文件也是二进制文件,可以使用gdb调试
- 虽然我们知道进程在coredump的时候会产生core文件
- 但是有时候却发现进程虽然core了,但是我们却找不到core文件
- 在ubuntu系统中需要进行设置
- ulimit -c 可以设置core文件的大小
- 如果这个值为0.则不会产生core文件
- 这个值太小,则core文件也不会产生
- 因为core文件一般都比较大
- 使用**ulimit -c unlimited**来设置无限大
- 则任意情况下都会产生core文件
- gdb打开core文件时,有显示没有调试信息
- 因为之前编译的时候没有带上-g选项
- 没有调试信息是正常的
- 实际上它也不影响调试core文件
- 因为调试core文件时,符号信息都来自符号表
- 用不到调试信息
- 如下为加上调试信息的效果
- 调试步骤:
- ＄gdb program core_file 进入
- $ bt或者where # 查看coredump位置
- 当程序带有调试信息的情况下
- 我们实际上是可以看到core的地方和代码行的匹配位置
- 但往往正常发布环境是不会带上调试信息的
- 因为调试信息通常会占用比较大的存储空间
- 一般都会在编译的时候把-g选项去掉
- 这种情况啊也是可以通过core_dump文件找到错误位置的
- 但这个过程比较复杂

### gdb调试线上死锁问题
- 如果你的程序是一个服务程序
- 那么你可以指定这个服务程序运行时的进程ID
- gdb会自动attach上去,并调试
- 对于服务进程,我们除了使用gdb调试之外
- 还可以使用pstack跟踪进程栈
- 这个命令在排查进程问题时非常有用
- 比如我们发现一个服务一直处于work状态
- (如假死状态,好似死循环)
- 使用这个命令就能轻松定位问题所在
- 可以在一段时间内,多执行几次pstack
- 若发现代码栈总是停在同一个位置
- 那个位置就需要重点关注
- 很可能就是出问题的地方
- gdb比pstack更加强大
- gdb可以随意进入进程、线程中改变程序的运行状态和查看程序的运行信息
- 思考:如何调试死锁?
- $gdb <program> <PID>
- $pstack pid

### undefined symbol问题解决步骤
- file 检查so或者可执行文件的架构
```
$ file _visp.so 
_visp.so: ELF 64-bit LSB pie executable, x86-64, version 1 (GNU/Linux), dynamically linked, BuildID[sha1]=6503ba6b7545e38e669ab9ed31f86449d8a5f78b, stripped
```
- ldd -r _visp.so 命令查看so库链接状态和错误信息
```
undefined symbol: __itt_api_version_ptr__3_0	(./_visp.so)
undefined symbol: __itt_id_create_ptr__3_0	(./_visp.so)
```
- c++filt symbol 定位错误在那个C++文件中
```
base) terse@ubuntu:~/code/terse-visp$ c++filt __itt_domain_create_ptr__3_0
__itt_domain_create_ptr__3_0
```
- 还可以使用grep -R __itt_domain_create_ptr__3_0 ./
最终发现这个符号来自XXX/opencv-3.4.6/build/share/OpenCV/3rdparty/libittnotify.a

- 通过nm命令也能看出该符号确实未定义
```
$ nm _visp.so | grep __itt_domain_create_ptr__3_0
      U __itt_domain_create_ptr__3_0
```

### pkg-config找第三方库的头文件和库文件
- pkg-config能方便使用第三方库和头文件和库文件
- 其运行原理
- 它首先根据PKG_CONFIG_PATH环境变量下寻找库对应的pc文件  
- 然后从pc文件中获取该库对应的头文件和库文件的位置信息
- 例如在项目中需要使用opencv库
- 该库包含的头文件和库文件比较多  
- 首先查看是否有对应的opencv.pc find /usr -name opencv.pc  
- 查看该路径是否包含在PKG_CONFIG_PATH  
- 使用pkg-config --cflags --libs opencv 查看库对应的头文件和库文件信息  
- pkg-config --modversion opencv 查看版本信息
参考链接：[https://blog.csdn.net/luotuo44/article/details/24836901](https://blog.csdn.net/luotuo44/article/details/24836901)

### cmake中的find_package
- find_package原理
- 首先明确一点,cmake本身不提供任何搜索库的便捷方法
- 所有搜索库并给变量赋值的操作必须由cmake代码完成
- 比如下面将要提到的FindXXX.cmake和XXXConfig.cmake
- 只不过,库的作者通常会提供这两个文件
- 以方便使用者调用
- find_package采用两种模式搜索库:

Module模式：搜索CMAKE_MODULE_PATH指定路径下的FindXXX.cmake文件，执行该文件从而找到XXX库。其中，具体查找库并给XXX_INCLUDE_DIRS和XXX_LIBRARIES两个变量赋值的操作由FindXXX.cmake模块完成。

Config模式：搜索XXX_DIR指定路径下的XXXConfig.cmake文件，执行该文件从而找到XXX库。其中具体查找库并给XXX_INCLUDE_DIRS和XXX_LIBRARIES两个变量赋值的操作由XXXConfig.cmake模块完成。

两种模式看起来似乎差不多，不过cmake默认采取Module模式，如果Module模式未找到库，才会采取Config模式。如果XXX_DIR路径下找不到XXXConfig.cmake文件，则会找/usr/local/lib/cmake/XXX/中的XXXConfig.cmake文件。总之，Config模式是一个备选策略。通常，库安装时会拷贝一份XXXConfig.cmake到系统目录中，因此在没有显式指定搜索路径时也可以顺利找到。

### ldd解决运行时问题
- 现象:
- error while loading shared libraries: libopencv_cudabgsegm.so.3.4: cannot open shared object file: No such file or directory  
- ldd ./xxx，发现库文件not found  

      libopencv_cudaobjdetect.so.3.4 => not found  
      libopencv_cudalegacy.so.3.4 => not found

- ld.so 动态共享库搜索顺序：  
- ELF可执行文件中动态段DT_RPATH指定；gcc加入链接参数"-Wl,-rpath"指定动态库搜索路径；  
- 环境变量LD_LIBRARY_PATH指定路径；  
- /etc/ld.so.cache中缓存的动态库路径。可以通过修改配置文件/etc/ld.so.conf 增删路径（修改后需要运行ldconfig命令）；  
- 默认的 /lib/;  
- 默认的 /usr/lib/  

- 解决办法：  
- 确认系统中是包含这个库文件的  
- pkg-config --libs opencv 查看opencv库的路径  
- export LD_LIBRARY_PATH=/usr/local/lib64，增加运行时加载路径  

 参考链接：[https://www.cnblogs.com/amyzhu/p/8871475.html](https://www.cnblogs.com/amyzhu/p/8871475.html)

### makefile和cmake的使用
- 跟我学些makefile
- CMake入门实战

## 性能分析与优化

### time

#### shell time
- time非常方便获取程序运行的时间
- 包括用户态时间user、内核态时间sys和实际运行的时间real
- 我们可以通过(user+sys)/real计算程序CPU占用率
- 判断程序时CPU密集型还是IO密集型程序

#### /usr/bin/time
- Linux中除了shell time,还有/usr/bin/time
- 它能获取程序运行更多的信息
- 通常带有-v参数

### top
- top是linux系统的任务管理器
- 它既能看系统所有任务信息
- 也能帮助查看单个进程资源使用情况
- 主要有以下几个功能:
- 查看系统任务信息
- 查看CPU使用情况
- 查看内存使用情况
- 查看单个进程资源使用情况
- 除此之外top还提供了一些交互命令

### perf

#### perf stat
- 做任何事都最好有条有理
- 老手往往能够做到不慌不忙,循序渐进
- 而新手则往往东一下,西一下,不知所措
- 面对一个问题程序,最好采用自顶向下的策略
- 先整体看看该程序运行时各种统计事件的大概
- 再针对某些方向深入细节
- 而不要一下子扎进琐碎细节,会一叶障目的
- 有些程序慢是因为计算量太大
- 其多数时间都应该在使用CPU进行计算
- 这叫做CPU bound型
- 有些程序慢是因为过多的IO
- 这种时候其CPU利用率应该不高
- 这叫做IO bound型
- 对于CPU bound程序的调优和IO bound的调优是不同的
- 如果您认同这些说法的话
- Perf stat应该是您最先使用的一个工具
- 它通过概括精简的方式提供被调试程序运行的整体情况和汇总数据

#### perf top
- Perf top用于实时显示当前系统的性能统计信息
- 该命令主要用来观察整个系统当前的状态
- 比如可以通过查看该命令的输出来查看当前系统最耗时的内核函数或某个用户进程

#### perf record/perf report
- 使用top和stat之后
- 这时对程序基本性能有了一个大致的了解
- 为了优化程序,便需要一些粒度更细的信息
- 比如说您已经断定目标程序计算量较大
- 也许是因为有些代码写的不够精简
- 那么面对长长的代码文件
- 究竟哪几行代码需要进一步修改呢
- 这便需要使用perf record记录单个函数级别的统计信息
- 并使用perf report来显示统计结果
- 您的调优应该将注意力集中到百分比高的热点代码片段上
- 假如一段代码只占用整个程序运行时间的0.1%
- 即使您将其优化到仅剩一条机器指令
- 恐怕也只能将整体的程序性能提高0.1%
- 俗话说,好钢用在刀刃上
- 要优化热点函数

### gprof
- gprof是GNU profiler工具
- 可以显示程序运行的"flat profile"
- 包括每个函数的调用次数
- 每个函数消耗的处理器时间
- 也可以显示"调用图"
- 包括函数的调用关系
- 每个函数调用花费了多少时间
- 还可以显示"注释的源代码"
- 是程序源代码的一个复本
- 标记有程序中每行代码的执行次数

### 内存问题与valgrind

#### 常见的内存问题
- 使用未初始化的变量
- 内存访问越界
- 内存覆盖
- 动态内存管理错误
- 内存泄露

#### valgrind内存检测
- valgrind是一个工具集
- 其中最有名的是Memcheck
- 它可以帮助我们检查程序中的内存问题
- 如内存泄漏、越界访问、重复释放等

### 自定义timer计时器
- 自己写一个计时器
- 计算局部函数的时间

## 常见问题与面试题

### 如何让类对象只在栈(堆)上分配空间?
- 只能在栈上建立对象
- 只能在堆上建立对象

### C++不可继承类的实现?
- 使用final关键字
- 使用私有构造函数
- 使用虚析构函数

### 如何定义和实现一个类的成员函数为回调函数?
- 友元函数/静态成员函数消除this指针的影响

### C++复制构造函数的参数为什么是引用类型?
- 编译时报错
- 需要首先调用该类的拷贝构造函数来初始化形参(局部对象)
- 造成无线循环递归

### C++全局对象如何在main函数之前构造和析构?
- 全局对象的构造和析构顺序
- 如何控制全局对象的构造和析构顺序

### 其它常见的问题
- vector和map的内存释放问题
- 容器删除数据的时候注意迭代器失效
- vector和map正确的内存释放
- C++的iostream的局限
- STL::list::sort链表归并排序

## 参考资料

- [了解google C++编码规范](https://zh-google-styleguide.readthedocs.io/en/latest/)
- [跟我学些makefile](https://github.com/wxquare/programming/blob/master/document/%E8%B7%9F%E6%88%91%E4%B8%80%E8%B5%B7%E5%86%99Makefile-%E9%99%88%E7%9A%93.pdf)
- [CMake入门实战](https://www.hahack.com/codes/cmake/)
- [gdb调试利器](https://linuxtools-rst.readthedocs.io/zh_CN/latest/tool/gdb.html)
- [陈皓专栏gdb调试系列](https://blog.csdn.net/haoel/article/details/2879)
- [gdb core_dump调试](https://blog.csdn.net/u014403008/article/details/54174109)
- [进程调试,死循环和死锁卡死](https://blog.csdn.net/guowenyan001/article/details/46238355)