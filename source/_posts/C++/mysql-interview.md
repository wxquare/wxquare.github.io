---
title: mysql 必知必会
categories:
- C/C++
---



## 数据库三大范式
- 第一范式：每个列都不可以再拆分。
- 第二范式：在第一范式的基础上，非主键列完全依赖于主键，而不能是依赖于主键的一部分。
- 第三范式：在第二范式的基础上，非主键列只依赖于主键，不依赖于其他非主键。
在设计数据库结构的时候，要尽量遵守三范式，如果不遵守，必须有足够的理由。比如性能。事实上我们经常会为了性能而妥协数据库的设计。

## 建表的类型和约束
- int,tinyint,int(10)什么意思
- float，double
- varchar，char（定长，根据需要使用空格填充）
- datetime，timestamp
- 约束：NOT UNLL,UNIQUE,PRIMARY KEY,DEFAULT,FOREIGN KEY约束

## mysql常用存储引擎以及区别
- mysql存储引擎是插件式的，支持多种存储引擎，比较常用的是innodb和myisam
- 存储结构上的不同：innodb数据和索引时集中存储的，myism数据和索引是分开存储的
- 数据插入顺序不同：innodb插入记录时是按照主键大小有序插入，myism插入数据时是按照插入顺序保存的
- 事务的支持：Innodb提供了对数据库ACID事务的支持，并且还提供了行级锁和外键的约束。MyIASM引擎不提供事务的支持，也不支持行级锁和外键
- 索引的不同：innodb主键索引是聚簇索引，非主键索引是非聚簇索引，myisam是非聚簇索引。聚簇索引的叶子节点就是数据节点，而myism索引的叶子节点仍然是索引节点，只不过是指向对应数据块的指针,InnoDB的非聚簇索引叶子节点存储的是主键，需要再寻址一次才能得到数据


## 索引index优化查询效率
- 什么是索引，对索引的理解，什么场景下使用索引，怎么用索引优化查询性能。索引的出现就是为了提高查询效率，就像书的目录。其实说白了，索引要解决的就是查询问题。
- 怎么创建和删除索引？
- 在什么场景下用索引，where，order by,join on
- 主键索引和普通索引。为什么innodb表必须有主键，并且推荐使用整型的自增主键？因为在innodb表中，数据和主键索引用B+Tree来组织的，没有主键innodb会生成唯一列，类似于rowid。InnoDB非主键索引的叶子节点存储的是主键
- 单列索引和联合索引，联合索引的存储结构，联合索引的左前缀原则
- 聚簇索引和非聚簇索引
- 覆盖索引。覆盖索引（covering index）指一个查询语句的执行只用从索引中就能够取得，不必从数据表中读取。也可以称之为实现了索引覆盖。
如果一个索引包含了（或覆盖了）满足查询语句中字段与条件的数据就叫做覆盖索引
- [索引的数据结构，红黑树、B树、B+树的比较](https://mp.weixin.qq.com/s?__biz=MzUxNTQyOTIxNA==&mid=2247484041&idx=1&sn=76d3bf1772f9e3c796ad3d8a089220fa&chksm=f9b784b8cec00dae3d52318f6cb2bdee39ad975bf79469b72a499ceca1c5d57db5cbbef914ea&token=2025456560&lang=zh_CN#rd)
- 列出索引失效的几种场景？
    - 条件中包含or
    - 条件中包含%like
    - 联调索引，违背最左匹配原则
    - 在索引列上有一些额外的计算操作



## 事务Transaction与锁
- 什么是数据库事务？事务是一个不可分割的数据库操作序列，也是数据库并发控制的基本单位，其执行的结果必须使数据库从一种一致性状态变到另一种一致性状态。事务是逻辑上的一组操作，要么都执行，要么都不执行
- [innodb事务的ACID特性，以及其对应的实现原理?](https://www.cnblogs.com/kismetv/p/10331633.html)   
    - 原子性：语句要么全执行，要么全不执行，是事务最核心的特性，事务本身就是以原子性来定义的；实现主要基于undo log
    - 持久性：保证事务提交后不会因为宕机等原因导致数据丢失；实现主要基于redo log
    - 隔离性：保证事务执行尽可能不受其他事务影响；InnoDB默认的隔离级别是RR，RR的实现主要基于锁机制（包含next-key lock）、MVCC（包括数据的隐藏列、基于undo log的版本链、ReadView）
    - 一致性：事务追求的最终目标，一致性的实现既需要数据库层面的保障，也需要应用层面的保障

- innodb四种隔离属性以及分别会产生什么问题?
- [MySQL InnoDB MVCC 机制的原理及实现](https://zhuanlan.zhihu.com/p/64576887)
- mvcc. https://zhuanlan.zhihu.com/p/64576887

- 四个隔离属性以及脏读，不可重复读和幻读,READ-UNCOMMITTED，READ-COMMITTED，READ-Repeatable，SERIALIZABLE(可串行化)
- 锁，在Read Uncommitted级别下，读取数据不需要加共享锁，这样就不会跟被修改的数据上的排他锁冲突；在Read Committed级别下，读操作需要加共享锁，但是在语句执行完以后释放共享锁；在Repeatable Read级别下，读操作需要加共享锁，但是在事务提交之前并不释放共享锁，也就是必须等待事务执行完毕以后才释放共享锁。SERIALIZABLE 是限制性最强的隔离级别，因为该级别锁定整个范围的键，并一直持有锁，直到事务完成
- 死锁。死锁是指两个或多个事务在同一资源上相互占用，并请求锁定对方的资源，从而导致恶性循环的现象
- 悲观锁和乐观锁。悲观锁：假定会发生并发冲突，屏蔽一切可能违反数据完整性的操作。在查询完数据的时候就把事务锁起来，直到提交事务。实现方式：使用数据库中的锁机制。：假设不会发生并发冲突，只在提交操作时检查是否违反数据完整性。在修改数据的时候把事务锁起来，通过version的方式来进行锁定。实现方式：乐一般会使用版本号机制或CAS算法实现。乐观锁适合多读的场景，悲观锁适合多写的场景
- MySQL中InnoDB引擎的行锁是怎么实现的？
- innodb支持行级索，事务和聚簇索引
- innodb的行级锁是基于索引实现的，即：加锁的对象是索引而非具体的数据。当加锁操作是使用聚簇索引时，innodb会先锁住非主键索引，再锁定非聚簇索引所对应的聚簇索引。行级锁的加锁条件必须有对应的索引项，否则会退化为表级索
- 表级锁：lock table tbl_name
- 页级锁：锁住指定数据空间。select id from table_name where age between 1 and 10 for update
- 行级锁：select id from table where age=12 for update
- 乐观锁：先对数据进行操作，提交时再校验权限的状态
- 悲观锁：现获取操作权限，再对数据进行操作
- 显示加锁：select ... for upate,隐式加锁，insert/update/delete
- [mysql事务的面试](https://blog.csdn.net/qq_43255017/article/details/106442887?utm_medium=distribute.pc_feed.none-task-blog-alirecmd-3.nonecase&depth_1-utm_source=distribute.pc_feed.none-task-blog-alirecmd-3.nonecase&request_id=)
- mysql mysql 逻辑存储结构


## 视图view
- 什么是视图，什么场景下使用，为什么使用？有什么优缺点
- 数据安全性，简化sql查询；
- 视图是由基本表(实表)产生的表(虚表)，视图的列可以来自不同的表，是表的抽象和在逻辑意义上建立的新关系，视图的建立和删除不影响基本表
- 数据库必须把视图的查询转化成对基本表的查询，如果这个视图是由一个复杂的多表查询所定义，那么，即使是视图的一个简单查询，数据库也把它变成一个复杂的结合体，需要花费一定的时间


## 优化
- 如何定位及优化SQL语句的性能问题？创建的索引有没有被使用到?或者说怎么才可以知道这条语句运行很慢的原因？explain
- 大表数据查询，怎么优化？a. 优化shema、sql语句+索引；b.第二加缓存，memcached, redis；
主从复制，读写分离；垂直拆分，根据你模块的耦合度，将一个大的系统分为多个小的系统，也就是分布式系统；水平切分，针对数据量大的表，这一步最麻烦，最能考验技术水平，要选择一个合理的sharding key, 为了有好的查询效率，表结构也要改动，做一定的冗余，应用也要改，sql中尽量带sharding key，将数据定位到限定的表上去查，而不是扫描全部的表
- 慢查询日志,用于记录执行时间超过某个临界值的SQL日志，用于快速定位慢查询，为我们的优化做参考
- 关心过业务系统里面的sql耗时吗？统计过慢查询吗？对慢查询都怎么优化过？分析sql语句，分析执行计划。慢查询的优化首先要搞明白慢的原因是什么？ 是查询条件没有命中索引？是load了不需要的数据列？还是数据量太大？
- 数据库分页了解吗？
- 不同DB库的表如何联表查询，怎么优化？
- 一个6亿的表a，一个3亿的表b，通过外间tid关联，你如何最快的查询出满足条件的第50000到第50200中的这200条数据记录。
1、如果A表TID是自增长,并且是连续的,B表的ID为索引 select * from a,b where a.tid = b.id and a.tid>500000 limit 200;
2、如果A表的TID不是连续的,那么就需要使用覆盖索引.TID要么是主键,要么是辅助索引,B表ID也需要有索引。select * from b , (select tid from a limit 50000,200) a where b.id = a .tid;
- 谈谈MySQL的Explain. index > ALL
- 读写分离常见方案？ 应用程序根据业务逻辑来判断，增删改等写操作命令发给主库，查询命令发给备库。利用中间件来做代理，负责对数据库的请求识别出读还是写，并分发到不同的数据库中。（如：amoeba，mysql-proxy）
- 关心过业务系统里面的sql耗时吗？统计过慢查询吗？对慢查询都怎么优化过？
我们平时写Sql时，都要养成用explain分析的习惯。
慢查询的统计，运维会定期统计给我们
优化慢查询：
分析语句，是否加载了不必要的字段/数据。
分析SQl执行句话，是否命中索引等。
如果SQL很复杂，优化SQL结构
如果表数据量太大，考虑分表
- SQL 约束有哪几种呢？
NOT NULL: 约束字段的内容一定不能为NULL。
UNIQUE: 约束字段唯一性，一个表允许有多个 Unique 约束。
PRIMARY KEY: 约束字段唯一，不可重复，一个表只允许存在一个。
FOREIGN KEY: 用于预防破坏表之间连接的动作，也能防止非法数据插入外键。
CHECK: 用于控制字段的值范围。
- MySQL中InnoDB引擎的行锁是怎么实现的？
- 主键和外键。数据库表中对储存数据对象予以唯一和完整标识的数据列或属性的组合。一个数据列只能有一个主键，且主键的取值不能缺失，即不能为空值（Null）。外键：在一个表中存在的另一个表的主键称此表的外键。
- 主键是数据库确保数据行在整张表唯一性的保障，即使业务上本张表没有主键，也建议添加一个自增长的ID列作为主键。设定了主键之后，在后续的删改查的时候可能更加快速以及确保操作数据范围安全。
- 关联查询、联合查询、子查询（嵌套查询）
- 隔离级别与锁的关系.可以先阐述四种隔离级别，再阐述它们的实现原理。隔离级别就是依赖锁和MVCC实现的。
- 可以先阐述四种隔离级别，再阐述它们的实现原理。隔离级别就是依赖锁和MVCC实现的。
- B+树在满足聚簇索引和覆盖索引的时候不需要回表查询数据
- 一条sql执行过长的时间，你如何优化，从哪些方面入手？
查看是否涉及多表和子查询，优化Sql结构，如去除冗余字段，是否可拆表等
优化索引结构，看是否可以适当添加索引
数量大的表，可以考虑进行分离/分表（如交易流水表）
数据库主从分离，读写分离
explain分析sql语句，查看执行计划，优化sql
查看mysql执行日志，分析是否有其他方面的问题
- Redis和mysql数据怎么保持数据一致的？ https://juejin.im/post/6844903805641818120
- [怎么处理线上DDL变更?](https://zhuanlan.zhihu.com/p/247939271)

## sql 练习
[leetcode sql](https://juejin.cn/post/6844903827934560263#heading-3)

## 推荐阅读:
1. https://thinkwon.blog.csdn.net/article/details/104778621
2. [MySQL索引那些事](https://mp.weixin.qq.com/s?__biz=MzUxNTQyOTIxNA==&mid=2247484041&idx=1&sn=76d3bf1772f9e3c796ad3d8a089220fa&chksm=f9b784b8cec00dae3d52318f6cb2bdee39ad975bf79469b72a499ceca1c5d57db5cbbef914ea&token=2025456560&lang=zh_CN#rd)
3. [MySQL foreign key](https://draveness.me/whys-the-design-database-foreign-key/)
4. [mysql auto increment primary key](https://draveness.me/whys-the-design-mysql-auto-increment/)
5. https://juejin.cn/post/6844903655439597582?hmsr=joyk.com&utm_source=joyk.com&utm_source=joyk.com&utm_medium=referral%3Fhmsr%3Djoyk.com&utm_medium=referral
6. https://www.cnblogs.com/kyoner/p/11366805.html
