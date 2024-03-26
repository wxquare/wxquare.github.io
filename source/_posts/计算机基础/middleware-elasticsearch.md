---
title: Middleware - Elasticsearch
categories:
- 计算机基础
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

