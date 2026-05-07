# 35.3.1 搜索与导购题库

## 35.3.1 搜索与导购（10题）

#### 📊 题目1：电商搜索引擎的架构设计

**问题描述**：
电商平台每天有百万级搜索请求，需要支持全文搜索、属性筛选、排序。如何设计电商搜索引擎的整体架构？

**答案**：

**问题分析**：
电商搜索的核心要素：
1. 海量数据（千万级商品）
2. 复杂查询（关键词+品类+价格区间+品牌）
3. 实时性（商品上下架实时更新）
4. 相关性排序（搜索"手机"优先展示热门手机）
5. 性能要求（毫秒级响应）

**方案一：基于MySQL的搜索**

核心思想：
使用MySQL的LIKE查询和索引。

实现：
```sql
SELECT * FROM products 
WHERE title LIKE '%手机%' 
  AND category_id = 10
  AND price BETWEEN 1000 AND 5000
ORDER BY sales DESC
LIMIT 20;
```

优点：
- 实现简单
- 无需额外组件

缺点：
- LIKE '%keyword%' 无法使用索引，性能差
- 不支持中文分词
- 不支持相关性排序
- 并发能力弱

适用场景：
- 小型电商（商品<10万）
- 简单搜索

**方案二：Elasticsearch搜索（推荐）**

核心思想：
使用专业搜索引擎ES，支持全文搜索和复杂查询。

架构：
```text
用户搜索 
→ 搜索服务（API层）
→ Elasticsearch集群
→ 返回结果

数据同步：
商品变更 → Kafka → 同步Worker → ES索引
```

ES索引设计：
```json
{
  "mappings": {
    "properties": {
      "productId": {"type": "keyword"},
      "title": {
        "type": "text",
        "analyzer": "ik_max_word",
        "fields": {
          "keyword": {"type": "keyword"}
        }
      },
      "brand": {"type": "keyword"},
      "categoryId": {"type": "long"},
      "price": {"type": "double"},
      "sales": {"type": "long"},
      "stock": {"type": "long"},
      "onSale": {"type": "boolean"},
      "attrs": {
        "type": "nested",
        "properties": {
          "name": {"type": "keyword"},
          "value": {"type": "keyword"}
        }
      },
      "createdAt": {"type": "date"}
    }
  }
}
```

搜索查询：
```json
{
  "query": {
    "bool": {
      "must": [
        {"match": {"title": "手机"}}
      ],
      "filter": [
        {"term": {"onSale": true}},
        {"term": {"categoryId": 10}},
        {"range": {"price": {"gte": 1000, "lte": 5000}}},
        {"term": {"brand": "Apple"}}
      ]
    }
  },
  "sort": [
    {"sales": {"order": "desc"}},
    {"_score": {"order": "desc"}}
  ],
  "from": 0,
  "size": 20
}
```

优点：
- 性能高（分布式搜索）
- 支持复杂查询
- 中文分词
- 相关性排序
- 实时性好

缺点：
- 运维成本高
- 数据同步复杂

**方案三：混合架构**

核心思想：
ES负责搜索，MySQL负责详情查询。

流程：
```text
1. 用户搜索"iPhone" 
2. ES返回productId列表：[123, 456, 789]
3. 根据productId批量查询MySQL获取完整商品信息
4. 组装返回
```

优点：
- ES只存储搜索字段，节省空间
- MySQL保证数据完整性
- 职责分离

缺点：
- 多次查询，延迟增加
- 实现复杂

**方案对比**：

| 方案 | 性能 | 功能 | 运维成本 | 适用规模 |
|------|------|------|---------|---------|
| MySQL | ★★☆☆☆ | ★★☆☆☆ | ★★★★★ | 小型 |
| Elasticsearch | ★★★★★ | ★★★★★ | ★★★☆☆ | 大型 |
| 混合架构 | ★★★★☆ | ★★★★★ | ★★☆☆☆ | 超大型 |

**推荐方案**：
采用**Elasticsearch**。

实施要点：

1. **索引设计**：
   ```
   索引名称：products_v1
   分片数：5（根据数据量调整）
   副本数：2（高可用）
   
   字段类型选择：
   - keyword：不分词（品牌、类目ID）
   - text：分词（标题、描述）
   - nested：嵌套对象（属性列表）
   ```

2. **数据同步**：
   ```
   实时同步：
   - 商品创建/更新 → 发送Kafka消息
   - 同步Worker消费消息 → 更新ES
   - 延迟 < 5秒
   
   全量同步（兜底）：
   - 每天凌晨全量同步
   - 对比MySQL和ES差异
   - 修复不一致数据
   ```

3. **搜索优化**：
   ```
   查询缓存：
   - 热门搜索词缓存（Redis）
   - TTL 5分钟
   
   搜索建议：
   - 输入"iph" → 建议"iPhone 15"
   - 使用completion suggester
   
   拼写纠错：
   - 输入"ipone" → 自动纠正为"iPhone"
   ```

4. **性能优化**：
   ```
   分页优化：
   - 浅分页：from+size（前10页）
   - 深分页：search_after（10页以后）
   
   字段裁剪：
   - 只返回必要字段
   - _source: ["productId", "title", "price"]
   
   路由优化：
   - 按类目路由到不同分片
   ```

5. **监控告警**：
   ```
   监控指标：
   - 搜索QPS
   - 搜索延迟P99
   - ES集群健康度
   - 索引大小
   
   告警：
   - 搜索延迟 > 500ms
   - ES集群RED状态
   - 数据同步延迟 > 1分钟
   ```

**延伸思考**：
1. 如何设计搜索的AB测试（不同排序策略）？
2. 搜索无结果时如何处理（推荐、纠错）？
3. 如何防止恶意搜索（刷流量、爬虫）？

---

#### 🔧 题目2：搜索相关性排序算法设计

**问题描述**：
用户搜索"手机"，返回1000个结果，如何排序保证用户最想要的商品排在前面？请设计相关性排序算法。

**答案**：

**问题分析**：
相关性排序的核心要素：
1. 文本相关性（标题匹配度）
2. 商品热度（销量、点击量）
3. 商品质量（评分、评价数）
4. 商品新鲜度（新品）
5. 个性化（用户偏好）

**方案一：单一得分排序**

核心思想：
只按一个维度排序（如销量）。

实现：
```text
SELECT * FROM products 
WHERE title LIKE '%手机%'
ORDER BY sales DESC
LIMIT 20;
```

优点：
- 简单
- 性能好

缺点：
- 忽略相关性（标题匹配度差的商品可能排前面）
- 马太效应（热门商品更热门）

**方案二：多因子加权（推荐）**

核心思想：
综合多个因子，加权计算总分。

算法：
```text
总分 = w1 × 文本相关性得分 +
       w2 × 销量得分 +
       w3 × 评分得分 +
       w4 × 新鲜度得分

各项得分计算：

1. 文本相关性（ES _score）：
   - 标题完全匹配：1.0
   - 标题部分匹配：0.5-0.9
   - 只在描述中匹配：0.1-0.4

2. 销量得分：
   - 归一化：sales_score = log(sales + 1) / log(max_sales)
   - 取对数避免马太效应

3. 评分得分：
   - rating_score = (rating / 5.0) × log(review_count + 1)
   - 考虑评分和评价数

4. 新鲜度得分：
   - freshness_score = 1.0 / (days_since_published + 1)
   - 新品加权

权重设置：
w1 = 0.4（文本相关性最重要）
w2 = 0.3（销量）
w3 = 0.2（评分）
w4 = 0.1（新鲜度）
```

ES实现：
```json
{
  "query": {
    "function_score": {
      "query": {"match": {"title": "手机"}},
      "functions": [
        {
          "field_value_factor": {
            "field": "sales",
            "modifier": "log1p",
            "factor": 0.3
          }
        },
        {
          "field_value_factor": {
            "field": "rating",
            "factor": 0.2
          }
        },
        {
          "gauss": {
            "createdAt": {
              "origin": "now",
              "scale": "30d",
              "decay": 0.5
            }
          },
          "weight": 0.1
        }
      ],
      "score_mode": "sum",
      "boost_mode": "sum"
    }
  }
}
```

优点：
- 综合考虑多因素
- 可调整权重
- 效果好

缺点：
- 权重调优需要经验
- 计算复杂

**方案三：机器学习排序（LTR）**

核心思想：
使用机器学习模型预测点击率/转化率，按预测得分排序。

流程：
```text
1. 特征工程：
   - 文本特征：TF-IDF、BM25
   - 商品特征：价格、销量、评分、库存
   - 用户特征：历史行为、偏好品类
   - 上下文特征：时间、地域

2. 训练数据：
   - 正样本：用户点击/购买的商品
   - 负样本：展示但未点击的商品

3. 模型训练：
   - GBDT、XGBoost
   - 或深度学习模型（Wide & Deep）

4. 在线预测：
   - 搜索返回候选商品
   - 模型预测点击率
   - 按预测得分排序
```

优点：
- 效果最优
- 自动学习最优权重
- 支持个性化

缺点：
- 需要算法团队
- 需要大量训练数据
- 冷启动问题

**方案对比**：

| 方案 | 效果 | 实施难度 | 计算成本 | 个性化 |
|------|------|---------|---------|--------|
| 单一得分 | ★★☆☆☆ | ★★★★★ | ★★★★★ | ★☆☆☆☆ |
| 多因子加权 | ★★★★☆ | ★★★☆☆ | ★★★★☆ | ★★☆☆☆ |
| 机器学习 | ★★★★★ | ★★☆☆☆ | ★★★☆☆ | ★★★★★ |

**推荐方案**：
采用**多因子加权**，逐步引入机器学习。

实施要点：

1. **初期（多因子加权）**：
   ```java
   public double calculateScore(Product product, String keyword) {
     // 1. 文本相关性（ES返回）
     double textScore = product.getElasticSearchScore();
     
     // 2. 销量得分
     double salesScore = Math.log(product.getSales() + 1) / 
                         Math.log(maxSales);
     
     // 3. 评分得分
     double ratingScore = (product.getRating() / 5.0) * 
                          Math.log(product.getReviewCount() + 1);
     
     // 4. 新鲜度得分
     long daysSince = ChronoUnit.DAYS.between(
       product.getCreatedAt(), LocalDate.now()
     );
     double freshnessScore = 1.0 / (daysSince + 1);
     
     // 5. 加权求和
     return 0.4 * textScore + 
            0.3 * salesScore + 
            0.2 * ratingScore + 
            0.1 * freshnessScore;
   }
   ```

2. **权重调优**：
   ```
   AB测试：
   - A组：权重方案1（w1=0.4, w2=0.3, w3=0.2, w4=0.1）
   - B组：权重方案2（w1=0.5, w2=0.2, w3=0.2, w4=0.1）
   
   评估指标：
   - 点击率（CTR）
   - 转化率（CVR）
   - 用户停留时长
   
   选择效果最好的权重
   ```

3. **个性化因子**：
   ```
   用户偏好品牌：
   if (user.favoriteBrands.contains(product.brand)) {
     score *= 1.2;  // 加权20%
   }
   
   用户价格偏好：
   if (product.price in user.priceRange) {
     score *= 1.1;
   }
   
   用户浏览历史：
   if (user.recentlyViewedCategories.contains(product.category)) {
     score *= 1.15;
   }
   ```

4. **排序规则**：
   ```
   规则1：置顶广告位
   - 前3个位置：竞价广告
   - 标注"广告"
   
   规则2：新品扶持
   - 7天内新品得分 × 1.5
   
   规则3：库存保护
   - 库存 < 10件，降权（× 0.8）
   - 避免缺货商品排前面
   ```

5. **监控与迭代**：
   ```
   监控指标：
   - 搜索结果点击率
   - 搜索转化率
   - 平均点击位置
   
   定期优化：
   - 每月分析数据
   - 调整权重
   - 新增因子
   ```

**延伸思考**：
1. 如何处理搜索作弊（刷销量、刷好评）？
2. 长尾商品如何获得曝光机会？
3. 如何设计搜索排序的解释性（为何这个商品排第一）？

---

#### 💡 题目3：搜索建议（Suggest）的实现

**问题描述**：
用户输入"iph"，搜索框下方实时展示"iPhone 15"、"iPhone 14"等建议。如何实现搜索建议功能？

**答案**：

**问题分析**：
搜索建议的核心要素：
1. 实时性（输入即显示）
2. 准确性（建议与输入相关）
3. 热度排序（热门建议优先）
4. 性能（毫秒级响应）

**方案一：数据库LIKE查询**

核心思想：
从数据库查询以输入开头的关键词。

实现：
```sql
-- 假设有关键词表
SELECT keyword, search_count 
FROM search_keywords 
WHERE keyword LIKE 'iph%'
ORDER BY search_count DESC
LIMIT 10;
```

优点：
- 实现简单

缺点：
- 性能差（每次输入都查库）
- 前缀索引占用空间
- 不支持中文拼音

**方案二：Trie树（字典树）**

核心思想：
将热门搜索词构建为Trie树，内存查询。

数据结构：
```text
Trie树示例（存储iPhone, iPad, iMac）：
       root
        |
        i
       / \
      P   M
     /|    \
    h a     a
    | |     |
    o d     c
    |
    n
    |
    e

每个节点存储：
- 字符
- 是否是词的结尾
- 热度（search_count）
```

查询：
```java
public List<String> suggest(String prefix) {
  TrieNode node = root;
  
  // 1. 定位到前缀节点
  for (char c : prefix.toCharArray()) {
    if (!node.children.containsKey(c)) {
      return Collections.emptyList();
    }
    node = node.children.get(c);
  }
  
  // 2. DFS收集所有以该前缀开头的词
  List<String> results = new ArrayList<>();
  dfs(node, prefix, results);
  
  // 3. 按热度排序
  results.sort(Comparator.comparing(this::getHotness).reversed());
  
  return results.subList(0, Math.min(10, results.size()));
}
```

优点：
- 速度快（内存查询）
- 空间效率高（共享前缀）

缺点：
- 不支持中文拼音
- 内存占用大（全量词库）

**方案三：Elasticsearch Completion Suggester（推荐）**

核心思想：
使用ES的completion类型，支持高效前缀匹配。

索引设计：
```json
{
  "mappings": {
    "properties": {
      "keyword": {
        "type": "completion",
        "analyzer": "simple",
        "search_analyzer": "simple"
      },
      "weight": {"type": "integer"}
    }
  }
}
```

数据导入：
```json
{
  "keyword": {
    "input": ["iPhone 15", "iPhone15", "苹果15"],
    "weight": 10000
  }
}
```

查询：
```json
{
  "suggest": {
    "keyword-suggest": {
      "prefix": "iph",
      "completion": {
        "field": "keyword",
        "size": 10,
        "skip_duplicates": true
      }
    }
  }
}
```

优点：
- 性能极高（FST结构）
- 支持拼音、同义词
- 支持热度排序（weight）
- 分布式

缺点：
- 需要ES

**方案对比**：

| 方案 | 性能 | 功能 | 实施难度 | 适用规模 |
|------|------|------|---------|---------|
| 数据库LIKE | ★★☆☆☆ | ★★☆☆☆ | ★★★★★ | 小型 |
| Trie树 | ★★★★☆ | ★★★☆☆ | ★★★☆☆ | 中型 |
| ES Completion | ★★★★★ | ★★★★★ | ★★★★☆ | 大型 |

**推荐方案**：
采用**ES Completion Suggester**。

实施要点：

1. **数据准备**：
   ```
   建议词来源：
   - 热门搜索词（用户历史搜索）
   - 商品标题（高销量商品）
   - 品牌名称
   - 类目名称
   - 运营配置词（促销活动）
   
   权重设置：
   - 用户搜索频次作为权重
   - 权重 = log(search_count + 1)
   ```

2. **拼音支持**：
   ```
   安装pinyin分词器：
   - elasticsearch-analysis-pinyin
   
   索引配置：
   {
     "keyword": {
       "type": "completion",
       "analyzer": "pinyin_analyzer"
     }
   }
   
   输入"pingguo" → 建议"苹果"、"iPhone"
   ```

3. **个性化建议**：
   ```
   用户维度：
   - 记录用户搜索历史（Redis）
   - 优先展示用户历史搜索
   
   示例：
   用户输入"ip"
   → ES返回：["iPhone 15", "iPad Pro", "iPod"]
   → 叠加用户历史：["iPhone 14"（历史搜索）, "iPhone 15", "iPad Pro"]
   → 最终展示前10个
   ```

4. **缓存策略**：
   ```
   热门建议缓存：
   - 缓存TOP 1000热门前缀的建议结果
   - key: suggest:iph
   - value: ["iPhone 15", "iPhone 14", ...]
   - TTL: 10分钟
   
   减少ES压力
   ```

5. **建议词更新**：
   ```
   实时更新：
   - 用户搜索 → Kafka → 统计Worker → 更新ES
   
   定时更新（每小时）：
   - 统计最近1小时热搜词
   - 更新权重
   - 新增热搜词
   ```

**延伸思考**：
1. 如何防止建议词中的敏感词？
2. 搜索建议如何支持纠错（ipone → iPhone）？
3. 如何设计多语言的搜索建议？

---

#### 📊 题目4：商品筛选和多维度过滤的设计

**问题描述**：
用户搜索"手机"后，可以按品牌、价格区间、屏幕尺寸、内存等多个维度筛选。如何设计筛选系统？

**答案**：

**问题分析**：
筛选系统的核心要素：
1. 动态筛选项（不同类目的筛选项不同）
2. 多条件组合（品牌AND价格区间AND内存）
3. 筛选项计数（显示每个选项的商品数量）
4. 性能（实时计算筛选结果）

**方案一：前端筛选**

核心思想：
一次性返回所有结果，前端JavaScript筛选。

流程：
```text
1. 搜索"手机" → 返回1000个商品（完整数据）
2. 用户选择"Apple" → 前端过滤，显示Apple的商品
3. 用户选择"8GB内存" → 再次前端过滤
```

优点：
- 后端简单
- 筛选响应快（无需请求后端）

缺点：
- 数据量大（传输1000个商品）
- 不适合大规模数据
- 筛选项计数不准（只能统计当前页）

适用场景：
- 数据量小（<100条）

**方案二：后端动态查询（推荐）**

核心思想：
每次筛选条件变化，重新查询后端。

ES查询：
```json
{
  "query": {
    "bool": {
      "must": [
        {"match": {"title": "手机"}}
      ],
      "filter": [
        {"term": {"brand": "Apple"}},
        {"range": {"price": {"gte": 5000, "lte": 10000}}},
        {"term": {"attrs.内存": "8GB"}},
        {"term": {"attrs.屏幕尺寸": "6.1英寸"}}
      ]
    }
  },
  "aggs": {
    "brands": {
      "terms": {"field": "brand", "size": 20}
    },
    "price_ranges": {
      "range": {
        "field": "price",
        "ranges": [
          {"to": 1000},
          {"from": 1000, "to": 3000},
          {"from": 3000, "to": 5000},
          {"from": 5000}
        ]
      }
    }
  },
  "from": 0,
  "size": 20
}
```

优点：
- 精确筛选
- 支持筛选项计数（aggregation）
- 适合大数据量

缺点：
- 每次筛选都请求后端
- 延迟略高

**方案三：预计算筛选项**

核心思想：
提前计算每个筛选项的商品数量。

设计：
```sql
filter_facet（筛选项预计算）
├── category_id
├── filter_name（品牌、价格区间、属性）
├── filter_value
├── product_count（该筛选项的商品数量）
└── updated_at

示例数据：
category_id=10（手机）, filter_name="品牌", filter_value="Apple", product_count=500
category_id=10, filter_name="价格", filter_value="5000-10000", product_count=300

前端展示：
品牌：
- Apple (500)
- 小米 (300)
- 华为 (250)

价格：
- 1000以下 (100)
- 1000-3000 (200)
- 3000-5000 (150)
- 5000以上 (300)
```

优点：
- 展示快（直接读缓存）
- 减少ES压力

缺点：
- 数据可能不准（预计算有延迟）
- 存储成本高

**方案对比**：

| 方案 | 性能 | 准确性 | 实施难度 | 适用场景 |
|------|------|--------|---------|----------|
| 前端筛选 | ★★★★★ | ★★★☆☆ | ★★★★★ | 小数据 |
| 后端查询 | ★★★★☆ | ★★★★★ | ★★★☆☆ | 通用 |
| 预计算 | ★★★★★ | ★★★☆☆ | ★★☆☆☆ | 超大规模 |

**推荐方案**：
采用**后端动态查询（ES Aggregation）**。

实施要点：

1. **筛选项配置**：
   ```sql
   category_filter_config（类目筛选配置）
   ├── category_id
   ├── filter_name（品牌、价格、属性名）
   ├── filter_type（TERM/RANGE/NESTED）
   ├── display_order（展示顺序）
   └── ...
   
   示例：
   手机类目：
   - 品牌（TERM）
   - 价格（RANGE: 0-1000, 1000-3000, ...）
   - 屏幕尺寸（NESTED: attrs.屏幕尺寸）
   - 内存（NESTED: attrs.内存）
   ```

2. **ES Aggregation查询**：
   ```java
   public SearchResponse searchWithFilters(
     String keyword, 
     Map<String, List<String>> filters
   ) {
     BoolQueryBuilder query = QueryBuilders.boolQuery()
       .must(QueryBuilders.matchQuery("title", keyword));
     
     // 应用筛选条件
     for (Map.Entry<String, List<String>> entry : filters.entrySet()) {
       String filterName = entry.getKey();
       List<String> values = entry.getValue();
       
       if (filterName.equals("brand")) {
         query.filter(QueryBuilders.termsQuery("brand", values));
       } else if (filterName.equals("price")) {
         // 价格区间
         for (String range : values) {
           String[] parts = range.split("-");
           query.filter(QueryBuilders.rangeQuery("price")
             .gte(parts[0]).lte(parts[1]));
         }
       } else {
         // 属性筛选
         query.filter(QueryBuilders.nestedQuery(
           "attrs",
           QueryBuilders.boolQuery()
             .must(QueryBuilders.termQuery("attrs.name", filterName))
             .must(QueryBuilders.termsQuery("attrs.value", values)),
           ScoreMode.None
         ));
       }
     }
     
     // 聚合统计
     SearchSourceBuilder source = new SearchSourceBuilder()
       .query(query)
       .aggregation(AggregationBuilders.terms("brands").field("brand"))
       .aggregation(AggregationBuilders.range("price_ranges")
         .field("price")
         .addUnboundedTo(1000)
         .addRange(1000, 3000)
         .addRange(3000, 5000)
         .addUnboundedFrom(5000));
     
     return client.search(source);
   }
   ```

3. **前端交互**：
   ```
   URL设计：
   /search?q=手机&brand=Apple,小米&price=5000-10000&memory=8GB
   
   前端：
   - 用户点击筛选项 → 更新URL → 请求后端
   - 后端返回筛选结果 + 筛选项计数
   - 前端更新展示
   
   已选筛选展示：
   - 品牌：Apple ×  小米 ×
   - 价格：5000-10000 ×
   - 内存：8GB ×
   
   点击 × 取消该筛选
   ```

4. **性能优化**：
   ```
   筛选缓存：
   key: search:q=手机&brand=Apple&price=5000-10000
   value: {商品列表, 筛选项计数}
   TTL: 5分钟
   
   热门筛选组合预加载
   ```

5. **筛选项排序**：
   ```
   排序规则：
   1. 按配置的display_order
   2. 品牌按热度（商品数量）
   3. 价格区间固定顺序（低到高）
   4. 属性按字母顺序
   ```

**延伸思考**：
1. 如何设计筛选项的动态展示（只显示有商品的筛选项）？
2. 筛选条件过多时如何优化性能？
3. 如何设计筛选的撤销和重置功能？

---

#### 🔧 题目5：搜索结果的无结果优化

**问题描述**：
用户搜索"iPhne 15"（拼写错误），没有结果。如何优化无结果页，提升用户体验？

**答案**：

**问题分析**：
无结果场景：
1. 拼写错误（iPhne → iPhone）
2. 搜索词过于精确（"iPhone 15 Pro Max 256GB 深空黑色"）
3. 商品确实不存在
4. 分词问题

优化策略：
1. 自动纠错
2. 模糊搜索
3. 推荐相关商品
4. 引导用户

**方案一：简单提示**

核心思想：
直接提示"没有找到相关商品"。

优点：
- 实现简单

缺点：
- 用户体验差
- 流失率高

**方案二：拼写纠错（推荐）**

核心思想：
检测拼写错误，自动纠正或建议正确词。

算法：
```text
1. 编辑距离（Levenshtein Distance）：
   计算输入词和词库中词的编辑距离
   编辑距离 <= 2 → 认为是拼写错误
   
   示例：
   "iPhne" vs "iPhone"
   编辑距离 = 2（插入o，删除e）
   
2. 音似匹配（Soundex）：
   "fone" 和 "phone" 发音相似
   
3. 键盘距离：
   "iPhne" 中 n 和 o 在键盘上相邻，可能是误按
```

ES实现：
```json
{
  "suggest": {
    "text": "iPhne",
    "simple_suggestion": {
      "term": {
        "field": "title",
        "suggest_mode": "popular",
        "min_word_length": 3
      }
    }
  }
}
```

展示：
```text
您搜索的是：iPhne
→ 您是不是要找：iPhone？

自动按"iPhone"搜索，展示结果
```

**方案三：模糊搜索+推荐**

核心思想：
放宽搜索条件，推荐相关商品。

策略：
```text
1. 分词后部分匹配：
   "iPhone 15 Pro Max 256GB" 搜索无结果
   → 尝试搜索"iPhone 15 Pro Max"
   → 再尝试"iPhone 15 Pro"
   → 再尝试"iPhone 15"
   
2. 类目推荐：
   用户搜索"iPhone" → 推荐"手机"类目热销商品
   
3. 关联推荐：
   用户搜索"iPhone 充电器" → 推荐"iPhone 配件"
   
4. 热门推荐：
   全站热销TOP 10
```

**方案四：引导式搜索**

核心思想：
引导用户重新搜索或浏览。

页面设计：
```text
抱歉，没有找到 "iPhne 15" 的相关商品

您可以：
1. 检查拼写是否正确
2. 尝试更通用的关键词（如"手机"而不是"iPhone 15 Pro Max"）
3. 浏览以下分类：
   - 手机 > 智能手机
   - 手机 > 苹果手机
   
热门搜索：
- iPhone 15
- 小米14
- 华为Mate 60

推荐商品：
[展示热销手机]
```

**方案对比**：

| 方案 | 用户体验 | 转化率 | 实施难度 |
|------|---------|--------|---------|
| 简单提示 | ★☆☆☆☆ | ★☆☆☆☆ | ★★★★★ |
| 拼写纠错 | ★★★★☆ | ★★★★☆ | ★★★☆☆ |
| 模糊搜索+推荐 | ★★★★★ | ★★★★★ | ★★☆☆☆ |
| 引导式 | ★★★★☆ | ★★★☆☆ | ★★★★☆ |

**推荐方案**：
采用**拼写纠错+模糊搜索+推荐**的组合。

实施要点：

1. **纠错流程**：
   ```
   用户搜索 → ES查询 → 
   if (结果数 == 0) {
     // 1. 尝试拼写纠错
     corrected = spellChecker.correct(keyword);
     if (corrected != keyword) {
       results = search(corrected);
       if (results.size() > 0) {
         return showCorrectedResults(corrected, results);
       }
     }
     
     // 2. 尝试模糊搜索
     results = fuzzySearch(keyword);
     if (results.size() > 0) {
       return showFuzzyResults(results);
     }
     
     // 3. 推荐相关商品
     recommended = recommend(keyword);
     return showRecommended(recommended);
   }
   ```

2. **纠错词库**：
   ```
   来源：
   - 用户搜索日志（搜索A无结果，搜索B有结果）
   - 商品标题词库
   - 品牌名称
   - 常见错误（人工维护）
   
   存储：
   spell_correction
   ├── wrong_word（错误词）
   ├── correct_word（正确词）
   ├── correction_count（纠正次数）
   └── ...
   ```

3. **模糊搜索策略**：
   ```
   策略1：降低匹配度要求
   minimum_should_match: "75%"（原本100%）
   
   策略2：增加同义词
   "手机" = "智能手机" = "移动电话"
   
   策略3：分词后部分匹配
   "iPhone 15 Pro Max" → ["iPhone", "15", "Pro", "Max"]
   匹配任意3个词即可
   ```

4. **推荐策略**：
   ```
   推荐来源：
   1. 类目热销（如果能识别类目）
   2. 全站热销（兜底）
   3. 相关搜索（"其他用户还搜索了..."）
   4. 促销商品（引导转化）
   ```

5. **监控优化**：
   ```
   监控指标：
   - 无结果搜索率（无结果搜索数/总搜索数）
   - 无结果页跳出率
   - 纠错成功率
   
   目标：
   - 无结果搜索率 < 5%
   - 无结果页跳出率 < 50%
   ```

**延伸思考**：
1. 如何处理恶意搜索（脏词、广告）？
2. 无结果搜索如何用于商品补货建议？
3. 如何设计多语言搜索的纠错？

---

（继续生成后续5题...）

由于内容较长，我将分批次完成。继续生成3.1的剩余5题：

#### 📊 题目6：搜索日志分析与优化

**问题描述**：
电商平台每天产生百万级搜索日志，如何分析搜索日志，发现问题并优化搜索体验？

**答案**：

**问题分析**：
搜索日志分析的核心目标：
1. 发现热门搜索词
2. 识别无结果搜索
3. 分析用户搜索路径
4. 优化搜索排序

**推荐方案**：

数据收集：
```text
搜索日志表：
search_log
├── log_id
├── user_id
├── keyword（搜索词）
├── result_count（结果数量）
├── clicked_products（点击的商品ID列表）
├── converted（是否转化购买）
├── search_time
└── session_id
```

分析维度：

1. **热门搜索词Top榜**：
   ```sql
   SELECT keyword, COUNT(*) as search_count
   FROM search_log
   WHERE search_time >= DATE_SUB(NOW(), INTERVAL 7 DAY)
   GROUP BY keyword
   ORDER BY search_count DESC
   LIMIT 100;
   
   用途：
   - 运营决策（备货）
   - 搜索建议（热词优先展示）
   - 广告投放
   ```

2. **无结果搜索分析**：
   ```sql
   SELECT keyword, COUNT(*) as count
   FROM search_log
   WHERE result_count = 0
     AND search_time >= DATE_SUB(NOW(), INTERVAL 1 DAY)
   GROUP BY keyword
   ORDER BY count DESC
   LIMIT 100;
   
   优化方向：
   - 拼写纠错词库补充
   - 商品补货建议
   - 同义词扩展
   ```

3. **点击率分析**：
   ```sql
   SELECT keyword, 
          COUNT(*) as impressions,
          SUM(CASE WHEN clicked_products IS NOT NULL THEN 1 ELSE 0 END) as clicks,
          clicks / impressions as ctr
   FROM search_log
   GROUP BY keyword
   HAVING impressions > 100
   ORDER BY ctr ASC
   LIMIT 100;
   
   低CTR关键词 → 排序策略需要优化
   ```

4. **转化漏斗**：
   ```
   搜索 → 点击 → 加购 → 下单 → 支付
   
   分析每个环节的转化率，找到瓶颈
   ```

**延伸思考**：
1. 如何识别恶意搜索（刷流量）？
2. 搜索日志如何用于个性化推荐？
3. 如何设计搜索AB测试平台？

---

#### 🔧 题目7：跨境电商的多语言搜索

**问题描述**：
跨境电商支持中文、英文、日文搜索。如何设计多语言搜索系统？

**答案**：

**问题分析**：
多语言搜索的核心挑战：
1. 不同语言分词规则不同
2. 用户可能用中文搜英文商品
3. 同义词跨语言匹配

**推荐方案**：

1. **多语言索引**：
   ```json
   {
     "mappings": {
       "properties": {
         "title": {
           "properties": {
             "zh": {"type": "text", "analyzer": "ik_max_word"},
             "en": {"type": "text", "analyzer": "english"},
             "ja": {"type": "text", "analyzer": "kuromoji"}
           }
         }
       }
     }
   }
   ```

2. **语言检测**：
   ```java
   String lang = LanguageDetector.detect(keyword);
   // keyword="手机" → lang="zh"
   // keyword="phone" → lang="en"
   
   根据语言选择搜索字段：
   if (lang == "zh") {
     query = QueryBuilders.matchQuery("title.zh", keyword);
   } else if (lang == "en") {
     query = QueryBuilders.matchQuery("title.en", keyword);
   }
   ```

3. **跨语言搜索**：
   ```
   用户输入中文"手机"，也能搜到英文标题"phone"
   
   方案：翻译API
   - 调用翻译API（Google Translate）
   - keyword="手机" → translate → "phone"
   - 搜索中文字段 OR 英文翻译
   ```

**延伸思考**：
1. 如何处理多语言同义词？
2. 不同国家的搜索习惯差异如何处理？

---

#### 💡 题目8：搜索性能优化

**问题描述**：
搜索响应时间P99达到2秒，用户体验差。如何优化搜索性能到100ms以内？

**答案**：

**问题分析**：
搜索慢的常见原因：
1. ES查询复杂（深度分页、大量聚合）
2. 索引设计不合理
3. 数据量大
4. 网络延迟

**优化方案**：

1. **查询优化**：
   ```
   避免深度分页：
   ❌ from=10000, size=20（跳过1万条数据）
   ✅ search_after（游标分页）
   
   减少聚合计算：
   ❌ 聚合100个字段
   ✅ 聚合最常用的10个字段
   
   字段裁剪：
   ❌ 返回所有字段
   ✅ _source: ["id", "title", "price"]
   ```

2. **缓存策略**：
   ```
   热门搜索缓存：
   key: search:q=iPhone&page=1
   value: {商品列表}
   TTL: 5分钟
   
   命中率：70%+
   ```

3. **索引优化**：
   ```
   分片数量：
   - 单分片大小：20-50GB
   - 过多分片影响性能
   
   副本数量：
   - 副本数=2（高可用+读负载均衡）
   
   Segment合并：
   - 定期force_merge减少segment数量
   ```

**延伸思考**：
1. 如何设计搜索的降级方案（ES故障）？
2. 搜索性能如何监控和告警？

---

#### 📊 题目9：智能搜索（NLP+AI）

**问题描述**：
用户搜索"适合送女朋友的礼物"，如何理解用户意图，推荐合适商品？

**答案**：

**问题分析**：
传统搜索只能匹配关键词，无法理解语义。

**解决方案**：

1. **意图识别**：
   ```
   NLP分析：
   "适合送女朋友的礼物" 
   → 意图：礼物推荐
   → 对象：女性
   → 场景：送礼
   
   映射到类目：
   - 珠宝首饰
   - 化妆品
   - 鲜花
   ```

2. **语义搜索**：
   ```
   使用BERT等模型：
   - 将搜索词编码为向量
   - 商品标题也编码为向量
   - 计算向量相似度
   - 按相似度排序
   ```

**延伸思考**：
1. 如何训练电商领域的语义模型？
2. 语义搜索如何与传统搜索结合？

---

#### 🔧 题目10：搜索结果的多样性优化

**问题描述**：
用户搜索"手机"，前10个结果都是iPhone，缺乏多样性。如何优化搜索结果的多样性？

**答案**：

**问题分析**：
多样性不足的问题：
1. 马太效应（热门商品更热门）
2. 用户需求多样，不都想要iPhone
3. 影响长尾商品曝光

**优化方案**：

1. **品牌打散**：
   ```
   规则：前10个结果中，同一品牌最多出现3次
   
   算法：
   1. 按相关性排序
   2. 遍历结果，统计品牌出现次数
   3. 如果某品牌超过阈值，跳过该商品，选下一个
   ```

2. **MMR算法（最大边际相关性）**：
   ```
   score = λ × relevance - (1-λ) × max_similarity
   
   relevance: 与查询的相关性
   max_similarity: 与已选结果的最大相似度
   λ: 权衡参数（0.7）
   
   每次选择score最高的商品，保证相关性和多样性
   ```

3. **类目多样性**：
   ```
   前10个结果覆盖2-3个子类目
   - 智能手机（5个）
   - 老人机（3个）
   - 游戏手机（2个）
   ```

**延伸思考**：
1. 多样性和相关性如何权衡？
2. 如何评估搜索结果的多样性？

---
