---
title: 如何用好缓存？（redis）
categories:
- C/C++
---

## redis 使用场景分析
1. 分布式锁. 在分布式场景下，无法使用单机环境下的锁来对多个节点上的进程进行同步。可以使用 Redis 自带的 SETNX 命令实现分布式锁，除此之外，还可以使用官方提供的 RedLock 分布式锁实现
2. [[如何使用Redis实现微信步数排行榜？](https://www.cnblogs.com/zwwhnly/p/13041641.html)]

## redis 使用注意事项
1. 五种数据结构选择string/list/hashmap/set/zset
2. 容量与淘汰策略
3. 过期键的删除策略
4. 缓存持久化策略
5. 如何做的缓存初始，缓存预热？
6. 如何处理可能遇到的缓存雪崩，缓存穿透和热点数据缓存击穿问题？


## redis 5种数据类型和底层数据结构
[!redis数据类型](/images/redis-data-type)
- redis中zset的底层原理
- [面试：原来Redis的五种数据类型底层结构是这样的](https://my.oschina.net/ccwwlx/blog/3120883)
https://juejin.cn/post/6844904192042074126#heading-8


## redis 数据持久化
- AOF: AOF持久化(即Append Only File持久化)
- RDB: 是Redis DataBase缩写快照

## redis 怎么扩容扩容和收缩
- https://www.infoq.cn/article/uiqypvrtnq4buerrm3dc

## redis 过期键的删除策略
过期策略通常有以下三种：
- 定时过期：每个设置过期时间的key都需要创建一个定时器，到过期时间就会立即清除。该策略可以立即清除过期的数据，对内存很友好；但是会占用大量的CPU资源去处理过期的数据，从而影响缓存的响应时间和吞吐量。
- 惰性过期：只有当访问一个key时，才会判断该key是否已过期，过期则清除。该策略可以最大化地节省CPU资源，却对内存非常不友好。极端情况可能出现大量的过期key没有再次被访问，从而不会被清除，占用大量内存。
- 定期过期：每隔一定的时间，会扫描一定数量的数据库的expires字典中一定数量的key，并清除其中已过期的key。该策略是前两者的一个折中方案。通过调整定时扫描的时间间隔和每次扫描的限定耗时，可以在不同情况下使得CPU和内存资源达到最优的平衡效果。
(expires字典会保存所有设置了过期时间的key的过期时间数据，其中，key是指向键空间中的某个键的指针，value是该键的毫秒精度的UNIX时间戳表示的过期时间。键空间是指该Redis集群中保存的所有键。)

## redis 内存淘汰策略
- MySQL里有2000w数据，redis中只存20w的数据，如何保证redis中的数据都是热点数据
- https://juejin.cn/post/6844903927037558792

## redis 为什么使用单线程模型
- https://draveness.me/whys-the-design-redis-single-thread/

## 缓存异常与对应的解决办法
- 缓存雪崩
- 缓存穿透
- 缓存击穿
缓存预热
缓存降级
热点数据和冷数据
缓存热点key
- 使用redis计数限制mdb并发访问的次数
- 缓存穿透，缓存击穿，缓存雪崩解决方案分析，https://juejin.im/post/6844903651182542856

## redis 为什么这么快
- 1、完全基于内存，绝大部分请求是纯粹的内存操作，非常快速。数据存在内存中，类似于 HashMap，HashMap 的优势就是查找和操作的时间复杂度都是O(1)；
- 2、数据结构简单，对数据操作也简单，Redis 中的数据结构是专门进行设计的；
- 3、采用单线程，避免了不必要的上下文切换和竞争条件，也不存在多进程或者多线程导致的切换而消耗 CPU，不用去考虑各种锁的问题，不存在加锁释放锁操作，没有因为可能出现死锁而导致的性能消耗；
- 4、使用多路 I/O 复用模型，非阻塞 IO；
- 5、使用底层模型不同，它们之间底层实现方式以及与客户端之间通信的应用协议不一样，Redis 直接自己构建了 VM 机制 ，因为一般的系统调用系统函数的话，会浪费一定的时间去移动和请求；


## Redis实现分布式锁
- Redis为单进程单线程模式，采用队列模式将并发访问变成串行访问，且多客户端对Redis的连接并不存在竞争关系Redis中可以使用SETNX命令实现分布式锁。当且仅当 key 不存在，将 key 的值设为 value。 若给定的 key 已经存在，则 SETNX 不做任何动作SETNX 是『SET if Not eXists』(如果不存在，则 SET)的简写。返回值：设置成功，返回 1 。设置失败，返回 0

## 如何使用Redis实现微信步数排行榜？
- https://www.cnblogs.com/zwwhnly/p/13041641.html

## 如何将db里面的数据同步到redis中去，以减小数据库的压力
- 



## 缓存使用注意事项
1. 缓存使用注意事项


## 推荐阅读:
1. https://blog.csdn.net/ThinkWon/article/details/103522351
2. https://tech.meituan.com/2017/03/17/cache-about.html