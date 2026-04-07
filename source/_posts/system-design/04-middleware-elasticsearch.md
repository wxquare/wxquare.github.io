---
title: 中间件 - 搜索和 Elasticsearch
date: 2024-03-07
categories:
- 系统设计
tags:
- Elasticsearch
- 搜索引擎
- 倒排索引
- DSL
toc: true
---

## 速查导航

**阅读时间**: 40 分钟 | **难度**: ⭐⭐⭐⭐ | **面试频率**: 高

**核心考点速查**:
- [基本使用](#基本使用) - Index 创建、Mapping 设计、Query DSL
- [倒排索引原理](#倒排索引深度解析面试必问) - 倒排索引结构、Analyzer 工作流程、评分算法
- [深分页问题](#深分页问题与解决方案) - scroll、search_after、分页优化
- [高可用与脑裂](#高可用与脑裂问题) - 脑裂原因、quorum 机制、解决方案
- [性能优化](#性能优化) - 写入性能、查询性能、索引优化
- [面试高频 20 题](#面试高频-20-题) - 标准答案 + 追问应对

---

## 基本使用
### 创建index，setting和mapping
```
curl -XPUT -H'Content-Type: application/json'  host/index_name?pretty=true -d@index_mapping.json 
```

<details>
  <summary>es index example</summary>
  ```json
 {
    "settings": {
        "index": {
            "number_of_shards": "5",
            "number_of_replicas": "2"
        },
        "analysis": {
            "filter": {
                "t2sconvert": {
                    "convert_type": "t2s",
                    "type": "stconvert"
                }
            },
            "analyzer": {
                "traditional_chinese_analyzer": {
                    "filter": "t2sconvert",
                    "type": "custom",
                    "tokenizer": "ik_smart"
                }
            },
            "normalizer": {
                "lowercase": {
                    "type": "custom",
                    "filter": [
                        "lowercase"
                    ]
                }
            }
        }
    },
    "mappings": {
        "_doc": {
            "properties": {
                "key_type": {
                    "type": "integer"
                },
                "country": {
                    "type": "keyword"
                },
                "language_code": {
                    "type": "keyword"
                },
                "level": {
                    "type": "integer"
                },
                "code": {
                    "type": "keyword"
                },
                "is_available": {
                    "type": "boolean"
                },
                "is_popular": {
                    "type": "boolean"
                },
                "pop_rank": {
                    "type": "integer"
                },
                "name": {
                    "properties": {
                        "value_in_chinese": {
                            "type": "text",
                            "fields": {
                                "keyword": {
                                    "type": "keyword"
                                }
                            },
                            "analyzer": "ik_max_word",
                            "search_analyzer": "ik_smart"
                        },
                        "value_in_english": {
                            "type": "text",
                            "fields": {
                                "keyword": {
                                    "type": "keyword",
                                    "normalizer": "lowercase"
                                }
                            }
                        }
                    }
                },
                "display_name": {
                    "properties": {
                        "value_in_chinese": {
                            "type": "text",
                            "fields": {
                                "keyword": {
                                    "type": "keyword"
                                }
                            },
                            "analyzer": "ik_max_word",
                            "search_analyzer": "ik_smart"
                        },
                        "value_in_english": {
                            "type": "text",
                            "fields": {
                                "keyword": {
                                    "type": "keyword",
                                    "normalizer": "lowercase"
                                }
                            }
                        }
                    }
                },
                "address": {
                    "properties": {
                        "value_in_chinese": {
                            "type": "text",
                            "fields": {
                                "keyword": {
                                    "type": "keyword"
                                }
                            },
                            "analyzer": "ik_max_word",
                            "search_analyzer": "ik_smart"
                        },
                        "value_in_english": {
                            "type": "text",
                            "fields": {
                                "keyword": {
                                    "type": "keyword",
                                    "normalizer": "lowercase"
                                }
                            }
                        }
                    }
                },
                "city_code": {
                    "type": "keyword"
                },
                "city_name": {
                    "properties": {
                        "value_in_chinese": {
                            "type": "text",
                            "fields": {
                                "keyword": {
                                    "type": "keyword"
                                }
                            },
                            "analyzer": "ik_max_word",
                            "search_analyzer": "ik_smart"
                        },
                        "value_in_english": {
                            "type": "text",
                            "fields": {
                                "keyword": {
                                    "type": "keyword",
                                    "normalizer": "lowercase"
                                }
                            }
                        }
                    },
                    "updated": {
                        "type": "date",
                        "format": "strict_date_optional_time||epoch_millis"
                    }
                }
            }
        }
    }
}
  ```
</details>

### 查看index,_cat 基本信息
```
curl -XGET 'host/_cat/indices/*hotel_basic_info_v2_live*(支持正则表达式）?v=true&pretty=true'
```

### 查看索引mapping信息
```
curl -XGET 'host/index_name/_mapping?pretty=true'
```

### 查看索引的setting信息
```
curl -XGET 'host/index_name/_settings?pretty=true'
```

### 通过doc id 正向查询
```
curl -XGET  'host/index/_doc/doc_id?pretty=true'
```

### query,search，倒排查询
```
curl -XPOST -H'Content-Type: application/json' 'host/index_name/_search?pretty=true' -d '{
"query":{}}'
```


### update
```
curl -XPOST  -H'Content-Type: application/json' 'host/index/_doc/doc_id/_update' -d '{
"doc": {
    "price": "6500000001"
}
}'
```

### 聚合count查询
```
curl -XPOST -H'Content-Type: application/json' 'host/index_name/_count' -d '{
    "query": {
        "term": {
            "city_name.value_in_english.keyword": "Jakarta"
        }
    }
}'
```

### 增加字段
```
curl -XPOST -H'Content-Type: application/json' 'host/index_name/_doc/_mapping' -d '{
    "properties": {
        "facility_codes": {
            "type":"keyword"
        }
    }
}'

```


### analyzer
- 参考：https://www.elastic.co/guide/en/elasticsearch/reference/current/analysis-index-search-time.html

    Elastic Search 在处理 Text 类型数据的时候，会把数据交给分词器处理。然后根据分词器给的词，建立倒排索引，通常一句话都由若干词语组成，分词结果会极大的影响到查询结果的质量
在 Elastic Search 中，分词器起到了非常重要的作用，在定义文档结构、录入和更新文档、查询文档的时候都会用到它。例如：
```
武汉市长江大桥欢迎您

默认分词器：
[武, 汉, 市, 长, 江, 大, 桥, 欢, 迎, 您]

普通分词器：
[武汉, 市, 武汉市, 长江, 大桥,长江大桥, 欢迎, 您, 欢迎您]

二哈分词器：
[武汉, 市长, 江大桥, 欢迎, 您]
```

### normalizer 
- 参考：https://www.elastic.co/guide/en/elasticsearch/reference/current/analysis-index-search-time.html



### alias
```
POST /_aliases
{
  "actions": [
    {"remove": {"index": "l1", "alias": "a1"}},
    {"add": {"index": "l1", "alias": "a2"}}
  ]
}
```

```
curl -XPUT  host/index_nane/_alias/index_alias_name
```
### query DSL
- term level queries
	- keyword term
    - https://www.elastic.co/guide/en/elasticsearch/reference/6.7/term-level-queries.html
- full text queries
	- Match Phrase Query
    - Mathc Query
- compound queries
   - dismax
   - bool
   - function score
   - boosting query

- https://www.elastic.co/guide/en/elasticsearch/reference/6.7/query-dsl.html

<details>
  <summary>es query dsl</summary>
```
{
    "query": {
        "function_score": {
            "functions": [
                {
                    "filter": {
                        "term": {
                            "key_type": 1
                        }
                    },
                    "weight": 3
                },
                {
                    "filter": {
                        "term": {
                            "key_type": 2
                        }
                    },
                    "weight": 2
                }
            ],
            "min_score": 0,
            "query": {
                "dis_max": {
                    "queries": [
                        {
                            "function_score": {
                                "functions": [
                                    {
                                        "filter": {
                                            "term": {
                                                "search_key.filed1.keyword": "querywords"
                                            }
                                        },
                                        "weight": 100
                                    },
                                    {
                                        "filter": {
                                            "term": {
                                                "search_key.filed2.keyword": "querywords"
                                            }
                                        },
                                        "weight": 100
                                    }
                                ],
                                "score_mode": "max"
                            }
                        },
                        {
                            "dis_max": {
                                "queries": [
                                    {
                                        "match_phrase": {
                                            "search_key.filed1": {
                                                "boost": 50,
                                                "query": "querywords"
                                            }
                                        }
                                    },
                                    {
                                        "match_phrase": {
                                            "search_key.field2": {
                                                "boost": 50,
                                                "query": "querywords"
                                            }
                                        }
                                    }
                                ]
                            }
                        },
                        {
                            "dis_max": {
                                "queries": [
                                    {
                                        "prefix": {
                                            "search_key.filed1.keyword": {
                                                "boost": 30,
                                                "value": "querywords"
                                            }
                                        }
                                    },
                                    {
                                        "prefix": {
                                            "search_key.field2.keyword": {
                                                "boost": 30,
                                                "value": "querywords"
                                            }
                                        }
                                    }
                                ]
                            }
                        },
                        {
                            "dis_max": {
                                "queries": [
                                    {
                                        "match": {
                                            "search_key.field1": {
                                                "boost": 10,
                                                "fuzziness": "auto:6,20",
                                                "minimum_should_match": "3>75%",
                                                "query": "querywords"
                                            }
                                        }
                                    },
                                    {
                                        "match": {
                                            "search_key.field2": {
                                                "boost": 10,
                                                "fuzziness": "auto:6,20",
                                                "minimum_should_match": "3>75%",
                                                "query": "querywords"
                                            }
                                        }
                                    }
                                ]
                            }
                        }
                    ]
                }
            }
        }
    },
    "size": 30,
    "sort": [
        {
            "_score": {
                "order": "desc"
            }
        },
        {
            "key_type": {
                "order": "asc"
            }
        },
        {
            "others": {
                "order": "desc"
            }
        }
    ]
}
```
</details>

## 原理
### 基本概念
- 节点：分布系统都有的master节点和普通节点。类似于kafka集群都会存在的一种节点
- master节点:用于管理索引（创建索引、删除索引）、分配分片，维护元数据
- 协调节点：ES的特殊性，需要由一个节点汇总多个分片的query结果。节点是否担任协调节点可通过配置文件配置。例如某个节点只想做协调节点：node.master=false，node.data=false
- ES的读写流程主要是协调节点，主分片节点、副分片节点间的相互协调。
- ES的读取分为GET和Search两种操作。GET根据文档id从正排索引中获取内容；Search不指定id，根据关键字从倒排索引中获取内容。

### 写单个文档的流程
1. 客户端向集群中的某个节点发送写请求，该节点就作为本次请求的协调节点
2. 协调节点使用文档ID来确定文档属于某个分片，再通过集群状态中的内容路由表信息获知该分片的主分片位置，将请求转发到主分片所在节点；
3. 主分片节点上的主分片执行写操作。如果写入成功，则它将请求并行转发到副分片所在的节点，等待副分片写入成功。所有副分片写入成功后，主分片节点向协调节点报告成功，协调节点向客户端报告成功。

### 读取单个文档的流程
1. 客户端向集群中的某个节点发送读取请求，该节点就作为本次请求的协调节点；
2. 协调节点使用文档ID来确定文档属于某个分片，再通过集群状态中的内容路由表信息获知该分片的副本信息，此时它可以把请求转发到有副分片的任意节点读取数据。
3. 协调节点会将客户端请求轮询发送到集群的所有副本来实现负载均衡。
4. 收到读请求的节点将文档返回给协调节点，协调节点将文档返回给客户端

### Search流程
ES的Search操作分为两个阶段：query then fetch。需要两阶段完成搜索的原因是：在查询时不知道文档位于哪个分片，因此索引的所有分片都要参与搜索，然后协调节点将结果合并，在根据文档ID获取文档内容。

**Query查询阶段** 
1. 客户端向集群中的某个节点发送Search请求，该节点就作为本次请求的协调节点；
2. 协调节点将查询请求转发到索引的每个主分片或者副分片中；
3. 每个分片在本地执行查询，并使用本地的Term/Document Frequency信息进行打分，添加结果到大小为from+size的本地有序优先队列中；
4. 每个分片返回各自优先队列中所有文档的ID和排序值给协调节点，协调节点合并这些值到自己的优先队列中，产生一个全局排序后的列表。

**Fetch拉取阶段**
query节点知道了要获取哪些信息，但是没有具体的数据，fetch阶段要去拉取具体的数据。相当于执行多次上面的GET流程
1. 协调节点向相关的节点发送GET请求；
2. 分片所在节点向协调节点返回数据；
3. 协调阶段等待所有的文档被取得，然后返回给客户端。


## 倒排索引深度解析（面试必问）

### 什么是倒排索引？

**标准答案（30 秒）**：
> 倒排索引是搜索引擎的核心数据结构，类似书籍的"索引"。传统数据库是"文档 → 关键词"的映射（正排），倒排索引是"关键词 → 文档列表"的映射，可以快速定位包含指定关键词的所有文档。

**对比**：

| 索引类型 | 映射关系 | 查询方式 | 示例 |
|---------|---------|---------|------|
| **正排索引** | 文档 ID → 内容 | 根据 ID 查内容 | MySQL 主键索引 |
| **倒排索引** | 关键词 → 文档 ID 列表 | 根据关键词查文档 | Elasticsearch |

### 倒排索引结构

**示例文档**：

```json
{"id": 1, "content": "Elasticsearch is a search engine"}
{"id": 2, "content": "Lucene is a search library"}
{"id": 3, "content": "Elasticsearch is built on Lucene"}
```

**倒排索引表**：

| Term（关键词） | Document IDs（文档列表） | Frequency（词频） |
|---------------|------------------------|-----------------|
| elasticsearch | [1, 3] | 2 |
| search | [1, 2] | 2 |
| engine | [1] | 1 |
| lucene | [2, 3] | 2 |
| library | [2] | 1 |
| built | [3] | 1 |

**查询 "elasticsearch search"**：
1. 查倒排索引表：`elasticsearch` → [1,3]，`search` → [1,2]
2. 求交集：[1,3] ∩ [1,2] = [1]
3. 返回文档 ID=1

### Analyzer 工作流程

**Analyzer 三大组件**：

| 组件 | 作用 | 示例 |
|------|------|------|
| **Character Filter** | 字符预处理（去特殊字符） | `<html>` → 空 |
| **Tokenizer** | 分词 | `Elasticsearch is` → [`Elasticsearch`, `is`] |
| **Token Filter** | 词项后处理（小写、停用词） | [`Elasticsearch`, `is`] → [`elasticsearch`] |

**示例**：

```json
// 输入
"Elasticsearch is a SEARCH Engine!!!"

// Character Filter: 去掉特殊字符
"Elasticsearch is a SEARCH Engine"

// Tokenizer: 分词
["Elasticsearch", "is", "a", "SEARCH", "Engine"]

// Token Filter: 小写 + 去停用词(is, a)
["elasticsearch", "search", "engine"]
```

**常用 Analyzer**：

| Analyzer | 适用场景 |
|----------|---------|
| **standard** | 英文（按空格分词 + 小写） |
| **ik_smart** | 中文（粗粒度分词） |
| **ik_max_word** | 中文（细粒度分词） |

### 评分算法（BM25）

**面试追问：Elasticsearch 如何对搜索结果排序？**

Elasticsearch 使用 **BM25 算法**（Best Matching 25）计算相关性得分。

**核心公式**：

```
score(D, Q) = Σ IDF(qi) × TF(qi, D)
```

- **TF（Term Frequency）**：词频，词在文档中出现次数
- **IDF（Inverse Document Frequency）**：逆文档频率，词的稀有程度
  - `IDF = log((总文档数 - 包含该词的文档数 + 0.5) / (包含该词的文档数 + 0.5))`

**示例**：

查询 "elasticsearch search"，文档库有 100 篇文档：
- `elasticsearch` 出现在 10 篇文档中
- `search` 出现在 50 篇文档中

**IDF 计算**：
- `IDF(elasticsearch) = log((100-10+0.5)/(10+0.5)) ≈ 2.18`（稀有词，权重高）
- `IDF(search) = log((100-50+0.5)/(50+0.5)) ≈ 0.69`（常见词，权重低）

**结论**：包含 `elasticsearch` 的文档得分更高。


## 深分页问题与解决方案

### 深分页问题

**面试标准答案**：
> 深分页是指查询"第 1000 页，每页 10 条"时，Elasticsearch 需要在每个分片上查询 `10000 条`数据（from=10000, size=10），然后在协调节点排序后取 10 条。数据量大时会导致内存溢出和性能急剧下降。

**示例**：

```json
GET /index/_search
{
  "from": 10000,  // 第 1000 页
  "size": 10
}
```

**问题**：
- 5 个分片，每个分片需查询 10010 条数据
- 协调节点需汇总 `5 × 10010 = 50050` 条数据
- 内存占用：`50050 × 1KB ≈ 50MB`（单次查询）

### 解决方案1：Scroll API（快照滚动）

**适用场景**：导出大量数据、批量处理

**原理**：创建快照，保持查询上下文，依次滚动获取数据。

**示例**：

```json
// 1. 创建 scroll
POST /index/_search?scroll=5m  // 保持 5 分钟
{
  "size": 1000,
  "query": {"match_all": {}}
}

// 2. 滚动获取下一批
POST /_search/scroll
{
  "scroll": "5m",
  "scroll_id": "DXF1ZXJ5QW5kRmV0Y2gBAAAAAAAAAD4WYm9laVYtZndUQlNsdDcwakFMNjU1QQ=="
}

// 3. 清理 scroll
DELETE /_search/scroll
{
  "scroll_id": "DXF1ZXJ5QW5kRmV0Y2gBAAAAAAAAAD4WYm9laVYtZndUQlNsdDcwakFMNjU1QQ=="
}
```

**缺点**：
- 占用大量内存（保持查询上下文）
- 不支持实时数据（快照固定）
- 需要手动清理 scroll

### 解决方案2：search_after（推荐）

**适用场景**：实时滚动分页、无限滚动

**原理**：使用上一次查询的最后一条数据的排序值作为下一次查询的起点。

**示例**：

```json
// 1. 第一次查询
GET /index/_search
{
  "size": 10,
  "sort": [
    {"timestamp": "desc"},
    {"_id": "asc"}  // 唯一字段，避免排序值相同
  ]
}

// 返回：
// "hits": [
//   {"_id": "1", "sort": [1609459200000, "1"]},
//   {"_id": "2", "sort": [1609459199000, "2"]},
//   ...
//   {"_id": "10", "sort": [1609459190000, "10"]}
// ]

// 2. 下一页查询（使用上一页最后一条的 sort 值）
GET /index/_search
{
  "size": 10,
  "search_after": [1609459190000, "10"],  // 上一页最后一条的 sort
  "sort": [
    {"timestamp": "desc"},
    {"_id": "asc"}
  ]
}
```

**优势**：
- 内存占用小（无需保持上下文）
- 支持实时数据
- 性能稳定（不受深度影响）

### 对比总结

| 方案 | 适用场景 | 优点 | 缺点 |
|------|---------|------|------|
| **from+size** | 前 100 页 | 简单、支持跳页 | 深分页性能差 |
| **scroll** | 数据导出、批量处理 | 全量遍历 | 占用内存、不实时 |
| **search_after** | 实时分页、无限滚动 | 高性能、实时 | 不支持跳页 |


## 高可用与脑裂问题

### 脑裂问题

**面试标准答案**：
> 脑裂（Split Brain）是指集群因网络分区，导致选举出多个 Master 节点，各自管理集群的一部分，造成数据不一致。

**原因**：
1. 网络分区：Master 节点与其他节点网络不通
2. 其他节点认为 Master 挂了，重新选举新 Master
3. 旧 Master 恢复后，集群出现两个 Master

**示例**：

```
正常状态：
Node1 (Master) - Node2 - Node3

网络分区：
[Node1 (Master)] | [Node2 (新Master) - Node3]
     ↓                      ↓
  写入数据A               写入数据B
     ↓                      ↓
   数据不一致！
```

### 解决方案：quorum 机制

**配置**：

```yaml
discovery.zen.minimum_master_nodes: (N/2 + 1)  // N 为 Master 候选节点数
```

**示例**：
- 3 个 Master 候选节点：`minimum_master_nodes = 2`
- 5 个 Master 候选节点：`minimum_master_nodes = 3`

**原理**：
- 只有获得 **半数以上** 节点认可，才能成为 Master
- 网络分区后，只有"多数派"能选出 Master，"少数派"无法选举

**示例**：

```
3 个节点，minimum_master_nodes=2

网络分区：
[Node1] | [Node2 - Node3]
   ↓           ↓
 只有1个节点   有2个节点
 无法选举     可以选举新Master
```

### 其他高可用措施

1. **Discovery.zen.ping_timeout**：心跳超时时间（默认 3 秒）
2. **Discovery.zen.fd.ping_interval**：心跳间隔（默认 1 秒）
3. **分片副本**：每个主分片至少 1 个副本分片


## 面试高频 20 题

### 1. Elasticsearch 是什么？

**标准答案**：
Elasticsearch 是基于 Lucene 的分布式搜索引擎，支持全文检索、日志分析、实时监控等场景。核心特点：分布式、RESTful API、实时搜索、高可用。

### 2. 倒排索引是什么？

**标准答案**：
倒排索引是"关键词 → 文档列表"的映射，与传统数据库的"文档 ID → 内容"相反。通过倒排索引可以快速定位包含指定关键词的所有文档。

### 3. Elasticsearch 的写入流程？

**标准答案**：
1. 客户端请求协调节点
2. 协调节点根据文档 ID 路由到主分片
3. 主分片写入，并行转发到副分片
4. 所有副分片写入成功后，向协调节点报告
5. 协调节点向客户端返回结果

### 4. Elasticsearch 的查询流程？

**标准答案**：
分为 Query 和 Fetch 两阶段：
- **Query**：协调节点向所有分片发送查询，各分片返回 top N 的文档 ID 和得分
- **Fetch**：协调节点合并结果，根据文档 ID 向分片请求完整文档

### 5. 深分页问题如何解决？

**标准答案**：
1. **scroll**：适用于数据导出，创建快照滚动获取
2. **search_after**：适用于实时分页，使用上一页最后一条的排序值作为起点
3. **避免**：使用 from+size 深分页（性能差）

### 6. 脑裂问题是什么？如何解决？

**标准答案**：
脑裂是指网络分区导致选举出多个 Master，造成数据不一致。解决方案：设置 `minimum_master_nodes = N/2 + 1`，确保只有"多数派"能选举 Master。

### 7. Elasticsearch 和 Solr 的区别？

| 维度 | Elasticsearch | Solr |
|------|--------------|------|
| **实时性** | 近实时（1 秒） | 需要手动 commit |
| **分布式** | 原生支持 | 需要 ZooKeeper |
| **社区** | 更活跃 | 较老牌 |
| **适用场景** | 日志分析、实时搜索 | 企业搜索 |

### 8. Elasticsearch 的分片和副本是什么？

**标准答案**：
- **主分片（Primary Shard）**：数据的横向分区，创建后不可修改数量
- **副本分片（Replica Shard）**：主分片的备份，提供高可用和负载均衡

### 9. Elasticsearch 如何实现高可用？

**标准答案**：
1. **分片副本**：每个主分片至少 1 个副本
2. **quorum 机制**：避免脑裂
3. **自动故障转移**：Master 挂了自动选举新 Master

### 10. Mapping 是什么？

**标准答案**：
Mapping 是 Elasticsearch 的 schema，定义字段类型（text、keyword、date 等）和分析器。类似于数据库的表结构。

### 11. text 和 keyword 的区别？

| 字段类型 | 分词 | 适用场景 |
|---------|------|---------|
| **text** | 是 | 全文检索（如文章内容） |
| **keyword** | 否 | 精确匹配（如 ID、标签） |

### 12. Analyzer 是什么？

**标准答案**：
Analyzer 是分词器，包含 3 个组件：
1. **Character Filter**：字符预处理
2. **Tokenizer**：分词
3. **Token Filter**：词项后处理（小写、停用词）

### 13. 如何优化 Elasticsearch 写入性能？

**标准答案**：
1. **批量写入**：使用 bulk API
2. **调整 refresh_interval**：从 1s 调整为 30s
3. **关闭副本**：写入时临时关闭副本，写入完成后再开启
4. **增加 indexing buffer**：调整 `indices.memory.index_buffer_size`

### 14. 如何优化 Elasticsearch 查询性能？

**标准答案**：
1. **使用 filter 而非 query**：filter 可缓存
2. **避免深分页**：使用 search_after
3. **合理设计 Mapping**：减少字段数量
4. **预热查询**：使用 warmers

### 15. 如何监控 Elasticsearch？

**标准答案**：
1. **集群健康**：`GET /_cluster/health`
2. **节点统计**：`GET /_nodes/stats`
3. **索引统计**：`GET /<index>/_stats`
4. **慢查询日志**：`slowlog`

### 16. Elasticsearch 的 routing 机制？

**标准答案**：
- 默认 routing：`shard_num = hash(_id) % num_primary_shards`
- 自定义 routing：可以指定 routing 参数，将相关文档路由到同一分片

### 17. Elasticsearch 的聚合（Aggregation）是什么？

**标准答案**：
聚合是 Elasticsearch 的分析功能，类似 SQL 的 GROUP BY，支持：
- **Bucket Aggregation**：分桶（如按时间分桶）
- **Metric Aggregation**：指标（如求和、平均值）

### 18. Elasticsearch 如何实现实时搜索？

**标准答案**：
- 写入数据后，默认 1 秒后 refresh，数据从内存刷到文件系统缓存，变为可搜索
- 调整 `refresh_interval` 可控制实时性

### 19. Elasticsearch 的段（Segment）是什么？

**标准答案**：
段是 Lucene 的存储单元，一个索引由多个段组成。写入数据时先写入内存，refresh 后生成新段。段会定期合并（merge），减少段数量，提高查询性能。

### 20. Elasticsearch 和 MySQL 的区别？

| 维度 | Elasticsearch | MySQL |
|------|--------------|-------|
| **查询类型** | 全文检索、模糊查询 | 精确查询、关联查询 |
| **事务** | 不支持 | 支持 |
| **扩展性** | 水平扩展（分片） | 垂直扩展（分库分表） |
| **适用场景** | 搜索、日志分析 | 事务、关系型数据 |

---

### es 更新和乐观锁控制
-  "_version" : 1,
- "_seq_no" : 426,
- "_primary_term" : 1,


## 性能优化
### 关注哪些性能指标
- （读）query latency 1-2ms，复杂的查询可能到几十ms
- （读）fetch latency ，QPS，读数据量，延时
- （写）index rate，QPS，数据量，延时
- （写）index latency
- 存储数据量
- 集群读写QPS，CPU、内存、存储、网络IO的监控
- 节点维度的监控
- index维度的监控


### 集群规划
- 业务存储量，期望的SLA指标
- 节点数量、内存、CPU数量，是否需要SSD等
- 预留buffer,磁盘使用率达到85%、90%、95%
- CPU使用率
- 内存使用率
- 冷热数据，灾备方案

### settings 索引优化实践
- 分片数量：number_of_shards，经验值：建议每个分片大小不要超过30GB。建议根据集群节点的个数规模，分片个数建议>=集群节点的个数。5节点的集群，5个分片就比较合理。注意：除非reindex操作，分片数是不可以修改的
- 副本数量：number_of_replicas。除非你对系统的健壮性有异常高的要求，比如：银行系统。可以考虑2个副本以上。否则，1个副本足够。注意：副本数是可以通过配置随时修改的
- refresh_interval 是一个参数，用于配置 Elasticsearch 中的索引刷新间隔。索引刷新是将内存中的数据写入磁盘以使其可搜索的过程。刷新操作会将新的文档和更新的文档写入磁盘，并使其在搜索结果中可见。默认值表示每秒执行一次刷新操作
- 按照日期规划索引是个很好的习惯
- 务必使用别名，ES不像mysql方面的更改索引名称。使用别名就是一个相对灵活的选择
- setting中定义繁体全文检索时的traditional_chinese_analyzer以及一个名为lowercase的normalizer，常用于keyword类型的匹配
- 结合profile、explain api 分析query慢的原因。[search profile api](https://www.elastic.co/guide/en/elasticsearch/reference/7.17/search-profile.html)

```
{
    "hotel_index_20220810": {
        "settings": {
            "index": {
                "refresh_interval": "1s",
                "number_of_shards": "5",
                "provided_name": "hotel_index_20220810",
                "creation_date": "1660127508475",
                "analysis": {
                    "filter": {
                        "t2sconvert": {
                            "convert_type": "t2s",
                            "type": "stconvert"
                        }
                    },
                    "normalizer": {
                        "lowercase": {
                            "filter": [
                                "lowercase"
                            ],
                            "type": "custom"
                        }
                    },
                    "analyzer": {
                        "traditional_chinese_analyzer": {
                            "filter": "t2sconvert",
                            "type": "custom",
                            "tokenizer": "ik_smart"
                        }
                    }
                },
                "number_of_replicas": "2",
                "uuid": "afdjafkdlaf",
                "version": {
                    "created": "6080599"
                }
            }
        }
    }
}

```

### mapping 数据模型优化
- 不要使用默认的mapping.默认Mapping的字段类型是系统自动识别的。其中：string类型默认分成：text和keyword两种类型。如果你的业务中不需要分词、检索，仅需要精确匹配，仅设置为keyword即可。根据业务需要选择合适的类型，有利于节省空间和提升精度，如：浮点型的选择.
-  Mapping各字段的选型流程
- 选择合理的分词器。常见的开源中文分词器包括：ik分词器、ansj分词器、hanlp分词器、结巴分词器、海量分词器、“ElasticSearch最全分词器比较及使用方法” 搜索可查看对比效果。如果选择ik，建议使用ik_max_word。因为：粗粒度的分词结果基本包含细粒度ik_smart的结果。
- 一个字段包含多种语言：分别设置了不同的分词器。中文：ik_max_word，英语：english等
- analyzer：表示文档写入时的分词，search_analyzer表示检索时query的分词
- type:text，type:keyword，不分词
- normalizer 表示英文keyword判断时不区分大小写
- "dynamic" : "strict"
- https://www.elastic.co/guide/en/elasticsearch/reference/current/mapping-types.html

```json
"properties": {
    "accommodation": {
        "properties": {
            "value_in_chinese": {
                "type": "text",
                "fields": {
                    "keyword": {
                        "type": "keyword"
                    }
                },
                "analyzer": "ik_max_word",
                "search_analyzer": "ik_smart"
            },
            "value_in_english": {
                "type": "text",
                "fields": {
                    "keyword": {
                        "type": "keyword",
                        "normalizer": "lowercase"
                    }
                },
                "analyzer": "english"
            },
            "value_in_filipino": {
                "type": "text",
                "fields": {
                    "keyword": {
                        "type": "keyword"
                    }
                },
                "analyzer": "standard"
            },
            "value_in_indonesian": {
                "type": "text",
                "fields": {
                    "keyword": {
                        "type": "keyword"
                    }
                },
                "analyzer": "indonesian"
            },
            "value_in_malay": {
                "type": "text",
                "fields": {
                    "keyword": {
                        "type": "keyword"
                    }
                },
                "analyzer": "standard"
            },
            "value_in_thai": {
                "type": "text",
                "fields": {
                    "keyword": {
                        "type": "keyword"
                    }
                },
                "analyzer": "thai"
            },
            "value_in_tw_chinese": {
                "type": "text",
                "fields": {
                    "keyword": {
                        "type": "keyword"
                    }
                },
                "analyzer": "traditional_chinese_analyzer"
            },
            "value_in_vietnamese": {
                "type": "text",
                "fields": {
                    "keyword": {
                        "type": "keyword"
                    }
                },
                "analyzer": "standard"
            }
        }
    }
}
```








### 数据写入优化
1. 要不要秒级响应？Elasticsearch近实时的本质是：最快1s写入的数据可以被查询到。如果refresh_interval设置为1s，势必会产生大量的segment，检索性能会受到影响。所以，非实时的场景可以调大，设置为30s，甚至-1
2. 能批量就不单条写入
3. 减少副本，提升写入性能。写入前，副本数设置为0，写入后，副本数设置为原来值



### 读优化
1. 分析dsl
2. 禁用 wildcard模糊匹配,通过match_phrase和slop结合查询。
3. 极小的概率使用match匹配
4. 结合业务场景，大量使用filter过滤器
5. 控制返回字段和结果,同理，ES中，_source 返回全部字段也是非必须的。要通过_source 控制字段的返回，只返回业务相关的字段。
6. 分页深度查询和遍历.分页查询使用：from+size;遍历使用：scroll；并行遍历使用：scroll+slice


### 业务优化
1. 字段抽取、倾向性分析、分类/聚类、相关性判定放在写入ES之前的ETL阶段进行
2. 产品经理基于各种奇葩业务场景可能会提各种无理需求



## SDK 使用
- github.com/olivere/elastic
- https://github.com/elastic/go-elasticsearch


## es migrate tools
- https://github.com/medcl/esm
- https://github.com/medcl/esm/tree/0.1.0

## 拓展阅读
- [普通搜索和向量搜索介绍](https://blog.csdn.net/weixin_40601534/article/details/122435858?spm=1001.2014.3001.5501)
- [广告索引（定向）的布尔表达式](https://www.cnblogs.com/chenny7/p/14765412.html)
- [official document](https://www.elastic.co/guide/en/elasticsearch/reference/8.3/index.html)
- [scroll使用和Elasticsearch的深度翻页问题](https://www.jianshu.com/p/eb7f11e178b3)
- [ES 更新并发控制问题](https://www.jianshu.com/p/d4da0182a67a)
- [ES 读写流程](https://www.cnblogs.com/upupfeng/p/13488120.html)
- [达观数据搜索引擎的Query自动纠错技术和架构](http://www.datagrand.com/blog/search-query.html)
- [Elasticsearch基础之相关性介绍](https://donggeitnote.com/2021/09/19/elasticsearch-tfidf/)
- [ElasticSearch进阶之拼写错误](https://donggeitnote.com/2022/01/02/elasticsearch-typo/)
- [ElasticSearch进阶之输入匹配](https://donggeitnote.com/2021/11/06/elasticsearch-typematch/)
- [ElasticSearch进阶之多域搜索](https://donggeitnote.com/2021/10/02/elasticsearch-multiplesearch/)
- [ElasticSearch进阶之Shard/segment内部原理](https://donggeitnote.com/2021/09/29/elasticsearch-shard/)
- analysizer,normalizer,常用分词器介绍和评估。https://blog.csdn.net/Q176782/article/details/119054132
- [Kafka VS ElasticSearch 的相似性和比较](https://juejin.cn/post/6844904008432402440)
- [理解ES的refresh、flush、merge](https://blog.csdn.net/weixin_37692493/article/details/108182161)
  - 节点
  - index/topic
  - shard/partiion
  - 副本机制
- [让Elasticsearch飞起来!——性能优化实践干货](https://developer.aliyun.com/article/706990)

