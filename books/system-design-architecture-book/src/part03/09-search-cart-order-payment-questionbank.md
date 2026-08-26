# 第 41 章 电商架构面试题精选（三）：搜索、购物车、订单与支付

> 本章是电商架构面试题库的第三部分，题库使用说明与面试官导航见[第 43 章](./03-ecommerce-architecture-interview.md)。

## 41.1 搜索、购物车、订单与支付题库

本专题聚焦电商交易核心链路，按题型拆成四个子章节：

- [35.3.1 搜索与导购](./03-ecommerce-architecture-interview.md)
- [35.3.2 购物车与结算](./03-ecommerce-architecture-interview.md)
- [35.3.3 订单系统](./03-ecommerce-architecture-interview.md)
- [35.3.4 支付系统](./03-ecommerce-architecture-interview.md)

建议按用户链路顺序阅读：搜索发现、加入购物车、结算创单、支付闭环。

---

### 41.1.1 搜索与导购（10题）

##### 📊 题目1：电商搜索引擎的架构设计

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

##### 🔧 题目2：搜索相关性排序算法设计

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

##### 💡 题目3：搜索建议（Suggest）的实现

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

##### 📊 题目4：商品筛选和多维度过滤的设计

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

##### 🔧 题目5：搜索结果的无结果优化

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

##### 📊 题目6：搜索日志分析与优化

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

##### 🔧 题目7：跨境电商的多语言搜索

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

##### 💡 题目8：搜索性能优化

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

##### 📊 题目9：智能搜索（NLP+AI）

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

##### 🔧 题目10：搜索结果的多样性优化

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

---

### 41.1.2 购物车与结算（15题）

##### 📊 题目1：购物车的数据存储设计

**问题描述**：
用户将商品加入购物车，需要跨设备同步（手机APP、Web、小程序）。如何设计购物车的存储方案？

**答案**：

**问题分析**：
购物车的核心要素：
1. 跨设备同步
2. 用户未登录也能加购
3. 数据持久化
4. 高并发读写

**方案一：Cookie存储**

核心思想：
购物车数据存储在浏览器Cookie。

优点：
- 无需服务器存储
- 减轻服务器压力

缺点：
- 不能跨设备
- Cookie大小限制（4KB）
- 不安全（可被篡改）

适用场景：
- 简单电商
- 临时购物车

**方案二：数据库存储（推荐）**

核心思想：
购物车存储在MySQL/Redis。

设计：
```sql
shopping_cart
├── cart_id
├── user_id
├── sku_id
├── quantity
├── selected（是否选中，用于结算）
├── added_at
└── updated_at

索引：
- PRIMARY KEY (cart_id)
- UNIQUE KEY (user_id, sku_id)
- INDEX (user_id)
```

优点：
- 跨设备同步
- 数据持久化
- 支持复杂操作

缺点：
- 服务器存储成本

**方案三：Redis+MySQL双写**

核心思想：
Redis提供高性能，MySQL保证持久化。

架构：
```text
写操作：
1. 写Redis（立即返回）
2. 异步写MySQL

读操作：
1. 优先读Redis
2. Redis不存在，读MySQL
3. 回写Redis
```

优点：
- 性能高
- 数据安全

缺点：
- 数据同步复杂

**推荐方案**：
采用**Redis+MySQL双写**。

实施要点：

1. **未登录用户**：
   ```
   未登录：
   - 生成临时cart_id（存Cookie）
   - 购物车数据存Redis
   - key: cart:temp:{cart_id}
   
   登录后：
   - 合并临时购物车到用户购物车
   - 删除临时购物车
   ```

2. **购物车合并**：
   ```java
   public void mergeCart(String tempCartId, Long userId) {
     List<CartItem> tempItems = getTempCart(tempCartId);
     List<CartItem> userItems = getUserCart(userId);
     
     for (CartItem temp : tempItems) {
       CartItem exist = findItem(userItems, temp.getSkuId());
       if (exist != null) {
         // 已存在，数量相加
         exist.setQuantity(exist.getQuantity() + temp.getQuantity());
       } else {
         // 不存在，添加
         userItems.add(temp);
       }
     }
     
     saveUserCart(userId, userItems);
     deleteTempCart(tempCartId);
   }
   ```

3. **失效商品处理**：
   ```
   商品失效场景：
   - 商品下架
   - 商品删除
   - 库存不足
   
   展示：
   - 失效商品置灰
   - 提示"商品已下架"
   - 提供"删除"或"移入收藏"选项
   ```

4. **购物车清理**：
   ```
   定时任务（每天凌晨）：
   - 删除90天未更新的购物车
   - 减少存储成本
   ```

5. **购物车同步**：
   ```
   跨设备同步：
   - 用户在APP加购 → 写Redis+MySQL
   - 用户在Web打开 → 读Redis → 显示购物车
   
   实时同步（WebSocket）：
   - 用户在设备A加购
   - 推送到设备B
   - 设备B实时更新购物车数量
   ```

**延伸思考**：
1. 购物车数量显示在导航栏，如何实时更新？
2. 如何处理购物车中的促销信息过期？
3. 购物车数据如何备份和恢复？

---

##### 🔧 题目2：购物车的价格计算

**问题描述**：
购物车中有多个商品，每个商品可能有不同促销（满减、折扣、优惠券）。如何设计购物车的实时价格计算？

**答案**：

**问题分析**：
购物车价格计算的复杂性：
1. 多商品组合
2. 多种促销叠加
3. 实时计算（用户修改数量即刻更新）
4. 价格明细展示

**推荐方案**：

价格计算引擎：
```java
public CartPrice calculateCart(Cart cart) {
  BigDecimal originalPrice = BigDecimal.ZERO;
  BigDecimal discountAmount = BigDecimal.ZERO;
  
  // 1. 计算商品级优惠
  for (CartItem item : cart.getItems()) {
    originalPrice = originalPrice.add(
      item.getPrice().multiply(new BigDecimal(item.getQuantity()))
    );
    
    // 商品折扣
    if (item.hasDiscount()) {
      BigDecimal itemDiscount = calculateItemDiscount(item);
      discountAmount = discountAmount.add(itemDiscount);
    }
  }
  
  // 2. 计算订单级优惠
  BigDecimal subtotal = originalPrice.subtract(discountAmount);
  
  // 满减
  BigDecimal fullReduceDiscount = calculateFullReduce(subtotal);
  discountAmount = discountAmount.add(fullReduceDiscount);
  
  // 优惠券
  if (cart.hasCoupon()) {
    BigDecimal couponDiscount = calculateCoupon(cart.getCoupon(), subtotal);
    discountAmount = discountAmount.add(couponDiscount);
  }
  
  // 3. 最终价格
  BigDecimal finalPrice = originalPrice.subtract(discountAmount);
  
  return new CartPrice(originalPrice, discountAmount, finalPrice);
}
```

实时计算触发：
```text
触发时机：
- 用户修改商品数量
- 用户选择/取消优惠券
- 用户勾选/取消商品
- 商品价格变动（后台推送）

性能优化：
- 防抖（用户停止操作500ms后计算）
- 缓存（相同购物车缓存5分钟）
```

**延伸思考**：
1. 购物车价格和下单后价格不一致如何处理？
2. 大促时购物车价格计算如何优化性能？

---

##### 💡 题目3：购物车的推荐功能

**问题描述**：
用户购物车中有商品A，如何推荐相关商品B，提升客单价？

**答案**：

**推荐策略**：

1. **关联推荐**：
   ```
   "买了还买"：
   - 统计购买商品A的用户还购买了哪些商品
   - 推荐高频商品
   
   示例：
   购物车有"iPhone 15" → 推荐"手机壳"、"钢化膜"、"充电器"
   ```

2. **凑单推荐**：
   ```
   购物车总价¥180
   满¥200减¥30
   
   推荐：再买¥20-30的商品，即可享受优惠
   ```

3. **替代推荐**：
   ```
   购物车中商品缺货 → 推荐同类商品
   ```

**延伸思考**：
1. 购物车推荐如何避免打扰用户？
2. 推荐商品点击率如何提升？

---

##### 📊 题目4：购物车的库存校验

**问题描述**：
用户加购物车时商品有货，结算时可能已无货。如何设计购物车的库存校验机制？

**答案**：

**校验时机**：

1. **加购时校验**：
   ```
   用户点击"加入购物车" → 检查库存
   库存充足 → 允许加购
   库存不足 → 提示"库存不足"
   ```

2. **结算时校验**：
   ```
   用户点击"去结算" → 
   1. 批量查询购物车所有商品库存
   2. 标记缺货商品
   3. 展示：
      - 有货商品（可结算）
      - 缺货商品（置灰，不可结算）
   ```

3. **实时推送**：
   ```
   商品库存变化（如售罄） → WebSocket推送
   前端实时更新购物车状态
   ```

**延伸思考**：
1. 购物车中的商品是否需要预占库存？
2. 库存不足时如何引导用户？

---

##### 🔧 题目5：购物车的性能优化

**问题描述**：
大促期间，购物车服务QPS达10万+，如何优化购物车性能？

**答案**：

**优化方案**：

1. **读写分离**：
   ```
   写操作（加购、删除）：
   - 写MySQL主库
   - 异步同步到Redis
   
   读操作（查询购物车）：
   - 读Redis（快）
   - 未命中读MySQL从库
   ```

2. **批量操作**：
   ```
   ❌ 单个加购：N次请求
   ✅ 批量加购：1次请求
   
   POST /api/cart/batch-add
   {
     "items": [
       {"skuId": "123", "quantity": 2},
       {"skuId": "456", "quantity": 1}
     ]
   }
   ```

3. **本地缓存**：
   ```
   热点用户购物车：
   - 加载到应用服务器内存
   - 减少Redis访问
   ```

4. **限流降级**：
   ```
   限流：
   - 单用户购物车操作频率限制（10次/分钟）
   
   降级：
   - Redis故障 → 降级到MySQL
   - MySQL故障 → 只读模式（不能加购）
   ```

**延伸思考**：
1. 购物车数据如何分片（sharding）？
2. 购物车服务如何实现高可用？

---

##### 📊 题目6：购物车商品失效的处理策略

**问题描述**：
用户购物车中的商品可能因为下架、删除、库存清零而失效。如何设计失效商品的处理策略，优化用户体验？

**答案**：

**问题分析**：
商品失效场景：
1. 商品下架（运营操作）
2. 商品删除（商品不再销售）
3. 库存售罄（暂时缺货）
4. 商品涨价（价格变动）
5. 促销过期（活动结束）

**方案一：定时批量检测**

核心思想：
定时任务扫描购物车，标记失效商品。

实现：
```text
定时任务（每小时）：
1. 查询所有购物车商品
2. 批量查询商品状态
3. 标记失效商品
4. 更新购物车
```

优点：
- 批量处理，效率高
- 服务器压力均匀

缺点：
- 实时性差（最长延迟1小时）
- 用户可能看到失效商品

**方案二：实时校验（推荐）**

核心思想：
用户打开购物车时，实时校验商品状态。

流程：
```text
用户打开购物车 →
1. 查询购物车商品列表
2. 批量查询商品最新状态（Redis缓存）
3. 分类展示：
   - 正常商品（可结算）
   - 失效商品（置灰，不可结算）
4. 标注失效原因
```

失效商品展示：
```text
[置灰显示]
iPhone 15 Pro 256GB
¥7999
状态：该商品已下架
操作：[删除] [移入收藏夹]
```

优点：
- 实时性好
- 用户体验清晰

缺点：
- 每次打开购物车都校验
- QPS增加

**方案三：消息推送**

核心思想：
商品状态变化时，主动推送更新购物车。

架构：
```text
商品下架 → 
发布事件（Kafka）→ 
购物车Worker消费 →
1. 查询包含该商品的购物车
2. 标记商品为失效
3. WebSocket推送用户（如果在线）
```

优点：
- 实时性最好
- 用户感知及时

缺点：
- 架构复杂
- 需要消息队列

**方案对比**：

| 方案 | 实时性 | 用户体验 | 实施难度 | 系统负载 |
|------|--------|---------|---------|---------|
| 定时检测 | ★★☆☆☆ | ★★★☆☆ | ★★★★★ | ★★★★☆ |
| 实时校验 | ★★★★☆ | ★★★★★ | ★★★★☆ | ★★★☆☆ |
| 消息推送 | ★★★★★ | ★★★★★ | ★★☆☆☆ | ★★★★☆ |

**推荐方案**：
采用**实时校验+消息推送**的组合。

实施要点：

1. **商品状态缓存**：
   ```
   Redis存储商品状态：
   key: product:status:{skuId}
   value: {
     "onSale": true,
     "stock": 100,
     "price": 7999,
     "promotionId": "xxx",
     "updatedAt": 1679800000
   }
   TTL: 10分钟
   
   商品变更时主动刷新
   ```

2. **批量校验优化**：
   ```java
   public Map<String, ProductStatus> batchCheckStatus(List<String> skuIds) {
     // 1. 批量查询Redis
     List<String> keys = skuIds.stream()
       .map(id -> "product:status:" + id)
       .collect(Collectors.toList());
     
     List<ProductStatus> cached = redis.mget(keys);
     
     // 2. 未命中的查数据库
     Set<String> missingIds = findMissingIds(cached);
     if (!missingIds.isEmpty()) {
       Map<String, ProductStatus> fromDB = queryFromDB(missingIds);
       // 写回Redis
       cacheToRedis(fromDB);
       cached.addAll(fromDB.values());
     }
     
     return toMap(cached);
   }
   ```

3. **失效商品操作**：
   ```
   用户操作：
   1. 删除：直接从购物车删除
   2. 移入收藏夹：
      - 加入收藏
      - 从购物车删除
      - 商品恢复上架时通知用户
   3. 查看替代品：
      - 推荐同类商品
      - 一键替换
   ```

4. **主动通知**：
   ```
   通知策略：
   - 商品下架 → App推送
     "您购物车中的【iPhone 15】已下架"
   - 商品降价 → App推送
     "您购物车中的【iPhone 15】降价了"
   - 库存恢复 → 收藏夹商品有货通知
   ```

5. **失效原因分类**：
   ```
   原因分类：
   - 已下架：运营下架
   - 已售罄：库存为0
   - 已删除：商品不存在
   - 已涨价：价格变动超过10%
   - 活动结束：促销过期
   
   针对性提示：
   - 已售罄 → "补货中，可先收藏"
   - 已涨价 → "当前价格¥xxx，加购时¥xxx"
   ```

**延伸思考**：
1. 如何设计购物车的自动清理（失效商品30天后自动删除）？
2. 失效商品是否计入购物车数量显示？
3. 如何处理部分失效（如只有某个规格缺货）？

---

##### 🔧 题目7：购物车的跨平台同步设计

**问题描述**：
用户在手机APP加购商品，打开电脑Web也能看到。如何实现购物车的跨平台实时同步？

**答案**：

**问题分析**：
跨平台同步的核心要素：
1. 数据一致性（同一购物车）
2. 实时性（秒级同步）
3. 冲突处理（同时操作）
4. 离线支持

**方案一：轮询同步**

核心思想：
客户端定时轮询服务器，获取最新购物车。

实现：
```javascript
// 前端定时轮询
setInterval(() => {
  fetch('/api/cart')
    .then(res => res.json())
    .then(cart => {
      if (cart.version > localVersion) {
        updateLocalCart(cart);
      }
    });
}, 5000); // 每5秒轮询一次
```

优点：
- 实现简单
- 兼容性好

缺点：
- 实时性差（5秒延迟）
- 浪费带宽（大部分请求无变化）
- 服务器压力大

**方案二：WebSocket推送（推荐）**

核心思想：
客户端与服务器建立长连接，服务器主动推送更新。

架构：
```text
用户A在APP加购 →
1. APP发送请求到服务器
2. 服务器更新购物车
3. 服务器通过WebSocket推送到用户A的所有设备
4. Web端接收推送，更新购物车显示

WebSocket消息格式：
{
  "type": "CART_UPDATE",
  "action": "ADD_ITEM",
  "data": {
    "skuId": "123",
    "quantity": 2
  },
  "version": 10,
  "timestamp": 1679800000
}
```

实现：
```java
// 服务端
@Service
public class CartService {
  @Autowired
  private WebSocketPushService pushService;
  
  public void addToCart(Long userId, String skuId, int quantity) {
    // 1. 更新购物车
    Cart cart = updateCart(userId, skuId, quantity);
    
    // 2. 推送到该用户所有在线设备
    CartUpdateMessage msg = new CartUpdateMessage(
      "ADD_ITEM", skuId, quantity, cart.getVersion()
    );
    pushService.pushToUser(userId, msg);
  }
}

// 客户端
websocket.onmessage = (event) => {
  const msg = JSON.parse(event.data);
  if (msg.type === 'CART_UPDATE') {
    // 更新本地购物车
    if (msg.version > localCartVersion) {
      applyCartUpdate(msg);
    }
  }
};
```

优点：
- 实时性好（秒级）
- 双向通信
- 节省带宽

缺点：
- 需要维护长连接
- 服务器成本高
- 需要心跳保活

**方案三：长轮询**

核心思想：
客户端发起请求，服务器hold住请求，有更新时返回。

实现：
```javascript
function longPoll() {
  fetch('/api/cart/poll?version=' + localVersion)
    .then(res => res.json())
    .then(cart => {
      if (cart.version > localVersion) {
        updateLocalCart(cart);
      }
      // 立即发起下一次轮询
      longPoll();
    })
    .catch(() => {
      // 失败后延迟重试
      setTimeout(longPoll, 5000);
    });
}
```

优点：
- 实时性较好
- 兼容性好（不需要WebSocket）

缺点：
- 服务器需要hold请求
- 连接可能超时

**方案对比**：

| 方案 | 实时性 | 服务器成本 | 兼容性 | 实施难度 |
|------|--------|-----------|--------|---------|
| 轮询 | ★★☆☆☆ | ★★☆☆☆ | ★★★★★ | ★★★★★ |
| WebSocket | ★★★★★ | ★★★☆☆ | ★★★★☆ | ★★★☆☆ |
| 长轮询 | ★★★★☆ | ★★☆☆☆ | ★★★★★ | ★★★★☆ |

**推荐方案**：
采用**WebSocket推送**（支持WebSocket）+ **轮询兜底**（不支持时降级）。

实施要点：

1. **连接管理**：
   ```java
   // 用户连接映射
   Map<Long, Set<WebSocketSession>> userSessions = new ConcurrentHashMap<>();
   
   // 用户连接时
   public void onConnect(Long userId, WebSocketSession session) {
     userSessions.computeIfAbsent(userId, k -> new ConcurrentHashSet<>())
       .add(session);
   }
   
   // 用户断开时
   public void onDisconnect(Long userId, WebSocketSession session) {
     Set<WebSocketSession> sessions = userSessions.get(userId);
     if (sessions != null) {
       sessions.remove(session);
     }
   }
   
   // 推送消息
   public void pushToUser(Long userId, Object message) {
     Set<WebSocketSession> sessions = userSessions.get(userId);
     if (sessions != null) {
       for (WebSocketSession session : sessions) {
         if (session.isOpen()) {
           session.sendMessage(new TextMessage(JSON.toJSONString(message)));
         }
       }
     }
   }
   ```

2. **版本控制**：
   ```
   购物车版本号：
   - 每次修改version+1
   - 客户端记录本地version
   - 接收推送时检查version
   - 如果本地version更新，忽略旧推送
   
   冲突解决：
   - 客户端操作携带version
   - 服务端CAS更新
   - 失败则拉取最新数据重试
   ```

3. **心跳保活**：
   ```javascript
   // 客户端定时发送心跳
   setInterval(() => {
     if (websocket.readyState === WebSocket.OPEN) {
       websocket.send(JSON.stringify({type: 'PING'}));
     }
   }, 30000); // 每30秒
   
   // 服务端响应心跳
   if (message.type === 'PING') {
     session.sendMessage(new TextMessage('{"type":"PONG"}'));
   }
   ```

4. **降级策略**：
   ```javascript
   // 检测WebSocket支持
   if ('WebSocket' in window) {
     connectWebSocket();
   } else {
     // 降级到轮询
     setInterval(pollCart, 10000);
   }
   
   // WebSocket断开时降级
   websocket.onclose = () => {
     console.log('WebSocket断开，降级到轮询');
     setInterval(pollCart, 10000);
   };
   ```

5. **离线支持**：
   ```
   离线操作：
   1. 用户离线时，操作保存到本地队列
   2. 用户上线后，批量同步到服务器
   3. 服务器合并操作，返回最终购物车
   
   冲突处理：
   - 添加：合并数量
   - 删除：以最新操作为准
   - 修改：以最新操作为准
   ```

**延伸思考**：
1. 如何处理网络不稳定导致的频繁重连？
2. 跨平台同步如何支持多账号（家庭共享）？
3. WebSocket服务如何实现横向扩展？

---

##### 💡 题目8：购物车推荐算法设计

**问题描述**：
用户购物车有"iPhone 15"，如何推荐相关商品（配件、保险、AppleCare）提升客单价？

**答案**：

**问题分析**：
购物车推荐的核心目标：
1. 提升客单价（关联销售）
2. 提升转化率（凑单满减）
3. 提升用户体验（需要的商品）

**推荐策略**：

1. **关联推荐（Frequently Bought Together）**：
   ```sql
   -- 统计商品关联
   SELECT b.sku_id, COUNT(*) as frequency
   FROM order_items a
   JOIN order_items b ON a.order_id = b.order_id
   WHERE a.sku_id = 'iPhone15' 
     AND b.sku_id != 'iPhone15'
   GROUP BY b.sku_id
   ORDER BY frequency DESC
   LIMIT 10;
   
   结果：
   - 手机壳（购买率80%）
   - 钢化膜（购买率70%）
   - 充电器（购买率60%）
   ```

2. **凑单推荐**：
   ```
   购物车总价：¥180
   满减活动：满¥200减¥30
   
   推荐策略：
   - 推荐价格在¥20-¥50的商品
   - 优先推荐与购物车商品相关的
   - 标注"再买¥20即享满减"
   ```

3. **类目互补推荐**：
   ```
   购物车有"相机" → 推荐：
   - 存储卡
   - 相机包
   - 三脚架
   
   购物车有"婴儿奶粉" → 推荐：
   - 奶瓶
   - 尿不湿
   - 湿巾
   ```

4. **个性化推荐**：
   ```
   基于用户历史：
   - 用户A经常买Apple产品
     → 推荐AppleCare+、AirPods
   - 用户B价格敏感
     → 推荐高性价比配件
   ```

**实施要点**：

1. **关联规则挖掘**：
   ```python
   # 使用Apriori算法
   from mlxtend.frequent_patterns import apriori, association_rules
   
   # 构建购物篮矩阵
   basket = orders.groupby(['order_id', 'sku_id'])['quantity'].sum().unstack().fillna(0)
   basket = basket.applymap(lambda x: 1 if x > 0 else 0)
   
   # 挖掘频繁项集
   frequent_itemsets = apriori(basket, min_support=0.01, use_colnames=True)
   
   # 生成关联规则
   rules = association_rules(frequent_itemsets, metric="confidence", min_threshold=0.5)
   
   # iPhone15 -> 手机壳 (confidence=0.8, lift=2.5)
   ```

2. **推荐展示位置**：
   ```
   位置1：购物车下方
   "买了还买"：展示3-5个商品
   
   位置2：结算页
   "凑单优惠"：满减差额商品
   
   位置3：加购弹窗
   用户加购商品A → 弹窗推荐配件B
   ```

3. **推荐排序**：
   ```
   score = w1 × 关联度 + 
           w2 × 利润率 + 
           w3 × 库存充足度 +
           w4 × 用户个性化得分
   
   w1=0.4, w2=0.3, w3=0.2, w4=0.1
   ```

4. **AB测试**：
   ```
   测试维度：
   - A组：展示3个推荐
   - B组：展示5个推荐
   - C组：不展示推荐
   
   评估指标：
   - 推荐点击率
   - 推荐加购率
   - 客单价提升
   ```

**延伸思考**：
1. 推荐商品如何避免干扰用户（显得推销）？
2. 推荐算法如何冷启动（新商品无关联数据）？
3. 推荐效果如何评估和持续优化？

---

##### 📊 题目9：购物车的结算流程设计

**问题描述**：
用户点击"去结算"，进入结算页面，需要选择地址、优惠券、支付方式。如何设计结算流程？

**答案**：

**问题分析**：
结算流程的核心环节：
1. 确认商品（数量、价格）
2. 选择收货地址
3. 选择配送方式
4. 应用优惠（优惠券、积分）
5. 选择支付方式
6. 提交订单

**方案一：单页结算**

核心思想：
所有信息在一个页面完成。

页面布局：
```text
结算页：
┌─────────────────┐
│ 1. 收货地址      │
│ [北京市朝阳区...] │
├─────────────────┤
│ 2. 商品清单      │
│ iPhone 15 × 1   │
│ ¥7999           │
├─────────────────┤
│ 3. 配送方式      │
│ ○ 标准配送（免费）│
│ ○ 次日达（¥10）  │
├─────────────────┤
│ 4. 优惠         │
│ 优惠券：¥30     │
│ 积分抵扣：¥10   │
├─────────────────┤
│ 5. 支付方式      │
│ ○ 支付宝        │
│ ○ 微信支付      │
├─────────────────┤
│ 总计：¥7959     │
│ [提交订单]       │
└─────────────────┘
```

优点：
- 流程简洁
- 一目了然
- 减少跳转

缺点：
- 页面信息多
- 移动端显示困难

**方案二：分步结算（推荐）**

核心思想：
分多个步骤完成结算。

流程：
```text
步骤1：选择地址
→ 步骤2：确认商品和配送
→ 步骤3：选择优惠
→ 步骤4：支付
```

优点：
- 逻辑清晰
- 移动端友好
- 可保存中间状态

缺点：
- 步骤多
- 可能流失

**推荐方案**：
PC端使用**单页结算**，移动端使用**分步结算**。

实施要点：

1. **结算前校验**：
   ```java
   public CheckoutResult preCheckout(Long userId) {
     // 1. 获取购物车
     Cart cart = getCart(userId);
     
     // 2. 校验商品状态
     List<String> invalidItems = new ArrayList<>();
     for (CartItem item : cart.getItems()) {
       Product product = productService.getProduct(item.getSkuId());
       if (!product.isOnSale()) {
         invalidItems.add(item.getSkuId() + "：已下架");
       } else if (product.getStock() < item.getQuantity()) {
         invalidItems.add(item.getSkuId() + "：库存不足");
       }
     }
     
     if (!invalidItems.isEmpty()) {
       return CheckoutResult.fail("部分商品无法结算", invalidItems);
     }
     
     // 3. 计算价格
     PriceDetail price = calculatePrice(cart);
     
     // 4. 返回结算信息
     return CheckoutResult.success(cart, price);
   }
   ```

2. **地址选择**：
   ```
   展示用户地址列表：
   - 默认地址（置顶）
   - 最近使用地址
   - 其他地址
   
   新增地址：
   - 省市区三级联动
   - 详细地址输入
   - 联系人和电话
   - 设为默认地址
   ```

3. **优惠券选择**：
   ```
   展示可用优惠券：
   - 按优惠力度排序
   - 标注"最优"推荐
   - 显示使用门槛
   
   自动选择：
   - 默认选择优惠最大的券
   - 用户可手动切换
   
   不可用优惠券：
   - 置灰显示
   - 标注不可用原因（如"不满足使用条件"）
   ```

4. **价格实时计算**：
   ```javascript
   // 监听用户操作
   onChange = () => {
     // 防抖：用户停止操作500ms后计算
     clearTimeout(this.timer);
     this.timer = setTimeout(() => {
       this.calculatePrice();
     }, 500);
   };
   
   calculatePrice = async () => {
     const params = {
       items: this.state.cartItems,
       addressId: this.state.selectedAddress,
       couponId: this.state.selectedCoupon,
       usePoints: this.state.usePoints
     };
     
     const result = await API.post('/api/order/calculate-price', params);
     this.setState({ priceDetail: result });
   };
   ```

5. **订单确认信息**：
   ```
   最终确认页展示：
   - 收货人：张三 138****1234
   - 收货地址：北京市朝阳区xxx
   - 商品清单：iPhone 15 × 1
   - 配送方式：标准配送（预计3天送达）
   - 优惠明细：
     * 商品折扣：-¥100
     * 满减优惠：-¥30
     * 优惠券：-¥20
   - 实付金额：¥7849
   
   用户确认无误后点击"提交订单"
   ```

**延伸思考**：
1. 如何设计结算页的防重复提交？
2. 结算过程中价格变动如何处理？
3. 结算流程如何优化转化率？

---

##### 🔧 题目10：购物车的分享功能设计

**问题描述**：
用户想分享购物车给朋友（如"帮我看看这些商品怎么样"），如何设计购物车分享功能？

**答案**：

**问题分析**：
购物车分享的核心场景：
1. 征求意见（送礼选择）
2. 代购（帮朋友买）
3. 拼单（一起买更便宜）

**方案一：生成分享链接**

核心思想：
生成唯一URL，包含购物车商品信息。

实现：
```text
生成分享：
1. 用户点击"分享购物车"
2. 服务端生成分享ID
3. 保存分享内容到数据库/Redis
4. 返回分享链接

分享链接：
https://example.com/cart/share/abc123

接收分享：
1. 朋友点击链接
2. 展示分享者的购物车商品
3. 可一键导入到自己购物车
```

数据设计：
```sql
cart_share
├── share_id（唯一ID）
├── user_id（分享者）
├── cart_snapshot（JSON，购物车快照）
├── expire_at（过期时间）
├── view_count（查看次数）
└── created_at
```

优点：
- 实现简单
- 支持任意平台

缺点：
- 链接可能泄露
- 分享内容是快照（不会实时更新）

**方案二：生成二维码**

核心思想：
生成二维码，扫码查看购物车。

实现：
```text
生成二维码：
1. 生成分享链接（同方案一）
2. 将链接转为二维码
3. 展示二维码供分享

扫码查看：
1. 扫描二维码
2. 跳转到分享页面
3. 展示商品列表
```

优点：
- 线下分享方便
- 移动端友好

缺点：
- 仍是快照

**方案三：实时共享购物车（推荐）**

核心思想：
创建共享购物车，多人实时协同。

实现：
```text
创建共享：
1. 用户创建共享购物车
2. 生成共享ID和密码（可选）
3. 邀请朋友加入

实时同步：
- 任何人添加/删除商品
- 通过WebSocket实时同步给所有成员
- 显示"张三添加了iPhone 15"

共享购物车表：
shared_cart
├── shared_cart_id
├── creator_id
├── name（如"周末采购清单"）
├── password（可选）
├── members（成员列表）
├── items（商品列表）
└── created_at
```

优点：
- 实时协同
- 支持多人编辑
- 适合家庭、团队采购

缺点：
- 实现复杂
- 需要冲突处理

**推荐方案**：
采用**分享链接+实时共享**的组合。

实施要点：

1. **分享类型**：
   ```
   类型1：只读分享
   - 生成分享链接
   - 朋友只能查看，不能修改
   - 可一键导入到自己购物车
   
   类型2：协同编辑
   - 创建共享购物车
   - 邀请成员
   - 成员可添加/删除商品
   ```

2. **分享页面设计**：
   ```
   分享页头部：
   "张三分享了购物车给你"
   
   商品列表：
   [展示所有商品]
   
   操作按钮：
   - [全部加入我的购物车]
   - [选择部分加入]
   - [保存为我的收藏清单]
   ```

3. **隐私控制**：
   ```
   隐私选项：
   - 公开：任何人都可查看
   - 仅好友：需要登录且是好友
   - 密码保护：需要输入密码
   
   敏感信息隐藏：
   - 不显示价格（可选）
   - 不显示数量（可选）
   ```

4. **分享统计**：
   ```
   统计指标：
   - 分享次数
   - 查看人数
   - 转化人数（查看后购买）
   - 传播路径（A分享给B，B分享给C）
   ```

5. **场景化推荐**：
   ```
   场景1：送礼征询
   "想送女朋友礼物，帮我选一个"
   → 展示多个候选商品
   → 朋友投票或评论
   
   场景2：拼单
   "一起买，更便宜"
   → 共享购物车
   → 凑满减金额
   → 分摊运费
   ```

**延伸思考**：
1. 如何设计购物车的协同冲突解决（同时删除同一商品）？
2. 分享购物车如何防止恶意刷单？
3. 共享购物车如何拆单结算（各付各的）？

---

##### 💡 题目11：购物车的满减凑单提示

**问题描述**：
购物车总价¥180，有满¥200减¥30活动。如何设计智能凑单提示，引导用户加购？

**答案**：

**推荐方案**：

1. **差额计算**：
   ```
   当前金额：¥180
   满减门槛：¥200
   差额：¥20
   
   提示："再买¥20，立减¥30"
   ```

2. **智能商品推荐**：
   ```
   推荐商品筛选条件：
   - 价格在¥20-¥50之间（差额附近）
   - 与购物车商品相关（配件、同类目）
   - 库存充足
   - 高评分
   
   排序：
   - 优先推荐价格接近差额的
   - 优先推荐关联度高的
   ```

3. **视觉引导**：
   ```
   进度条展示：
   [████████░░] 90% (¥180/¥200)
   "再买¥20，立减¥30，相当于打8.5折"
   
   推荐商品卡片：
   ┌───────────┐
   │ 手机壳     │
   │ ¥29       │
   │ [加入购物车]│
   └───────────┘
   ```

4. **多档位满减**：
   ```
   满减档位：
   - 满¥100减¥10（已达成✓）
   - 满¥200减¥30（差¥20）
   - 满¥500减¥100（差¥320）
   
   提示优先显示最接近的下一档
   ```

**延伸思考**：
1. 凑单推荐如何避免过度营销（让用户反感）？
2. 多个满减活动同时存在时如何提示？

---

##### 📊 题目12：购物车的批量操作设计

**问题描述**：
用户购物车有50个商品，想批量删除、批量加入收藏。如何设计批量操作功能？

**答案**：

**推荐方案**：

1. **批量选择**：
   ```
   界面设计：
   [全选] 已选0件
   
   ☑ 商品A  ¥100
   ☑ 商品B  ¥200
   ☐ 商品C  ¥300
   
   批量操作：
   [删除选中] [加入收藏] [移除失效商品]
   ```

2. **批量接口**：
   ```java
   POST /api/cart/batch-delete
   {
     "skuIds": ["123", "456", "789"]
   }
   
   POST /api/cart/batch-move-to-favorite
   {
     "skuIds": ["123", "456"]
   }
   ```

3. **事务处理**：
   ```
   批量操作的事务性：
   - 部分成功部分失败如何处理？
   
   方案A：全量事务
   - 全部成功才提交
   - 任一失败全部回滚
   
   方案B：部分成功（推荐）
   - 成功的操作提交
   - 失败的返回错误信息
   - 前端展示"成功X件，失败Y件"
   ```

4. **性能优化**：
   ```
   批量删除50个商品：
   ❌ for循环50次DELETE
   ✅ 一次DELETE WHERE sku_id IN (...)
   
   批量更新库存：
   ❌ 50次UPDATE
   ✅ 批量UPDATE CASE WHEN
   ```

**延伸思考**：
1. 批量操作如何支持撤销（Undo）？
2. 批量操作的进度如何展示？

---

##### 🔧 题目13：购物车的收藏夹联动

**问题描述**：
购物车和收藏夹如何联动？商品从购物车移入收藏，或从收藏加入购物车。

**答案**：

**推荐方案**：

1. **数据模型**：
   ```sql
   favorite
   ├── favorite_id
   ├── user_id
   ├── sku_id
   ├── source（CART/BROWSE）
   ├── added_at
   └── ...
   ```

2. **互相转换**：
   ```
   购物车 → 收藏夹：
   1. 用户点击"移入收藏"
   2. 加入收藏夹
   3. 从购物车删除
   4. 提示"已移入收藏夹"
   
   收藏夹 → 购物车：
   1. 用户点击"加入购物车"
   2. 加入购物车
   3. 保留在收藏夹（不删除）
   ```

3. **降价提醒**：
   ```
   收藏商品降价：
   - 监控收藏商品价格
   - 降价时推送通知
   - 引导用户加购
   ```

**延伸思考**：
1. 收藏夹和购物车的区别是什么？
2. 如何设计收藏夹的分组功能？

---

##### 💡 题目14：购物车的历史记录

**问题描述**：
用户删除了购物车商品，想恢复。如何设计购物车的历史记录功能？

**答案**：

**推荐方案**：

1. **软删除**：
   ```sql
   shopping_cart
   ├── ...
   ├── deleted_at（软删除标记）
   └── deleted（是否删除）
   
   查询购物车：
   SELECT * FROM shopping_cart 
   WHERE user_id=? AND deleted=0
   
   查询历史：
   SELECT * FROM shopping_cart 
   WHERE user_id=? AND deleted=1
   ORDER BY deleted_at DESC
   ```

2. **恢复功能**：
   ```
   历史记录页面：
   最近删除：
   - 商品A（3天前删除）[恢复]
   - 商品B（7天前删除）[恢复]
   
   恢复操作：
   UPDATE shopping_cart 
   SET deleted=0, deleted_at=NULL
   WHERE cart_id=?
   ```

3. **自动清理**：
   ```
   定时任务：
   - 删除30天后的历史记录
   - 减少存储成本
   ```

**延伸思考**：
1. 购物车历史记录是否需要版本控制（记录每次修改）？
2. 如何设计购物车的快照功能（保存多个购物清单）？

---

##### 📊 题目15：购物车的AB测试设计

**问题描述**：
想测试新的购物车布局对转化率的影响。如何设计购物车的AB测试？

**答案**：

**推荐方案**：

1. **分流策略**：
   ```java
   public String getCartVersion(Long userId) {
     // 基于用户ID哈希分流
     int hash = userId.hashCode();
     if (hash % 2 == 0) {
       return "A"; // 对照组
     } else {
       return "B"; // 实验组
     }
   }
   ```

2. **实验设计**：
   ```
   对照组A（50%用户）：
   - 旧购物车布局
   
   实验组B（50%用户）：
   - 新购物车布局（优化后）
   
   评估指标：
   - 加购率
   - 结算率
   - 转化率
   - 客单价
   ```

3. **数据埋点**：
   ```javascript
   // 购物车页面浏览
   track('cart_view', {
     version: 'A', // 或 'B'
     cartItemCount: 5
   });
   
   // 点击结算
   track('cart_checkout_click', {
     version: 'A',
     cartTotal: 1000
   });
   
   // 完成下单
   track('order_created', {
     version: 'A',
     orderAmount: 1000
   });
   ```

4. **结果分析**：
   ```
   结果对比：
   | 指标 | A组 | B组 | 提升 |
   |------|-----|-----|------|
   | 结算率 | 60% | 65% | +8.3% |
   | 转化率 | 40% | 45% | +12.5% |
   | 客单价 | ¥800 | ¥850 | +6.25% |
   
   结论：B组效果更好，全量发布
   ```

**延伸思考**：
1. AB测试如何保证结果的统计显著性？
2. 多个AB测试同时进行时如何隔离影响？
3. 如何设计购物车的渐进式发布（灰度发布）？

---

---

### 41.1.3 订单系统（15题）

##### 📊 题目1：订单状态机的设计

**问题描述**：
订单从创建到完成，经历多个状态（待支付、待发货、待收货、已完成）。如何设计订单状态机，保证状态流转的正确性？

**答案**：

**问题分析**：
订单状态流转的核心要素：
1. 状态定义清晰
2. 流转规则明确
3. 防止非法跳转
4. 支持异常流程（取消、退款）

**状态定义**：
```text
正向流程：
PENDING_PAYMENT（待支付）
→ PAID（已支付/待发货）
→ SHIPPED（已发货/待收货）
→ RECEIVED（已收货/待评价）
→ COMPLETED（已完成）

逆向流程：
CANCELLED（已取消）
REFUNDING（退款中）
REFUNDED（已退款）

特殊状态：
TIMEOUT（超时关闭）
```

**状态机实现**：

方案一：If-Else判断
```java
public void updateOrderStatus(Order order, OrderStatus newStatus) {
  OrderStatus currentStatus = order.getStatus();
  
  if (currentStatus == PENDING_PAYMENT) {
    if (newStatus == PAID || newStatus == CANCELLED || newStatus == TIMEOUT) {
      order.setStatus(newStatus);
    } else {
      throw new IllegalStateException("非法状态转换");
    }
  } else if (currentStatus == PAID) {
    if (newStatus == SHIPPED || newStatus == REFUNDING) {
      order.setStatus(newStatus);
    } else {
      throw new IllegalStateException("非法状态转换");
    }
  }
  // ... 更多判断
}
```

缺点：
- 代码冗长
- 难以维护
- 状态多时复杂度爆炸

方案二：状态转换表（推荐）
```java
// 定义状态转换规则
private static final Map<OrderStatus, Set<OrderStatus>> TRANSITIONS = Map.of(
  PENDING_PAYMENT, Set.of(PAID, CANCELLED, TIMEOUT),
  PAID, Set.of(SHIPPED, REFUNDING),
  SHIPPED, Set.of(RECEIVED, REFUNDING),
  RECEIVED, Set.of(COMPLETED, REFUNDING),
  REFUNDING, Set.of(REFUNDED)
);

public void updateOrderStatus(Order order, OrderStatus newStatus) {
  OrderStatus currentStatus = order.getStatus();
  
  Set<OrderStatus> allowedTransitions = TRANSITIONS.get(currentStatus);
  if (allowedTransitions == null || !allowedTransitions.contains(newStatus)) {
    throw new IllegalStateException(
      String.format("不允许从%s转换到%s", currentStatus, newStatus)
    );
  }
  
  // 记录状态变更历史
  OrderStatusHistory history = new OrderStatusHistory();
  history.setOrderId(order.getId());
  history.setFromStatus(currentStatus);
  history.setToStatus(newStatus);
  history.setOperator(getCurrentUser());
  history.setReason(reason);
  historyRepository.save(history);
  
  // 更新订单状态
  order.setStatus(newStatus);
  orderRepository.save(order);
  
  // 发布状态变更事件
  eventPublisher.publish(new OrderStatusChangedEvent(order, currentStatus, newStatus));
}
```

优点：
- 规则清晰
- 易于维护
- 可扩展

**状态流转图**：
```text
                    ┌─> CANCELLED
                    │
PENDING_PAYMENT ──┬─┴─> PAID ───> SHIPPED ───> RECEIVED ───> COMPLETED
                  │                  │            │
                  └─> TIMEOUT        │            │
                                     │            │
                                     └─> REFUNDING <─┘
                                            │
                                            └─> REFUNDED
```

**延伸思考**：
1. 如何设计订单的子状态（如待发货细分为待拣货、待打包、待出库）？
2. 订单状态变更如何触发后续操作（如发货后通知物流）？
3. 如何处理状态流转的并发冲突？

---

##### 🔧 题目2：订单号生成规则

**问题描述**：
订单号需要唯一、有序、不易被猜测。如何设计订单号生成规则？

**答案**：

**订单号设计要求**：
1. 全局唯一
2. 趋势递增（便于分库分表）
3. 信息可读（包含时间、业务类型）
4. 安全性（不易被遍历）
5. 长度适中（15-20位）

**方案一：数据库自增ID**

优点：
- 简单
- 唯一

缺点：
- 连续，易被猜测
- 分布式环境难实现
- 信息量少

**方案二：UUID**

优点：
- 全局唯一
- 无需中心化

缺点：
- 无序（影响索引性能）
- 长度太长（36位）
- 无业务含义

**方案三：Snowflake算法（推荐）**

结构：
```text
64位Long型：
1位符号位 + 41位时间戳 + 10位机器ID + 12位序列号

示例：
0 - 00000000000000000000000000000000000000000 - 0000000000 - 000000000000
│   └─────────────41位时间戳─────────────────┘   └10位机器┘   └12位序列┘
符号位

生成的订单号：1234567890123456789（19位）
```

实现：
```java
public class SnowflakeIdGenerator {
  // 起始时间戳（2020-01-01）
  private final long epoch = 1577836800000L;
  
  // 机器ID（数据中心ID + 机器ID）
  private final long workerId;
  
  // 序列号
  private long sequence = 0L;
  
  // 上次生成ID的时间戳
  private long lastTimestamp = -1L;
  
  public synchronized long nextId() {
    long timestamp = System.currentTimeMillis();
    
    // 时钟回拨检测
    if (timestamp < lastTimestamp) {
      throw new RuntimeException("时钟回拨");
    }
    
    // 同一毫秒内
    if (timestamp == lastTimestamp) {
      sequence = (sequence + 1) & 4095; // 4095=2^12-1
      if (sequence == 0) {
        // 序列号用完，等待下一毫秒
        timestamp = waitNextMillis(lastTimestamp);
      }
    } else {
      sequence = 0;
    }
    
    lastTimestamp = timestamp;
    
    // 组装ID
    return ((timestamp - epoch) << 22) 
         | (workerId << 12) 
         | sequence;
  }
}
```

优点：
- 趋势递增
- 高性能
- 分布式友好

缺点：
- 依赖机器时钟
- 机器ID需要管理

**方案四：业务规则拼接**

结构：
```text
订单号格式：业务前缀 + 日期 + 随机数

示例：
OR20260418123456789
│  └────┘└───────┘
│   日期   随机数
业务前缀（OR=Order）

生成：
String orderId = "OR" 
               + LocalDate.now().format(DateTimeFormatter.BASIC_ISO_DATE)
               + RandomStringUtils.randomNumeric(9);
```

优点：
- 可读性强
- 包含业务信息
- 可自定义

缺点：
- 需要保证随机数不重复
- 长度较长

**推荐方案**：
使用**Snowflake算法**生成基础ID，再转为业务订单号。

实现：
```java
public String generateOrderNo() {
  long snowflakeId = idGenerator.nextId();
  
  // 转为订单号（添加业务前缀）
  return "OR" + snowflakeId;
}
```

**延伸思考**：
1. 如何设计订单号的校验规则（防止伪造）？
2. 订单号如何支持多业务类型（普通订单、预售订单、拼团订单）？
3. 分库分表场景下订单号如何设计路由键？

---

##### 💡 题目3：订单超时自动取消

**问题描述**：
用户下单30分钟未支付，订单自动关闭并释放库存。如何实现订单超时自动取消？

**答案**：

**方案一：定时任务扫描**

核心思想：
定时任务定期扫描超时订单。

实现：
```java
@Scheduled(fixedDelay = 60000) // 每分钟执行
public void cancelTimeoutOrders() {
  // 查询超时未支付订单
  List<Order> timeoutOrders = orderRepository.findByStatusAndCreateTimeBefore(
    OrderStatus.PENDING_PAYMENT,
    LocalDateTime.now().minus(30, ChronoUnit.MINUTES)
  );
  
  for (Order order : timeoutOrders) {
    try {
      // 取消订单
      orderService.cancel(order.getId(), "超时未支付自动取消");
      
      // 释放库存
      inventoryService.release(order.getItems());
      
      // 通知用户
      notificationService.send(order.getUserId(), "订单已超时关闭");
    } catch (Exception e) {
      log.error("取消订单失败", e);
    }
  }
}
```

优点：
- 实现简单
- 可靠性高

缺点：
- 实时性差（最长延迟1分钟）
- 数据库扫描压力大
- 定时任务单点故障

**方案二：延迟队列（推荐）**

核心思想：
订单创建时发送延迟消息，30分钟后消费取消订单。

使用RabbitMQ延迟队列：
```java
// 创建订单时
public void createOrder(Order order) {
  // 1. 保存订单
  orderRepository.save(order);
  
  // 2. 发送延迟消息（30分钟后）
  rabbitTemplate.convertAndSend(
    "order.cancel.exchange",
    "order.cancel.routing.key",
    order.getId(),
    message -> {
      message.getMessageProperties().setDelay(30 * 60 * 1000); // 30分钟
      return message;
    }
  );
}

// 消费延迟消息
@RabbitListener(queues = "order.cancel.queue")
public void handleOrderCancel(Long orderId) {
  Order order = orderRepository.findById(orderId);
  
  // 检查订单状态
  if (order.getStatus() == OrderStatus.PENDING_PAYMENT) {
    // 仍未支付，取消订单
    orderService.cancel(orderId, "超时未支付自动取消");
    inventoryService.release(order.getItems());
  }
  // 如果已支付，忽略
}
```

使用Redis实现延迟队列：
```java
// 创建订单时
public void createOrder(Order order) {
  orderRepository.save(order);
  
  // 添加到Redis有序集合（Sorted Set）
  long expireTime = System.currentTimeMillis() + 30 * 60 * 1000;
  redis.zadd("order:timeout", expireTime, order.getId());
}

// 定时消费
@Scheduled(fixedDelay = 1000) // 每秒执行
public void processTimeoutOrders() {
  long now = System.currentTimeMillis();
  
  // 获取已到期的订单ID
  Set<String> orderIds = redis.zrangeByScore("order:timeout", 0, now);
  
  for (String orderId : orderIds) {
    try {
      // 处理超时订单
      processTimeoutOrder(Long.parseLong(orderId));
      
      // 从集合中移除
      redis.zrem("order:timeout", orderId);
    } catch (Exception e) {
      log.error("处理超时订单失败", e);
    }
  }
}
```

优点：
- 准确到秒
- 分布式友好
- 性能好

缺点：
- 依赖消息队列
- 需要处理消息丢失

**方案三：时间轮算法**

核心思想：
使用时间轮数据结构管理超时任务。

实现（Netty HashedWheelTimer）：
```java
private final HashedWheelTimer timer = new HashedWheelTimer(
  1, TimeUnit.SECONDS,  // 每秒tick一次
  60                     // 60个槽位
);

public void createOrder(Order order) {
  orderRepository.save(order);
  
  // 添加超时任务
  timer.newTimeout(timeout -> {
    Order latestOrder = orderRepository.findById(order.getId());
    if (latestOrder.getStatus() == OrderStatus.PENDING_PAYMENT) {
      orderService.cancel(order.getId(), "超时未支付自动取消");
    }
  }, 30, TimeUnit.MINUTES);
}
```

优点：
- 高性能
- 精确度高

缺点：
- 内存占用（任务在内存）
- 单机方案（不支持分布式）
- 服务重启任务丢失

**方案对比**：

| 方案 | 实时性 | 可靠性 | 分布式 | 实施难度 |
|------|--------|--------|--------|---------|
| 定时扫描 | ★★☆☆☆ | ★★★★★ | ★★★★☆ | ★★★★★ |
| 延迟队列 | ★★★★★ | ★★★★☆ | ★★★★★ | ★★★☆☆ |
| 时间轮 | ★★★★★ | ★★★☆☆ | ★★☆☆☆ | ★★★☆☆ |

**推荐方案**：
采用**延迟队列（RabbitMQ或Redis）**。

实施要点：

1. **幂等性保证**：
   ```java
   @Transactional
   public void cancel(Long orderId, String reason) {
     Order order = orderRepository.findById(orderId);
     
     // 检查当前状态
     if (order.getStatus() != OrderStatus.PENDING_PAYMENT) {
       log.warn("订单{}状态不是待支付，跳过取消", orderId);
       return; // 已被其他线程处理
     }
     
     // CAS更新状态
     int updated = orderRepository.updateStatus(
       orderId, 
       OrderStatus.CANCELLED,
       OrderStatus.PENDING_PAYMENT // 期望的旧状态
     );
     
     if (updated == 0) {
       log.warn("订单{}取消失败，可能已被处理", orderId);
       return;
     }
     
     // 释放库存
     inventoryService.release(order.getItems());
   }
   ```

2. **异常重试**：
   ```
   取消失败的处理：
   - 消息重新入队，稍后重试
   - 最多重试3次
   - 仍失败则记录告警，人工处理
   ```

3. **监控告警**：
   ```
   监控指标：
   - 超时订单数量
   - 取消成功率
   - 延迟队列堆积量
   
   告警：
   - 取消失败率 > 1%
   - 延迟队列堆积 > 10000
   ```

**延伸思考**：
1. 如何设计不同订单类型的不同超时时间（普通30分钟，秒杀10分钟）？
2. 订单超时取消如何通知用户？
3. 大促期间超时订单激增如何处理？

---

##### 📊 题目4：订单拆单与合单策略

**问题描述**：
用户购买多个商品，可能来自不同仓库或不同商家。如何设计订单拆单与合单策略？

**答案**：

**问题分析**：
拆单场景：
1. 多仓库发货（就近发货）
2. 多商家发货（平台+第三方卖家）
3. 预售+现货（发货时间不同）
4. 自营+跨境（清关时间不同）

合单场景：
1. 同一地址多笔订单（节省运费）
2. 同一商家商品（方便发货）

**方案一：用户下单时拆单**

核心思想：
用户提交订单时，系统自动拆分为多个子订单。

流程：
```text
用户购物车：
- 商品A（北京仓）
- 商品B（上海仓）
- 商品C（北京仓）

拆单规则：
按仓库拆分：
→ 子订单1：商品A + C（北京仓）
→ 子订单2：商品B（上海仓）

数据结构：
parent_order（父订单）
├── parent_order_id
├── user_id
├── total_amount
└── status

sub_order（子订单）
├── sub_order_id
├── parent_order_id
├── warehouse_id
├── items
└── status
```

用户支付：
```text
用户支付父订单 → 分配金额到各子订单
子订单独立发货、收货
```

优点：
- 逻辑清晰
- 用户感知明确

缺点：
- 用户体验复杂（多个运单号）
- 退款复杂（部分退款）

**方案二：后台自动拆单（推荐）**

核心思想：
用户下单时是一个订单，后台根据规则自动拆分为多个发货单。

流程：
```text
用户下单：创建订单（单个）
↓
订单支付成功
↓
订单中心分析：需要拆单
↓
创建多个发货单（shipment）
- 发货单1：商品A+C → 北京仓
- 发货单2：商品B → 上海仓
↓
各仓库独立发货
```

数据结构：
```sql
order（订单）
├── order_id
├── user_id
├── total_amount
└── status

shipment（发货单）
├── shipment_id
├── order_id
├── warehouse_id
├── items（发货商品）
├── tracking_number（运单号）
└── status
```

优点：
- 用户无感知（看到的是一个订单）
- 退款简单（按订单退）
- 灵活（可随时调整拆单规则）

缺点：
- 实现复杂
- 需要维护订单和发货单的关系

**拆单规则**：

1. **按仓库拆分**：
   ```java
   public List<Shipment> splitByWarehouse(Order order) {
     // 1. 为每个商品选择最优仓库
     Map<String, Warehouse> itemWarehouse = new HashMap<>();
     for (OrderItem item : order.getItems()) {
       Warehouse warehouse = selectWarehouse(item.getSkuId(), order.getAddress());
       itemWarehouse.put(item.getSkuId(), warehouse);
     }
     
     // 2. 按仓库分组
     Map<Warehouse, List<OrderItem>> grouped = order.getItems().stream()
       .collect(Collectors.groupBy(item -> itemWarehouse.get(item.getSkuId())));
     
     // 3. 生成发货单
     List<Shipment> shipments = new ArrayList<>();
     for (Map.Entry<Warehouse, List<OrderItem>> entry : grouped.entrySet()) {
       Shipment shipment = new Shipment();
       shipment.setOrderId(order.getId());
       shipment.setWarehouseId(entry.getKey().getId());
       shipment.setItems(entry.getValue());
       shipments.add(shipment);
     }
     
     return shipments;
   }
   ```

2. **按商家拆分**：
   ```
   平台订单包含：
   - 自营商品（平台发货）
   - 第三方商品（商家发货）
   
   拆分：
   - 子订单1：自营商品
   - 子订单2：商家A的商品
   - 子订单3：商家B的商品
   ```

3. **按发货时间拆分**：
   ```
   订单包含：
   - 现货商品（立即发货）
   - 预售商品（15天后发货）
   
   拆分：
   - 发货单1：现货（立即发）
   - 发货单2：预售（延迟发）
   ```

**合单策略**：

1. **同地址合并**：
   ```
   用户A在1小时内下了3笔订单：
   - 订单1：商品A（北京仓）
   - 订单2：商品B（北京仓）
   - 订单3：商品C（上海仓）
   
   合单：
   - 发货单1：订单1+订单2的商品（北京仓合并发货）
   - 发货单2：订单3的商品（上海仓单独发货）
   
   好处：
   - 节省运费
   - 减少包裹数量
   ```

2. **运费优化**：
   ```
   规则：
   - 同一仓库、同一地址、24小时内的订单
   - 自动合并发货
   - 运费退还到用户余额
   ```

**推荐方案**：
采用**后台自动拆单**。

实施要点：

1. **拆单时机**：
   ```
   时机选择：
   - 订单支付后立即拆单（推荐）
   - 发货前拆单（更灵活）
   ```

2. **用户展示**：
   ```
   订单详情页：
   订单号：OR123456
   总金额：¥1000
   
   发货信息：
   - 包裹1：商品A+B（运单号：SF123）
     状态：已发货
   - 包裹2：商品C（运单号：SF456）
     状态：待发货
   ```

3. **退款处理**：
   ```
   部分商品退款：
   - 用户申请退商品A
   - 计算退款金额（商品价 + 分摊运费）
   - 只退部分金额
   - 其他商品正常履约
   ```

**延伸思考**：
1. 如何设计拆单的运费分摊规则？
2. 拆单后如何保证库存一致性？
3. 跨境订单的拆单有何特殊性？

---

##### 🔧 题目5：订单的并发创建与幂等性

**问题描述**：
用户可能重复点击"提交订单"按钮，导致创建多个订单。如何保证订单创建的幂等性？

**答案**：

**问题分析**：
重复下单的原因：
1. 用户重复点击
2. 网络超时重试
3. 前端未防抖
4. 恶意刷单

**方案一：前端防抖**

核心思想：
前端限制用户短时间内多次点击。

实现：
```javascript
let submitting = false;

function submitOrder() {
  if (submitting) {
    return; // 正在提交中，忽略
  }
  
  submitting = true;
  
  fetch('/api/order/create', {
    method: 'POST',
    body: JSON.stringify(orderData)
  })
  .then(res => {
    // 处理结果
  })
  .finally(() => {
    submitting = false; // 完成后恢复
  });
}
```

优点：
- 简单有效

缺点：
- 仅防止前端重复
- 无法防止恶意绕过前端

**方案二：唯一索引（推荐）**

核心思想：
数据库层面保证唯一性。

实现：
```sql
CREATE TABLE orders (
  order_id BIGINT PRIMARY KEY,
  user_id BIGINT NOT NULL,
  idempotent_key VARCHAR(64) UNIQUE, -- 幂等键
  ...
);

CREATE UNIQUE INDEX uk_user_idempotent ON orders(user_id, idempotent_key);
```

创建订单：
```java
@Transactional
public Order createOrder(OrderRequest request, String idempotentKey) {
  try {
    // 1. 构建订单
    Order order = new Order();
    order.setUserId(request.getUserId());
    order.setIdempotentKey(idempotentKey);
    order.setItems(request.getItems());
    // ...
    
    // 2. 保存订单（唯一索引保证幂等）
    orderRepository.save(order);
    
    // 3. 扣减库存
    inventoryService.deduct(order.getItems());
    
    return order;
  } catch (DuplicateKeyException e) {
    // 幂等键重复，说明订单已创建
    return orderRepository.findByIdempotentKey(idempotentKey);
  }
}
```

幂等键生成：
```java
// 方案1：前端生成UUID
String idempotentKey = UUID.randomUUID().toString();

// 方案2：后端生成（基于购物车内容）
String idempotentKey = DigestUtils.md5Hex(
  userId + ":" + cartItems.toString() + ":" + timestamp
);
```

优点：
- 数据库层面保证
- 可靠性高

缺点：
- 依赖唯一索引
- 需要生成幂等键

**方案三：分布式锁**

核心思想：
使用Redis分布式锁，同一用户同时只能创建一个订单。

实现：
```java
public Order createOrder(OrderRequest request) {
  String lockKey = "order:create:" + request.getUserId();
  
  // 尝试获取锁
  boolean locked = redisLock.tryLock(lockKey, 10, TimeUnit.SECONDS);
  if (!locked) {
    throw new BizException("正在创建订单，请勿重复提交");
  }
  
  try {
    // 创建订单
    Order order = doCreateOrder(request);
    return order;
  } finally {
    // 释放锁
    redisLock.unlock(lockKey);
  }
}
```

优点：
- 防止并发创建
- 灵活控制

缺点：
- 依赖Redis
- 锁超时需要处理

**方案四：Token机制**

核心思想：
用户进入结算页时，服务端生成唯一Token，提交订单时校验Token。

流程：
```text
1. 用户进入结算页
   → 请求服务端生成Token
   → 服务端生成Token并存Redis
   → 返回Token给前端

2. 用户提交订单
   → 携带Token
   → 服务端校验Token是否存在
   → 存在则删除Token，创建订单
   → 不存在则拒绝（重复提交）
```

实现：
```java
// 生成Token
public String generateOrderToken(Long userId) {
  String token = UUID.randomUUID().toString();
  String key = "order:token:" + token;
  redis.setex(key, 300, userId.toString()); // 5分钟有效
  return token;
}

// 创建订单（校验Token）
@Transactional
public Order createOrder(OrderRequest request, String token) {
  String key = "order:token:" + token;
  
  // 检查Token是否存在
  String userId = redis.get(key);
  if (userId == null) {
    throw new BizException("订单Token无效或已使用");
  }
  
  // 验证Token归属
  if (!userId.equals(request.getUserId().toString())) {
    throw new BizException("订单Token不匹配");
  }
  
  // 删除Token（保证一次性）
  redis.del(key);
  
  // 创建订单
  return doCreateOrder(request);
}
```

优点：
- 防止重复提交
- 安全性高（Token一次性）

缺点：
- 需要多次交互
- Token过期需要重新获取

**方案对比**：

| 方案 | 可靠性 | 易用性 | 性能 | 适用场景 |
|------|--------|--------|------|----------|
| 前端防抖 | ★★☆☆☆ | ★★★★★ | ★★★★★ | 辅助手段 |
| 唯一索引 | ★★★★★ | ★★★★☆ | ★★★★☆ | 通用 |
| 分布式锁 | ★★★★☆ | ★★★☆☆ | ★★★☆☆ | 高并发 |
| Token机制 | ★★★★★ | ★★★☆☆ | ★★★★☆ | 安全性要求高 |

**推荐方案**：
采用**唯一索引+Token机制**的组合。

实施要点：

1. **多层防护**：
   ```
   L1：前端防抖（用户体验）
   L2：Token机制（防恶意）
   L3：唯一索引（最后防线）
   ```

2. **幂等键设计**：
   ```
   幂等键组成：
   userId + cartVersion + timestamp
   
   例如：
   123_v10_1679800000
   
   说明：
   - userId：用户ID
   - cartVersion：购物车版本（购物车内容变化版本号+1）
   - timestamp：提交时间戳（精确到秒）
   ```

3. **异常处理**：
   ```java
   try {
     return createOrder(request, token);
   } catch (DuplicateKeyException e) {
     // 唯一索引冲突，查询已存在的订单
     Order existingOrder = findByIdempotentKey(idempotentKey);
     return existingOrder;
   } catch (BizException e) {
     // Token无效等业务异常
     throw e;
   }
   ```

**延伸思考**：
1. 如何设计订单创建的限流（防止刷单）？
2. 订单创建失败如何回滚库存？
3. 分布式事务下如何保证订单创建的一致性？

---

##### 📊 题目6：订单的分布式事务设计（Saga模式）

**问题描述**：
订单创建涉及多个服务（订单服务、库存服务、优惠券服务、积分服务）。如何使用Saga模式保证分布式事务一致性？

**答案**：

**问题分析**：
订单创建的分布式事务流程：
1. 扣减库存（库存服务）
2. 核销优惠券（营销服务）
3. 扣减积分（会员服务）
4. 创建订单（订单服务）

任一环节失败，已执行的操作需要回滚。

**Saga模式实现**（使用Go）：

```go
package saga

import (
	"context"
	"fmt"
)

// SagaStep 定义Saga步骤
type SagaStep struct {
	Name         string
	Execute      func(ctx context.Context, data interface{}) error
	Compensate   func(ctx context.Context, data interface{}) error
}

// SagaOrchestrator Saga编排器
type SagaOrchestrator struct {
	steps []SagaStep
}

// Execute 执行Saga
func (s *SagaOrchestrator) Execute(ctx context.Context, data interface{}) error {
	executedSteps := make([]int, 0)
	
	// 正向执行
	for i, step := range s.steps {
		if err := step.Execute(ctx, data); err != nil {
			// 执行失败，触发补偿
			s.compensate(ctx, data, executedSteps)
			return fmt.Errorf("步骤 %s 执行失败: %w", step.Name, err)
		}
		executedSteps = append(executedSteps, i)
	}
	
	return nil
}

// compensate 执行补偿
func (s *SagaOrchestrator) compensate(ctx context.Context, data interface{}, executedSteps []int) {
	// 反向补偿
	for i := len(executedSteps) - 1; i >= 0; i-- {
		stepIndex := executedSteps[i]
		step := s.steps[stepIndex]
		
		if err := step.Compensate(ctx, data); err != nil {
			// 补偿失败，记录日志，转人工处理
			log.Errorf("步骤 %s 补偿失败: %v", step.Name, err)
		}
	}
}

// 订单创建Saga示例
func CreateOrderSaga(orderReq *CreateOrderRequest) error {
	saga := &SagaOrchestrator{
		steps: []SagaStep{
			// 步骤1：扣减库存
			{
				Name: "DeductInventory",
				Execute: func(ctx context.Context, data interface{}) error {
					req := data.(*CreateOrderRequest)
					return inventoryService.Deduct(ctx, req.Items)
				},
				Compensate: func(ctx context.Context, data interface{}) error {
					req := data.(*CreateOrderRequest)
					return inventoryService.Release(ctx, req.Items)
				},
			},
			// 步骤2：核销优惠券
			{
				Name: "UseCoupon",
				Execute: func(ctx context.Context, data interface{}) error {
					req := data.(*CreateOrderRequest)
					if req.CouponID == "" {
						return nil // 无优惠券，跳过
					}
					return couponService.Use(ctx, req.UserID, req.CouponID)
				},
				Compensate: func(ctx context.Context, data interface{}) error {
					req := data.(*CreateOrderRequest)
					if req.CouponID == "" {
						return nil
					}
					return couponService.Release(ctx, req.UserID, req.CouponID)
				},
			},
			// 步骤3：扣减积分
			{
				Name: "DeductPoints",
				Execute: func(ctx context.Context, data interface{}) error {
					req := data.(*CreateOrderRequest)
					if req.PointsToUse == 0 {
						return nil
					}
					return pointsService.Deduct(ctx, req.UserID, req.PointsToUse)
				},
				Compensate: func(ctx context.Context, data interface{}) error {
					req := data.(*CreateOrderRequest)
					if req.PointsToUse == 0 {
						return nil
					}
					return pointsService.Refund(ctx, req.UserID, req.PointsToUse)
				},
			},
			// 步骤4：创建订单
			{
				Name: "CreateOrder",
				Execute: func(ctx context.Context, data interface{}) error {
					req := data.(*CreateOrderRequest)
					order := &Order{
						OrderID:   generateOrderID(),
						UserID:    req.UserID,
						Items:     req.Items,
						Status:    OrderStatusPending,
					}
					return orderRepo.Create(ctx, order)
				},
				Compensate: func(ctx context.Context, data interface{}) error {
					req := data.(*CreateOrderRequest)
					// 订单创建失败不需要补偿（未持久化）
					return nil
				},
			},
		},
	}
	
	return saga.Execute(context.Background(), orderReq)
}
```

**优点**：
- 逻辑清晰（正向+补偿）
- 解耦各服务
- 支持长事务

**缺点**：
- 实现复杂
- 补偿可能失败（需要人工介入）
- 中间状态可见（不是强一致性）

**延伸思考**：
1. Saga补偿失败如何处理？
2. 如何设计Saga的可视化监控？
3. Saga vs 2PC（两阶段提交）如何选择？

---

##### 🔧 题目7：订单数据的分库分表设计

**问题描述**：
订单表数据量达到亿级，单表查询性能下降。如何设计订单的分库分表方案？

**答案**：

**问题分析**：
订单分库分表的核心要素：
1. 分片键选择（user_id还是order_id）
2. 分片数量（16、32、64、128）
3. 跨片查询（如运营查询某时间段订单）
4. 数据扩容

**方案一：按user_id分片（推荐）**

核心思想：
同一用户的订单存储在同一分片。

分片规则：
```go
// 分片数量
const ShardCount = 64

// 计算分片
func GetShardIndex(userID int64) int {
	return int(userID % ShardCount)
}

// 路由到数据源
func GetDataSource(userID int64) *sql.DB {
	shardIndex := GetShardIndex(userID)
	return dataSources[shardIndex]
}
```

表结构：
```sql
-- 64个库，每个库有orders表
database_00.orders
database_01.orders
...
database_63.orders

订单ID生成：
order_id = snowflake_id
不包含分片信息（通过user_id路由）
```

优点：
- 用户维度查询高效（"我的订单"）
- 单用户订单聚合容易
- 避免跨库JOIN

缺点：
- 按订单ID查询需要广播（查所有分片）
- 数据可能不均匀（大客户订单多）

**方案二：按order_id分片**

核心思想：
按订单ID散列分片。

分片规则：
```go
func GetShardIndex(orderID int64) int {
	return int(orderID % ShardCount)
}
```

订单ID生成（包含分片信息）：
```go
// 订单ID结构：分片位 + Snowflake ID
// 前6位：分片号（0-63）
// 后13位：Snowflake ID

func GenerateOrderID(userID int64) int64 {
	shardIndex := GetShardIndex(userID)
	snowflakeID := snowflake.Generate()
	
	// 组装：分片号（6位） + snowflake（13位）
	return int64(shardIndex)*1e13 + snowflakeID
}

// 解析分片
func ParseShard(orderID int64) int {
	return int(orderID / 1e13)
}
```

优点：
- 按订单ID查询高效（直接定位分片）
- 数据均匀

缺点：
- 用户维度查询需要广播
- "我的订单"查询慢

**方案三：复合分片**

核心思想：
主表按user_id分片，建立order_id到分片的映射表。

设计：
```text
主表（按user_id分片）：
shard_00.orders
shard_01.orders

映射表（不分片，单独集群）：
order_routing
├── order_id（主键）
├── shard_index（分片号）
└── user_id

查询流程：
1. 按订单ID查询：
   - 查询order_routing获取分片号
   - 路由到对应分片查询

2. 按用户ID查询：
   - 直接路由到用户分片
```

优点：
- 支持多种查询方式
- 灵活

缺点：
- 映射表是单点
- 实现复杂

**方案对比**：

| 方案 | 用户查询 | 订单查询 | 数据均匀度 | 实施难度 |
|------|---------|---------|-----------|---------|
| 按user_id | ★★★★★ | ★★☆☆☆ | ★★★☆☆ | ★★★★☆ |
| 按order_id | ★★☆☆☆ | ★★★★★ | ★★★★★ | ★★★★☆ |
| 复合分片 | ★★★★★ | ★★★★★ | ★★★★★ | ★★☆☆☆ |

**推荐方案**：
采用**按user_id分片**。

实施要点（Go实现）：

1. **分片路由中间件**：
   ```go
   package sharding
   
   import (
   	"context"
   	"database/sql"
   )
   
   // ShardingManager 分片管理器
   type ShardingManager struct {
   	dataSources []*sql.DB
   	shardCount  int
   }
   
   // NewShardingManager 创建分片管理器
   func NewShardingManager(dsns []string) (*ShardingManager, error) {
   	dbs := make([]*sql.DB, len(dsns))
   	for i, dsn := range dsns {
   		db, err := sql.Open("mysql", dsn)
   		if err != nil {
   			return nil, err
   		}
   		dbs[i] = db
   	}
   	
   	return &ShardingManager{
   		dataSources: dbs,
   		shardCount:  len(dsns),
   	}, nil
   }
   
   // GetDB 根据用户ID获取数据库连接
   func (sm *ShardingManager) GetDB(userID int64) *sql.DB {
   	shardIndex := userID % int64(sm.shardCount)
   	return sm.dataSources[shardIndex]
   }
   
   // ExecuteOnShard 在指定分片执行查询
   func (sm *ShardingManager) ExecuteOnShard(ctx context.Context, userID int64, 
   	fn func(*sql.DB) error) error {
   	db := sm.GetDB(userID)
   	return fn(db)
   }
   
   // Broadcast 广播到所有分片执行
   func (sm *ShardingManager) Broadcast(ctx context.Context, 
   	fn func(*sql.DB) error) []error {
   	errors := make([]error, 0)
   	for _, db := range sm.dataSources {
   		if err := fn(db); err != nil {
   			errors = append(errors, err)
   		}
   	}
   	return errors
   }
   ```

2. **订单Repository实现**：
   ```go
   type OrderRepository struct {
   	shardingMgr *ShardingManager
   }
   
   // Create 创建订单
   func (r *OrderRepository) Create(ctx context.Context, order *Order) error {
   	return r.shardingMgr.ExecuteOnShard(ctx, order.UserID, func(db *sql.DB) error {
   		query := `INSERT INTO orders (order_id, user_id, total_amount, status, created_at)
   		          VALUES (?, ?, ?, ?, ?)`
   		_, err := db.ExecContext(ctx, query, 
   			order.OrderID, order.UserID, order.TotalAmount, 
   			order.Status, time.Now())
   		return err
   	})
   }
   
   // FindByUserID 查询用户订单（单分片）
   func (r *OrderRepository) FindByUserID(ctx context.Context, userID int64, 
   	page, size int) ([]*Order, error) {
   	var orders []*Order
   	
   	err := r.shardingMgr.ExecuteOnShard(ctx, userID, func(db *sql.DB) error {
   		query := `SELECT * FROM orders 
   		          WHERE user_id=? 
   		          ORDER BY created_at DESC 
   		          LIMIT ? OFFSET ?`
   		rows, err := db.QueryContext(ctx, query, userID, size, (page-1)*size)
   		if err != nil {
   			return err
   		}
   		defer rows.Close()
   		
   		for rows.Next() {
   			order := &Order{}
   			// 扫描数据...
   			orders = append(orders, order)
   		}
   		return nil
   	})
   	
   	return orders, err
   }
   
   // FindByOrderID 按订单ID查询（需要广播）
   func (r *OrderRepository) FindByOrderID(ctx context.Context, orderID int64) (*Order, error) {
   	// 方案1：广播到所有分片查询（慢）
   	for _, db := range r.shardingMgr.dataSources {
   		order, err := queryFromDB(db, orderID)
   		if err == nil && order != nil {
   			return order, nil
   		}
   	}
   	return nil, ErrOrderNotFound
   	
   	// 方案2：维护order_id -> user_id映射（推荐）
   	// userID := r.getOrderUserMapping(orderID)
   	// return r.FindByUserAndOrderID(ctx, userID, orderID)
   }
   ```

3. **订单ID包含分片信息**：
   ```go
   // 订单ID结构：6位分片号 + 13位Snowflake
   
   func GenerateOrderIDWithShard(userID int64) int64 {
   	shardIndex := userID % ShardCount
   	snowflakeID := snowflake.NextID()
   	
   	// 组装：前6位是分片号
   	return shardIndex*1e13 + snowflakeID
   }
   
   // 解析分片号
   func ParseShardFromOrderID(orderID int64) int {
   	return int(orderID / 1e13)
   }
   
   // 直接定位查询
   func (r *OrderRepository) FindByOrderIDFast(ctx context.Context, orderID int64) (*Order, error) {
   	shardIndex := ParseShardFromOrderID(orderID)
   	db := r.shardingMgr.dataSources[shardIndex]
   	
   	query := `SELECT * FROM orders WHERE order_id=?`
   	row := db.QueryRowContext(ctx, query, orderID)
   	
   	order := &Order{}
   	err := row.Scan(&order.OrderID, &order.UserID, ...) 
   	return order, err
   }
   ```

4. **扩容方案**：
   ```
   扩容策略（64 → 128分片）：
   
   方案A：双写期
   1. 新建64个分片（总共128个）
   2. 新订单写入新分片规则
   3. 老订单保留在老分片
   4. 查询时先查新分片，未命中再查老分片
   
   方案B：一致性哈希
   1. 使用一致性哈希算法
   2. 扩容时只需迁移部分数据
   3. 数据迁移期间双写
   ```

**延伸思考**：
1. 如何设计分库分表的全局查询（如运营后台）？
2. 订单归档如何设计（冷热数据分离）？
3. 分库分表如何支持跨库JOIN？

---

##### 💡 题目8：订单履约流程的编排

**问题描述**：
订单支付成功后，需要依次执行：分配仓库、创建拣货单、打包、出库、创建运单、发货。如何设计订单履约流程的编排？

**答案**：

**推荐方案**：事件驱动+状态机

架构（Go实现）：
```go
package fulfillment

import (
	"context"
)

// FulfillmentEvent 履约事件
type FulfillmentEvent struct {
	OrderID   int64
	EventType string
	Data      map[string]interface{}
}

// FulfillmentOrchestrator 履约编排器
type FulfillmentOrchestrator struct {
	eventBus EventBus
}

// OnOrderPaid 订单支付事件处理
func (o *FulfillmentOrchestrator) OnOrderPaid(ctx context.Context, orderID int64) error {
	// 1. 分配仓库
	warehouse, err := o.allocateWarehouse(ctx, orderID)
	if err != nil {
		return err
	}
	
	// 2. 创建拣货单
	pickingOrder, err := o.createPickingOrder(ctx, orderID, warehouse.ID)
	if err != nil {
		return err
	}
	
	// 3. 发布拣货事件
	o.eventBus.Publish(&FulfillmentEvent{
		OrderID:   orderID,
		EventType: "PickingOrderCreated",
		Data: map[string]interface{}{
			"pickingOrderID": pickingOrder.ID,
			"warehouseID":    warehouse.ID,
		},
	})
	
	return nil
}

// OnPickingCompleted 拣货完成事件处理
func (o *FulfillmentOrchestrator) OnPickingCompleted(ctx context.Context, event *FulfillmentEvent) error {
	orderID := event.OrderID
	
	// 1. 打包
	if err := o.pack(ctx, orderID); err != nil {
		return err
	}
	
	// 2. 出库
	if err := o.outbound(ctx, orderID); err != nil {
		return err
	}
	
	// 3. 创建物流运单
	trackingNumber, err := o.createShipment(ctx, orderID)
	if err != nil {
		return err
	}
	
	// 4. 发布发货事件
	o.eventBus.Publish(&FulfillmentEvent{
		OrderID:   orderID,
		EventType: "OrderShipped",
		Data: map[string]interface{}{
			"trackingNumber": trackingNumber,
		},
	})
	
	return nil
}

// 事件监听器
func (o *FulfillmentOrchestrator) Start() {
	o.eventBus.Subscribe("OrderPaid", o.OnOrderPaid)
	o.eventBus.Subscribe("PickingCompleted", o.OnPickingCompleted)
	o.eventBus.Subscribe("PackingCompleted", o.OnPackingCompleted)
	// ...
}
```

**履约状态机**：
```go
type FulfillmentStatus int

const (
	FulfillmentPending      FulfillmentStatus = 0  // 待履约
	FulfillmentWarehouseAllocated FulfillmentStatus = 1  // 已分配仓库
	FulfillmentPicking      FulfillmentStatus = 2  // 拣货中
	FulfillmentPacked       FulfillmentStatus = 3  // 已打包
	FulfillmentOutbound     FulfillmentStatus = 4  // 已出库
	FulfillmentShipped      FulfillmentStatus = 5  // 已发货
	FulfillmentReceived     FulfillmentStatus = 6  // 已签收
)

// 状态流转规则
var fulfillmentTransitions = map[FulfillmentStatus][]FulfillmentStatus{
	FulfillmentPending:            {FulfillmentWarehouseAllocated},
	FulfillmentWarehouseAllocated: {FulfillmentPicking},
	FulfillmentPicking:            {FulfillmentPacked},
	FulfillmentPacked:             {FulfillmentOutbound},
	FulfillmentOutbound:           {FulfillmentShipped},
	FulfillmentShipped:            {FulfillmentReceived},
}

// UpdateStatus 更新履约状态
func (o *FulfillmentOrchestrator) UpdateStatus(ctx context.Context, 
	orderID int64, newStatus FulfillmentStatus) error {
	// 1. 查询当前状态
	currentStatus, err := o.getStatus(ctx, orderID)
	if err != nil {
		return err
	}
	
	// 2. 检查状态流转是否合法
	allowedTransitions := fulfillmentTransitions[currentStatus]
	if !contains(allowedTransitions, newStatus) {
		return fmt.Errorf("不允许从%v转换到%v", currentStatus, newStatus)
	}
	
	// 3. 更新状态
	return o.updateStatusInDB(ctx, orderID, newStatus)
}
```

**延伸思考**：
1. 履约流程如何支持异常处理（缺货、商品损坏）？
2. 多个发货单如何协调履约进度？
3. 履约时效如何监控和告警？

---

##### 📊 题目9：订单的退款和售后流程设计

**问题描述**：
用户申请退款（仅退款、退货退款），如何设计售后流程，保证资金安全和用户体验？

**答案**：

**退款场景**：
1. 仅退款（未发货）
2. 退货退款（已发货）
3. 部分退款（退部分商品）
4. 售后退款（商品质量问题）

**推荐方案**（Go实现）：

退款状态机：
```go
type RefundStatus int

const (
	RefundPending   RefundStatus = 0  // 待审核
	RefundApproved  RefundStatus = 1  // 已同意
	RefundRejected  RefundStatus = 2  // 已拒绝
	RefundReturning RefundStatus = 3  // 退货中
	RefundReturned  RefundStatus = 4  // 已退货
	RefundCompleted RefundStatus = 5  // 已退款
)

// Refund 退款单
type Refund struct {
	RefundID     int64
	OrderID      int64
	UserID       int64
	RefundType   string  // REFUND_ONLY, RETURN_REFUND
	RefundAmount decimal.Decimal
	Reason       string
	Status       RefundStatus
	CreatedAt    time.Time
}

// RefundService 退款服务
type RefundService struct {
	orderRepo   OrderRepository
	paymentSvc  PaymentService
	inventorySvc InventoryService
}

// CreateRefund 创建退款申请
func (s *RefundService) CreateRefund(ctx context.Context, req *RefundRequest) (*Refund, error) {
	// 1. 校验订单状态
	order, err := s.orderRepo.FindByID(ctx, req.OrderID)
	if err != nil {
		return nil, err
	}
	
	if order.Status != OrderStatusPaid && order.Status != OrderStatusShipped {
		return nil, errors.New("订单状态不允许退款")
	}
	
	// 2. 校验退款金额
	if req.RefundAmount.GreaterThan(order.PaidAmount) {
		return nil, errors.New("退款金额超过实付金额")
	}
	
	// 3. 创建退款单
	refund := &Refund{
		RefundID:     generateRefundID(),
		OrderID:      req.OrderID,
		UserID:       req.UserID,
		RefundType:   req.RefundType,
		RefundAmount: req.RefundAmount,
		Reason:       req.Reason,
		Status:       RefundPending,
		CreatedAt:    time.Now(),
	}
	
	if err := s.refundRepo.Create(ctx, refund); err != nil {
		return nil, err
	}
	
	// 4. 自动审核（部分场景）
	if s.shouldAutoApprove(refund) {
		return s.Approve(ctx, refund.RefundID)
	}
	
	return refund, nil
}

// Approve 审核通过退款
func (s *RefundService) Approve(ctx context.Context, refundID int64) (*Refund, error) {
	refund, err := s.refundRepo.FindByID(ctx, refundID)
	if err != nil {
		return nil, err
	}
	
	// 1. 更新退款状态
	refund.Status = RefundApproved
	if err := s.refundRepo.Update(ctx, refund); err != nil {
		return nil, err
	}
	
	// 2. 根据退款类型处理
	if refund.RefundType == "REFUND_ONLY" {
		// 仅退款：直接退款
		return s.processRefund(ctx, refund)
	} else {
		// 退货退款：等待用户退货
		refund.Status = RefundReturning
		s.refundRepo.Update(ctx, refund)
		// 生成退货地址和快递单号
		s.generateReturnLabel(ctx, refund)
		return refund, nil
	}
}

// processRefund 执行退款
func (s *RefundService) processRefund(ctx context.Context, refund *Refund) (*Refund, error) {
	// 1. 调用支付服务退款
	if err := s.paymentSvc.Refund(ctx, refund.OrderID, refund.RefundAmount); err != nil {
		return nil, fmt.Errorf("退款失败: %w", err)
	}
	
	// 2. 回补库存
	order, _ := s.orderRepo.FindByID(ctx, refund.OrderID)
	if err := s.inventorySvc.Return(ctx, order.Items); err != nil {
		log.Errorf("回补库存失败: %v", err)
		// 不阻塞退款流程，记录异常任务
		s.createCompensationTask(ctx, "ReturnInventory", refund.RefundID)
	}
	
	// 3. 更新退款状态
	refund.Status = RefundCompleted
	if err := s.refundRepo.Update(ctx, refund); err != nil {
		return nil, err
	}
	
	// 4. 更新订单状态
	s.orderRepo.UpdateStatus(ctx, refund.OrderID, OrderStatusRefunded)
	
	// 5. 发送通知
	s.notifySvc.Send(ctx, refund.UserID, "退款已到账")
	
	return refund, nil
}
```

**自动审核规则**：
```go
func (s *RefundService) shouldAutoApprove(refund *Refund) bool {
	// 自动同意条件：
	// 1. 订单未发货
	// 2. 退款金额 < 500元
	// 3. 用户信用良好
	
	order, _ := s.orderRepo.FindByID(context.Background(), refund.OrderID)
	
	if order.Status == OrderStatusPaid &&
		refund.RefundAmount.LessThan(decimal.NewFromInt(500)) &&
		s.userSvc.IsTrusted(refund.UserID) {
		return true
	}
	
	return false
}
```

**延伸思考**：
1. 退款失败如何重试和补偿？
2. 恶意退款如何识别和防范？
3. 部分退款如何计算退款金额（商品价+运费分摊）？

---

##### 🔧 题目10：订单的异常处理（缺货、地址错误）

**问题描述**：
订单履约过程中可能出现异常（缺货、地址无法送达、商品损坏）。如何设计异常处理流程？

**答案**：

**异常场景及处理方案**：

1. **库存不足（超卖）**：
   ```go
   // 发现超卖
   func (s *FulfillmentService) HandleOutOfStock(ctx context.Context, orderID int64) error {
   	// 1. 联系用户
   	s.notifySvc.Send(ctx, order.UserID, "商品暂时缺货，为您申请退款")
   	
   	// 2. 创建退款
   	refund := &Refund{
   		OrderID:      orderID,
   		RefundType:   "OUT_OF_STOCK",
   		RefundAmount: order.PaidAmount,
   		AutoApprove:  true,
   	}
   	return s.refundSvc.CreateRefund(ctx, refund)
   }
   ```

2. **地址无法送达**：
   ```go
   func (s *FulfillmentService) HandleUndeliverableAddress(ctx context.Context, 
   	orderID int64) error {
   	// 1. 通知用户修改地址
   	s.notifySvc.Send(ctx, order.UserID, "收货地址无法送达，请修改地址")
   	
   	// 2. 订单挂起
   	s.orderRepo.UpdateStatus(ctx, orderID, OrderStatusAddressError)
   	
   	// 3. 用户修改地址后重新履约
   	// 或超时自动退款
   	s.scheduleAutoRefund(ctx, orderID, 48*time.Hour)
   	
   	return nil
   }
   ```

3. **商品损坏**：
   ```go
   func (s *FulfillmentService) HandleDamaged(ctx context.Context, 
   	orderID int64, itemID string) error {
   	// 1. 记录损坏
   	s.logDamage(ctx, orderID, itemID)
   	
   	// 2. 检查是否有替代品
   	if hasReplace, err := s.inventorySvc.CheckStock(ctx, itemID); err == nil && hasReplace {
   		// 有替代品，重新拣货
   		return s.repick(ctx, orderID, itemID)
   	}
   	
   	// 3. 无替代品，部分退款
   	item := s.getOrderItem(ctx, orderID, itemID)
   	return s.refundSvc.CreatePartialRefund(ctx, orderID, item.Amount)
   }
   ```

**延伸思考**：
1. 异常订单如何统计和分析？
2. 如何设计异常的自动化处理规则？

---

##### 💡 题目11：订单的搜索和查询优化

**问题描述**：
用户需要查询历史订单（按时间、状态、商品筛选），运营需要查询全部订单。如何设计订单查询系统？

**答案**：

**方案一：主从分离**

用户查询（读从库）：
```go
// 查询我的订单
func (r *OrderRepository) FindUserOrders(ctx context.Context, 
	userID int64, filter *OrderFilter) ([]*Order, error) {
	// 路由到从库
	db := r.shardingMgr.GetReadDB(userID)
	
	query := `SELECT * FROM orders WHERE user_id=?`
	args := []interface{}{userID}
	
	// 添加筛选条件
	if filter.Status != "" {
		query += ` AND status=?`
		args = append(args, filter.Status)
	}
	
	if !filter.StartTime.IsZero() {
		query += ` AND created_at >= ?`
		args = append(args, filter.StartTime)
	}
	
	query += ` ORDER BY created_at DESC LIMIT ? OFFSET ?`
	args = append(args, filter.PageSize, filter.Offset)
	
	rows, err := db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	
	return scanOrders(rows)
}
```

**方案二：ES同步（推荐）**

架构：
```text
订单创建/更新 → Kafka → 同步Worker → Elasticsearch

ES索引设计：
{
  "order_id": "123",
  "user_id": 456,
  "status": "PAID",
  "total_amount": 1000,
  "created_at": "2024-04-18T10:00:00Z",
  "items": [
    {"sku_id": "789", "title": "iPhone 15"}
  ]
}
```

查询实现：
```go
// 复杂查询用ES
func (r *OrderRepository) SearchOrders(ctx context.Context, 
	query *OrderSearchQuery) (*SearchResult, error) {
	esQuery := elastic.NewBoolQuery()
	
	// 用户维度
	if query.UserID > 0 {
		esQuery.Must(elastic.NewTermQuery("user_id", query.UserID))
	}
	
	// 订单号
	if query.OrderID != "" {
		esQuery.Must(elastic.NewTermQuery("order_id", query.OrderID))
	}
	
	// 状态
	if len(query.Statuses) > 0 {
		esQuery.Must(elastic.NewTermsQuery("status", query.Statuses...))
	}
	
	// 时间范围
	if !query.StartTime.IsZero() || !query.EndTime.IsZero() {
		rangeQuery := elastic.NewRangeQuery("created_at")
		if !query.StartTime.IsZero() {
			rangeQuery.Gte(query.StartTime)
		}
		if !query.EndTime.IsZero() {
			rangeQuery.Lte(query.EndTime)
		}
		esQuery.Must(rangeQuery)
	}
	
	// 商品筛选（嵌套查询）
	if query.SkuID != "" {
		esQuery.Must(elastic.NewNestedQuery("items",
			elastic.NewTermQuery("items.sku_id", query.SkuID)))
	}
	
	// 执行查询
	searchResult, err := r.esClient.Search().
		Index("orders").
		Query(esQuery).
		From(query.From).
		Size(query.Size).
		Sort("created_at", false).
		Do(ctx)
	
	if err != nil {
		return nil, err
	}
	
	return parseESResult(searchResult), nil
}
```

**延伸思考**：
1. 订单数据如何归档（如1年前的订单）？
2. 分库分表+ES同步如何保证一致性？

---

##### 📊 题目12：订单的消息通知设计

**问题描述**：
订单状态变化时需要通知用户（下单成功、发货、签收）。如何设计消息通知系统？

**答案**：

**通知渠道**：
1. App推送
2. 短信
3. 微信公众号/服务号
4. 站内信
5. 邮件

**推荐方案**（Go实现）：

```go
package notification

import (
	"context"
)

// NotificationService 通知服务
type NotificationService struct {
	pushSvc     PushService     // App推送
	smsSvc      SMSService      // 短信
	wechatSvc   WechatService   // 微信
	emailSvc    EmailService    // 邮件
	inboxSvc    InboxService    // 站内信
}

// NotifyOrderStatusChanged 订单状态变更通知
func (s *NotificationService) NotifyOrderStatusChanged(ctx context.Context, 
	order *Order, oldStatus, newStatus OrderStatus) error {
	
	// 根据状态确定通知内容
	template := s.getTemplate(newStatus)
	
	// 并行发送多渠道通知
	errChan := make(chan error, 5)
	
	// 1. App推送（必发）
	go func() {
		errChan <- s.pushSvc.Push(ctx, order.UserID, PushMessage{
			Title:   template.Title,
			Content: template.Content,
			Data:    map[string]interface{}{"order_id": order.OrderID},
		})
	}()
	
	// 2. 短信（重要状态才发）
	if s.shouldSendSMS(newStatus) {
		go func() {
			phone := s.getUserPhone(ctx, order.UserID)
			errChan <- s.smsSvc.Send(ctx, phone, template.SMSContent)
		}()
	} else {
		errChan <- nil
	}
	
	// 3. 微信（用户已绑定才发）
	go func() {
		if openID := s.getUserWechatOpenID(ctx, order.UserID); openID != "" {
			errChan <- s.wechatSvc.SendTemplateMessage(ctx, openID, template.WechatTemplate)
		} else {
			errChan <- nil
		}
	}()
	
	// 4. 站内信（必发）
	go func() {
		errChan <- s.inboxSvc.Create(ctx, &InboxMessage{
			UserID:  order.UserID,
			Title:   template.Title,
			Content: template.Content,
			Type:    "ORDER_UPDATE",
		})
	}()
	
	// 5. 邮件（用户订阅才发）
	go func() {
		if s.userHasEmailSubscription(ctx, order.UserID) {
			email := s.getUserEmail(ctx, order.UserID)
			errChan <- s.emailSvc.Send(ctx, email, template.EmailContent)
		} else {
			errChan <- nil
		}
	}()
	
	// 收集结果（至少一个渠道成功即可）
	successCount := 0
	for i := 0; i < 5; i++ {
		if err := <-errChan; err == nil {
			successCount++
		}
	}
	
	if successCount == 0 {
		return errors.New("所有通知渠道都失败")
	}
	
	return nil
}

// 通知模板
func (s *NotificationService) getTemplate(status OrderStatus) *NotificationTemplate {
	templates := map[OrderStatus]*NotificationTemplate{
		OrderStatusPaid: {
			Title:       "订单支付成功",
			Content:     "您的订单已支付成功，我们将尽快为您发货",
			SMSContent:  "【京东】您的订单已支付成功，预计3天内送达",
		},
		OrderStatusShipped: {
			Title:       "订单已发货",
			Content:     "您的订单已发货，快递单号：SF1234567890",
			SMSContent:  "【京东】您的订单已发货，单号SF1234567890",
		},
		OrderStatusReceived: {
			Title:       "订单已签收",
			Content:     "您的订单已签收，期待您的评价",
		},
	}
	
	return templates[status]
}

// 是否发送短信
func (s *NotificationService) shouldSendSMS(status OrderStatus) bool {
	// 只有关键状态发短信（控制成本）
	importantStatuses := []OrderStatus{
		OrderStatusPaid,
		OrderStatusShipped,
		OrderStatusRefunded,
	}
	
	for _, s := range importantStatuses {
		if s == status {
			return true
		}
	}
	return false
}
```

**延伸思考**：
1. 通知失败如何重试？
2. 如何设计通知的用户偏好设置（关闭某些通知）？
3. 大批量通知如何限流（避免骚扰）？

---

##### 🔧 题目13：订单数据的冷热分离

**问题描述**：
订单数据90天后很少查询，但占用大量存储。如何设计订单数据的冷热分离？

**答案**：

**推荐方案**：

```go
// 冷热分离策略
type OrderArchiveService struct {
	hotDB  *sql.DB  // 热数据库（MySQL）
	coldDB *sql.DB  // 冷数据库（可以是低成本存储）
	ossClient OSSClient // 对象存储
}

// 归档策略
func (s *OrderArchiveService) ArchiveOrders(ctx context.Context) error {
	// 1. 查询90天前已完成的订单
	cutoffTime := time.Now().AddDate(0, 0, -90)
	
	query := `SELECT * FROM orders 
	          WHERE status IN ('COMPLETED', 'CANCELLED', 'REFUNDED')
	          AND updated_at < ?
	          LIMIT 1000`
	
	rows, err := s.hotDB.QueryContext(ctx, query, cutoffTime)
	if err != nil {
		return err
	}
	defer rows.Close()
	
	orders := make([]*Order, 0)
	for rows.Next() {
		order := &Order{}
		// 扫描数据...
		orders = append(orders, order)
	}
	
	// 2. 写入冷库
	for _, order := range orders {
		if err := s.writeToArchive(ctx, order); err != nil {
			log.Errorf("归档订单%d失败: %v", order.OrderID, err)
			continue
		}
		
		// 3. 删除热库数据
		if err := s.deleteFromHot(ctx, order.OrderID); err != nil {
			log.Errorf("删除热库订单%d失败: %v", order.OrderID, err)
		}
	}
	
	return nil
}

// 查询时智能路由
func (s *OrderArchiveService) FindByID(ctx context.Context, orderID int64) (*Order, error) {
	// 1. 先查热库
	order, err := s.queryFromHot(ctx, orderID)
	if err == nil && order != nil {
		return order, nil
	}
	
	// 2. 查冷库
	order, err = s.queryFromArchive(ctx, orderID)
	if err == nil && order != nil {
		return order, nil
	}
	
	return nil, ErrOrderNotFound
}
```

**延伸思考**：
1. 归档订单如何支持查询？
2. 冷数据恢复到热库的策略？

---

##### 💡 题目14：订单的限流和防刷

**问题描述**：
恶意用户频繁下单不支付，占用库存和系统资源。如何设计订单的限流和防刷机制？

**答案**：

**推荐方案**（Go实现）：

```go
package ratelimit

import (
	"context"
	"fmt"
	"time"
	
	"github.com/go-redis/redis/v8"
)

// OrderRateLimiter 订单限流器
type OrderRateLimiter struct {
	rdb *redis.Client
}

// CheckLimit 检查用户是否超过限流
func (l *OrderRateLimiter) CheckLimit(ctx context.Context, userID int64) error {
	// 限流规则：
	// 1. 每分钟最多下单5次
	// 2. 每小时最多下单20次
	// 3. 每天最多50个待支付订单
	
	// 规则1：每分钟限流
	key1 := fmt.Sprintf("order:limit:min:%d:%s", userID, time.Now().Format("200601021504"))
	count1, err := l.rdb.Incr(ctx, key1).Result()
	if err != nil {
		return err
	}
	if count1 == 1 {
		l.rdb.Expire(ctx, key1, time.Minute)
	}
	if count1 > 5 {
		return errors.New("下单太频繁，请稍后再试")
	}
	
	// 规则2：每小时限流
	key2 := fmt.Sprintf("order:limit:hour:%d:%s", userID, time.Now().Format("2006010215"))
	count2, err := l.rdb.Incr(ctx, key2).Result()
	if err != nil {
		return err
	}
	if count2 == 1 {
		l.rdb.Expire(ctx, key2, time.Hour)
	}
	if count2 > 20 {
		return errors.New("您今天下单次数过多，请明天再试")
	}
	
	// 规则3：待支付订单数量限制
	pendingCount, err := l.getPendingOrderCount(ctx, userID)
	if err != nil {
		return err
	}
	if pendingCount >= 50 {
		return errors.New("您有过多待支付订单，请先完成支付")
	}
	
	return nil
}

// 用户信用评分
type UserCreditService struct {
	repo UserCreditRepository
}

func (s *UserCreditService) CheckCredit(ctx context.Context, userID int64) error {
	credit := s.repo.GetCredit(ctx, userID)
	
	// 信用分低于60分，禁止下单
	if credit.Score < 60 {
		return errors.New("您的信用分过低，暂时无法下单")
	}
	
	return nil
}

// 信用分扣减规则
func (s *UserCreditService) UpdateCredit(ctx context.Context, userID int64, behavior string) {
	switch behavior {
	case "ORDER_TIMEOUT":
		// 订单超时未支付：-5分
		s.repo.DeductCredit(ctx, userID, 5, "订单超时未支付")
	case "MALICIOUS_REFUND":
		// 恶意退款：-10分
		s.repo.DeductCredit(ctx, userID, 10, "恶意退款")
	case "ORDER_COMPLETED":
		// 订单完成：+1分
		s.repo.AddCredit(ctx, userID, 1, "订单完成")
	}
}
```

**延伸思考**：
1. 如何识别黄牛和恶意用户？
2. 限流策略如何针对不同用户等级差异化？

---

##### 📊 题目15：订单的实时数据统计

**问题描述**：
运营大盘需要实时显示订单量、GMV、转化率。如何设计订单的实时统计系统？

**答案**：

**推荐方案**：Flink流式计算

```go
// 实时统计指标
type OrderMetrics struct {
	Timestamp      time.Time
	OrderCount     int64           // 订单数
	GMV            decimal.Decimal // 交易额
	PaidOrderCount int64           // 已支付订单数
	AvgOrderAmount decimal.Decimal // 客单价
}

// 指标计算Worker（消费Kafka）
func ConsumeOrderEvents(ctx context.Context) {
	consumer := kafka.NewConsumer(...)
	
	for {
		msg, err := consumer.ReadMessage(ctx)
		if err != nil {
			continue
		}
		
		event := parseOrderEvent(msg.Value)
		
		switch event.Type {
		case "OrderCreated":
			// 订单数+1
			metrics.IncrOrderCount()
			
		case "OrderPaid":
			// 已支付订单数+1
			metrics.IncrPaidOrderCount()
			// GMV累加
			metrics.AddGMV(event.Order.PaidAmount)
			
		case "OrderCancelled":
			// 订单数-1（或单独统计取消数）
			metrics.IncrCancelledOrderCount()
		}
		
		// 定期刷新到Redis
		if time.Now().Unix()%10 == 0 {
			metrics.FlushToRedis()
		}
	}
}

// 实时大盘查询
func GetRealTimeMetrics(ctx context.Context) (*OrderMetrics, error) {
	// 从Redis读取实时指标
	rdb := redis.NewClient(...)
	
	orderCount, _ := rdb.Get(ctx, "metrics:order:count").Int64()
	gmv, _ := rdb.Get(ctx, "metrics:order:gmv").Float64()
	paidCount, _ := rdb.Get(ctx, "metrics:order:paid_count").Int64()
	
	return &OrderMetrics{
		Timestamp:      time.Now(),
		OrderCount:     orderCount,
		GMV:            decimal.NewFromFloat(gmv),
		PaidOrderCount: paidCount,
		AvgOrderAmount: decimal.NewFromFloat(gmv).Div(decimal.NewFromInt(paidCount)),
	}, nil
}
```

**延伸思考**：
1. 实时统计如何保证准确性（与离线对账）？
2. 多维度统计（按类目、品牌）如何设计？

---

---

### 41.1.4 支付系统（10题）

##### 📊 题目1：支付系统的整体架构设计

**问题描述**：
电商平台需要支持多种支付方式（支付宝、微信、银行卡）。如何设计支付系统的整体架构？

**答案**：

**问题分析**：
支付系统的核心要素：
1. 多渠道接入（支付宝、微信、银联）
2. 支付安全性
3. 异步回调处理
4. 对账和资金安全

**架构设计**（Go实现）：

```go
package payment

import (
	"context"
	"time"
)

// PaymentChannel 支付渠道
type PaymentChannel string

const (
	ChannelAlipay PaymentChannel = "ALIPAY"
	ChannelWechat PaymentChannel = "WECHAT"
	ChannelUnion  PaymentChannel = "UNION"
)

// PaymentService 支付服务
type PaymentService struct {
	alipayAdapter  PaymentAdapter
	wechatAdapter  PaymentAdapter
	unionAdapter   PaymentAdapter
	paymentRepo    PaymentRepository
	orderSvc       OrderService
}

// PaymentAdapter 支付适配器接口（适配器模式）
type PaymentAdapter interface {
	// 创建支付
	CreatePayment(ctx context.Context, req *PaymentRequest) (*PaymentResponse, error)
	// 查询支付状态
	QueryPayment(ctx context.Context, paymentID string) (*PaymentStatus, error)
	// 申请退款
	Refund(ctx context.Context, req *RefundRequest) error
	// 验证回调签名
	VerifyCallback(callback *CallbackData) error
}

// Payment 支付单
type Payment struct {
	PaymentID       string
	OrderID         int64
	UserID          int64
	Channel         PaymentChannel
	Amount          decimal.Decimal
	Status          PaymentStatus
	ThirdPartyID    string  // 第三方支付单号
	CallbackData    string  // 回调原始数据
	CreatedAt       time.Time
	PaidAt          *time.Time
}

// CreatePayment 创建支付
func (s *PaymentService) CreatePayment(ctx context.Context, 
	orderID int64, channel PaymentChannel) (*PaymentResponse, error) {
	
	// 1. 查询订单
	order, err := s.orderSvc.GetOrder(ctx, orderID)
	if err != nil {
		return nil, err
	}
	
	// 2. 校验订单状态
	if order.Status != OrderStatusPending {
		return nil, errors.New("订单状态不正确")
	}
	
	// 3. 创建支付单
	payment := &Payment{
		PaymentID: generatePaymentID(),
		OrderID:   orderID,
		UserID:    order.UserID,
		Channel:   channel,
		Amount:    order.TotalAmount,
		Status:    PaymentStatusPending,
		CreatedAt: time.Now(),
	}
	
	if err := s.paymentRepo.Create(ctx, payment); err != nil {
		return nil, err
	}
	
	// 4. 调用支付渠道
	adapter := s.getAdapter(channel)
	resp, err := adapter.CreatePayment(ctx, &PaymentRequest{
		OutTradeNo:  payment.PaymentID,
		Amount:      payment.Amount,
		Subject:     fmt.Sprintf("订单%d支付", orderID),
		NotifyURL:   "https://api.example.com/payment/callback",
		ReturnURL:   "https://www.example.com/order/success",
	})
	
	if err != nil {
		return nil, err
	}
	
	// 5. 保存第三方支付单号
	payment.ThirdPartyID = resp.TradeNo
	s.paymentRepo.Update(ctx, payment)
	
	return resp, nil
}

// 获取支付适配器
func (s *PaymentService) getAdapter(channel PaymentChannel) PaymentAdapter {
	switch channel {
	case ChannelAlipay:
		return s.alipayAdapter
	case ChannelWechat:
		return s.wechatAdapter
	case ChannelUnion:
		return s.unionAdapter
	default:
		return nil
	}
}

// HandleCallback 处理支付回调
func (s *PaymentService) HandleCallback(ctx context.Context, 
	channel PaymentChannel, callback *CallbackData) error {
	
	// 1. 验证签名
	adapter := s.getAdapter(channel)
	if err := adapter.VerifyCallback(callback); err != nil {
		return fmt.Errorf("签名验证失败: %w", err)
	}
	
	// 2. 查询支付单
	payment, err := s.paymentRepo.FindByID(ctx, callback.OutTradeNo)
	if err != nil {
		return err
	}
	
	// 3. 幂等性检查
	if payment.Status == PaymentStatusSuccess {
		return nil // 已处理，直接返回
	}
	
	// 4. 更新支付单状态
	payment.Status = PaymentStatusSuccess
	payment.PaidAt = &callback.PayTime
	payment.CallbackData = callback.RawData
	
	if err := s.paymentRepo.Update(ctx, payment); err != nil {
		return err
	}
	
	// 5. 更新订单状态
	if err := s.orderSvc.MarkAsPaid(ctx, payment.OrderID); err != nil {
		// 支付成功但订单更新失败，记录补偿任务
		s.createCompensationTask(ctx, payment.PaymentID)
		return err
	}
	
	// 6. 发布支付成功事件
	s.eventBus.Publish(&PaymentSuccessEvent{
		OrderID:   payment.OrderID,
		PaymentID: payment.PaymentID,
		Amount:    payment.Amount,
	})
	
	return nil
}
```

**支付宝适配器示例**：
```go
type AlipayAdapter struct {
	client *alipay.Client
}

func (a *AlipayAdapter) CreatePayment(ctx context.Context, 
	req *PaymentRequest) (*PaymentResponse, error) {
	
	// 调用支付宝SDK
	payReq := alipay.TradeAppPay{
		OutTradeNo:  req.OutTradeNo,
		TotalAmount: req.Amount.String(),
		Subject:     req.Subject,
		NotifyURL:   req.NotifyURL,
	}
	
	orderStr, err := a.client.TradeAppPay(payReq)
	if err != nil {
		return nil, err
	}
	
	return &PaymentResponse{
		PayData: orderStr, // APP端拉起支付宝所需的参数
	}, nil
}

func (a *AlipayAdapter) VerifyCallback(callback *CallbackData) error {
	// 验证支付宝回调签名
	return a.client.VerifySign(callback.RawData)
}
```

**延伸思考**：
1. 支付系统如何实现高可用？
2. 支付渠道故障如何降级？
3. 支付回调丢失如何处理？

---

##### 🔧 题目2：支付回调的幂等性处理

**问题描述**：
支付回调可能重复发送（网络重试、第三方重推）。如何保证支付回调处理的幂等性？

**答案**：

**推荐方案**（Go实现）：

```go
// 幂等性处理
func (s *PaymentService) HandleCallbackIdempotent(ctx context.Context, 
	callback *CallbackData) error {
	
	paymentID := callback.OutTradeNo
	lockKey := fmt.Sprintf("payment:callback:lock:%s", paymentID)
	
	// 1. 获取分布式锁
	lock := redis.NewDistributedLock(s.rdb, lockKey)
	acquired, err := lock.TryLock(ctx, 30*time.Second)
	if err != nil {
		return err
	}
	if !acquired {
		// 其他请求正在处理，直接返回成功
		return nil
	}
	defer lock.Unlock(ctx)
	
	// 2. 查询支付单
	payment, err := s.paymentRepo.FindByID(ctx, paymentID)
	if err != nil {
		return err
	}
	
	// 3. 状态检查（幂等性）
	if payment.Status == PaymentStatusSuccess {
		log.Infof("支付单%s已处理，跳过", paymentID)
		return nil // 已成功，幂等返回
	}
	
	// 4. 使用数据库行锁+版本号
	affected, err := s.paymentRepo.UpdateStatusWithVersion(ctx, 
		paymentID, 
		PaymentStatusSuccess,
		payment.Version,
	)
	
	if err != nil {
		return err
	}
	
	if affected == 0 {
		// 版本号不匹配，说明已被其他请求处理
		log.Warnf("支付单%s已被处理，版本冲突", paymentID)
		return nil
	}
	
	// 5. 执行后续操作
	return s.postPaymentProcess(ctx, payment)
}

// 数据库更新（带版本号）
func (r *PaymentRepository) UpdateStatusWithVersion(ctx context.Context, 
	paymentID string, newStatus PaymentStatus, expectedVersion int) (int64, error) {
	
	query := `UPDATE payments 
	          SET status=?, version=version+1, updated_at=?
	          WHERE payment_id=? AND version=? AND status!=?`
	
	result, err := r.db.ExecContext(ctx, query, 
		newStatus, time.Now(), paymentID, expectedVersion, PaymentStatusSuccess)
	if err != nil {
		return 0, err
	}
	
	return result.RowsAffected()
}
```

**延伸思考**：
1. 如何设计支付回调的重试机制？
2. 回调处理失败如何人工介入？

---

##### 💡 题目3：支付的对账系统设计

**问题描述**：
每天需要与支付宝、微信对账，确保平台账和渠道账一致。如何设计支付对账系统？

**答案**：

**对账流程**（Go实现）：

```go
package reconciliation

import (
	"context"
	"time"
)

// ReconciliationService 对账服务
type ReconciliationService struct {
	paymentRepo PaymentRepository
	alipayClient *alipay.Client
	wechatClient *wechat.Client
}

// DailyReconciliation 每日对账
func (s *ReconciliationService) DailyReconciliation(ctx context.Context, date time.Time) error {
	// 1. 下载渠道对账单
	alipayBill, err := s.downloadAlipayBill(ctx, date)
	if err != nil {
		return err
	}
	
	wechatBill, err := s.downloadWechatBill(ctx, date)
	if err != nil {
		return err
	}
	
	// 2. 查询平台当日支付记录
	platformRecords, err := s.paymentRepo.FindByDate(ctx, date)
	if err != nil {
		return err
	}
	
	// 3. 三方对账
	diff := s.compare(platformRecords, alipayBill, wechatBill)
	
	// 4. 处理差异
	if err := s.handleDifferences(ctx, diff); err != nil {
		return err
	}
	
	// 5. 生成对账报告
	report := s.generateReport(diff)
	s.saveReport(ctx, report)
	
	return nil
}

// ReconciliationDiff 对账差异
type ReconciliationDiff struct {
	OnlyInPlatform   []*Payment  // 只在平台有
	OnlyInChannel    []*ChannelRecord  // 只在渠道有
	AmountMismatch   []*Mismatch  // 金额不一致
	StatusMismatch   []*Mismatch  // 状态不一致
}

// compare 比对数据
func (s *ReconciliationService) compare(platform []*Payment, 
	alipay, wechat []*ChannelRecord) *ReconciliationDiff {
	
	diff := &ReconciliationDiff{}
	
	// 构建平台数据map
	platformMap := make(map[string]*Payment)
	for _, p := range platform {
		platformMap[p.ThirdPartyID] = p
	}
	
	// 构建渠道数据map
	channelMap := make(map[string]*ChannelRecord)
	for _, c := range alipay {
		channelMap[c.TradeNo] = c
	}
	for _, c := range wechat {
		channelMap[c.TransactionID] = c
	}
	
	// 比对
	for tradeNo, channelRecord := range channelMap {
		platformRecord, exists := platformMap[tradeNo]
		
		if !exists {
			// 只在渠道有，平台无
			diff.OnlyInChannel = append(diff.OnlyInChannel, channelRecord)
		} else {
			// 金额比对
			if !platformRecord.Amount.Equal(channelRecord.Amount) {
				diff.AmountMismatch = append(diff.AmountMismatch, &Mismatch{
					TradeNo:        tradeNo,
					PlatformAmount: platformRecord.Amount,
					ChannelAmount:  channelRecord.Amount,
				})
			}
			
			// 状态比对
			if platformRecord.Status != channelRecord.Status {
				diff.StatusMismatch = append(diff.StatusMismatch, &Mismatch{
					TradeNo:       tradeNo,
					PlatformStatus: platformRecord.Status,
					ChannelStatus:  channelRecord.Status,
				})
			}
			
			delete(platformMap, tradeNo)
		}
	}
	
	// 只在平台有的
	for _, p := range platformMap {
		diff.OnlyInPlatform = append(diff.OnlyInPlatform, p)
	}
	
	return diff
}

// handleDifferences 处理差异
func (s *ReconciliationService) handleDifferences(ctx context.Context, 
	diff *ReconciliationDiff) error {
	
	// 1. 只在渠道有的（平台漏单）
	for _, record := range diff.OnlyInChannel {
		log.Warnf("平台漏单: %s", record.TradeNo)
		// 补单：创建支付记录
		s.createMissingPayment(ctx, record)
	}
	
	// 2. 只在平台有的（渠道无记录，可能未支付成功）
	for _, payment := range diff.OnlyInPlatform {
		log.Warnf("渠道无记录: %s", payment.PaymentID)
		// 主动查询第三方状态
		s.queryThirdPartyStatus(ctx, payment)
	}
	
	// 3. 金额不一致
	for _, mismatch := range diff.AmountMismatch {
		log.Errorf("金额不一致: %s, 平台=%v, 渠道=%v", 
			mismatch.TradeNo, mismatch.PlatformAmount, mismatch.ChannelAmount)
		// 转人工处理
		s.createManualTask(ctx, "AMOUNT_MISMATCH", mismatch)
	}
	
	// 4. 状态不一致
	for _, mismatch := range diff.StatusMismatch {
		log.Warnf("状态不一致: %s", mismatch.TradeNo)
		// 以渠道状态为准，更新平台状态
		s.syncStatus(ctx, mismatch)
	}
	
	return nil
}
```

**对账报告**：
```go
type ReconciliationReport struct {
	Date              time.Time
	TotalCount        int
	MatchCount        int
	MismatchCount     int
	OnlyInPlatform    int
	OnlyInChannel     int
	AmountMismatch    int
	TotalAmount       decimal.Decimal
	ChannelTotalAmount decimal.Decimal
}
```

**延伸思考**：
1. 对账差异如何自动修复？
2. 对账失败如何告警和处理？
3. 实时对账和T+1对账如何结合？

---

##### 📊 题目4：支付的异步回调处理

**问题描述**：
支付成功后，第三方通过回调通知平台。回调可能延迟、丢失、重复。如何设计健壮的回调处理机制？

**答案**：

**推荐方案**（Go实现）：

```go
// 回调处理器
type CallbackHandler struct {
	paymentSvc  *PaymentService
	orderSvc    *OrderService
	lockSvc     *DistributedLockService
}

// HandleCallback 处理回调
func (h *CallbackHandler) HandleCallback(ctx context.Context, 
	channel PaymentChannel, rawData []byte) error {
	
	// 1. 解析回调数据
	callback, err := parseCallback(channel, rawData)
	if err != nil {
		return fmt.Errorf("解析回调失败: %w", err)
	}
	
	// 2. 记录回调日志（用于排查问题）
	h.logCallback(ctx, callback)
	
	// 3. 验证签名
	adapter := h.paymentSvc.getAdapter(channel)
	if err := adapter.VerifyCallback(callback); err != nil {
		log.Errorf("回调签名验证失败: %v", err)
		return err
	}
	
	// 4. 幂等性处理（分布式锁）
	lockKey := fmt.Sprintf("payment:callback:%s", callback.OutTradeNo)
	acquired, err := h.lockSvc.TryLock(ctx, lockKey, 30*time.Second)
	if err != nil {
		return err
	}
	if !acquired {
		log.Infof("回调%s正在处理中，跳过", callback.OutTradeNo)
		return nil
	}
	defer h.lockSvc.Unlock(ctx, lockKey)
	
	// 5. 处理支付结果
	return h.paymentSvc.HandleCallback(ctx, channel, callback)
}

// 主动查询（回调超时补偿）
func (h *CallbackHandler) QueryPaymentStatus(ctx context.Context) {
	// 定时任务：查询10分钟前创建但未回调的支付单
	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()
	
	for range ticker.C {
		cutoffTime := time.Now().Add(-10 * time.Minute)
		
		// 查询超时支付单
		payments, err := h.paymentSvc.FindPendingPayments(ctx, cutoffTime)
		if err != nil {
			log.Errorf("查询超时支付单失败: %v", err)
			continue
		}
		
		for _, payment := range payments {
			// 主动查询第三方状态
			go func(p *Payment) {
				adapter := h.paymentSvc.getAdapter(p.Channel)
				status, err := adapter.QueryPayment(ctx, p.ThirdPartyID)
				if err != nil {
					log.Errorf("查询支付状态失败: %v", err)
					return
				}
				
				// 如果已支付，补偿处理
				if status.Status == "SUCCESS" {
					log.Warnf("支付单%s回调丢失，主动补偿", p.PaymentID)
					h.paymentSvc.MarkAsPaid(ctx, p.PaymentID)
				}
			}(payment)
		}
	}
}
```

**回调重试策略**：
```go
// 回调处理失败时的重试
func (h *CallbackHandler) retryCallback(ctx context.Context, 
	callback *CallbackData) error {
	
	maxRetries := 5
	backoff := []time.Duration{
		1 * time.Second,
		5 * time.Second,
		30 * time.Second,
		2 * time.Minute,
		10 * time.Minute,
	}
	
	for i := 0; i < maxRetries; i++ {
		err := h.HandleCallback(ctx, callback.Channel, callback.RawData)
		if err == nil {
			return nil // 成功
		}
		
		log.Warnf("回调处理失败，第%d次重试: %v", i+1, err)
		
		if i < maxRetries-1 {
			time.Sleep(backoff[i])
		}
	}
	
	// 所有重试失败，记录人工任务
	return h.createManualTask(ctx, "CALLBACK_FAILED", callback)
}
```

**延伸思考**：
1. 回调接口如何防止伪造（恶意请求）？
2. 回调处理超时如何设置？

---

##### 🔧 题目5：支付的分账系统设计（平台+商家）

**问题描述**：
B2B2C平台，用户支付100元，平台抽佣10%，商家获得90元。如何设计支付分账系统？

**答案**：

**推荐方案**（Go实现）：

```go
// Settlement 结算单
type Settlement struct {
	SettlementID   string
	OrderID        int64
	MerchantID     int64
	TotalAmount    decimal.Decimal  // 订单总额
	PlatformAmount decimal.Decimal  // 平台佣金
	MerchantAmount decimal.Decimal  // 商家收入
	Status         SettlementStatus
	SettledAt      *time.Time
}

// SettlementService 结算服务
type SettlementService struct {
	settlementRepo SettlementRepository
	paymentSvc     PaymentService
}

// CreateSettlement 创建结算单
func (s *SettlementService) CreateSettlement(ctx context.Context, 
	orderID int64) error {
	
	// 1. 查询订单
	order := s.orderSvc.GetOrder(ctx, orderID)
	
	// 2. 计算佣金
	commissionRate := s.getCommissionRate(ctx, order.MerchantID)
	platformAmount := order.TotalAmount.Mul(commissionRate)
	merchantAmount := order.TotalAmount.Sub(platformAmount)
	
	// 3. 创建结算单
	settlement := &Settlement{
		SettlementID:   generateSettlementID(),
		OrderID:        orderID,
		MerchantID:     order.MerchantID,
		TotalAmount:    order.TotalAmount,
		PlatformAmount: platformAmount,
		MerchantAmount: merchantAmount,
		Status:         SettlementPending,
	}
	
	return s.settlementRepo.Create(ctx, settlement)
}

// Settle 执行结算（T+N结算）
func (s *SettlementService) Settle(ctx context.Context, merchantID int64, date time.Time) error {
	// 1. 查询该商家待结算的订单
	settlements, err := s.settlementRepo.FindPendingByMerchant(ctx, merchantID, date)
	if err != nil {
		return err
	}
	
	// 2. 汇总金额
	totalAmount := decimal.Zero
	for _, s := range settlements {
		totalAmount = totalAmount.Add(s.MerchantAmount)
	}
	
	// 3. 调用支付渠道分账/转账
	if err := s.paymentSvc.Transfer(ctx, &TransferRequest{
		ToAccount: s.getMerchantAccount(ctx, merchantID),
		Amount:    totalAmount,
		Remark:    fmt.Sprintf("商家%d的%s结算", merchantID, date.Format("2006-01-02")),
	}); err != nil {
		return err
	}
	
	// 4. 更新结算单状态
	for _, settlement := range settlements {
		settlement.Status = SettlementCompleted
		settlement.SettledAt = timePtr(time.Now())
		s.settlementRepo.Update(ctx, settlement)
	}
	
	return nil
}

// 佣金率配置
func (s *SettlementService) getCommissionRate(ctx context.Context, merchantID int64) decimal.Decimal {
	// 根据商家等级、类目等确定佣金率
	merchant := s.merchantSvc.GetMerchant(ctx, merchantID)
	
	switch merchant.Level {
	case "VIP":
		return decimal.NewFromFloat(0.05) // 5%
	case "GOLD":
		return decimal.NewFromFloat(0.08) // 8%
	default:
		return decimal.NewFromFloat(0.10) // 10%
	}
}
```

**结算周期**：
```text
T+0：实时结算（高成本，高信用商家）
T+1：次日结算（平衡）
T+7：周结算（标准）
T+30：月结算（新商家）
```

**延伸思考**：
1. 如何设计结算的对账机制？
2. 商家提现如何设计？
3. 结算失败如何处理？

---

##### 📊 题目6：支付密码和安全设计

**问题描述**：
支付环节涉及资金安全，如何设计支付密码、短信验证码等安全机制？

**答案**：

**推荐方案**（Go实现）：

```go
// PaymentSecurityService 支付安全服务
type PaymentSecurityService struct {
	rdb        *redis.Client
	smsSvc     SMSService
	encryptSvc EncryptService
}

// VerifyPaymentPassword 验证支付密码
func (s *PaymentSecurityService) VerifyPaymentPassword(ctx context.Context, 
	userID int64, password string) error {
	
	// 1. 获取用户存储的支付密码（加密）
	user, err := s.userRepo.FindByID(ctx, userID)
	if err != nil {
		return err
	}
	
	if user.PaymentPassword == "" {
		return errors.New("请先设置支付密码")
	}
	
	// 2. 验证密码
	if !s.encryptSvc.VerifyPassword(password, user.PaymentPassword) {
		// 记录失败次数
		failCount := s.incrFailCount(ctx, userID)
		
		// 超过5次锁定账户
		if failCount >= 5 {
			s.lockAccount(ctx, userID, 30*time.Minute)
			return errors.New("密码错误次数过多，账户已锁定30分钟")
		}
		
		return fmt.Errorf("密码错误，还可尝试%d次", 5-failCount)
	}
	
	// 3. 清除失败计数
	s.clearFailCount(ctx, userID)
	
	return nil
}

// SendPaymentSMS 发送支付验证码
func (s *PaymentSecurityService) SendPaymentSMS(ctx context.Context, 
	userID int64, phone string) error {
	
	// 1. 限流检查（防止短信轰炸）
	key := fmt.Sprintf("sms:limit:%s", phone)
	count, err := s.rdb.Incr(ctx, key).Result()
	if err != nil {
		return err
	}
	if count == 1 {
		s.rdb.Expire(ctx, key, time.Hour)
	}
	if count > 5 {
		return errors.New("发送次数过多，请1小时后再试")
	}
	
	// 2. 生成6位验证码
	code := fmt.Sprintf("%06d", rand.Intn(1000000))
	
	// 3. 存储验证码（5分钟有效）
	codeKey := fmt.Sprintf("sms:code:%s", phone)
	s.rdb.SetEX(ctx, codeKey, code, 5*time.Minute)
	
	// 4. 发送短信
	return s.smsSvc.Send(ctx, phone, fmt.Sprintf("您的支付验证码是%s，5分钟内有效", code))
}

// VerifySMSCode 验证短信验证码
func (s *PaymentSecurityService) VerifySMSCode(ctx context.Context, 
	phone, code string) error {
	
	codeKey := fmt.Sprintf("sms:code:%s", phone)
	
	// 查询验证码
	storedCode, err := s.rdb.Get(ctx, codeKey).Result()
	if err == redis.Nil {
		return errors.New("验证码已过期")
	}
	if err != nil {
		return err
	}
	
	// 验证
	if storedCode != code {
		return errors.New("验证码错误")
	}
	
	// 验证成功，删除验证码（防止重复使用）
	s.rdb.Del(ctx, codeKey)
	
	return nil
}

// 风控检查
func (s *PaymentSecurityService) RiskCheck(ctx context.Context, 
	userID int64, amount decimal.Decimal) error {
	
	// 规则1：大额支付需要额外验证
	if amount.GreaterThan(decimal.NewFromInt(5000)) {
		// 需要短信验证码或支付密码
		return errors.New("REQUIRE_SMS_OR_PASSWORD")
	}
	
	// 规则2：新用户限额
	user := s.userSvc.GetUser(ctx, userID)
	if user.RegisterDays() < 7 && amount.GreaterThan(decimal.NewFromInt(1000)) {
		return errors.New("新用户单笔限额1000元")
	}
	
	// 规则3：异常IP检测
	ip := s.getRequestIP(ctx)
	if s.isBlacklistIP(ctx, ip) {
		return errors.New("异常IP，禁止支付")
	}
	
	// 规则4：高频支付检测
	recentPayments := s.getRecentPaymentCount(ctx, userID, 10*time.Minute)
	if recentPayments > 10 {
		return errors.New("支付频率异常")
	}
	
	return nil
}
```

**延伸思考**：
1. 支付密码如何加密存储？
2. 如何设计支付的二次确认（大额支付）？
3. 支付安全如何平衡用户体验？

---

##### 🔧 题目7：支付渠道的路由和降级

**问题描述**：
支付宝渠道故障时，如何自动切换到微信支付？如何设计支付渠道的路由和降级策略？

**答案**：

**推荐方案**（Go实现）：

```go
// ChannelRouter 支付渠道路由器
type ChannelRouter struct {
	healthChecker *ChannelHealthChecker
	config        *RoutingConfig
}

// SelectChannel 选择支付渠道
func (r *ChannelRouter) SelectChannel(ctx context.Context, 
	preferredChannel PaymentChannel) (PaymentChannel, error) {
	
	// 1. 检查首选渠道健康状态
	if r.healthChecker.IsHealthy(preferredChannel) {
		return preferredChannel, nil
	}
	
	log.Warnf("渠道%s不可用，尝试降级", preferredChannel)
	
	// 2. 降级到备用渠道
	fallbackChannels := r.config.GetFallback(preferredChannel)
	for _, channel := range fallbackChannels {
		if r.healthChecker.IsHealthy(channel) {
			log.Infof("降级到渠道%s", channel)
			return channel, nil
		}
	}
	
	// 3. 所有渠道都不可用
	return "", errors.New("支付渠道暂时不可用，请稍后再试")
}

// ChannelHealthChecker 渠道健康检查
type ChannelHealthChecker struct {
	rdb *redis.Client
}

func (c *ChannelHealthChecker) IsHealthy(channel PaymentChannel) bool {
	key := fmt.Sprintf("payment:channel:health:%s", channel)
	
	// 从Redis读取健康状态
	status, err := c.rdb.Get(context.Background(), key).Result()
	if err != nil || status != "UP" {
		return false
	}
	
	return true
}

// 健康检查任务（心跳）
func (c *ChannelHealthChecker) StartHealthCheck(ctx context.Context) {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()
	
	for range ticker.C {
		// 对每个渠道执行健康检查
		for _, channel := range AllChannels {
			go c.checkChannel(ctx, channel)
		}
	}
}

func (c *ChannelHealthChecker) checkChannel(ctx context.Context, 
	channel PaymentChannel) {
	
	adapter := getAdapter(channel)
	
	// 调用渠道健康检查接口（或创建1分钱订单测试）
	err := adapter.HealthCheck(ctx)
	
	key := fmt.Sprintf("payment:channel:health:%s", channel)
	if err != nil {
		// 不健康
		c.rdb.SetEX(ctx, key, "DOWN", 5*time.Minute)
		log.Errorf("渠道%s健康检查失败: %v", channel, err)
		
		// 告警
		c.alertSvc.Send(fmt.Sprintf("支付渠道%s故障", channel))
	} else {
		// 健康
		c.rdb.SetEX(ctx, key, "UP", 5*time.Minute)
	}
}

// 路由配置
type RoutingConfig struct {
	fallbacks map[PaymentChannel][]PaymentChannel
}

func NewRoutingConfig() *RoutingConfig {
	return &RoutingConfig{
		fallbacks: map[PaymentChannel][]PaymentChannel{
			ChannelAlipay: {ChannelWechat, ChannelUnion},  // 支付宝 → 微信 → 银联
			ChannelWechat: {ChannelAlipay, ChannelUnion},
			ChannelUnion:  {ChannelAlipay, ChannelWechat},
		},
	}
}
```

**延伸思考**：
1. 如何设计支付渠道的成本优化（选择手续费低的）？
2. 支付渠道限额如何处理？

---

##### 💡 题目8：支付的退款处理

**问题描述**：
用户申请退款，需要原路退回。如何设计退款流程，处理退款失败、部分退款等场景？

**答案**：

**推荐方案**（Go实现）：

```go
// RefundService 退款服务
type RefundService struct {
	paymentRepo PaymentRepository
}

// Refund 申请退款
func (s *RefundService) Refund(ctx context.Context, req *RefundRequest) error {
	// 1. 查询原支付记录
	payment, err := s.paymentRepo.FindByOrderID(ctx, req.OrderID)
	if err != nil {
		return err
	}
	
	// 2. 校验退款金额
	if req.RefundAmount.GreaterThan(payment.Amount) {
		return errors.New("退款金额超过支付金额")
	}
	
	// 3. 检查是否已退款
	totalRefunded, err := s.paymentRepo.GetTotalRefundedAmount(ctx, payment.PaymentID)
	if err != nil {
		return err
	}
	
	if totalRefunded.Add(req.RefundAmount).GreaterThan(payment.Amount) {
		return errors.New("累计退款金额超过支付金额")
	}
	
	// 4. 创建退款记录
	refund := &PaymentRefund{
		RefundID:    generateRefundID(),
		PaymentID:   payment.PaymentID,
		OrderID:     req.OrderID,
		Amount:      req.RefundAmount,
		Reason:      req.Reason,
		Status:      RefundStatusPending,
		CreatedAt:   time.Now(),
	}
	
	if err := s.refundRepo.Create(ctx, refund); err != nil {
		return err
	}
	
	// 5. 调用第三方退款接口
	adapter := s.getAdapter(payment.Channel)
	err = adapter.Refund(ctx, &ThirdPartyRefundRequest{
		OutRefundNo:   refund.RefundID,
		OutTradeNo:    payment.ThirdPartyID,
		RefundAmount:  req.RefundAmount,
		TotalAmount:   payment.Amount,
		RefundReason:  req.Reason,
	})
	
	if err != nil {
		refund.Status = RefundStatusFailed
		refund.FailReason = err.Error()
		s.refundRepo.Update(ctx, refund)
		return err
	}
	
	// 6. 更新退款状态
	refund.Status = RefundStatusSuccess
	refund.RefundedAt = timePtr(time.Now())
	s.refundRepo.Update(ctx, refund)
	
	return nil
}

// 退款重试（定时任务）
func (s *RefundService) RetryFailedRefunds(ctx context.Context) {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()
	
	for range ticker.C {
		// 查询失败的退款（创建时间<30分钟前）
		refunds, err := s.refundRepo.FindFailed(ctx, time.Now().Add(-30*time.Minute))
		if err != nil {
			log.Errorf("查询失败退款失败: %v", err)
			continue
		}
		
		for _, refund := range refunds {
			// 重试退款
			go func(r *PaymentRefund) {
				if r.RetryCount >= 5 {
					log.Errorf("退款%s重试次数过多，转人工处理", r.RefundID)
					s.createManualTask(ctx, r.RefundID)
					return
				}
				
				payment, _ := s.paymentRepo.FindByID(ctx, r.PaymentID)
				adapter := s.getAdapter(payment.Channel)
				
				err := adapter.Refund(ctx, &ThirdPartyRefundRequest{
					OutRefundNo:  r.RefundID,
					OutTradeNo:   payment.ThirdPartyID,
					RefundAmount: r.Amount,
					TotalAmount:  payment.Amount,
				})
				
				if err == nil {
					r.Status = RefundStatusSuccess
					r.RefundedAt = timePtr(time.Now())
				} else {
					r.RetryCount++
					r.FailReason = err.Error()
				}
				
				s.refundRepo.Update(ctx, r)
			}(refund)
		}
	}
}
```

**部分退款处理**：
```go
// 部分退款（一单多件商品，退部分）
func (s *RefundService) PartialRefund(ctx context.Context, 
	orderID int64, items []RefundItem) error {
	
	// 1. 计算退款金额
	var refundAmount decimal.Decimal
	for _, item := range items {
		itemAmount := item.Price.Mul(decimal.NewFromInt(int64(item.Quantity)))
		refundAmount = refundAmount.Add(itemAmount)
	}
	
	// 2. 分摊运费
	order := s.orderSvc.GetOrder(ctx, orderID)
	refundItemCount := len(items)
	totalItemCount := len(order.Items)
	
	shippingRefund := order.ShippingFee.
		Mul(decimal.NewFromInt(int64(refundItemCount))).
		Div(decimal.NewFromInt(int64(totalItemCount)))
	
	refundAmount = refundAmount.Add(shippingRefund)
	
	// 3. 执行退款
	return s.Refund(ctx, &RefundRequest{
		OrderID:      orderID,
		RefundAmount: refundAmount,
		RefundItems:  items,
		Reason:       "部分退货",
	})
}
```

**延伸思考**：
1. 退款失败如何通知用户？
2. 如何设计退款的限额控制（防止洗钱）？

---

##### 🔧 题目9：支付的容灾和降级

**问题描述**：
支付是核心链路，不能中断。如何设计支付系统的容灾和降级方案？

**答案**：

**推荐方案**（Go实现）：

```go
// PaymentFallbackService 支付降级服务
type PaymentFallbackService struct {
	primarySvc   *PaymentService
	fallbackMode bool
}

// Pay 支付（带降级）
func (s *PaymentFallbackService) Pay(ctx context.Context, 
	req *PaymentRequest) (*PaymentResponse, error) {
	
	// 1. 尝试正常支付
	resp, err := s.primarySvc.CreatePayment(ctx, req)
	if err == nil {
		return resp, nil
	}
	
	log.Warnf("支付失败: %v，尝试降级", err)
	
	// 2. 降级方案
	if s.shouldFallback(err) {
		return s.fallbackPay(ctx, req)
	}
	
	return nil, err
}

// 降级支付
func (s *PaymentFallbackService) fallbackPay(ctx context.Context, 
	req *PaymentRequest) (*PaymentResponse, error) {
	
	// 降级策略1：切换支付渠道
	if req.Channel == ChannelAlipay {
		req.Channel = ChannelWechat
		return s.primarySvc.CreatePayment(ctx, req)
	}
	
	// 降级策略2：使用货到付款
	if s.isCODAvailable(req) {
		return s.createCODOrder(ctx, req)
	}
	
	// 降级策略3：延迟支付（订单保留，稍后支付）
	return s.createDelayedPayment(ctx, req)
}

// 熔断器
type CircuitBreaker struct {
	failureThreshold int
	timeout          time.Duration
	state            CircuitState
	failureCount     int
	lastFailTime     time.Time
}

type CircuitState int

const (
	StateClosed CircuitState = 0  // 闭合（正常）
	StateOpen   CircuitState = 1  // 开启（熔断）
	StateHalfOpen CircuitState = 2  // 半开（尝试恢复）
)

func (cb *CircuitBreaker) Execute(ctx context.Context, 
	fn func() error) error {
	
	// 检查熔断器状态
	if cb.state == StateOpen {
		// 检查是否可以尝试恢复
		if time.Since(cb.lastFailTime) > cb.timeout {
			cb.state = StateHalfOpen
		} else {
			return errors.New("熔断器开启，拒绝请求")
		}
	}
	
	// 执行函数
	err := fn()
	
	if err != nil {
		cb.onFailure()
	} else {
		cb.onSuccess()
	}
	
	return err
}

func (cb *CircuitBreaker) onFailure() {
	cb.failureCount++
	cb.lastFailTime = time.Now()
	
	if cb.failureCount >= cb.failureThreshold {
		cb.state = StateOpen
		log.Warn("熔断器开启")
	}
}

func (cb *CircuitBreaker) onSuccess() {
	if cb.state == StateHalfOpen {
		// 半开状态成功，恢复到闭合
		cb.state = StateClosed
		cb.failureCount = 0
		log.Info("熔断器关闭，恢复正常")
	}
}
```

**延伸思考**：
1. 如何设计支付系统的多机房容灾？
2. 支付降级后如何通知用户？

---

##### 📊 题目10：预授权支付的设计（酒店、租车场景）

**问题描述**：
酒店预订需要预授权（冻结资金但不扣款），退房时根据实际消费扣款。如何设计预授权支付？

**答案**：

**推荐方案**（Go实现）：

```go
// PreAuthService 预授权服务
type PreAuthService struct {
	paymentAdapter PaymentAdapter
	preAuthRepo    PreAuthRepository
}

// PreAuthorize 预授权
func (s *PreAuthService) PreAuthorize(ctx context.Context, 
	req *PreAuthRequest) (*PreAuthResponse, error) {
	
	// 1. 创建预授权记录
	preAuth := &PreAuthorization{
		PreAuthID:  generatePreAuthID(),
		OrderID:    req.OrderID,
		UserID:     req.UserID,
		Amount:     req.Amount,  // 冻结金额
		Status:     PreAuthStatusFrozen,
		CreatedAt:  time.Now(),
		ExpireAt:   time.Now().Add(30 * 24 * time.Hour), // 30天有效期
	}
	
	if err := s.preAuthRepo.Create(ctx, preAuth); err != nil {
		return nil, err
	}
	
	// 2. 调用支付渠道预授权接口
	resp, err := s.paymentAdapter.PreAuthorize(ctx, &ThirdPartyPreAuthRequest{
		OutRequestNo: preAuth.PreAuthID,
		Amount:       req.Amount,
		ExpireTime:   preAuth.ExpireAt,
	})
	
	if err != nil {
		preAuth.Status = PreAuthStatusFailed
		s.preAuthRepo.Update(ctx, preAuth)
		return nil, err
	}
	
	// 3. 保存第三方预授权号
	preAuth.ThirdPartyID = resp.AuthNo
	s.preAuthRepo.Update(ctx, preAuth)
	
	return &PreAuthResponse{
		PreAuthID: preAuth.PreAuthID,
		AuthNo:    resp.AuthNo,
	}, nil
}

// Complete 完成预授权（实际扣款）
func (s *PreAuthService) Complete(ctx context.Context, 
	preAuthID string, actualAmount decimal.Decimal) error {
	
	// 1. 查询预授权
	preAuth, err := s.preAuthRepo.FindByID(ctx, preAuthID)
	if err != nil {
		return err
	}
	
	// 2. 校验金额
	if actualAmount.GreaterThan(preAuth.Amount) {
		return errors.New("实际金额超过预授权金额")
	}
	
	// 3. 调用支付渠道完成预授权
	err = s.paymentAdapter.CompletePreAuth(ctx, &CompletePreAuthRequest{
		AuthNo: preAuth.ThirdPartyID,
		Amount: actualAmount,
	})
	
	if err != nil {
		return err
	}
	
	// 4. 更新状态
	preAuth.Status = PreAuthStatusCompleted
	preAuth.ActualAmount = actualAmount
	preAuth.CompletedAt = timePtr(time.Now())
	s.preAuthRepo.Update(ctx, preAuth)
	
	// 5. 多余金额解冻
	if actualAmount.LessThan(preAuth.Amount) {
		unfreezeAmount := preAuth.Amount.Sub(actualAmount)
		log.Infof("解冻多余金额: %v", unfreezeAmount)
	}
	
	return nil
}

// Cancel 取消预授权
func (s *PreAuthService) Cancel(ctx context.Context, preAuthID string) error {
	preAuth, err := s.preAuthRepo.FindByID(ctx, preAuthID)
	if err != nil {
		return err
	}
	
	// 调用支付渠道取消预授权
	err = s.paymentAdapter.CancelPreAuth(ctx, preAuth.ThirdPartyID)
	if err != nil {
		return err
	}
	
	preAuth.Status = PreAuthStatusCancelled
	s.preAuthRepo.Update(ctx, preAuth)
	
	return nil
}
```

**延伸思考**：
1. 预授权过期如何自动解冻？
2. 预授权场景下的对账如何设计？

---

---
