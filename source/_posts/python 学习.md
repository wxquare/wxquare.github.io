---
title: python 日常学习
---


##1.环境
anaconda+sublime+Python的交互式环境

##2.python基础
###2.1数据类型
1. 整数
2. 浮点数，整数和浮点数在计算机内部存储方式是不同的，整数运算时精确的，而浮点数则可能包含四舍五入。
3. 字符串，使用``或者""括起来的任意文本，多行内容使用'''。  
4. 布尔值，只有True和False。逻辑运算and、or、not。
5. 空值，None。
6. 数据类型变换，int,float,str,bool
###2.2变量
1. python是动态类型语言，可以把任意数据类型赋值给变量，同一个变量可以反复赋值，而且可以是不同类型的变量。  
    a = 123 # a是整数  
    print(a)  
    a = 'ABC' # a变为字符串  
    print(a)  

###2.3常量
1. Python根本没有任何机制保证常量不会被改变，但是用全部大写的变量名表示常量是一个习惯上的用法。  
    PI = 3.14159265359   

###2.4内置数据类型
1. list，它是一种有序的集合，可以随时添加和删除其中的元素，元素的类型也可以不同，可以嵌套。
2. tuple，元组也是一种有序集合，但是一旦初始化，就不能更改。
3. dict，字典
4. set，集合

###2.5条件判断
1.if...elif...else,条件判断从上向下匹配，当满足条件时执行对应的快内语句，后续的elif和else都不会执行。

###2.6循环
1.for...in...
2.range
3.while..break

###2.7函数
- 定义函数时需要确定函数名和参数个数，函数参数非常灵活，默认参数，可变参数，关键字参数。
- 如果有必要，可以先对参数的数据类型做检查；
- 函数体内部可以用return随时返回函数结果；
- 函数执行完毕也没有return语句时，自动return None。
- 函数可以同时返回多个值，但其实就是一个tuple

###2.8高级特性
1. 切片操作
2. 迭代操作（对list、dict）
3. 列表生成式
4. 生成器
5. 迭代器


###2.9模板
1. 导入模板
2. 模板中函数和变量的作用域，通过"_"实现，类似_xxx和__xxx这样的函数或变量就是非公开的（private），不应该被直接引用，比如_abc，__abc等；private函数和变量不应该被直接引用，而不是“不能”被引用，python不限制private函数和变量，但是从变成习惯上不应该引用private函数䄦变量。


##3面向对象
1. 定义一个类
2. __init__函数
3. self参数
4. 访问限制，private,"_"
5. 继承和多态
6. 类属性和实例属性
7. 创建了一个class的实例后，可以给该实例绑定任何属性和方法，这就是动态语言的灵活性
8. Python内置的@property装饰器能把一个方法变成属性调用的
9. 多种继承

##4.错误处理、调试和测试
1. 返回错误码
2. 抛出错误，raise 
2. try...except...finally
3. 日志logging记录错误堆栈
4. 断言，assert
5. python调试器pdb



##5.python日志模块
日志器（logger）是入口，真正干活儿的是处理器（handler），处理器（handler）还可以通过过滤器（filter）和格式器（formatter）对要输出的日志内容做过滤和格式化等处理操作。


##6.数据库操作pymysql


##7.python函数式编程
1. 函数赋值给变量
2. 函数传入函数，高阶函数
3. map、reduce、filter、sorted
4. 返回一个函数、闭包
5. 匿名函数，lambda
6. 装饰器decorator，运行期动态增加函数功能
7. 偏函数



1.python日志处理
https://www.cnblogs.com/yyds/p/6901864.html
https://blog.csdn.net/yohohaha/article/details/77777864


2.python 对json的处理
https://www.cnblogs.com/loleina/p/5623968.html


3.python web.py
https://github.com/webpy/webpy
https://www.cnblogs.com/LD-linux/p/4089205.html
https://www.cnblogs.com/xiaoleiel/p/8301442.html
https://blog.csdn.net/five3/article/details/7732832/
https://www.jianshu.com/p/260fbb89d3a3
http://webpy.org/tutorial3.zh-cn

4.python 中的subprocess
https://blog.csdn.net/longshenlmj/article/details/45174363


5.python gprof2dot
https://github.com/jrfonseca/gprof2dot


6.python argparse 参数解析
https://www.jianshu.com/p/fef2d215b91d
https://www.cnblogs.com/yymn/p/8056487.html

7.python性能分析指南
https://www.imooc.com/article/4170


8. python环境
root用户：  python2.7、pip
个人用户 ： anaconda、phython3.6


如何优化python代码的效率？

## 4.效率优化技术
有些技术和编程方法可以我们大的发挥 Python 和 Numpy 的威力。 我们仅仅提一下相关的你可以接查找更多细信息。我们 的的一点是先用简单的方式实现你的算法结果正确当结 果正确后再使用上的提到的方法找到程序的瓶来优化它。

- 尽量不要使用循环尤其双层三层循环它们天生就是常慢的。
- 算法中尽量使用向操作因为 Numpy 和 OpenCV 对向操作了优化。
- 利用缓存一致性。
- 没有必的就不复制数组。使用图来代替复制。数组复制是常浪源的。
- 就算了上优化如果你的程序是很慢或者大的不可免的 你应尝使用其他的包比如 Cython来加你的程序。
