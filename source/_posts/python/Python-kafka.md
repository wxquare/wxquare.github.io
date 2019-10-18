---
title: Python 使用kafka
categories:
- Python
---

## 一、kafka使用场景
kafka是一个分布式消息队列，具有高性能、持久化、多副本备份，横向扩展能力。

1. 解耦：广告日志数据的上报
2. 发布/订阅：订单系统
3. kafka+AI算法：由于AI算法执行时间较长，不适合使用http协议。将请求消息暂存在kafka中，AI算法从kafka获取请求不断执行。

## 二、基本概念
- broker：
- 生产者producer：
- 消费者consumer：
- 主题topic：
- 分区partition：kafka通过分区将topic消息分散在不同server的多个磁盘上，来提高消息发送和消费的效率。
- 消费者分组group：kafka中同一个consumer group id下的多个消费者共同消费一个topic下的消息。
- 偏移量offset：消息messge存储在kafka的broker中，消费者拉去消息时需要知道消息在文件中偏移量。


## 三、Python使用kafka
安装kafka-python
pip3 install kafka-python
https://zhuanlan.zhihu.com/p/38330574
https://kafka-python.readthedocs.io/en/master/usage.html


## 四、最简单的demo
### producer
```
brokers = ['host1:9092','host2:9092','host3:9092']
cliend_id = 'test1_client'
topic = 'test1'
from kafka import KafkaProducer
producer = KafkaProducer(client_id=cliend_id,bootstrap_servers=brokers)
future = producer.send(topic, value= b'hello,world')
result = future.get(timeout= 10)
print(result)

```

### consumer
```
brokers = ['host1:9092','host2:9092','host3:9092']
consumer_group_id = 'test1_client'
topic = 'test1'
consumer = KafkaConsumer(topic, group_id=consumer_group_id, bootstrap_servers=brokers)
tasks = []
for msg in consumer:
    print(msg)
```

## 五、常见问题汇总
1. consumer设置超时时间：consumer_timeout_ms
2. 不同consumer group 消费同一个topic的数据
3. consumer reblance机制导致数据重复消费：https://blog.csdn.net/yhyr_ycy/article/details/88960072

