---
title: Middleware - MySQL
categories:
- C/C++
---



## 了解数据库三大范式
- 第一范式：每个列都不可以再拆分。
- 第二范式：在第一范式的基础上，非主键列完全依赖于主键，而不能是依赖于主键的一部分。
- 第三范式：在第二范式的基础上，非主键列只依赖于主键，不依赖于其他非主键。
在设计数据库结构的时候，要尽量遵守三范式，如果不遵守，必须有足够的理由。比如性能。事实上我们经常会为了性能而妥协数据库的设计。


[MySQL 5.7 Reference Manual](https://dev.mysql.com/doc/refman/5.7/en/null-values.html)

## 如何建表
```
CREATE TABLE `hotel_info_tab` (
  `id` bigint(20) NOT NULL AUTO_INCREMENT,
  `hotel_id` bigint(20) NOT NULL DEFAULT '0',
  `hotel_name` varchar(64) NOT NULL DEFAULT '',
  `area_code` varchar(64) NOT NULL DEFAULT '',
  `phone_no` varchar(24) NOT NULL DEFAULT '',
  `address` text,
  `star_rating` varchar(16) NOT NULL DEFAULT '',
  `popularity_score` int(11) NOT NULL DEFAULT '0',
  `longitude` varchar(64) NOT NULL DEFAULT '',
  `latitude` varchar(64) NOT NULL DEFAULT '',
  `policies` text,
  `ext_info` text,
  `update_time` bigint(20) NOT NULL DEFAULT '0',
  `create_time` bigint(20) NOT NULL DEFAULT '0',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uidx_hotel_id` (`hotel_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 ROW_FORMAT=COMPRESSED
```

- 数值类型：int,tinyint,int(10),bigint
- 定点数（exact-value），decimal，使用字符串存储
- 浮点数（approximate-value (floating-point)）：float，double
- string: varchar(24)，char(10)（定长，根据需要使用空格填充),text
- 时间time：建表时通常会带上create_time,update_time，[datetime，timestamp类型](https://segmentfault.com/a/1190000017393602?utm_source=tag-newest)，有时也会用int32和int64的时间戳类型
   ```
  `create_time` timestamp NULL DEFAULT CURRENT_TIMESTAMP,
  `update_time` timestamp NULL DEFAULT CURRENT_TIMESTAMP,
   ```
- 约束：NOT UNLL,DEFAULT、UNIQUE,PRIMARY KEY,,FOREIGN KEY约束
-  [9.1.7 NULL Values](https://dev.mysql.com/doc/refman/5.7/en/null-values.html)，除text类型外其它类型一般不使用null
- primary key,[自增主键还是UUID？优缺点？怎么生成UUID？](https://blog.csdn.net/rocling/article/details/83116950)
- [snowfake生成订单号](https://blog.csdn.net/fly910905/article/details/82054196)
- 主键和外键。数据库表中对储存数据对象予以唯一和完整标识的数据列或属性的组合。一个数据列只能有一个主键，且主键的取值不能缺失，即不能为空值（Null）。外键：在一个表中存在的另一个表的主键称此表的外键。主键是数据库确保数据行在整张表唯一性的保障，即使业务上本张表没有主键，也建议添加一个自增长的ID列作为主键。设定了主键之后，在后续的删改查的时候可能更加快速以及确保操作数据范围安全。
- engine，通常是innodb
- 字符集选择，mysql数据默认情况下是不区分大小写的，可通过设置字符集设置大小写敏感，charset,utf8mb4
- 行格式，row_format，(https://dev.mysql.com/doc/refman/5.7/en/innodb-row-format.html)
- [int(10) 零填充zerofill](https://blog.csdn.net/houwanle/article/details/123192185)
- [10.9.1 The utf8mb4 Character Set (4-Byte UTF-8 Unicode Encoding)](https://dev.mysql.com/doc/refman/5.7/en/charset-unicode-utf8mb4.html)
- 是否需要分表，分库?（https://blog.csdn.net/thekenofDIS/article/details/108577905）
- 是否需要增加index？
- 查看见表sql：show create table table_name;


## 怎么考虑分表，单表的size？
原文链接：https://juejin.cn/post/6844903872134135816
- 今天，探讨一个有趣的话题：MySQL 单表数据达到多少时才需要考虑分库分表？有人说 2000 万行，也有人说 500 万行。那么，你觉得这个数值多少才合适呢？
曾经在中国互联网技术圈广为流传着这么一个说法：MySQL 单表数据量大于 2000 万行，性能会明显下降。事实上，这个传闻据说最早起源于百度。具体情况大概是这样的，当年的 DBA 测试 MySQL性能时发现，当单表的量在 2000 万行量级的时候，SQL 操作的性能急剧下降，因此，结论由此而来。然后又据说百度的工程师流动到业界的其它公司，也带去了这个信息，所以，就在业界流传开这么一个说法。
再后来，阿里巴巴《Java 开发手册》提出单表行数超过 500 万行或者单表容量超过 2GB，才推荐进行分库分表。对此，有阿里的黄金铁律支撑，所以，很多人设计大数据存储时，多会以此为标准，进行分表操作。那么，你觉得这个数值多少才合适呢？为什么不是 300 万行，或者是 800 万行，而是 500 万行？也许你会说这个可能就是阿里的最佳实战的数值吧？那么，问题又来了，这个数值是如何评估出来的呢？稍等片刻，请你小小思考一会儿。事实上，这个数值和实际记录的条数无关，而与 MySQL 的配置以及机器的硬件有关。因为，MySQL 为了提高性能，会将表的索引装载到内存中。InnoDB buffer size 足够的情况下，其能完成全加载进内存，查询不会有问题。但是，当单表数据库到达某个量级的上限时，导致内存无法存储其索引，使得之后的 SQL 查询会产生磁盘 IO，从而导致性能下降。当然，这个还有具体的表结构的设计有关，最终导致的问题都是内存限制。这里，增加硬件配置，可能会带来立竿见影的性能提升哈。
那么，我对于分库分表的观点是，需要结合实际需求，不宜过度设计，在项目一开始不采用分库与分表设计，而是随着业务的增长，在无法继续优化的情况下，再考虑分库与分表提高系统的性能。对此，阿里巴巴《Java 开发手册》补充到：如果预计三年后的数据量根本达不到这个级别，请不要在创建表时就分库分表。那么，回到一开始的问题，你觉得这个数值多少才合适呢？我的建议是，根据自身的机器的情况综合评估，如果心里没有标准，那么暂时以 500 万行作为一个统一的标准，相对而言算是一个比较折中的数值。


## 存储引擎（Storage Engine) 选择
[Setting the Storage Engine](https://dev.mysql.com/doc/refman/5.7/en/storage-engine-setting.html)
- mysql存储引擎是插件式的，支持多种存储引擎，比较常用的是innodb和myisam
- 存储结构上的不同：innodb数据和索引时集中存储的，myism数据和索引是分开存储的
- 数据插入顺序不同：innodb插入记录时是按照主键大小有序插入，myism插入数据时是按照插入顺序保存的
- 事务的支持：Innodb提供了对数据库ACID事务的支持，并且还提供了行级锁和外键的约束。MyIASM引擎不提供事务的支持，支持表级锁，不支持行级锁和外键。
- 索引的不同：innodb主键索引是聚簇索引，非主键索引是非聚簇索引，myisam是非聚簇索引。聚簇索引的叶子节点就是数据节点，而myism索引的叶子节点仍然是索引节点，只不过是指向对应数据块的指针,InnoDB的非聚簇索引叶子节点存储的是主键，需要再寻址一次才能得到数据

InnoDB and MyISAM are two of the most commonly used storage engines in MySQL.

InnoDB is a transactional storage engine, which means that it supports the ACID (Atomicity, Consistency, Isolation, Durability) properties of database transactions. This makes InnoDB well-suited for applications that require data consistency and integrity, such as e-commerce and financial applications. InnoDB also supports row-level locking, which allows multiple transactions to access and modify different rows in the same table simultaneously. This results in higher concurrency and better performance for multi-user applications.

MyISAM, on the other hand, is a non-transactional storage engine. This means that it does not support transactions and does not enforce the ACID properties. MyISAM is optimized for fast read performance, and is often used for applications that need to read large amounts of data quickly, such as reporting and data warehousing applications. However, because MyISAM does not support transactions, it is not as well-suited for applications that require data consistency and integrity.


## 添加索引index，优化访问速度
- [关于MySQL索引那些事](https://mp.weixin.qq.com/s?__biz=MzUxNTQyOTIxNA==&mid=2247484041&idx=1&sn=76d3bf1772f9e3c796ad3d8a089220fa&chksm=f9b784b8cec00dae3d52318f6cb2bdee39ad975bf79469b72a499ceca1c5d57db5cbbef914ea&token=2025456560&lang=zh_CN#rd)
- 什么是索引，对索引的理解，索引时一种数据结构，通过增加索引通常可以提高数据库查询的效率，但是为了维护索引结构也会降低数据更新的效率和增加一些存储代价。
- **索引类型**
    ```
    普通索引(INDEX)：最基本的索引，没有任何限制
    唯一索引(UNIQUE)：与"普通索引"类似，不同的就是：索引列的值必须唯一，但允许有空值。
    主键索引(PRIMARY)：它 是一种特殊的唯一索引，不允许有空值。
    全文索引(FULLTEXT )：仅可用于 MyISAM 表， 用于在一篇文章中，检索文本信息的, 针对较大的数据，生成全文索引很耗时好空间。
    组合索引：为了更多的提高mysql效率可建立组合索引，遵循”最左前缀“原则。
    ```
- **理解主键索引和普通索引、聚簇索引和非聚簇索引、单列索引和联合索引、覆盖索引和回表**
```
    - 主键索引和普通索引。数据和主键索引用B+Tree来组织的，没有主键innodb会生成唯一列，类似于rowid。InnoDB非主键索引的叶子节点存储的是主键
    - 单列索引和联合索引，联合索引的存储结构，联合索引的左前缀原则
    - 聚簇索引和非聚簇索引，聚簇索引数据和索引一起存储，非聚簇索引在无法做到索引覆盖的情况下需要回表
    - 覆盖索引。覆盖索引（covering index）指一个查询语句的执行只用从索引中就能够取得，不必从数据表中读取。也可以称之为实现了索引覆盖。
    如果一个索引包含了（或覆盖了）满足查询语句中字段与条件的数据就叫做覆盖索引
```
- [索引的数据结构，红黑树、B树、B+树的比较](https://mp.weixin.qq.com/s?__biz=MzUxNTQyOTIxNA==&mid=2247484041&idx=1&sn=76d3bf1772f9e3c796ad3d8a089220fa&chksm=f9b784b8cec00dae3d52318f6cb2bdee39ad975bf79469b72a499ceca1c5d57db5cbbef914ea&token=2025456560&lang=zh_CN#rd)
- [面试题：InnoDB中一棵B+树能存多少行数据？计算innob的高度](https://cloud.tencent.com/developer/article/1443681)
- 列出索引失效的几种场景？
    - 条件中包含or
    - 条件中包含%like
    - 联合索引，违背最左匹配原则
    - 在索引列上有一些额外的计算操作
- **联合索引和最左匹配原则**
    - 最左匹配原则
    - 当遇到范围查询(>、<、between、like)就会停止匹配
    - 区分度高的字段放在前面，区分度低的字段放后面。像性别、状态这种字段区分度就很低，我们一般放后面
    - [结合实例理解联合索引与最左匹配原则](https://www.cnblogs.com/rjzheng/p/12557314.html)
    - https://dev.mysql.com/doc/refman/5.7/en/multiple-column-indexes.html
- sql
    ```
    - 增加index，
    - ALTER TABLE `table` ADD INDEX `product_id_index` (`product_id`)

    ```



## 事务Transaction与锁
- [精读mysql事务的面试](https://blog.csdn.net/qq_43255017/article/details/106442887?utm_medium=distribute.pc_feed.none-task-blog-alirecmd-3.nonecase&depth_1-utm_source=distribute.pc_feed.none-task-blog-alirecmd-3.nonecase&request_id=)
- 什么是数据库事务？事务是一个不可分割的数据库操作序列，也是数据库并发控制的基本单位，其执行的结果必须使数据库从一种一致性状态变到另一种一致性状态。事务是逻辑上的一组操作，要么都执行，要么都不执行
- [innodb事务的ACID特性，以及其对应的实现原理?](https://www.cnblogs.com/kismetv/p/10331633.html)   
    - 原子性：语句要么全执行，要么全不执行，是事务最核心的特性，事务本身就是以原子性来定义的；实现主要基于undo log
    - 持久性：保证事务提交后不会因为宕机等原因导致数据丢失；实现主要基于redo log
    - 隔离性：保证事务执行尽可能不受其他事务影响；InnoDB默认的隔离级别是RR，RR的实现主要基于锁机制（包含next-key lock）、MVCC（包括数据的隐藏列、基于undo log的版本链、ReadView）
    - 一致性：事务追求的最终目标，一致性的实现既需要数据库层面的保障，也需要应用层面的保障
- **innodb四种隔离属性以及分别会产生什么问题?分别举例说明**
  -  读未提交（READ UNCOMMITTED),会产生脏读问题
  -  读提交，READ-COMMITTED，会产生不可重复读问题
  -  可重复读 （REPEATABLE READ），幻读问题
  -  SERIALIZABLE(可串行化)
- **概念**
  - 共享锁/读锁（Shared Locks）
  - 排他锁/写锁（Exclusive Locks）
  - 间隙锁
  - 死锁。死锁是指两个或多个事务在同一资源上相互占用，并请求锁定对方的资源，从而导致恶性循环的现象
  - 悲观锁和乐观锁。悲观锁：假定会发生并发冲突，屏蔽一切可能违反数据完整性的操作。在查询完数据的时候就把事务锁起来，直到提交事务。实现方式：使用数据库中的锁机制。：假设不会发生并发冲突，只在提交操作时检查是否违反数据完整性。在修改数据的时候把事务锁起来，通过version的方式来进行锁定。实现方式：乐一般会使用版本号机制或CAS算法实现。乐观锁适合多读的场景，悲观锁适合多写的场景
  - 表级锁：lock table tbl_name
  - 页级锁：锁住指定数据空间。select id from table_name where age between 1 and 10 for update
  - 行级锁：select id from table where age=12 for update
  - 在数据库的增、删、改、查中，只有增、删、改才会加上排它锁，而只是查询并不会加锁，只能通过在select语句后显式加lock in share mode或者for update来加共享锁或者排它锁
  - 显示加锁：select ... for upate,隐式加锁，insert/update/delete
  - MySQL中InnoDB引擎的行锁是怎么实现的？InnoDB行锁是通过给索引上的索引项加锁来实现的，只有通过索引条件检索数据，InnoDB才使用行级锁，否则，InnoDB将使用表锁。
  - 隔离级别与锁的关系.可以先阐述四种隔离级别，再阐述它们的实现原理。隔离级别就是依赖锁和MVCC实现的。
- **事务的隔离属性底层实现原理**，关于锁和mvcc
  - 可以先阐述四种隔离级别，再阐述它们的实现原理。隔离级别就是依赖锁和MVCC实现的。
  - https://zhuanlan.zhihu.com/p/143866444
- **分布式事务**
  - https://juejin.cn/post/6844903647197806605

## mysql 乐观锁和悲观锁使用
- 悲观锁：
  悲观锁是一种保守的并发控制机制，它假设在并发访问中会发生冲突，因此在访问数据之前会锁定资源，阻止其他事务对资源进行修改。在MySQL中，悲观锁主要通过以下方式实现：

  - 使用SELECT ... FOR UPDATE语句：在读取数据时对所选行进行锁定，确保其他事务不能对这些行进行修改。
  - 使用LOCK TABLES语句：锁定整个表，防止其他事务对该表进行读取和修改。


- 乐观锁：
  乐观锁是一种乐观的并发控制机制，它假设在并发访问中不会发生冲突，允许多个事务同时访问资源。当提交事务时，系统会检查资源是否被其他事务修改，如果检测到冲突，则回滚事务。在MySQL中，乐观锁通常通过以下方式实现：
   - 使用版本号或时间戳：在数据表中增加一个版本号或时间戳字段，每次修改数据时更新该字段。在提交事务时，检查版本号或时间戳是否与开始事务时的值相同，如果不同则表示发生了冲突。
   - 使用CAS（Compare and Swap）操作：在编程语言层面，通过CAS操作来比较内存中的值与预期值是否相等，如果相等则修改，否则放弃修改。
使用乐观锁和悲观锁的选择取决于应用场景和需求：

悲观锁适合在并发冲突频繁的情况下，通过独占资源避免并发问题，但会对系统性能产生一定的影响。
乐观锁适合在并发冲突较少的情况下，通过乐观的并发控制机制提高系统性能，但需要处理冲突的情况。
在实际使用时，需要根据具体业务场景和需求选择适当的并发控制机制，并注意处理冲突和回滚事务的策略，以确保数据的一致性和完整性。

## 视图view
- 什么是视图，什么场景下使用，为什么使用？有什么优缺点
- 数据安全性，简化sql查询；
- 视图是由基本表(实表)产生的表(虚表)，视图的列可以来自不同的表，是表的抽象和在逻辑意义上建立的新关系，视图的建立和删除不影响基本表
- 数据库必须把视图的查询转化成对基本表的查询，如果这个视图是由一个复杂的多表查询所定义，那么，即使是视图的一个简单查询，数据库也把它变成一个复杂的结合体，需要花费一定的时间


## 数据库优化
- **关注监控指标**
  - read write qps
  - connections
    ```
    show variables like '%max_connection%'; 查看最大连接数
    show status like  'Threads%';
    show processlist;
    ```
  
- 存储空间information_schema
  ```
    -- desc information_schema.tables;
  
    -- 查看 MySQL「所有库」的容量大小
    SELECT table_schema AS '数据库', SUM(table_rows) AS '记录数', 
    SUM(truncate(data_length / 1024 / 1024, 2)) AS '数据容量(MB)',
    SUM(truncate(index_length / 1024 / 1024, 2)) AS '索引容量(MB)',
    SUM(truncate(DATA_FREE / 1024 / 1024, 2)) AS '碎片占用(MB)'
    FROM information_schema.tables
    GROUP BY table_schema
    ORDER BY SUM(data_length) DESC, SUM(index_length) DESC;
    
    -- 指定书库查看表的数据量
    SELECT
      table_schema as '数据库',
      table_name as '表名',
      table_rows as '记录数',
      truncate(data_length/1024/1024, 2) as '数据容量(MB)',
      truncate(index_length/1024/1024, 2) as '索引容量(MB)',
      truncate(DATA_FREE/1024/1024, 2) as '碎片占用(MB)'
    from 
      information_schema.tables
    where 
      table_schema='<数据库名>'
    order by 
      data_length desc, index_length desc;
  ```
  - [performance_schema](https://www.cnblogs.com/Courage129/p/14188422.html)
  - slow query
  - 读写分离架构需要考虑主从延时

- 查看数据timeout相关参数设置
```
show variables like '%timeout%';
{
"show variables like '%timeout%'": [
	{
		"Variable_name" : "connect_timeout",
		"Value" : "10"
	},
	{
		"Variable_name" : "delayed_insert_timeout",
		"Value" : "300"
	},
	{
		"Variable_name" : "have_statement_timeout",
		"Value" : "YES"
	},
	{
		"Variable_name" : "innodb_flush_log_at_timeout",
		"Value" : "1"
	},
	{
		"Variable_name" : "innodb_lock_wait_timeout",
		"Value" : "50"
	},
	{
		"Variable_name" : "innodb_print_lock_wait_timeout_info",
		"Value" : "OFF"
	},
	{
		"Variable_name" : "innodb_rollback_on_timeout",
		"Value" : "OFF"
	},
	{
		"Variable_name" : "interactive_timeout",
		"Value" : "28800"
	},
	{
		"Variable_name" : "lock_wait_timeout",
		"Value" : "31536000"
	},
	{
		"Variable_name" : "net_read_timeout",
		"Value" : "30"
	},
	{
		"Variable_name" : "net_write_timeout",
		"Value" : "60"
	},
	{
		"Variable_name" : "rpl_stop_slave_timeout",
		"Value" : "31536000"
	},
	{
		"Variable_name" : "slave_net_timeout",
		"Value" : "60"
	},
	{
		"Variable_name" : "thread_pool_idle_timeout",
		"Value" : "60"
	},
	{
		"Variable_name" : "wait_timeout",
		"Value" : "28800"
	}
]}

```
 
  
-  **优化的步骤**
    - 考虑数据量大导致的性能问题，访问量大导致的性能问题？
    - sql语句优化。分析执行计划，减少load的数据量
    - 考虑能否通过增加索引优化查询效率，检查索引是否生效
    - 是否有缓存
    - 根据场景来看，写操作多的情况下，考虑读写分离
    - 垂直分表、水平分表、分库
   
- **sql优化**
    - 分析数据sql的结构是否加载了不必要的字段和数据
    - [深度分页查询优化](https://juejin.cn/post/7012016858379321358)
    - 子查询和连接查询
    - 会查看sql执行计划explain
     ```
       id列：在复杂的查询语句中包含多个查询使用id标示
       select_type:select/subquery/derived/union
       table: 显示对应行正在访问哪个表
       type：访问类型，关联类型。非常重要，All,index,range,ref,const,
       possible_keys: 显示可以使用哪些索引列
       key列：显示mysql决定使用哪个索引来优化对该表的访问
       key_len：显示在索引里使用的字节数
       rows：为了找到所需要的行而需要读取的行数
     ```  

- **缓存优化**
    - 缓存更新、过期、淘汰的策略
    - 缓存可能遇到的三大问题，雪崩、穿透、击穿
    - 缓存和db的一致性问题，[缓存更新策略及其分析？](https://zhuanlan.zhihu.com/p/86396877),业界比较通用的先更新DB，再删除cache
    
- **读写分离优化**
    - 在写操作的较多的情况可以考虑数据库读写分离的方案
    - [业界的方案](https://www.cnblogs.com/wollow/p/10839890.html),代理实现和业务实现
- **分表、分库**
    - 垂直分表
    - 水平分表
    - 分库
    - 业界成熟的方案
    
## 常用命令
    mysql登陆：
        mysql -h主机 -P端口 -u用户 -p密码
        SET PASSWORD FOR 'root'@'localhost' = PASSWORD('root');
        create database wxquare_test;
        show databases;
        use wxquare_test;
        
## 常见问题
1. update json 文本需要转义
 ```sql
  update table set extinfo='{
    \"urls\": [
        {
            \"url\": \"/path1\",
            \"type\": \"type1\"
        },
        {
            \"url\": \"/path2\",
            \"type\": \"type2\"
        },
    ]
}' where id = 2;
 ```
 
2. truncate table 属于ddl语句，需要ddl的权限
3. rename database
     ```
        create database new_db; 
        rename table old_db.aaa to new_db. aaa;
        mysql -hhost -PPort -uuser_name -ppassword old_db -sNe 'show tables' | while read table; do mysql -hhost -PPort -uuser_name -ppassword -sNe "rename table old_db.$table to new_db.$table"; done
     ```
4. dump 库表结构
     ```
        mysqldump --column-statistics=0 -hhost -PPort -uuser_name -ppassword --databases -d db_name --skip-lock-tables --skip-add-drop-table --set-gtid-purged=OFF | sed 's/ AUTO_INCREMENT=[0-9]*//g' > db.sql

     ```

## 推荐阅读:
- [MySQL索引那些事](https://mp.weixin.qq.com/s?__biz=MzUxNTQyOTIxNA==&mid=2247484041&idx=1&sn=76d3bf1772f9e3c796ad3d8a089220fa&chksm=f9b784b8cec00dae3d52318f6cb2bdee39ad975bf79469b72a499ceca1c5d57db5cbbef914ea&token=2025456560&lang=zh_CN#rd)
- [MySQL foreign key](https://draveness.me/whys-the-design-database-foreign-key/)
- [mysql auto increment primary key](https://draveness.me/whys-the-design-mysql-auto-increment/)
- [SQL语句执行过程详解](https://juejin.cn/post/6844903655439597582?hmsr=joyk.com&utm_source=joyk.com&utm_source=joyk.com&utm_medium=referral%3Fhmsr%3Djoyk.com&utm_medium=referral)
- MySQL alter table的过程如下： 创建ALTER TABLE目的新表；将老表数据导入新表；删除老表。（https://blog.csdn.net/zhaiwx1987/article/details/6688970）
- [Mysql on duplicate key update 用法以及优缺点](https://www.cnblogs.com/better-farther-world2099/articles/11737376.html)
- [mysql upsert](https://stackoverflow.com/questions/6107752/how-to-perform-an-upsert-so-that-i-can-use-both-new-and-old-values-in-update-par)
- [腾讯面试：一条SQL语句执行得很慢的原因有哪些？---不看后悔系列](https://www.cnblogs.com/kubidemanong/p/10734045.html)
- [4种MySQL分页查询优化的方法](https://juejin.cn/post/6844903955470745614#heading-6)
- [怎么处理线上DDL变更?](https://zhuanlan.zhihu.com/p/247939271)
- [Redis和mysql数据怎么保持数据一致的？](https://coolshell.cn/articles/17416.html) 
- [MySQL数据库面试题（2020最新版）](https://thinkwon.blog.csdn.net/article/details/104778621)


