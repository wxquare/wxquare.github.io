---
title: Python 基础
categories:
- Python
---




## 一、基础

### 1.1 数据类型
1. 整数
2. 浮点数，整数和浮点数在计算机内部存储方式是不同的，整数运算时精确的，而浮点数则可能包含四舍五入。
3. 字符串，使用``或者""括起来的任意文本，多行内容使用'''。  
4. **布尔值，只有True和False**。逻辑运算and、or、not。
5. **空值，None**。
6. 数据类型变换，int,float,str,bool  

### 1.2 定义变量、全局变量  
1. python中的变量在使用之前必须给他赋值
2. python是**动态类型语言，可以把任意数据类型赋值给变量，同一个变量可以反复赋值，而且可以是不同类型的变量**。  
    a = 123 # a是整数  
    print(a)  
    a = 'ABC' # a变为字符串  
    print(a)  

### 1.3 常量
1. Python根本没有任何机制保证常量不会被改变，但是用全部大写的变量名表示常量是一个习惯上的用法。  
    PI = 3.14159265359

### 1.4 简单语法   
1. 条件判断 if...elif...else,条件判断从上向下匹配，当满足条件时执行对应的快内语句，后续的elif和else都不会执行。
2. 循环
	- for...in...
	- range
	- while..break


### 1.5. **除法运算**
- 在python2中数据相除，向下取整，例如3/2结果为0，在python3中相除为3/2为1.5,3//2为1,round(3/2)=2,floor(3/2)=1。

### 1.6 注释
- 单行注释用#
- 多行注释用'''或者"""


### 1.7 指定python解释器
```
	#！/usr/local/env python  
	# -*- coding： utf-8 -*-  
```

### 1.8 单引号和双引号相互转义
- "let's go"
- '"hello,world"'
- '''保留字符串格式

### 1.9 字符串格式化
- %
- format



## 二、数据结构
1. 字符串str
2. list
3. tuple
4. dict
5. set，并集、交集、子集
6. 堆
7. 双端队列，collections中包含几个collection类型


## 三、函数
1. 函数定义非常灵活，不需要指定类型，支持默认参数，可变参数，关键字参数，可变参数（*），不定关键字参数（**）
2. 函数内部如果有必要，可以先对参数的数据类型做检查；
3. 函数体内部可以用return随时返回函数结果；函数执行完毕也没有return语句时，自动return None，函数可以同时返回多个值，但其实就是一个tuple
4. 函数传入函数，高阶函数
5. 匿名函数，lambda
6. 装饰器decorator，运行期动态增加函数功能
7. 偏函数
8. 函数注释，文档
9. 传参，拷贝和引用，是否改变参数的值
10. 作用域（函数局部变量和如何使用全局变量global）
11. 函数式编程（map、filter、reduce）


## 四. 异常和错误的处理
1. 返回错误码
2. 抛出错误，raise 
2. try...except...else...finally
3. 日志logging记录错误堆栈
4. 断言，assert
5. python调试器pdb
6. 异常类型，自定义异常，抛出异常，捕获异常


## 五、高级特性
1. 迭代操作（对list、dict）
2. 列表推导
4. 迭代器（next）
5. 生成器（yield）
6. 装饰器


## 六、对象抽象
1. 封装、继承、多态
2. 类的定义
3. 私有属性和方法（"__xxx"）
4. 继承和指定超类（父类）
5. 内置函数issubclass、isinstance
6. 指定多个超类（多重继承），当通多个超类以不同的方式实现了同一个方法时，必须小心排列这些类，因为位于前面的类的方法将覆盖位于后面类的方法。
7. python接口？鸭子类型，hasattr检查是否有所需的方法。
8. 通过装饰器支持抽象基类
9. 构造函数（__init__(self)）
10. 继承，重写构造函数，super.__init__,调用超类（父类）的构造函数
11. 静态方法和类方法
12. 访问限制，private,"_"
13. 创建了一个class的实例后，可以给该实例绑定任何属性和方法，这就是动态语言的灵活性
14. Python内置的@property装饰器能把一个方法变成属性调用的


## 七、包和模板
1. 包为什么必须包含__init__函数
2. help 帮助查询模板和函数的使用方法
3. sys:解释器相关的变量和函数
4. time：asctime,strptime
5. datetime、timeit
6. random、urandom产生随机数
7. re正则表达式
8. argparse 命令行参数解析
9. logging
10. timeit、profile、trace
11. 模板中函数和变量的作用域，通过"_"实现，类似_xxx和__xxx这样的函数或变量就是非公开的（private），不应该被直接引用，比如_abc，__abc等；private函数和变量不应该被直接引用，而不是“不能”被引用，python不限制private函数和变量，但是从变成习惯上不应该引用private函数䄦变量。
12. import module,然后使用module.function调用函数
13. from module import function，直接调用函数


## 八、文件与流
1. 打开文件，文件打开的类型、文件简单读写，open，write和read。
2. 读写行。readline、readlines
3. 文件关闭close。（内容可能被缓冲了，没有写入磁盘，close或者flush）
4. 文件写的时候不需要判断是否存在文件，读的时候需要判断。
5. 多使用，with open(filename) as f 结构


## 九、Python优化
有些技术和编程方法可以我们大的发挥 Python 和 Numpy 的威力。 我们仅仅提一下相关的你可以接查找更多细信息。我们 的的一点是先用简单的方式实现你的算法结果正确当结 果正确后再使用上的提到的方法找到程序的瓶来优化它。

- 尽量不要使用循环尤其双层三层循环它们天生就是常慢的。
- 算法中尽量使用向操作因为 Numpy 和 OpenCV 对向操作了优化。
- 利用缓存一致性。
- 没有必的就不复制数组。使用图来代替复制。数组复制是常浪源的。
- 就算了上优化如果你的程序是很慢或者大的不可免的 你应尝使用其他的包比如 Cython来加你的程序。



参考：
- 1.python日志处理
- https://www.cnblogs.com/yyds/p/6901864.html
- https://blog.csdn.net/yohohaha/article/details/77777864

- 2.python 对json的处理
- https://www.cnblogs.com/loleina/p/5623968.html

- 3.python web.py
- https://github.com/webpy/webpy
- https://www.cnblogs.com/LD-linux/p/4089205.html
- https://www.cnblogs.com/xiaoleiel/p/8301442.html
- https://blog.csdn.net/five3/article/details/7732832/
- https://www.jianshu.com/p/260fbb89d3a3
- http://webpy.org/tutorial3.zh-cn

- 4.python 中的subprocess
- https://blog.csdn.net/longshenlmj/article/details/45174363

- 5.python gprof2dot
- https://github.com/jrfonseca/gprof2dot


- 6.python argparse 参数解析
- https://www.jianshu.com/p/fef2d215b91d
- https://www.cnblogs.com/yymn/p/8056487.html

- 7.python性能分析指南
- https://www.imooc.com/article/4170
- 































