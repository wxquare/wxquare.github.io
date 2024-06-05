
---
title: 互联网系统设计 - API 设计规范和管理
date: 2024-04-15
categories: 
- 系统设计
---


## API 设计规范和管理
### API 架构风格
- RESTful API
- GraphQL 
- RPC
- SOA

### RESTful API 
- 路径名称避免动词
```sh
路径名称避免动词
# Good
curl -X GET /orders
# Bad
curl -X GET /getOrders
```

- GET 获取指定 URI 的资源信息
```sh
# 代表获取当前系统的所有订单信息
curl -X GET /orders

curl -X GET /users/{user_id}/orders

# 代表获取指定订单编号为订单详情信息
curl -X GET /orders/{order_id}
```

- POST 通过指定的 URI 创建资源
```sh
curl -X POST /orders \
  -d '{"name": "awesome", region: "A"}' \
```

- PUT 创建或全量替换指定 URI 上的资源
```
curl -X PUT http://httpbin.org/orders/1 \
  -d '{"name": "new awesome", region: "B"}' \
```

- PATCH 执行一个资源的部分更新
```sh
# 代表将 id 为 1 的 order 中的 region 字段进行更改，其他数据保持不变
curl -X PATCH /orders/{order_id} \
  -d '{name: "nameB"}' \
curl -X order/{order_id}/name (用来重命名)
curl -X /order/{order_id}/status(用来更改用户状态)
```

- DELETE 通过指定的 URI 移除资源
```sh
# 代表将id的 order 删除
curl -X DELETE /orders/{order_id}
```

其它规则：
规则1：应使用连字符（ - ）来提高URI的可读性
规则2：不得在URI中使用下划线（_）
规则3：URI路径中全都使用小写字母


### API 错误码设计规范
1. 不论请求成功或失败，始终返回 200 http status code，在 HTTP Body 中包含用户账号没有找到的错误信息:

```
如: Facebook API 的错误 Code 设计，始终返回 200 http status code：
{
  "error": {
    "message": "Syntax error \"Field picture specified more than once. This is only possible before version 2.1\" at character 23: id,name,picture,picture",
    "type": "OAuthException",
    "code": 2500,
    "fbtrace_id": "xxxxxxxxxxx"
  }
}

缺点:
  对于每一次请求，我们都要去解析 HTTP Body，从中解析出错误码和错误信息

```

2. 返回 http 404 Not Found 错误码，并在 Body 中返回简单的错误信息:

```
如: Twitter API 的错误设计
根据错误类型，返回合适的 HTTP Code，并在 Body 中返回错误信息和自定义业务 Code

HTTP/1.1 400 Bad Request
{"errors":[{"code":215,"message":"Bad Authentication data."}]}
```

3. 返回 http 404 Not Found 错误码，并在 Body 中返回详细的错误信息:

```
如: 微软 Bing API 的错误设计，会根据错误类型，返回合适的 HTTP Code，并在 Body 中返回详尽的错误信息
HTTP/1.1 400
{
  "code": 100101,
  "message": "Database error",
  "reference": "https://github.com/xx/tree/master/docs/guide/faq/xxxx"
}
```

4. 业务 Code 码设计
- 纯数字表示
- 不同部位代表不同的服务
- 不同的模块（品类）

```
如: 错误代码说明：100101
10: 服务
01: 某个服务下的某个模块
01: 模块下的错误码序号，每个模块可以注册 100 个错误
建议 http status code 不要太多:

200 - 表示请求成功执行
400 - 表示客户端出问题
500 - 表示服务端出问题

如果觉得这 3 个错误码不够用，可以加如下 3 个错误码:
401 - 表示认证失败
403 - 表示授权失败
404 - 表示资源找不到，这里的资源可以是 URL 或者 RESTful 资源
```

### 接口幂等性设计
#### 幂等性的重要性
- 网络不稳定，导致的重试。在网络不稳定的情况下，可能会重试请求。幂等性确保重复请求不会导致意外的副作用。
- 简化客户端代码：客户端不需要担心重复请求的副作用，从而简化了错误处理逻辑。
- 人工误操作，确保用户操作的可预测性，避免因重复提交表单等操作导致的错误或重复数据。
- 重复请求是返回错误，还是返回成功一样的内容呢？

### 业务场景：
- 用户商品上传，同步供应商数据
- 订单创建、支付等
- 消费MQ，异步发货等

#### 怎么实现幂等性
- 前端防抖、置灰
- 前端生成一个唯一ID（比较少）
- 幂等键：后端生成唯一幂等键；根据请求参数，识别幂等键
- 幂等表：通常是redis，后端根据请求参数+时间戳的方式，判断是否是重复请求
- 在高并发的场景下， 牢记一锁二判三更新。锁：分布式锁、数据库锁。判：状态、版本检查。
- 数据库唯一性约束兜底


[接口幂等性设计]https://blog.dreamfactory.com/what-is-idempotency/