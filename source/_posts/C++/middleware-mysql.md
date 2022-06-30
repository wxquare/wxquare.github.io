---
title: Middleware - MySQL
categories:
- C/C++
---



## 数据库三大范式
- 第一范式：每个列都不可以再拆分。
- 第二范式：在第一范式的基础上，非主键列完全依赖于主键，而不能是依赖于主键的一部分。
- 第三范式：在第二范式的基础上，非主键列只依赖于主键，不依赖于其他非主键。
在设计数据库结构的时候，要尽量遵守三范式，如果不遵守，必须有足够的理由。比如性能。事实上我们经常会为了性能而妥协数据库的设计。

## 如何建表
- int,tinyint,int(10),bigint
- float，double
- varchar(24)，char(10)（定长，根据需要使用空格填充),text
- create_time,update_time
- [datetime，timestamp](https://segmentfault.com/a/1190000017393602?utm_source=tag-newest)
- 约束：NOT UNLL,UNIQUE,PRIMARY KEY,DEFAULT,FOREIGN KEY约束
- 建立一张表时需要注意什么？
```
    create table Student(
    -> Sno int not null auto_increment primary key,
    -> Sname varchar(10) not null,
    -> Sex char(1) not null,
    -> Sage tinyint(100) not null,
    -> Sdept char(4) not null)comment = '学生表';
```
- [int(10) 零填充zerofill](https://blog.csdn.net/houwanle/article/details/123192185)
- [怎么选择varchar，char,text](https://www.jianshu.com/p/a1ef006ade16)
- [自增主键还是UUID？优缺点？怎么生成UUID？](https://blog.csdn.net/rocling/article/details/83116950)
- [snowfake生成订单号](https://blog.csdn.net/fly910905/article/details/82054196)
- 主键和外键。数据库表中对储存数据对象予以唯一和完整标识的数据列或属性的组合。一个数据列只能有一个主键，且主键的取值不能缺失，即不能为空值（Null）。外键：在一个表中存在的另一个表的主键称此表的外键。主键是数据库确保数据行在整张表唯一性的保障，即使业务上本张表没有主键，也建议添加一个自增长的ID列作为主键。设定了主键之后，在后续的删改查的时候可能更加快速以及确保操作数据范围安全。

## mysql常用存储引擎以及区别
- mysql存储引擎是插件式的，支持多种存储引擎，比较常用的是innodb和myisam
- 存储结构上的不同：innodb数据和索引时集中存储的，myism数据和索引是分开存储的
- 数据插入顺序不同：innodb插入记录时是按照主键大小有序插入，myism插入数据时是按照插入顺序保存的
- 事务的支持：Innodb提供了对数据库ACID事务的支持，并且还提供了行级锁和外键的约束。MyIASM引擎不提供事务的支持，支持表级锁，不支持行级锁和外键。
- 索引的不同：innodb主键索引是聚簇索引，非主键索引是非聚簇索引，myisam是非聚簇索引。聚簇索引的叶子节点就是数据节点，而myism索引的叶子节点仍然是索引节点，只不过是指向对应数据块的指针,InnoDB的非聚簇索引叶子节点存储的是主键，需要再寻址一次才能得到数据


## 索引index优化查询效率
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
- 列出索引失效的几种场景？
    - 条件中包含or
    - 条件中包含%like
    - 联合索引，违背最左匹配原则
    - 在索引列上有一些额外的计算操作



## 事务Transaction与锁
- 什么是数据库事务？事务是一个不可分割的数据库操作序列，也是数据库并发控制的基本单位，其执行的结果必须使数据库从一种一致性状态变到另一种一致性状态。事务是逻辑上的一组操作，要么都执行，要么都不执行
- [innodb事务的ACID特性，以及其对应的实现原理?](https://www.cnblogs.com/kismetv/p/10331633.html)   
    - 原子性：语句要么全执行，要么全不执行，是事务最核心的特性，事务本身就是以原子性来定义的；实现主要基于undo log
    - 持久性：保证事务提交后不会因为宕机等原因导致数据丢失；实现主要基于redo log
    - 隔离性：保证事务执行尽可能不受其他事务影响；InnoDB默认的隔离级别是RR，RR的实现主要基于锁机制（包含next-key lock）、MVCC（包括数据的隐藏列、基于undo log的版本链、ReadView）
    - 一致性：事务追求的最终目标，一致性的实现既需要数据库层面的保障，也需要应用层面的保障
- innodb四种隔离属性以及分别会产生什么问题?
- 四个隔离属性以及脏读，不可重复读和幻读,READ-UNCOMMITTED，READ-COMMITTED，READ-Repeatable，SERIALIZABLE(可串行化)
- 锁，在Read Uncommitted级别下，读取数据不需要加共享锁，这样就不会跟被修改的数据上的排他锁冲突；在Read Committed级别下，读操作需要加共享锁，但是在语句执行完以后释放共享锁；在Repeatable Read级别下，读操作需要加共享锁，但是在事务提交之前并不释放共享锁，也就是必须等待事务执行完毕以后才释放共享锁。SERIALIZABLE 是限制性最强的隔离级别，因为该级别锁定整个范围的键，并一直持有锁，直到事务完成
- 死锁。死锁是指两个或多个事务在同一资源上相互占用，并请求锁定对方的资源，从而导致恶性循环的现象
- 悲观锁和乐观锁。悲观锁：假定会发生并发冲突，屏蔽一切可能违反数据完整性的操作。在查询完数据的时候就把事务锁起来，直到提交事务。实现方式：使用数据库中的锁机制。：假设不会发生并发冲突，只在提交操作时检查是否违反数据完整性。在修改数据的时候把事务锁起来，通过version的方式来进行锁定。实现方式：乐一般会使用版本号机制或CAS算法实现。乐观锁适合多读的场景，悲观锁适合多写的场景
- innodb支持行级索，事务和聚簇索引
- 表级锁：lock table tbl_name
- 页级锁：锁住指定数据空间。select id from table_name where age between 1 and 10 for update
- 行级锁：select id from table where age=12 for update
- 乐观锁：先对数据进行操作，提交时再校验权限的状态
- 悲观锁：现获取操作权限，再对数据进行操作
- 显示加锁：select ... for upate,隐式加锁，insert/update/delete
- [mysql事务的面试](https://blog.csdn.net/qq_43255017/article/details/106442887?utm_medium=distribute.pc_feed.none-task-blog-alirecmd-3.nonecase&depth_1-utm_source=distribute.pc_feed.none-task-blog-alirecmd-3.nonecase&request_id=)
- mysql mysql 逻辑存储结构
- MySQL中InnoDB引擎的行锁是怎么实现的？InnoDB行锁是通过给索引上的索引项加锁来实现的，只有通过索引条件检索数据，InnoDB才使用行级锁，否则，InnoDB将使用表锁。
- 隔离级别与锁的关系.可以先阐述四种隔离级别，再阐述它们的实现原理。隔离级别就是依赖锁和MVCC实现的。
- 可以先阐述四种隔离级别，再阐述它们的实现原理。隔离级别就是依赖锁和MVCC实现的。
- https://zhuanlan.zhihu.com/p/143866444


## 视图view
- 什么是视图，什么场景下使用，为什么使用？有什么优缺点
- 数据安全性，简化sql查询；
- 视图是由基本表(实表)产生的表(虚表)，视图的列可以来自不同的表，是表的抽象和在逻辑意义上建立的新关系，视图的建立和删除不影响基本表
- 数据库必须把视图的查询转化成对基本表的查询，如果这个视图是由一个复杂的多表查询所定义，那么，即使是视图的一个简单查询，数据库也把它变成一个复杂的结合体，需要花费一定的时间


## 数据库优化
- 如何定位及优化SQL语句的性能问题？创建的索引有没有被使用到?或者说怎么才可以知道这条语句运行很慢的原因？explain
-  **优化的步骤**
    - 检查mysql服务器资源使用情况，https://www.cnblogs.com/remember-forget/p/10400496.html
    - sql语句优化。分析执行计划，减少load的数据量
    - 增加索引，索引优化
    - 缓存，memcached, redis
    - 根据场景来看，写操作多的情况下，考虑读写分离
    - 垂直分表、水平分表、分库
- **sql优化**
    - 分析数据sql的结构是否加载了不必要的字段和数据
    - 分页查询优化
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
    - 什么是慢查询
- **索引优化**
    - 怎么建索引
    - 注意索引是否生效
- **缓存优化**
    - 缓存更新和淘汰的策略
    - 缓存可能遇到的三大问题，雪崩、穿透、击穿
    - 缓存和db的一致性问题
    - [缓存更新策略及其分析？](https://zhuanlan.zhihu.com/p/86396877),业界比较通用的先更新DB，在删除cache
- **读写分离优化**
    - 在写操作的较多的情况可以考虑数据库读写分离的方案
    - [业界的方案](https://www.cnblogs.com/wollow/p/10839890.html),代理实现和业务实现
- **分表、分库**
    - 垂直分表
    - 水平分表
    - 分库
    - 业界成熟的方案
- [腾讯面试：一条SQL语句执行得很慢的原因有哪些？---不看后悔系列](https://www.cnblogs.com/kubidemanong/p/10734045.html)
- [4种MySQL分页查询优化的方法](https://juejin.cn/post/6844903955470745614#heading-6)
- [不同DB库的表如何联表查询](https://www.modb.pro/db/27539)，怎么优化？
- [一个跨库复杂查询的SQL优化的案例](https://blog.csdn.net/waste_land_wolf/article/details/76419207)
- [怎么处理线上DDL变更?](https://zhuanlan.zhihu.com/p/247939271)
- [Redis和mysql数据怎么保持数据一致的？](https://coolshell.cn/articles/17416.html) 

- 一个6亿的表a，一个3亿的表b，通过外间tid关联，你如何最快的查询出满足条件的第50000到第50200中的这200条数据记录。
1、如果A表TID是自增长,并且是连续的,B表的ID为索引 select * from a,b where a.tid = b.id and a.tid>500000 limit 200;
2、如果A表的TID不是连续的,那么就需要使用覆盖索引.TID要么是主键,要么是辅助索引,B表ID也需要有索引。select * from b , (select tid from a limit 50000,200) a where b.id = a .tid;


## sql 练习
1. 常用命令
    mysql登陆：
        mysql -h主机 -P端口 -u用户 -p密码
        SET PASSWORD FOR 'root'@'localhost' = PASSWORD('root');
        create database wxquare_test;
        show databases;
        use wxquare_test;
[leetcode sql](https://juejin.cn/post/6844903827934560263#heading-3)

## 推荐阅读:
1. https://thinkwon.blog.csdn.net/article/details/104778621
2. [MySQL索引那些事](https://mp.weixin.qq.com/s?__biz=MzUxNTQyOTIxNA==&mid=2247484041&idx=1&sn=76d3bf1772f9e3c796ad3d8a089220fa&chksm=f9b784b8cec00dae3d52318f6cb2bdee39ad975bf79469b72a499ceca1c5d57db5cbbef914ea&token=2025456560&lang=zh_CN#rd)
3. [MySQL foreign key](https://draveness.me/whys-the-design-database-foreign-key/)
4. [mysql auto increment primary key](https://draveness.me/whys-the-design-mysql-auto-increment/)
5. https://juejin.cn/post/6844903655439597582?hmsr=joyk.com&utm_source=joyk.com&utm_source=joyk.com&utm_medium=referral%3Fhmsr%3Djoyk.com&utm_medium=referral
6. https://www.cnblogs.com/kyoner/p/11366805.html
7. [一文精通MYSQL](http://km.oa.com/articles/show/491871?kmref=author_post)
8. MySQL alter table的过程如下： 创建ALTER TABLE目的新表；将老表数据导入新表；删除老表。（https://blog.csdn.net/zhaiwx1987/article/details/6688970）