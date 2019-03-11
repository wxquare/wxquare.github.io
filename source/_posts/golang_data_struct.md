---
title: golang常用数据结构和容器
---

### 1.字符串string
1. 基本数组类型s := "hello,world"
2. 一旦初始化后不允许修改字符串的内容
3. 常用函数s1+s2,len(s1)等
4. <font color=red>字符串与数值类型的不能强制转化，要使用strconv包中的函数</font>
5. 标准库strings提供了许多字符串操作的函数,例如Split、HasPrefix,Trim。

### 2.数组array: [3]int{1,2,3}
1. <font color=red>**数组是值类型**</font>，数组传参发生拷贝
2. 定长
3. 数组的创建、初始化、访问和遍历range，len(arr)求数组的长度
  
### 3.数组切片slice: make([]int,len,cap)
1. <font color=red>**slice是引用类型**</font>
2. 变长，用容量和长度的区别，分别使用cap和len函数获取
3. 内存结构：指针、cap、size共24字节
4. 常用函数，append，cap，len
5. 切片动态扩容，拷贝

### 4.存储kv的哈希表map：make(map[string]int,5) 
1.  map的创建，为了避免频繁的扩容和迁移，创建map时应指定适当的大小
2.  无序
3.  赋值，相同键值会覆盖
4.  遍历，range
5.  [如何实现顺序遍历？](https://blog.csdn.net/slvher/article/details/44779081)
6.  [内部hashmap的实现原理](https://ninokop.github.io/2017/10/24/Go-Hashmap%E5%86%85%E5%AD%98%E5%B8%83%E5%B1%80%E5%92%8C%E5%AE%9E%E7%8E%B0/)。内部结构（bucket），扩容与迁移，删除。 
7.  如何保证map的协程安全性？[sync.map](https://colobu.com/2017/07/11/dive-into-sync-Map/)? 


### 5.集合set
1. golang中本身没有提供set，但可以通过map自己实现
2. 利用map键值不可重复的特性实现set，value为空结构体。 map[interface{}]struct{} 
3. [如何自己实现set？](https://studygolang.com/articles/11179)

  
### 6.容器container/heap、list、ring
1. heap与优先队列，最小堆
2. 链表list，双向列表
3. 循环队列ring
4. <font color=red>golang没有提供stack，可自己实现</font>
5. <font color=red>golang没有提供queue，但可以通过channel替换或者自己实现</font>


### 7.延伸问题：
#### 1.如何比较struct/slice/map?
- struct没有slice和map类型时可直接判断
- slice和map本身不可比较，需要使用reflect.DeepEqual()。
- truct中包含slice和map等字段时，也要使用reflect.DeepEqual().
- [https://stackoverflow.com/questions/24534072/how-to-compare-struct-slice-map-are-equal](https://stackoverflow.com/questions/24534072/how-to-compare-struct-slice-map-are-equal)
- [https://studygolang.com/articles/11342](https://studygolang.com/articles/11342)