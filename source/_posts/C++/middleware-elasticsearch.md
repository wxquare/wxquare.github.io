---
title: Middleware - elasticSearch
categories:
- C/C++
---


官方文档：https://www.elastic.co/guide/en/elasticsearch/reference/8.3/index.html



## 常用命令
### 创建索引
curl -XPUT -H'Content-Type: application/json'  host/index_name?pretty=true -d@index_mapping.json 

### 查看index
curl -XGET 'host/_cat/indices/*hotel_basic_info_v2_live*(支持正则表达式）?v=true&pretty=true'

### 查看索引mapping信息
curl -XGET 'host/index_name/_mapping?pretty=true'

### 通过doc id 查询
curl -XGET  'host/index/_doc/doc_id?pretty=true'

### query
curl -XPOST -H'Content-Type: application/json' 'http://es.i.dp.online_es.sz.shopee.io:9201/shopee_digital_purchase_hotel_search_key_v2_live_id/_search?pretty=true' -d '{
"query":{}}'


### update
curl -XPOST  -H'Content-Type: application/json' 'host/index/_doc/doc_id/_update' -d '{
"doc": {
    "price": "6500000001"
}
}'

### 聚合count查询
curl -XPOST -H'Content-Type: application/json' 'host/index_name/_count' -d '{
    "query": {
        "term": {
            "city_name.value_in_english.keyword": "Jakarta"
        }
    }
}'

