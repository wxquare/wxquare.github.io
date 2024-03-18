


## 总体架构图
![Alt text](image.png)

展示层：各个端口：买家卖家管理端，已经三方应用：支付宝微信。
负载均衡：可以使用nginx或其他中间件做负载均衡
网关：使用spring gateway作为网关，使用Sentinel做限流、熔断、降级
微服务：拆分为会员、订单、商品等服务，使用nacos作为服务注册和配置中心，用openfeign做服务调用，使用seata做分布式事务
数据层：Mysql、Redis 、RabbitMq、Elasticsearch和XXL-Job(任务调度)




Admin管理测
商品类目管理 category
模型，category_tab
admin 页面:https://dp-admin.test.shopee.io/id/category/list
RESTful的增删查改：add(post),update(put),delete(delete),get(GET)
8个一级类目、12个二级类目，57个三级类目
三级类目放在同一张表中，用parent_category 关联？（思考：分三张表存储的优劣势）
供应商管理 carrier
模型：carrier_tab
2000多个carrier
设计时和category一样
DP 首页entrance管理
fe_group_tab
fe_category_tab
fe_configuration_snatshot_tab
fe_operation_tab
entrance：
group和category的管理
快照表和操作表（Snapshot Table 和 Operation Log Table）
DB→redis→CDN
商品批量上传、编辑
订单列表、详情、数据下载
用户APP测核心功能：首页entrance设计和实现
DB数据模型：
fe_group_tab
fe_category_tab
fe_configuration_snatshot_tab
fe_operation_tab
接口设计：
entracne：
group的编辑
group 新增活着减少category
group 排序
entracne group的release，redis，cdn文件，为了避免热键，这里会根据用户id哈希指定到不同的redis
在线接口访问，访问redis，注意防止热键
CDN 数据地址：
https://deo.shopeemobile.com/shopee/test/digital-product/entrance/render/json/entrance-group-test-my-v2.json
rediskey:
get dp:entrance_snapshot_1_5:test:my
用户APP测核心功能：商品检索、列表、详情设计和实现
商品模型，分类化管理
category + carrier + item
category + carrier + item + 业务表
难点：不同品类商品模型差异较大，如果管理商品的信息？
类似自营商品，local上传上传商品、描述信息、价格、库存等（有些商品没有库存的概念，例如bill，有些是有券是有库存的）存储商品静态信息+价格+库存
只管理商品content，不管理价格和库存（hotel，movie)，存储商品静态信息+缓存价格+库存
既不管理content，也不管理价格和库存，只存储必要信息（train，bus，flight)，只存必要的信息，缓存供应商信息
商品快照和操作日志
local编辑日志，方便追溯问题，提供check机制
商品快照表
商品操作日志
供应商商品静态信息同步
难点：如何保证供应商数据的实时性？
全量同步
增量同步
热门商品，热门商品定期同步。刷新规则类似
供应商动态报价缓存同步刷新规则 
难点：如何保证供应商数据的实时性？
特征集合：酒店、提前购买时间（1-30）、间夜（1,2,3,4,5)、成人数量(1,2,3,4)、是否节假日
最小刷新间隔10min，最大刷新间隔 7 *days  (10080min)
归一化：
酒店百分位10%（1），
提前购买时间（1-2）(1),
间夜（2).
成人数量。2
是否节假日。是（1），不是（0.6）
key: hotelid_checkin_nights_adult_isholiday.  val: price, last_updatetime.
hbase+redis的二级缓存
检索、过滤、排序
ES增量索引+索引
内存过滤和排序
用户APP测核心功能：订单系统设计和实现
订单模型
一个支付单pay_order_id，对应多个order_tab
一个order对应多个item
订单状态机
service/common/utils/order_status.go
创单过程（checkout)
创单限制频率
请求参数校验
价格校验、库存校验
订单ID,pay_order_id,order_id 生成，时间戳+ 机器pod ID. + 用户ID后三位 + 每秒生成的id，src/git.garena.com/shopee/digital-purchase/service/order-server/util/order_id/order_id_generator.go
营销活动查询、（有些品类evoucher）库存扣减
支付(paynow)
参数校验
支付渠道校验和优惠查询
营销、coins 扣减，(不同团队，分布式事务)
请求spm支付初始化
更新订单状态为pending P1
用户触发 user → pay → spm → PayOrderProcessor.            internal/payment/result
支付回调
成功
失败
cancel
支付过期
用户长时间不支付
订单cancel
用户主动cancel订单
履约发货
支付完成事件触发，先请求hub获取Hub Reference Id
然后请求api/v2/orders 发货
发货结果回调
   callback api/v1/fulfillment/status
   成功
  失败
退货
履约失败或者用户主动发起
call hub api/v1/orders/return
退货回调
hub callback refund/notification
成功
失败
退款
请求spm  api/v3/transaction/refund
退款回调
请求spm  api/v3/transaction/refund
中间件和使用场景
数据库
为什么不使用mysql的join功能，而是业务join？
ES:
redis。有哪些缓存
kafka
Hive
hbase

## 可用性做了哪些工作？
首页的分布式存储、降级策略
性能方面做了哪些工作？
效率方便做了哪些工作？
难点1:   如何解决不同品类差距较大的商品和订单系统？
难点2：如何解决大规模商品数据近实时同步问题？（监控指标）
全量同步（保证全）
增量同步 （大部分没有这个接口）
根据商品的热门程度设置不同的刷新规则？（保证热门数据的准确性）
难点3：如何解决数据一致性问题？（核心监控指标）
微服务架构中的数据一致性：解决方案与实践| 得物技术
系统内部：
不同表之间事物
更新时使用乐观锁，where条件检查状态
系统之间：spm 支付 （最终一致性）
发起支付，等待回调
支付过期
补偿任务：CreateAutoUpdatePayOrderPaymentStatusJob，listPaymentPendingOrders->handlePaymentPendingOrders→GetTransactionV3。查询支付状态，更新订单状态
支付幂等设计
// 退货退款类任务
系统之间：hub 履约 （最终一致性）
获取referenceid
调用hub v1/orders履约
失败重试
补偿任务：CreateReFulfillmentJob，分别处理 processNotFulfilled，processPending 任务，继续调用 v1/orders 接口创单
hub 发货幂等设计
系统之间：coin 扣减 （tcc 事务）
https://boxiaoyang.club/article/163
coin消费行为 API 迁移至 TCC API
coin 二阶段confirm 火车cancel 失败不影响支付，提供查询接口给css团队回调
秒杀系统，库存问题：mysql和redis数据一致性问题，乐观锁更新mysql，然后更新redis，设置过期时间
https://gongfukangee.github.io/2019/06/09/SecondsKill/
限流，redis + mysql，库存缓存，乐观锁
12306的异步机制，接受请求，异步创单

## 难点4:  在系统可用性方便做了哪些工作？（核心指标）
浅谈系统稳定性与高可用保障的几种思路
架构上设计和服务拆分：
支持异地多活
支持横向扩容
中间件选型
容量评估，mysql 一写多读，codis，kafka，ES
灾备，快速恢复
资源个隔离和拆分：
mysql分库，分表、kafka topic，es 集群
逻辑架构和物理架构分离，订单系统支持根据业务类型路由
离线和在线分离
功能设计时：
接口维度的限流、用户维度限流
避免单点：比如在主页设计时，主页配置数据需要写在多个redis中
核心功能降级策略：redis→cdn
监控和告警的设计：
metric->trace->log
变更流程：
可灰度
可回滚
资损防控



## 在实际开发中遇到哪些挑战以及使用了哪些框架、技术、组件？

### API开发规范和管理
- RESTful API + swagger + yapi
- https://www.restapitutorial.com/
- https://juejin.cn/post/7126802030944878600

### DDD 
- 模型设计
- https://tech.meituan.com/2017/12/22/ddd-in-practice.html
- 

### 限流保护
- 单机限流
- 分布式限流

### 分布式锁
- redis 分布式锁
- 分布式锁：src/git.garena.com/shopee/digital-purchase/service/common/distributed-lock-task/distributed_lock.go

### 在实际开发中遇到哪些挑战以及使用了哪些组件
- nginx: http://nginx.org/en/download.html
- gin: https://github.com/gin-gonic/gin
- rpc 框架
- zookeeper 
- mysql: https://www.mysql.com/
    - 数据库客户端工具: https://dbeaver.io
    - xorm: https://pkg.go.dev/xorm.io/xorm
    - 建标、索引
    - 分库、分表
    - 快照表（Snapshot Table）和操作日志表（Operation Log Table）是常见的数据库设计模式
- redis
    - 缓存
    - 分布式锁
    - bloomfilter
- elasticsearch
    - es客户端工具：https://github.com/mobz/elasticsearch-head
- kafka
    - kafka 客户端工具：https://kafka.apache.org/downloads，下载之后bin目录下面的脚本
- 日志采集
    - Elasticsearch + logstash + kibana
- 监控prometheus+grafna
    


在实际开发中遇到哪些挑战以及使用了哪些工具？
- 绘图工具：https://app.diagrams.net/
- uml: 时序图，对象图等 https://plantuml.com。vscode+platuml
- Postman: https://www.postman.com/
- Charles: https://www.charlesproxy.com/


参考：
- http://doc.javashop.cn/docs/7.3.0/achitecture/overview
- https://github.com/Exrick/xmall
- https://github.com/linlinjava/litemall
- https://github.com/macrozheng/mall
