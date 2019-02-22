#golang内存管理和垃圾回收

##1.为什么需要自主管理内存？
- 完成类似预分配、内存池等操作，以避开系统调用带来的性能问题
- 更好的配合垃圾回收

##1.内存分配器解决哪些问题？
- 内存碎片  
- 多核处理器能够扩展  

##2.如何分配定长记录？

##3.如何分配变长的记录？
取整导致“内存碎片”问题



##4.如何处理小对象<=32k？？
thread cache->central free list->central page allocator

##5.如何处理大对象>32k？？
直接从central page heap中分配。

##5.Central page Heap以span为管理对象。
span list。



##4.大对象如何分配？
分配对象时，大的对象直接分配 Span，小的对象从Span中分配


TCMalloc 是 Google 开发的内存分配器，在不少项目中都有使用，例如在 Golang 中就使用了类似的算法进行内存分配。它具有现代化内存分配器的基本特征：对抗内存碎片、在多核处理器能够 scale。据称，它的内存分配速度是 glibc2.3 中实现的 malloc的数倍。




##7.垃圾回收
- 标记清扫算法：标记阶段和清扫阶段  
- 精确垃圾回收  

golang使用标记清扫的垃圾回收算法，标记位图是非侵入式的

golang实现了精确的垃圾回收，在精确的垃圾回收中，先通过扫描整个内存块区域，定位对象的类型信息，得到该类型信息，得到其中的gc域。然后得到该类型中的垃圾回收的指令码，通过一个状态机解释这段指令码来执行特定类型的垃圾回收工作。

对于堆中的任意地址的对象，先通过它所在的内存页找到它所属的Mspan，然后通过MSpan中的类型信息找到它的类型信息。

golang并行垃圾回收

垃圾回收的触发是由一个gcpercent的变量控制的，当新分配的内存占已在使用中的内存的比例超过gcprecent时就会触发。比如，gcpercent=100，当前使用了4M的内存，那么当内存分配到达8M时就会再次gc。如果回收完毕后，内存的使用量为5M，那么下次回收的时机则是内存分配达到10M的时候。也就是说，并不是内存分配越多，垃圾回收频率越高，这个算法使得垃圾回收的频率比较稳定，适合应用的场景。



gcpercent的值是通过环境变量GOGC获取的，如果不设置这个环境变量，默认值是100。如果将它设置成off，则是关闭垃圾回收。



1.Go的垃圾回收机制在实践中有哪些需要注意的地方？？？
- 尽量不要创建大量的对象，也尽量不要频繁的创建对象  
- gc执行的时间跟数量是相关的
- 1、尽早的用memprof、cpuprof、GCTRACE来观察程序。 
- 2、关注请求处理时间，特别是开发新功能的时候，有助于发现设计上的问题。  
- 3、尽量避免频繁创建对象(&abc{}、new(abc{})、make())，在频繁调用的地方可以做对象重用。    
- 4、尽量不要用go管理大量对象，内存数据库可以完全用c实现好通过cgo来调用。

作者：达达
链接：https://www.zhihu.com/question/21615032/answer/18781477
来源：知乎
著作权归作者所有。商业转载请联系作者获得授权，非商业转载请注明出处。



Golang程序调优：memprof、cpuprof。



【1】.https://zhuanlan.zhihu.com/p/29216091  
【2】.http://goog-perftools.sourceforge.net/doc/tcmalloc.html  
【3】.https://cloud.tencent.com/developer/article/1072602
[4].http://www.opscoder.info/golang_gc.html

https://studygolang.com/articles/9389（现代垃圾回收）
https://segmentfault.com/a/1190000016828394

https://studygolang.com/articles/14497



