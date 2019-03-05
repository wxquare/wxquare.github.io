---
title: redis简单使用
---



http://www.runoob.com/redis/redis-tutorial.html

##1.Redis简介
- 主要用于key-value缓存
- 支持数据持久化，将内存数据保存在磁盘中，重启后再次加载使用
- 值（value）类型丰富，支持字符串（String）、哈希（Map）、列表（list）、集合（sets）和有序集合（sorted sets）
- 原子性，Redis的所有操作都是原子性的，意思就是要么成功执行要么失败完全不执行。单个操作是原子性的。多个操作也支持事务，即原子性，通过MULTI和EXEC指令包起来


##2.Redis安装和使用
	ubuntu直接使用命令行安装  
	`$sudo apt-get update`  
	`$sudo apt-get install redis-server`  
	启动服务器：`$redis-server`  
	启动命令行客户端：`redis-cli`  

##3.五种数据类型
1. String（字符串）：二进制安全，可以包含任何数据，比如图片或者序列化的对象，一个键最多存储512MB
2. Hash（字典）：键值对集合，编程语言中的map，例如用于存储、读取、修改用户属性
3. List（列表）：链表（双向链表），例如用于记录朋友圈的时间线数据
4. Set（集合）：哈希表实现，常用于求交集，例如共同好友等应用
5. zset(sorted set:有序集合)：数据插入时已经排序，常用于排行榜和带有权重的消息队列



##4.常用命令


##5.C++使用Redis
C++使用redis用hiredis完成，hiredis使用c语言编写。
官网：[https://github.com/redis/hiredis](https://github.com/redis/hiredis)




##6.Golang使用Redis
[https://github.com/go-redis/redis](https://github.com/go-redis/redis)



