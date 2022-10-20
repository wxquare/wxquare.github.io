---
title: Middleware - Redis
categories:
- C/C++
---

## 前言
1. redis有哪些使用场景？
2. redis五种数据结构选择以及其底层实现原理？
3. 如何处理可能遇到的缓存雪崩，缓存穿透和热点数据缓存击穿问题？是否需要缓存预热
4. 如何考虑缓存和数据库一致性的问题？
5. redis 超过使用容量时的内存淘汰策略
6. redis 过期键的删除策略
7. redis 数据持久化是怎么做的？RDB和AOF机制？
8. redis 分布式架构，codis，rdis cluster？


## redis 使用场景
1. 缓存数据（db，service) 的数据，提高访问效率
2. incr + expire 实现滑动窗口计数器限流
```go
      package main

      import (
        "fmt"
        "time"

        "github.com/go-redis/redis"
      )

      var pool redis.UniversalClient

      func simpleIncr(key string) {
        ok, err := pool.SetNX(key, 1, time.Duration(10)*time.Second).Result()
        fmt.Printf("%+v,%+v\n", ok, err)
        val, err := pool.Incr(key).Result()
        fmt.Printf("%+v,%+v\n", val, err)
        //手动过期
        ok, err = pool.Expire(key, time.Duration(60)*time.Second).Result()
        fmt.Printf("%+v,%+v\n", ok, err)
      }

      func simpleUserRateLimit(userID string) bool {
        /*
          限制用户1秒钟最多访问10次，对于异常用户冻结2分钟
        */
        key := userID
        var maxRequestPerSecond int64 = 10
        var frozenTime int = 2 * 60
        ok, _ := pool.SetNX(key, 1, time.Duration(1)*time.Second).Result()
        if ok {
          return true
        }
        val, _ := pool.Incr(key).Result()
        if val <= maxRequestPerSecond {
          return true
        }
        if val > maxRequestPerSecond {
          pool.Expire(key, time.Duration(frozenTime)*time.Second).Result()
          return false
        }
        return false
      }

      func main() {
        pool = redis.NewUniversalClient(&redis.UniversalOptions{
          Addrs:        []string{"127.0.0.1:6379"},
          MaxRetries:   3,
          DialTimeout:  time.Duration(300) * time.Millisecond,
          ReadTimeout:  time.Duration(300) * time.Millisecond,
          WriteTimeout: time.Duration(300) * time.Millisecond,
          PoolSize:     10,
          IdleTimeout:  time.Duration(10) * time.Second,
        })
        _, err := pool.Ping().Result()
        if err != nil {
          fmt.Printf("%+v", err)
        }

        for i := 0; i < 20; i++ {
          islimited := simpleUserRateLimit("12345678")
          fmt.Println(islimited)
        }

      }

```
3. 延时队列
   - 使用 ZSET+ 定时轮询的方式实现延时队列机制，任务集合记为 taskGroupKey
   - 生成任务以 当前时间戳 与 延时时间 相加后得到任务真正的触发时间，记为 time1，任务的 uuid 即为 taskid，当前时间戳记为 curTime
   - 使用 ZADD taskGroupKey time1 taskid 将任务写入 ZSET
   - 主逻辑不断以轮询方式 ZRANGE taskGroupKey curTime MAXTIME withscores 获取 [curTime,MAXTIME) 之间的任务，记为已经到期的延时任务（集）
   - 处理延时任务，处理完成后删除即可
   - 保存当前时间戳 curTime，作为下一次轮询时的 ZRANGE 指令的范围起点
   - https://github.com/bitleak/lmstfy
   
4. 消息队列
   - redis 支持 List 数据结构，有时也会充当消息队列。使用生产者：LPUSH；消费者：RBPOP 或 RPOP 模拟队列
5. incr计数器
6. 分布式锁：https://juejin.cn/post/6936956908007850014
7. 基于redis的分布式限流：https://pandaychen.github.io/2020/09/21/A-DISTRIBUTE-GOREDIS-RATELIMITER-ANALYSIS/
8. bloomfilter: https://juejin.cn/post/6844903862072000526



## redis 5种数据类型和底层数据结构
- [面试：原来Redis的五种数据类型底层结构是这样的]https://juejin.cn/post/6844904192042074126#heading-8
- [最详细的Redis五种数据结构详解](https://juejin.cn/post/6844904192042074126)


## 计算所需的缓存的容量，当容量超过限制时的淘汰策略
```
- noeviction(默认策略)：对于写请求不再提供服务，直接返回错误（DEL请求和部分特殊请求除外）
- allkeys-lru：从所有key中使用LRU算法进行淘汰
- volatile-lru：从设置了过期时间的key中使用LRU算法进行淘汰
- allkeys-random：从所有key中随机淘汰数据
- volatile-random：从设置了过期时间的key中随机淘汰
- volatile-ttl：在设置了过期时间的key中，根据key的过期时间进行淘汰，越早过期的越优先被淘汰
LFU算法是Redis4.0里面新加的一种淘汰策略。它的全称是Least Frequently Used

```
[redis 内存淘汰策略解析](https://juejin.cn/post/6844903927037558792)

## redis 过期键的删除策略
过期策略通常有以下三种：
- 定时过期：每个设置过期时间的key都需要创建一个定时器，到过期时间就会立即清除。该策略可以立即清除过期的数据，对内存很友好；但是会占用大量的CPU资源去处理过期的数据，从而影响缓存的响应时间和吞吐量。
- 惰性过期：只有当访问一个key时，才会判断该key是否已过期，过期则清除。该策略可以最大化地节省CPU资源，却对内存非常不友好。极端情况可能出现大量的过期key没有再次被访问，从而不会被清除，占用大量内存。
- 定期过期：每隔一定的时间，会扫描一定数量的数据库的expires字典中一定数量的key，并清除其中已过期的key。该策略是前两者的一个折中方案。通过调整定时扫描的时间间隔和每次扫描的限定耗时，可以在不同情况下使得CPU和内存资源达到最优的平衡效果。
(expires字典会保存所有设置了过期时间的key的过期时间数据，其中，key是指向键空间中的某个键的指针，value是该键的毫秒精度的UNIX时间戳表示的过期时间。键空间是指该Redis集群中保存的所有键。)


## redis 两种数据持久化的原理以及优缺点
- AOF: AOF持久化(即Append Only File持久化)，文本日志，记录增删改
- RDB: 是Redis DataBase缩写快照，紧凑的二进制数据
- [Redis持久化是如何做的？RDB和AOF对比分析] (http://kaito-kidd.com/2020/06/29/redis-persistence-rdb-aof/)


## 怎么考虑缓存和db数据一致性的问题
- 当使用redis缓存db数据时，db数据会发生update，如何考虑redis和db数据的一致性问题呢？
- 通常来说，对于流量较小的业务来说，可以设置较小的expire time,可以将redis和db的不一致的时间控制在一定的范围内部
- 对于缓存和db一致性要求较高的场合，通常采用的是先更新db，再删除或者更新redis，考虑到并发性和两个操作的原子性（删除或者更新可能会失败），可以增加重试机制（双删除），如果考虑主从延时，可以引入mq做延时双删
- http://kaito-kidd.com/2021/09/08/how-to-keep-cache-and-consistency-of-db/

## redis 怎么扩容扩容和收缩
- https://www.infoq.cn/article/uiqypvrtnq4buerrm3dc


## redis 为什么使用单线程模型
- https://draveness.me/whys-the-design-redis-single-thread/

## 缓存异常与对应的解决办法
- 缓存雪崩问题，大面积键失效或删除
- 缓存穿透问题，不存在key的攻击行为
- 热点数据缓存击穿，热门key失效
- 是否需要缓存预热
- 缓存穿透，缓存击穿，缓存雪崩解决方案分析，https://juejin.im/post/6844903651182542856

## redis 为什么这么快
- 1、完全基于内存，绝大部分请求是纯粹的内存操作，非常快速。数据存在内存中，类似于 HashMap，HashMap 的优势就是查找和操作的时间复杂度都是O(1)；
- 2、数据结构简单，对数据操作也简单，Redis 中的数据结构是专门进行设计的；
- 3、采用单线程，避免了不必要的上下文切换和竞争条件，也不存在多进程或者多线程导致的切换而消耗 CPU，不用去考虑各种锁的问题，不存在加锁释放锁操作，没有因为可能出现死锁而导致的性能消耗；
- 4、使用多路 I/O 复用模型，非阻塞 IO；
- 5、使用底层模型不同，它们之间底层实现方式以及与客户端之间通信的应用协议不一样，Redis 直接自己构建了 VM 机制 ，因为一般的系统调用系统函数的话，会浪费一定的时间去移动和请求；


## Redis实现分布式锁
- Redis为单进程单线程模式，采用队列模式将并发访问变成串行访问，且多客户端对Redis的连接并不存在竞争关系Redis中可以使用SETNX命令实现分布式锁。当且仅当 key 不存在，将 key 的值设为 value。 若给定的 key 已经存在，则 SETNX 不做任何动作SETNX 是『SET if Not eXists』(如果不存在，则 SET)的简写。返回值：设置成功，返回 1 。设置失败，返回 0



## 分布式redis
1. 单机版，并发访问有限，存储有限，单点故障。
2. 数据持久化
4. 主从复制。主库（写）同步到从库（读）的延时会造成数据的不一致；主从模式不具备自动容错，需要大量的人工操作
5. 哨兵模式sentinel。在主从的基础上，实现哨兵模式就是为了监控主从的运行状况，对主从的健壮进行监控，就好像哨兵一样，只要有异常就发出警告，对异常状况进行处理。当master出现故障时，哨兵通过raft选举，leader哨兵选择优先级最高的slave作为新的master，其它slaver从新的master同步数据。哨兵解决和主从不能自动故障恢复的问题，但是同时也存在难以扩容以及单机存储、读写能力受限的问题，并且集群之前都是一台redis都是全量的数据，这样所有的redis都冗余一份，就会大大消耗内存空间
6. **codis**: https://github.com/CodisLabs/codis
7. **redis cluster集群模式**：集群模式时一个无中心的架构模式，将数据进行分片，分不到对应的槽中，每个节点存储不同的数据内容，通过路由能够找到对应的节点负责存储的槽，能够实现高效率的查询。并且集群模式增加了横向和纵向的扩展能力，实现节点加入和收缩，集群模式时哨兵的升级版，哨兵的优点集群都有
8. [redis 分布式架构演进](https://blog.csdn.net/QQ1006207580/article/details/103243281)
9. [Redis集群化方案对比：Codis、Twemproxy、Redis Cluster](http://kaito-kidd.com/2020/07/07/redis-cluster-codis-twemproxy/)




## 推荐阅读:
1. https://blog.csdn.net/ThinkWon/article/details/103522351
2. https://tech.meituan.com/2017/03/17/cache-about.html
3. [一不小心肝出了4W字的Redis面试教程](https://juejin.cn/post/6868409018151337991)
4. [你的 Redis 为什么变慢了？](https://cloud.tencent.com/developer/article/1724076)
5. redis 为什么快？
6. redis数据结构，详细说说一种。字典如何实现的？
7. redis内存淘汰策略？
8. redis dbindex. https://blog.csdn.net/lsm135/article/details/52945197
9. [颠覆认知——Redis会遇到的15个「坑」，你踩过几个？](http://kaito-kidd.com/2021/03/14/redis-trap/)
