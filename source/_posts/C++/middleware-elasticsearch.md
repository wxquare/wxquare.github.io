---
title: Middleware - elasticSearch
categories:
- C/C++
---



## 1、基本概念和文档导读
- [普通搜索和向量搜索介绍](https://blog.csdn.net/weixin_40601534/article/details/122435858?spm=1001.2014.3001.5501)
- [official document](https://www.elastic.co/guide/en/elasticsearch/reference/8.3/index.html)


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


## Golang SDK
- github.com/olivere/elastic
- https://github.com/elastic/go-elasticsearch


## es migrate tools
- https://github.com/medcl/esm
- https://github.com/medcl/esm/tree/0.1.0

## es scroll 和 深度翻页问题
- [如何解决Elasticsearch的深度翻页问题](https://www.jianshu.com/p/eb7f11e178b3)
