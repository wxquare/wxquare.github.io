---
title: Middleware - Elasticsearch
categories:
- C/C++
---



## 1、基本概念和文档导读
- [普通搜索和向量搜索介绍](https://blog.csdn.net/weixin_40601534/article/details/122435858?spm=1001.2014.3001.5501)
- [official document](https://www.elastic.co/guide/en/elasticsearch/reference/8.3/index.html)
- [scroll使用和Elasticsearch的深度翻页问题](https://www.jianshu.com/p/eb7f11e178b3)
- [ES 读写流程](https://www.cnblogs.com/upupfeng/p/13488120.html)
- [达观数据搜索引擎的Query自动纠错技术和架构](http://www.datagrand.com/blog/search-query.html)


## 2、基本用法
### 创建索引
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
## 3、query DSL
- compound queries
- full text queries
- term level queries

## 4、原理和实现
ES的读写流程主要是协调节点，主分片节点、副分片节点间的相互协调。

ES的读取分为GET和Search两种操作。GET根据文档id从正排索引中获取内容；Search不指定id，根据关键字从倒排索引中获取内容。

### 写单个文档的流程
1. 客户端向集群中的某个节点发送写请求，该节点就作为本次请求的协调节点；
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



## 5、性能优化

### 关注哪些性能指标
- （读）query latency 1-2ms，复杂的查询可能到几十ms
- （读）fetch latency
- （写）index rate
- （写）index latency

**方法**
1. 结合profile、explain api 分析query慢的原因。[search profile api](https://www.elastic.co/guide/en/elasticsearch/reference/7.17/search-profile.html)


## 6、SDK 使用
- github.com/olivere/elastic
- https://github.com/elastic/go-elasticsearch


## es migrate tools
- https://github.com/medcl/esm
- https://github.com/medcl/esm/tree/0.1.0

