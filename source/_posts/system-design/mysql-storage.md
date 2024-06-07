---
title: 互联网系统设计 - 存储与Mysql数据库
date: 2024-03-04
categories:
- 系统设计
---

多查看文档
[MySQL 5.7 Reference Manual](https://dev.mysql.com/doc/refman/5.7/en/null-values.html)


## 基本使用
### 如何建表？
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
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci ROW_FORMAT=COMPRESSED
```

#### 类型选择？
- 数值类型：int,tinyint,int(10),bigint
- 定点数（exact-value），decimal，使用字符串存储，精度
- 浮点数（approximate-value (floating-point)）：float，double，精度缺失
- 字符串: varchar(256)，char(10)（定长，根据需要使用空格填充)
- 文本: text,json
  ```
  JSON 数据类型提供了数据格式验证和以及一些内置函数帮助查询和检索。
  JSON数据类型更适合存储和处理结构化的JSON数据，而TEXT数据类型更适合存储纯文本字符串。如果你需要在数据库中存储和操作JSON数据，并且使用MySQL 5.7及更高版本，那么JSON数据类型是更好的选择。如果你只需要存储普通的文本字符串，而不需要对JSON数据进行特殊处理，那么TEXT数据类型就足够了
  ```
- 时间time：建表时通常会带上create_time,update_time，[datetime，timestamp类型](https://segmentfault.com/a/1190000017393602?utm_source=tag-newest)，有时也会用int32和int64的时间戳类型
   ```
  `create_time` timestamp NULL DEFAULT CURRENT_TIMESTAMP,
  `update_time` timestamp NULL DEFAULT CURRENT_TIMESTAMP,
   ```
   **通常存储的都是时间戳，需要考虑使用mysql服务器的时间还是业务的时间戳，考虑使用mysql时间戳是否会有不利的影响**

#### primary key
- **主键PRIMARY KEY**。数据库表中对储存数据对象予以唯一和完整标识的数据列或属性的组合。一个数据列只能有一个主键，且主键的取值不能缺失，即不能为空值（Null）。主键是数据库确保数据行在整张表唯一 性的保障，即使业务上本张表没有主键，也建议添加一个自增长的ID列作为主键。设定了主键之后，在后续的删改查的时候可能更加快速以及确保操作数据范围安全。
- [自增主键还是UUID？优缺点？怎么生成UUID？](https://blog.csdn.net/rocling/article/details/83116950)，比如item表使用自增ID，order表使用订单id，订单id可以认为是uuid。

#### unique key
- **唯一性约束UNIQUE KEY**：唯一性约束是很重要的特性，防止重复插入数据


#### 关于FOREIGN KEY约束,不建议使用
- **外键**：在一个表中存在的另一个表的主键称此表的外键。外键约束能保证好的保证的数据的完整性，但是会影响数据插入的性能，并且不方便后续的shard，所以一般不建议使用。
- [为什么不推荐使用外键约束，而是业务代码来实现？](https://www.zhihu.com/question/21863571)


#### int(10),bigint(20)
- **整数类型的括号中的数字仅用于指定显示宽度，并不会影响存储范围或存储空**
- 显示宽度：括号中的数字用于指定在查询结果中显示整数类型字段时的字符个数。它可以控制字段在查询结果中的对齐和显示格式。例如，如果将一个整数字段定义为 int(3)，并插入值 100，在查询时该字段将以 '100' 的形式显示，左侧用空格填充以达到指定的宽度
- 零填充：括号中的数字还可以与 ZEROFILL 属性一起使用，以实现零填充的效果。当整数类型字段定义为 int(3) ZEROFILL 时，如果插入的值不足指定的宽度，MySQL 将在左侧用零进行填充


#### 编码方式 
- **编码方式**：**utf8mb4**：通过 show variables like 'character_set_%'; 可以查看系统默认字符集。mysql中有utf8和utf8mb4两种编码，在mysql中请大家忘记**utf8**，永远使用**utf8mb4**。这是mysql的一个遗留问题，mysql中的utf8最多只能支持3bytes长度的字符编码，对于一些需要占据4bytes的文字，mysql的utf8就不支持了，要使用utf8mb4才行
- **COLLATE=utf8mb4_unicode_ci**,所谓utf8_unicode_ci，其实是用来排序的规则。对于mysql中那些字符类型的列，如VARCHAR，CHAR，TEXT类型的列，都需要有一个COLLATE类型来告知mysql如何对该列进行排序和比较。简而言之，COLLATE会影响到ORDER BY语句的顺序，会影响到WHERE条件中大于小于号筛选出来的结果，会影响**DISTINCT**、**GROUP BY**、**HAVING**语句的查询结果。另外，mysql建索引的时候，如果索引列是字符类型，也会影响索引创建，只不过这种影响我们感知不到。总之，凡是涉及到字符类型比较或排序的地方，都会和COLLATE有关。
- **行格式**，row_format，(https://dev.mysql.com/doc/refman/5.7/en/innodb-row-format.html)
- [10.9.1 The utf8mb4 Character Set (4-Byte UTF-8 Unicode Encoding)](https://dev.mysql.com/doc/refman/5.7/en/charset-unicode-utf8mb4.html)

#### 关于 null 的使用
- **除text类型外其它类型一般不使用null，都应该指定默认值**
   在MySQL和许多其他数据库系统中，**NULL是一个特殊的值，表示缺少值或未知值**。虽然NULL在某些情况下是有用的，但由于它的特殊性，使用NULL可能会带来一些问题，因此在某些情况下不建议过度使用NULL。一般只有text类型回用到，其它都应该制定默认值
1. 逻辑判断和比较的复杂性：由于NULL表示未知或缺少值，它的比较结果不是true也不是false，而是NULL。这意味着使用NULL进行逻辑判断和比较时需要额外的注意，可能需要使用IS NULL或IS NOT NULL等特殊的操作符。
2. 聚合函数的结果处理：在使用聚合函数（如SUM、AVG、COUNT等）进行计算时，NULL的处理可能会产生意外的结果。通常情况下，聚合函数会忽略NULL值，因此如果某列中有NULL值，可能会导致计算结果不准确。
3. 索引的使用限制：某些类型的索引在处理NULL值时可能会受到限制。例如，对于普通索引（B-tree索引）来说，NULL值并不会被索引，因此在查询时可能无法充分利用索引的性能优势。
4. 查询语句的复杂性增加：当使用NULL值进行查询时，可能需要编写更复杂的查询语句来处理NULL的情况，这会增加查询的复杂性和维护成本。

虽然NULL有其合理的用途，例如表示缺失的数据或未知的值，但过度使用NULL可能会导致代码的复杂性增加、查询的不准确性和性能问题。在设计数据库模式和数据模型时，需要根据实际需求和业务逻辑合理使用NULL，并考虑到其带来的潜在问题。


#### 存储引擎（Storage Engine) 选择
[Setting the Storage Engine](https://dev.mysql.com/doc/refman/5.7/en/storage-engine-setting.html)
MySQL支持多种存储引擎，每种存储引擎都有其特点和适用场景。以下是几种常见的MySQL存储引擎对比：
- InnoDB：
	- 事务支持：InnoDB是MySQL默认的事务性存储引擎，支持ACID事务特性，适用于需要强一致性和事务支持的应用。
	- 行级锁定：InnoDB支持行级锁定，提供更好的并发性能。
	- 外键约束：InnoDB支持外键约束，可以保持数据完整性。
	- Crash Recovery：InnoDB具有崩溃恢复机制，能够在故障恢复时保证数据的一致性。
	- 适用场景：适用于高并发、需要事务支持和数据完整性的应用，如电子商务、在线交易等。

- MyISAM：
	- 速度和性能：MyISAM对于读取操作有很好的性能表现，适用于读取频繁的应用。
	- 表级锁定：MyISAM使用表级锁定，对并发性能有一定影响。
	- 不支持事务：MyISAM不支持事务和崩溃恢复机制，不保证数据的完整性和一致性。
	- 全文索引：MyISAM支持全文索引，适用于对文本内容进行高效搜索的应用。
	- 适用场景：适用于读取频繁、对事务和数据完整性要求不高的应用，如博客、新闻等。
- mysql存储引擎是插件式的，支持多种存储引擎，比较常用的是innodb和myisam
- 存储结构上的不同：innodb数据和索引时集中存储的，myism数据和索引是分开存储的
- 数据插入顺序不同：innodb插入记录时是按照主键大小有序插入，myism插入数据时是按照插入顺序保存的
- 事务的支持：Innodb提供了对数据库ACID事务的支持，并且还提供了行级锁和外键的约束。MyIASM引擎不提供事务的支持，支持表级锁，不支持行级锁和外键。
- 索引的不同：innodb主键索引是聚簇索引，非主键索引是非聚簇索引，myisam是非聚簇索引。聚簇索引的叶子节点就是数据节点，而myism索引的叶子节点仍然是索引节点，只不过是指向对应数据块的指针,InnoDB的非聚簇索引叶子节点存储的是主键，需要再寻址一次才能得到数据
总结：
- 是否需要支持事务？innodb
- 并发写是不是很多？innoda
- 读多，写少，追求读速度？myisam


#### 索引选择
- 唯一性约束
- 联合索引
- 见索引

#### mysql隐式类型变换（有一次面试题：存储类型和查询类型不一致会发生什么？）
在MySQL中，隐式类型转换是指在表达式或操作中自动将一个数据类型转换为另一个数据类型。MySQL会根据一组规则来执行隐式类型转换，以便执行操作或比较不同类型的数据。以下是MySQL中的一些常见的隐式类型转换规则：
- 数值类型之间的转换：MySQL会自动将不同数值类型之间进行隐式转换，例如将整数转换为浮点数，或将较小的数值类型转换为较大的数值类型。
- 字符串和数值类型之间的转换：MySQL会尝试将字符串转换为数值类型，或将数值类型转换为字符串。如果字符串可以解析为有效的数值，那么它将被转换为相应的数值类型。
- 日期和时间类型之间的转换：MySQL会自动将日期和时间类型转换为其他日期和时间类型。例如，可以将日期类型转换为字符串，或将字符串转换为日期类型。
- NULL的处理：在与其他数据类型进行操作时，MySQL会将NULL隐式转换为适当的数据类型。例如，NULL与数值类型相加时会被转换为0


## mysql 线上DDL表结构变更注意事项

在MySQL中进行字段类型修改、增加字段、增加索引和删除索引时，需要注意以下事项：

- 数据备份：在进行任何结构变更之前，务必备份数据库的数据。这样可以在出现意外情况或错误时恢复数据。
- 考虑数据类型转换：如果要修改字段的数据类型，需要考虑可能的数据类型转换问题。确保目标数据类型能够容纳原有数据，并且进行数据类型转换时不会导致数据丢失或截断。
- 处理依赖关系：在修改字段类型、增加字段或删除字段时，需要考虑是否存在其他对象（如视图、存储过程或触发器）依赖于该字段。如果存在依赖关系，需要先处理这些依赖关系，以免操作失败或导致不一致性。
- 使用ALTER TABLE语句：对于字段类型修改、增加字段和删除字段操作，可以使用ALTER TABLE语句来执行。确保在执行ALTER TABLE语句之前，先检查表的当前状态和结构，以避免不必要的错误。
- 考虑数据量和性能：在进行结构变更操作时，特别是增加字段或增加索引时，需要考虑表中的数据量和性能影响。某些操作可能需要较长时间来完成，或者会对数据库的性能产生影响。在进行这些操作时，要谨慎评估和测试，以确保不会对正常运行产生负面影响。
- 索引的选择和删除：在增加索引时，需要根据查询需求和数据访问模式选择合适的索引类型（如B-tree索引、哈希索引等）。而在删除索引时，需要确保不会影响到相关查询的性能。在进行索引的修改和删除操作时，最好事先进行性能测试和评估。
- 注意并发操作和锁定：某些结构变更操作可能需要锁定表或行，以确保数据的一致性。在进行这些操作时，要注意可能的并发访问冲突，并在必要时进行合理的调度和通知，以避免对系统的影响。
- 测试和验证：在进行结构变更之后，务必进行充分的测试和验证，以确保数据库的功能和性能没有受到不良影响。验证包括执行常见的查询、操作和业务逻辑，以确保一切正常。
- 一般要求先变更DB，再发布代码
总之，在进行MySQL的字段类型修改、增加字段、增加索引和删除索引时，需要谨慎行事，提前做好充分的准备、备份和测试，以确保操作的成功和数据的安全性
- 表锁定和影响：某些DDL操作可能需要锁定整个表，这可能会对其他用户的操作产生影响。请在合适的时机执行DDL操作，避免对关键业务时间或频繁访问的表造成过多的阻塞。
- 大型表操作：对于大型表的DDL操作（如ALTER TABLE），可能会涉及大量的数据移动和重建，可能会导致长时间的操作和额外的存储空间使用。在执行这些操作之前，请确保对表的大小和操作的影响进行评估
- 错误处理和回滚：在执行DDL操作时，要注意捕获和处理可能的错误。如果DDL操作失败，确保有适当的错误处理机制和回滚策略，以保持数据的一致性
- 数据库备份：在执行重要的DDL操作之前，请确保对数据库进行备份，以防操作出现问题导致数据丢失或不可恢复。这可以帮助你在需要时还原到先前的状态


## mysql架构扩展

关系型数据库扩展包括许多技术：**主从复制**、**主主复制**、**联合**、**分片**、**非规范化**和 **SQL调优**。
- [扩展你的用户数到第一个一千万](https://www.youtube.com/watch?v=w95murBkYmU)

### 主从复制
<p align="center">
  <img src="/images/C9ioGtn.png" width=600 height=400>
  <br/>
  <strong><a href="http://www.slideshare.net/jboner/scalability-availability-stability-patterns/">资料来源：可扩展性、可用性、稳定性、模式</a></strong>
</p>
主库同时负责读取和写入操作，并复制写入到一个或多个从库中，从库只负责读操作。树状形式的从库再将写入复制到更多的从库中去。如果主库离线，系统可以以只读模式运行，直到某个从库被提升为主库或有新的主库出现。主要的优缺点：
- 读写分离提供集群的性能
- 主、从多节点，宕机容灾
- 将从库提升为主库需要额外的逻辑
- 主从延时问题，需要监控

### 主主复制,多主复制
<p align="center">
  <img src="/images/krAHLGg.png" width=600 height=400>
  <br/>
  <strong><a href="http://www.slideshare.net/jboner/scalability-availability-stability-patterns/">资料来源：可扩展性、可用性、稳定性、模式</a></strong>
</p>

两个主库都负责读操作和写操作，写入操作时互相协调。如果其中一个主库挂机，系统可以继续读取和写入。
- [多主复制](https://en.wikipedia.org/wiki/Multi-master_replication)
优缺点：
- 你需要添加负载均衡器或者在应用逻辑中做改动，来确定写入哪一个数据库。
- 多数主-主系统要么不能保证一致性（违反 ACID），要么因为同步产生了写入延迟。
- 随着更多写入节点的加入和延迟的提高，**如何解决冲突显得越发重要**
- 多活架构

### 联合（垂直分实例，比如商品实例、订单实例等分开）

<p align="center">
  <img src="/images/U3qV33e.png" width=600 height=400>
  <br/>
  <strong><a href="https://www.youtube.com/watch?v=w95murBkYmU">资料来源：扩展你的用户数到第一个一千万</a></strong>
</p>

优缺点：
联合（或按功能划分）将数据库按对应功能分割。例如，你可以有三个数据库：**论坛**、**用户**和**产品**，而不仅是一个单体数据库，从而减少每个数据库的读取和写入流量，减少复制延迟。较小的数据库意味着更多适合放入内存的数据，进而意味着更高的缓存命中几率。没有只能串行写入的中心化主库，你可以并行写入，提高负载能力。
- 如果你的数据库模式需要大量的功能和数据表，联合的效率并不好。
- 你需要更新应用程序的逻辑来确定要读取和写入哪个数据库。
- 从两个库联结数据更复杂。
- 联合需要更多的硬件和额外的复杂度。


### 分片 (水平分实例，比如订单按照用户shard)
<p align="center">
  <img src="/images/wU8x5Id.png" width=600 height=400>
  <br/>
  <strong><a href="http://www.slideshare.net/jboner/scalability-availability-stability-patterns/">资料来源：可扩展性、可用性、稳定性、模式</a></strong>
</p>

分片将数据分配在不同的数据库上，使得每个数据库仅管理整个数据集的一个子集。以用户数据库为例，随着用户数量的增加，越来越多的分片会被添加到集群中。
类似[联合](#联合)的优点，分片可以减少读取和写入流量，减少复制并提高缓存命中率。也减少了索引，通常意味着查询更快，性能更好。如果一个分片出问题，其他的仍能运行，你可以使用某种形式的冗余来防止数据丢失。类似联合，没有只能串行写入的中心化主库，你可以并行写入，提高负载能力。
常见的做法是用户姓氏的首字母或者用户的地理位置来分隔用户表。
- 你需要修改应用程序的逻辑来实现分片，这会带来复杂的 SQL 查询。
- 分片不合理可能导致数据负载不均衡。例如，被频繁访问的用户数据会导致其所在分片的负载相对其他分片高。
- 再平衡会引入额外的复杂度。基于[一致性哈希](http://www.paperplanes.de/2011/12/9/the-magic-of-consistent-hashing.html)的分片算法可以减少这种情况。
- 联结多个分片的数据操作更复杂。
- 分片需要更多的硬件和额外的复杂度。
- [分片时代来临](http://highscalability.com/blog/2009/8/6/an-unorthodox-approach-to-database-design-the-coming-of-the.html)
- [数据库分片架构](https://en.wikipedia.org/wiki/Shard_(database_architecture))
- [一致性哈希](http://www.paperplanes.de/2011/12/9/the-magic-of-consistent-hashing.html)


## 分表/分库/历史数据归档和路由
原文链接：https://juejin.cn/post/6844903872134135816
- 今天，探讨一个有趣的话题：MySQL 单表数据达到多少时才需要考虑分库分表？有人说 2000 万行，也有人说 500 万行。那么，你觉得这个数值多少才合适呢？
曾经在中国互联网技术圈广为流传着这么一个说法：MySQL 单表数据量大于 2000 万行，性能会明显下降。事实上，这个传闻据说最早起源于百度。具体情况大概是这样的，当年的 DBA 测试 MySQL性能时发现，当单表的量在 2000 万行量级的时候，SQL 操作的性能急剧下降，因此，结论由此而来。然后又据说百度的工程师流动到业界的其它公司，也带去了这个信息，所以，就在业界流传开这么一个说法。
再后来，阿里巴巴《Java 开发手册》提出单表行数超过 500 万行或者单表容量超过 2GB，才推荐进行分库分表。对此，有阿里的黄金铁律支撑，所以，很多人设计大数据存储时，多会以此为标准，进行分表操作。那么，你觉得这个数值多少才合适呢？为什么不是 300 万行，或者是 800 万行，而是 500 万行？也许你会说这个可能就是阿里的最佳实战的数值吧？那么，问题又来了，这个数值是如何评估出来的呢？稍等片刻，请你小小思考一会儿。事实上，这个数值和实际记录的条数无关，而与 MySQL 的配置以及机器的硬件有关。因为，MySQL 为了提高性能，会将表的索引装载到内存中。InnoDB buffer size 足够的情况下，其能完成全加载进内存，查询不会有问题。但是，当单表数据库到达某个量级的上限时，导致内存无法存储其索引，使得之后的 SQL 查询会产生磁盘 IO，从而导致性能下降。当然，这个还有具体的表结构的设计有关，最终导致的问题都是内存限制。这里，增加硬件配置，可能会带来立竿见影的性能提升哈。
那么，我对于分库分表的观点是，需要结合实际需求，不宜过度设计，在项目一开始不采用分库与分表设计，而是随着业务的增长，在无法继续优化的情况下，再考虑分库与分表提高系统的性能。对此，阿里巴巴《Java 开发手册》补充到：如果预计三年后的数据量根本达不到这个级别，请不要在创建表时就分库分表。那么，回到一开始的问题，你觉得这个数值多少才合适呢？我的建议是，根据自身的机器的情况综合评估，如果心里没有标准，那么暂时以 500 万行作为一个统一的标准，相对而言算是一个比较折中的数值。

**案例1. 酒店分表：**
- 酒店数量100w, 支持8中语言，2000kw种房型，1亿的图片。支持未来3年可能扩展成：酒店数量500w, 支持8钟语言，房型1亿，图片5亿
- 分表方式：hotel 1张表，多语言表10张表，房型表20张，图片表：100张表
- 酒店和多语言文本垂直分表
- 根据酒店id水平分表。
- 如果还要继续扩展，可以重新搞一个库，酒店id从500w开始，不断扩展。增加一个数据路由的模块。

**案例2. 订单分表和历史订单归档（3个月或者更长时间）**
- 订单每天新增1000w。按照用户维度分1000张表。一年下来，平均每张表360w。
- 超过1年的历史订单归档，将时间超过1年的订单归档存储到hbase中
- 如何实现历史订单表数据归档，冷热数据的路由？
- [订单系统设计方案之如何做历史订单和归档](https://www.80wz.com/wfwstudy/1084.html)
- [订单数据归档方案](https://zq99299.github.io/note-book/back-end-storage/02/07.html#%E5%AD%98%E6%A1%A3%E5%8E%86%E5%8F%B2%E8%AE%A2%E5%8D%95%E6%95%B0%E6%8D%AE%E6%8F%90%E5%8D%87%E6%9F%A5%E8%AF%A2%E6%80%A7%E8%83%BD)
- <p align="center">
  <img src="/images/mysql-order-archive.png" width=600 height=200>
</p>


**案例3. 数据历史版本记录、快照表**
- 在有些场景中，数据变更不回特别频繁，特别是人工变更时，记录数据版本和快照是非常好的习惯，方便追溯历史行为记录
- 数据变更时通常会先写入快照表或者历史记录表，通常在业务代码中实现
- 有时也会采用mysql 存储过程实现：https://blog.csdn.net/wcdunf/article/details/129792810

**案例4. 商品库存扣减方案**
- 乐观索和悲观锁
- https://zhuanlan.zhihu.com/p/143866444


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
    - 对于联合索引c1、c2、c3，跳过c1 字段会导致无法命中index
    - 对于联合索引c1、c2、c3，不按照创建索引顺序也可以命中索引，innodb有索引优化
    - 当遇到范围查询(>、<、between、like)就会停止匹配
    - 区分度高的字段放在前面，区分度低的字段放后面。像性别、状态这种字段区分度就很低，我们一般放后面
    - [结合实例理解联合索引与最左匹配原则](https://www.cnblogs.com/rjzheng/p/12557314.html)
    - https://dev.mysql.com/doc/refman/5.7/en/multiple-column-indexes.html

## ACID、事务、数据库锁（数据准确性和并发安全，一锁二判三更新）
- [精读mysql事务](https://blog.csdn.net/qq_43255017/article/details/106442887?utm_medium=distribute.pc_feed.none-task-blog-alirecmd-3.nonecase&depth_1-utm_source=distribute.pc_feed.none-task-blog-alirecmd-3.nonecase&request_id=)
- [innodb事务的ACID特性，以及其对应的实现原理?](https://www.cnblogs.com/kismetv/p/10331633.html)   
    - 原子性：在很多场景中，一个操作需要执行多条 update/insert SQL。原子性保证了SQL语句要么全执行，要么全不执行，是事务最核心的特性，事务本身就是以原子性来定义的；实现主要基于undolog/redolog
    - 持久性：保证事务提交后不会因为宕机等原因导致数据丢失；实现主要基于redo log
    - 隔离性：保证事务执行尽可能不受其他事务影响；InnoDB默认的隔离级别是RR，RR的实现主要基于锁机制（包含next-key lock）、MVCC（包括数据的隐藏列、基于undo log的版本链、ReadView）
    - 一致性：事务追求的最终目标，一致性的实现既需要数据库层面的保障，也需要应用层面的保障
- **innodb四种隔离属性以及分别会产生什么问题?分别举例说明**
  -  读未提交（READ UNCOMMITTED),会产生脏读问题
  -  读提交，READ-COMMITTED，会产生不可重复读问题
  -  可重复读 （REPEATABLE READ），幻读问题(insert)，**mysql 默认的事务隔离级别**
  -  SERIALIZABLE(可串行化)
- **事务的隔离属性底层实现原理**，关于锁和mvcc
  - 可以先阐述四种隔离级别，再阐述它们的实现原理。隔离级别就是依赖锁和MVCC实现的。
- **悲观锁与乐观锁**
  - [Select for update使用详解](https://zhuanlan.zhihu.com/p/143866444) 及在库存和金钱系统上的应用
  - 悲观锁：悲观锁是一种保守的并发控制机制，它假设在并发访问中会发生冲突，因此在访问数据之前会锁定资源，阻止其他事务对资源进行修改。在MySQL中，悲观锁主要通过以下方式实现：
  	- 使用SELECT ... FOR UPDATE语句：在读取数据时对所选行进行锁定，确保其他事务不能对这些行进行修改。
  	- 使用LOCK TABLES语句：锁定整个表，防止其他事务对该表进行读取和修改。
  - 乐观锁：乐观锁是一种乐观的并发控制机制，它假设在并发访问中不会发生冲突，允许多个事务同时访问资源。当提交事务时，系统会检查资源是否被其他事务修改，如果检测到冲突，则回滚事务。在MySQL中，乐观锁通常通过以下方式实现：
   	- 使用版本号或时间戳：在数据表中增加一个版本号或时间戳字段，每次修改数据时更新该字段。在提交事务时，检查版本号或时间戳是否与开始事务时的值相同，如果不同则表示发生了冲突。
   	- 使用CAS（Compare and Swap）操作：在编程语言层面，通过CAS操作来比较内存中的值与预期值是否相等，如果相等则修改，否则放弃修改。
   使用乐观锁和悲观锁的选择取决于应用场景和需求：悲观锁适合在并发冲突频繁的情况下，通过独占资源避免并发问题，但会对系统性能产生一定的影响。乐观锁适合在并发冲突较少的情况下，通过乐观的并发控制机制提高系统性能，但需要处理冲突的情况。在实际使用时，需要根据具体业务场景和需求选择适当的并发控制机制，并注意处理冲突和回滚事务的策略，以确保数据的一致性和完整性。
- 死锁问题，如何避免死锁
	- 死锁的条件：
		- 事务并发执行：多个事务同时操作相同的数据，请求相同或不同的锁资源。
		- 锁竞争：事务之间竞争相同的资源而产生死锁。
		- 不同的锁顺序：不同的事务以不同的顺序请求锁资源，导致死锁。
	- 避免死锁的方法：
		- 统一锁资源访问顺序：对于需要操作多个锁资源的事务，保持统一的访问顺序，避免不同事务之间出现交叉的锁请求顺序
		- 减少事务持有时间：尽量将事务的持有时间缩短，减少锁资源的占用时间，降低死锁的概率。
		- 使用合理的索引：合理的索引设计可以减少查询中的锁竞争，提高并发性能，减少死锁的可能性。
		- 限制事务并发度：通过调整事务的并发度，限制同时执行的事务数量，减少锁竞争的机会。
- **分布式事务**
  - https://juejin.cn/post/6844903647197806605
  - https://www.cnblogs.com/jajian/p/10014145.html

 
 ## 数据库调优
 -  **优化的步骤**
    - 考虑数据量大导致的性能问题，访问量大导致的性能问题？
    - sql语句优化。分析执行计划，减少load的数据量
    - 考虑能否通过增加索引优化查询效率，检查索引是否生效
    - 是否有缓存
    - 垂直分表、水平分表、分库
    - 根据场景来看，写操作多的情况下，考虑读写分离
    - [数据归档](https://www.cnblogs.com/goodAndyxublog/p/14994451.html)：数据是否有冷热的区别，例如订单数据有比较明显的时间冷热的区别，可以考虑冷数据归档。比如半年前的订单数据可以写入hbase
    - 池化
  
- **架构优化**
    - 分库，分表。垂直分，水平分。依据QPS和耗时，服务端最大并非连接数量
    - 读写分离
    - 批量读写，批量更新
    - 异步写，写平滑
    - 缓存优化
    - 历史数据归档

- **连接池的配置和使用**
    - 连接池能减少连接创建和释放带来的开销，大多数SDK也支持是支持连接池的，通常实际生产环境中也都会使用到连接池，需要关注一下几个参数
    - max_idle_connections: 最大空闲连接数
    - max_open_connections: 最大连接数
    - connection_max_lifetime: 连接最大可重用时间
    - 要使用好连接池，除了关注客户端的配置还需要关注mysql服务端的配置
    - 服务端最大连接数量：show variables like '%connection%'; max_connections
    - 服务端连接最大生命周期：show variables like '%wait_timeout%'
      ```
	      最大空闲连接数 =（QPS*请求平均耗时）/ 应用节点个数
	      最大连接数 =（QPS*请求最大耗时）/ 应用节点个数
              客户端连接maxlifetime < 数据库服务端设置的connection_max_lifttime
      ```

- **慢sql优化**
    - 慢查询问题，查看慢查询设置的阈值。show variables like '%long_query%';
    - 打开慢查询日志
    - 分析数据sql的结构是否加载了不必要的字段和数据
    - [深度分页查询优化](https://juejin.cn/post/7012016858379321358)
    - 子查询和连接查询
     ```
       	explain select * from test_xxxx_tab txt order by id limit 10000,10;
	explain SELECT * from test_xxxx_tab txt where id >= (select id from test_xxxx_tab txt order by id limit 10,1) limit 10;
       id列：在复杂的查询语句中包含多个查询使用id标示
       select_type:select/subquery/derived/union
       table: 显示对应行正在访问哪个表
       type：访问类型，关联类型。非常重要，All,index,range,ref,const,
       possible_keys: 显示可以使用哪些索引列
       key列：显示mysql决定使用哪个索引来优化对该表的访问
       key_len：显示在索引里使用的字节数
       rows：为了找到所需要的行而需要读取的行数
     ```
   - 慢查询日志样例子
   ```
   	# Time: 2022-05-10T10:15:32.123456Z
    # User@Host: myuser[192.168.0.1] @ localhost []  Id: 12345
    # Query_time: 3.456789  Lock_time: 0.123456 Rows_sent: 10  Rows_examined: 100000
    SET timestamp=1657475732;
    SELECT * FROM orders WHERE customer_id = 1001 ORDER BY order_date DESC LIMIT 10;
    这个慢查询日志示例包含以下重要的信息：

    时间戳（Time）: 日志记录的时间，以 UTC 时间表示。
    用户和主机（User@Host）: 执行查询的用户和主机地址。
    连接 ID（Id）: 表示执行查询的连接 ID。
    查询时间（Query_time）: 查询执行所花费的时间，以秒为单位。
    锁定时间（Lock_time）: 在执行查询期间等待锁定资源所花费的时间，以秒为单位。
    返回行数（Rows_sent）: 查询返回的结果集中的行数。
    扫描行数（Rows_examined）: 在执行查询过程中扫描的行数。
    时间戳（SET timestamp）: 查询开始执行的时间戳。
    查询语句（SELECT * FROM orders WHERE customer_id = 1001 ORDER BY order_date DESC LIMIT 10）: 实际执行的查询语句
   ```
- **index优化** 
    - 会查看sql执行计划explain
    - 关注：type、const、ref
    - 关注：extra等字段

- **使用缓存优化DB需要考虑的问题**
    - 缓存更新、过期、淘汰的策略
    - 缓存可能遇到的三大问题，雪崩、穿透、击穿
    - 缓存和db的一致性问题，[缓存更新策略及其分析？](https://zhuanlan.zhihu.com/p/86396877),业界比较通用的先更新DB，再删除cache

- **库表优化/分表/分库**
    - 垂直分表
    - 水平分表
    - 分库
    - 业界成熟的方案

- **架构优化读写分离优化**
    - 在写操作的较多的情况可以考虑数据库读写分离的方案
    - [业界的方案](https://www.cnblogs.com/wollow/p/10839890.html),代理实现和业务实现

- **核心监控告警指标**
  - read write qps 监控/select/update/insert
  - connections
  - thread
  - InnoDB buffer pool
  - 慢查询监控
  - 网络流量IO
  - 读写分离架构时需要监控主从延时


- **关键配置查看**
    ```
    show global variables;
    show variables like '%max_connection%'; 查看最大连接数
    show status like  'Threads%';
    show processlist;
    show variables like '%connection%';
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


    
## MySQL多表关联查询 vs 多次单表查询service组装
- 多次单表查询+Service组装：
	- 灵活性：多次单表查询+Service组装方式更加灵活，可以根据具体需求灵活组装和调整查询逻辑，适应各种复杂的查询需求。
	- 可扩展性：通过多次单表查询和Service组装，可以将查询逻辑分解为多个简单的查询，有助于代码的模块化和可扩展性，方便后续的维护和修改。
	- 缓存利用：多次单表查询+Service组装方式可以更好地利用缓存，针对每个单表查询的结果进行缓存，提高查询性能
  https://www.zhihu.com/question/68258877

## mysql binlog
- https://zhuanlan.zhihu.com/p/33504555
- show global variables like "binlog%";

## show processlist;
- https://zhuanlan.zhihu.com/p/30743094

## 常用命令
   - mysql登陆：
        mysql -h主机 -P端口 -u用户 -p密码
        SET PASSWORD FOR 'root'@'localhost' = PASSWORD('root');
        create database wxquare_test;
        show databases;
        use wxquare_test;
   - 查看见表sql：show create table table_name;
   - show variables like '%timeout%';
   - update json 文本需要转义
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
   - truncate table 属于ddl语句，需要ddl的权限
   - mysqldump 库表结构
	```
      mysqldump --column-statistics=0 -hhost -PPort -uuser_name -ppassword --databases -d db_name --skip-lock-tables --skip-add-drop-table --set-gtid-purged=OFF | sed 's/ AUTO_INCREMENT=	[0-9]*//g' > db.sql
     ```
	- 批量更新
       	```
	UPDATE employees
	SET salary = CASE
	    WHEN grade = 'A' THEN salary * 1.1
	    WHEN grade = 'B' THEN salary * 1.05
	    WHEN grade = 'C' THEN salary * 1.03
	    ELSE salary
	END
	WHERE department = 'IT';
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
- https://cyborg2077.github.io/2023/05/06/InQMySQL/

