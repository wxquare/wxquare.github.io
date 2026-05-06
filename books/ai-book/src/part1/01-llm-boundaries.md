# 第1章 LLM 能力边界与架构约束

> "Understanding the limits of AI is as important as understanding its capabilities." （理解AI的边界与理解它的能力同样重要）

## 引言

在进入 Prompt、Context、Harness 和 Agent 架构之前，我们需要先理解 LLM 的**能力边界**和**工程化要点**。许多 AI 工程问题并不是工具不够强，而是我们把模型当成了确定性系统。

许多 Agent 系统失败的根源，不是架构设计问题，而是对 LLM 能力的**误解**或**过度期待**。本章将系统梳理 LLM 的能力边界、常见陷阱和工程化最佳实践。

---

## 1.1 LLM的能力边界

### LLM 擅长什么？

基于 GPT-4、Claude 3.5、Gemini 等现代 LLM 的实际表现：

**✅ 强项：**

| 能力 | 准确率 | 适用场景 |
|------|--------|---------|
| **文本生成** | 90%+ | 写作、总结、翻译 |
| **代码生成** | 85%+ | 标准算法、常见框架 |
| **信息提取** | 85-90% | 结构化数据提取 |
| **逻辑推理** | 80-85% | 简单推理、常识判断 |
| **对话交互** | 90%+ | 客服、助手、问答 |
| **创意生成** | 90%+ | 头脑风暴、创意写作 |

**实际例子：**

```text
任务: 将客户反馈分类
输入: "订单号12345还没发货，已经等了3天了"
LLM输出:
{
  "category": "物流问题",
  "sentiment": "不满",
  "urgency": "中",
  "order_id": "12345"
}

准确率: ~95%
```

### LLM 不擅长什么？

**❌ 弱项：**

| 限制 | 表现 | 原因 |
|------|------|------|
| **精确计算** | 60-70% | 非符号推理 |
| **实时信息** | 0%（知识截止） | 训练数据有时效性 |
| **长期记忆** | 依赖上下文 | 无状态设计 |
| **复杂多步推理** | 60-75% | 推理链容易断裂 |
| **自我修正** | 50-60% | 缺少验证机制 |
| **领域专业知识** | 70-85% | 依赖训练数据 |

**实际例子：**

```text
任务: 计算复利
输入: "本金10000元，年利率5%，复利计算20年后是多少？"

❌ LLM 直接计算:
"大约是 25000 元"
正确答案: 26532.98 元
错误率: ~6%

✅ 正确做法:
LLM 生成公式 → 调用计算工具
FV = 10000 × (1 + 0.05)^20 = 26532.98
准确率: 100%
```

### 能力边界总结

```text
LLM 的核心能力：
┌─────────────────────────────────────┐
│  模式识别 + 序列生成               │
├─────────────────────────────────────┤
│  基于训练数据中的模式，           │
│  生成最可能的下一个 token         │
└─────────────────────────────────────┘

推论：
1. 擅长：常见模式、标准任务
2. 不擅长：罕见场景、精确计算
3. 不可靠：幻觉、一致性
```

---

## 1.2 幻觉问题与应对策略

### 什么是幻觉？

**定义**：LLM 生成看似合理但实际错误的内容。

**典型案例：**

```text
问题: "2022年诺贝尔物理学奖得主是谁？"

LLM 幻觉回答:
"2022年诺贝尔物理学奖授予了 John Doe 和 Jane Smith，
表彰他们在量子纠缠领域的贡献。"

问题：
1. 名字是编造的
2. 研究方向是猜测的
3. 表述非常自信

正确答案:
Alain Aspect, John Clauser, Anton Zeilinger
（量子信息科学的奠基实验）
```

### 幻觉的类型

**1. 事实性幻觉（Factual Hallucination）**

```text
输入: "介绍一下 TensorFlow 2.0 的新特性"
幻觉: "TensorFlow 2.0 引入了自动微分功能"
事实: TensorFlow 1.x 就有自动微分
```

**2. 逻辑性幻觉（Logical Hallucination）**

```text
输入: "如果 A > B 且 B > C，那么 A 和 C 的关系？"
幻觉: "无法确定 A 和 C 的关系"
事实: 必然 A > C（传递性）
```

**3. 引用性幻觉（Citation Hallucination）**

```text
输入: "引用一篇关于 Transformer 的论文"
幻觉: "根据 Smith et al. (2023) 的研究..."
事实: 这篇论文不存在
```

### 应对策略

**策略 1：工具调用（Tool Use）**

```python
# ❌ 直接让 LLM 计算
prompt = "计算 sin(45°) × cos(30°)"
result = llm.generate(prompt)  # 不可靠

# ✅ 调用计算工具
prompt = "生成 Python 代码计算 sin(45°) × cos(30°)"
code = llm.generate(prompt)
result = execute_code(code)  # 可靠
```

**策略 2：检索增强（RAG）**

```python
# ❌ 直接询问 LLM
answer = llm.generate("2022年诺贝尔物理学奖得主？")

# ✅ 先检索再回答
docs = search_wikipedia("2022 Nobel Prize Physics")
answer = llm.generate(f"基于以下资料回答：\n{docs}\n问题：...")
```

**策略 3：Self-Consistency（自我一致性）**

```python
# 多次采样，选择一致的答案
answers = []
for _ in range(5):
    answer = llm.generate(question, temperature=0.7)
    answers.append(answer)

# 投票选择最一致的答案
final_answer = most_common(answers)
```

**策略 4：Chain-of-Thought 验证**

```python
prompt = """
问题：{question}

请分步推理：
1. 列出已知条件
2. 列出推理步骤
3. 给出最终答案
4. 验证答案是否合理

如果发现矛盾，请指出并重新推理。
"""
```

**策略 5：External Verification（外部验证）**

```python
class VerifiedAnswer:
    def answer(self, question: str):
        # 1. LLM 生成答案
        answer = self.llm.generate(question)

        # 2. 提取可验证的事实
        claims = self.extract_claims(answer)

        # 3. 外部验证
        for claim in claims:
            if not self.verify_claim(claim):
                # 标记不可靠
                answer = self.add_warning(answer, claim)

        return answer

    def verify_claim(self, claim: str) -> bool:
        # 通过搜索引擎、数据库等验证
        search_results = search(claim)
        return check_consistency(claim, search_results)
```

---

## 1.3 Prompt Engineering 核心原则

### 原则 1：明确性（Clarity）

**❌ 模糊的 Prompt：**
```text
"帮我写个函数"
```

**✅ 明确的 Prompt：**
```text
请用 Python 写一个函数，功能如下：
- 函数名：calculate_discount
- 输入参数：
  - price: float (原价)
  - discount_rate: float (折扣率，0-1之间)
- 返回：float (折后价)
- 要求：
  - 参数验证（价格非负，折扣率在0-1之间）
  - 保留两位小数
  - 添加 docstring
```

**效果对比：**
- 模糊 Prompt：成功率 ~50%
- 明确 Prompt：成功率 ~95%

### 原则 2：结构化（Structure）

**❌ 无结构：**
```text
我想知道 Transformer 的工作原理以及它和 RNN 的区别还有它的优缺点
```

**✅ 结构化：**
```text
关于 Transformer 架构，请回答以下问题：

## 1. 工作原理
- Self-Attention 机制如何工作？
- Positional Encoding 的作用是什么？

## 2. 与 RNN 对比
- 主要区别是什么？
- 各自的优势场景？

## 3. 优缺点
- 优点（至少3个）
- 缺点（至少2个）

请用 Markdown 格式回答。
```

### 原则 3：示例驱动（Few-Shot Learning）

**Zero-Shot（无示例）：**
```text
将以下客户反馈分类：
"产品质量不错，但物流太慢了"
```

**Few-Shot（有示例）：**
```text
将客户反馈分类为：物流、产品质量、客服、价格

示例 1:
输入: "订单还没发货，已经等了5天"
分类: 物流

示例 2:
输入: "产品做工粗糙，不值这个价"
分类: 产品质量

示例 3:
输入: "客服态度很好，帮我解决了问题"
分类: 客服

现在分类：
输入: "产品质量不错，但物流太慢了"
分类: ?
```

**效果提升：**
- Zero-Shot: ~75% 准确率
- Few-Shot: ~90% 准确率

### 原则 4：约束条件（Constraints）

**❌ 无约束：**
```text
生成一篇关于 AI 的文章
```

**✅ 有约束：**
```text
生成一篇关于 AI 的技术文章，要求：

格式约束：
- 字数：800-1000字
- 结构：引言 + 3个小节 + 总结
- 使用 Markdown 格式

内容约束：
- 目标读者：有编程基础的工程师
- 深度：中级（不要太基础，不要太学术）
- 必须包含：实际代码示例

风格约束：
- 语言：中文
- 风格：技术准确，表达简洁
- 避免：营销话术、夸大其词
```

### 原则 5：输出格式（Output Format）

**❌ 自由格式：**
```text
提取这段文本中的关键信息
```

**✅ 指定格式：**
```text
从以下文本提取关键信息，返回 JSON 格式：

{
  "name": "人名",
  "email": "邮箱",
  "phone": "电话",
  "company": "公司名称"
}

文本：...
```

### Prompt 模板库

```python
# 模板 1：任务分解
TASK_DECOMPOSITION_TEMPLATE = """
任务：{task}

请将此任务分解为可执行的子任务：

1. 子任务 1
   - 输入：...
   - 输出：...
   - 验证标准：...

2. 子任务 2
   ...

最终输出：...
"""

# 模板 2：错误处理
ERROR_HANDLING_TEMPLATE = """
执行任务时发生错误：

任务：{task}
错误信息：{error}

请分析：
1. 错误原因是什么？
2. 如何修复？
3. 给出修复后的代码/方案

不要重复之前的错误。
"""

# 模板 3：Self-Critique（自我批评）
SELF_CRITIQUE_TEMPLATE = """
你刚才给出的答案是：
{previous_answer}

请批判性地审查这个答案：
1. 是否有事实错误？
2. 逻辑是否严密？
3. 是否遗漏重要信息？
4. 是否有更好的表达方式？

如果有问题，请给出改进后的答案。
"""
```

---

## 1.4 模型选择与权衡

### 主流模型对比（2026年）

| 模型 | 优势 | 劣势 | 适用场景 | 成本 |
|------|------|------|---------|------|
| **GPT-4 Turbo** | 推理能力强、工具调用准确 | 成本较高 | 复杂推理、Agent 系统 | $$ |
| **Claude 3.5 Sonnet** | 代码生成强、上下文长 | API 限速 | 代码生成、长文本处理 | $$ |
| **Gemini 1.5 Pro** | 多模态、上下文超长 | 推理稍弱 | 多模态任务、超长文档 | $ |
| **GPT-3.5 Turbo** | 速度快、成本低 | 能力有限 | 简单任务、高并发 | $ |
| **Llama 3.1 (70B)** | 可本地部署、无API成本 | 需要GPU资源 | 隐私敏感、高频调用 | 硬件成本 |

### 选择决策树

```text
是否需要本地部署？
├─ 是 → Llama 3.1 (70B/405B)
└─ 否 ↓

是否需要多模态？
├─ 是 → Gemini 1.5 Pro / GPT-4V
└─ 否 ↓

是否需要超长上下文（>100k tokens）？
├─ 是 → Claude 3.5 Sonnet / Gemini 1.5 Pro
└─ 否 ↓

任务复杂度？
├─ 高（复杂推理、Agent）→ GPT-4 Turbo / Claude 3.5
├─ 中（代码生成、总结）→ Claude 3.5 / GPT-4
└─ 低（分类、简单问答）→ GPT-3.5 Turbo
```

### 成本优化策略

**策略 1：模型分层（Model Tiering）**

```python
class AdaptiveModelRouter:
    """根据任务复杂度选择模型"""

    def __init__(self):
        self.models = {
            "simple": GPT35Turbo(),      # $0.0015/1K tokens
            "medium": Claude35Sonnet(),   # $0.015/1K tokens
            "complex": GPT4Turbo()        # $0.03/1K tokens
        }

    def route(self, task: str):
        complexity = self.classify_complexity(task)
        return self.models[complexity]

    def classify_complexity(self, task: str) -> str:
        """用便宜的模型分类任务复杂度"""
        prompt = f"评估任务复杂度（simple/medium/complex）：{task}"
        result = self.models["simple"].generate(prompt)
        return result
```

**策略 2：Prompt 缓存（Prompt Caching）**

```python
# Claude 3.5 支持 Prompt Caching
# 缓存长 System Prompt，节省 90% 的输入 token 成本

response = anthropic.messages.create(
    model="claude-3-5-sonnet-20241022",
    system=[
        {
            "type": "text",
            "text": LONG_SYSTEM_PROMPT,  # 缓存这部分
            "cache_control": {"type": "ephemeral"}
        }
    ],
    messages=[{"role": "user", "content": user_input}]
)
```

**策略 3：批量处理（Batch Processing）**

```python
# OpenAI Batch API - 50% 折扣
# 适用于非实时任务

batch_jobs = [
    {"custom_id": "task-1", "method": "POST", "url": "/v1/chat/completions", ...},
    {"custom_id": "task-2", ...},
    ...
]

# 提交批量任务
batch = client.batches.create(
    input_file_id=upload_batch_file(batch_jobs),
    endpoint="/v1/chat/completions",
    completion_window="24h"
)

# 24小时内完成，成本减半
```

---

## 1.5 上下文管理

### Token 限制

| 模型 | 上下文长度 | 输入成本 | 输出成本 |
|------|-----------|---------|---------|
| GPT-4 Turbo | 128k tokens | $0.01/1K | $0.03/1K |
| Claude 3.5 Sonnet | 200k tokens | $0.003/1K | $0.015/1K |
| Gemini 1.5 Pro | 1M tokens | $0.00125/1K | $0.005/1K |
| GPT-3.5 Turbo | 16k tokens | $0.0005/1K | $0.0015/1K |

### 上下文溢出问题

**问题场景：**

```python
# Agent 循环执行多次工具调用
context = ""
for i in range(10):
    context += f"Step {i}: {tool_result}\n"  # 累积上下文
    response = llm.generate(context)

# 问题：
# - 第10次迭代时，context 可能超过 token 限制
# - 早期步骤可能不再相关，但仍占用 token
```

### 解决策略

**策略 1：滑动窗口（Sliding Window）**

```python
class SlidingWindowMemory:
    """保留最近 N 条消息"""

    def __init__(self, max_messages: int = 10):
        self.messages = []
        self.max_messages = max_messages

    def add(self, message: dict):
        self.messages.append(message)
        if len(self.messages) > self.max_messages:
            # 保留 system message + 最近的消息
            system_msg = self.messages[0]  # 假设第一条是 system
            self.messages = [system_msg] + self.messages[-self.max_messages+1:]

    def get_context(self):
        return self.messages
```

**策略 2：摘要压缩（Summarization）**

```python
class SummarizingMemory:
    """定期压缩历史对话"""

    def __init__(self, llm, max_tokens: int = 4000):
        self.llm = llm
        self.messages = []
        self.max_tokens = max_tokens

    def add(self, message: dict):
        self.messages.append(message)

        # 检查 token 数量
        if self.estimate_tokens() > self.max_tokens:
            self.compress()

    def compress(self):
        """压缩旧消息"""
        # 保留最近 5 条完整消息
        recent = self.messages[-5:]

        # 压缩更早的消息
        old = self.messages[:-5]
        summary = self.llm.generate(
            f"总结以下对话，保留关键信息：\n{old}"
        )

        # 用摘要替换旧消息
        self.messages = [
            {"role": "system", "content": f"之前的对话摘要：{summary}"}
        ] + recent
```

**策略 3：相关性过滤（Relevance Filtering）**

```python
class RelevanceFilteredMemory:
    """只保留与当前问题相关的历史"""

    def get_relevant_context(self, current_query: str, history: List):
        """检索相关的历史消息"""
        relevant = []

        for msg in history:
            relevance_score = self.calculate_relevance(current_query, msg)
            if relevance_score > 0.7:
                relevant.append(msg)

        return relevant

    def calculate_relevance(self, query: str, message: dict) -> float:
        """计算相关性（简化版，实际可用 embedding）"""
        # 使用 LLM 评估相关性
        prompt = f"""
        问题：{query}
        历史消息：{message['content']}

        这条历史消息与当前问题的相关性（0-1）？
        只返回数字。
        """
        score = float(self.llm.generate(prompt))
        return score
```

---

## 1.6 质量保证与测试

### LLM 系统的测试策略

**1. 单元测试（固定输入输出）**

```python
def test_sentiment_analysis():
    """测试情感分析功能"""

    test_cases = [
        {
            "input": "这个产品太棒了！",
            "expected": "positive"
        },
        {
            "input": "质量很差，非常失望",
            "expected": "negative"
        },
        {
            "input": "还可以吧",
            "expected": "neutral"
        }
    ]

    for case in test_cases:
        result = sentiment_agent.analyze(case["input"])
        assert result == case["expected"], \
            f"Failed: {case['input']} -> {result} (expected {case['expected']})"
```

**2. 基于 LLM 的评估（Evaluation with LLM）**

```python
class LLMEvaluator:
    """用 LLM 评估 LLM 输出"""

    def evaluate_answer(self, question: str, answer: str, reference: str) -> dict:
        """评估答案质量"""

        prompt = f"""
        评估以下答案的质量：

        问题：{question}
        参考答案：{reference}
        待评估答案：{answer}

        评分标准（1-5分）：
        1. 准确性：答案是否正确？
        2. 完整性：是否覆盖所有要点？
        3. 简洁性：表达是否简洁清晰？

        返回 JSON：
        {{
          "accuracy": 1-5,
          "completeness": 1-5,
          "conciseness": 1-5,
          "overall": 1-5,
          "feedback": "具体反馈"
        }}
        """

        result = self.llm.generate(prompt)
        return parse_json(result)
```

**3. A/B 测试（在线评估）**

```python
class ABTestingFramework:
    """A/B 测试框架"""

    def __init__(self):
        self.model_a = GPT4()
        self.model_b = Claude35()
        self.results = []

    def route_request(self, user_id: int, query: str):
        """随机分配用户到不同模型"""

        if hash(user_id) % 2 == 0:
            model, variant = self.model_a, "A"
        else:
            model, variant = self.model_b, "B"

        start_time = time.time()
        response = model.generate(query)
        latency = time.time() - start_time

        # 记录结果
        self.results.append({
            "variant": variant,
            "latency": latency,
            "response": response,
            "user_id": user_id
        })

        return response

    def analyze_results(self):
        """分析 A/B 测试结果"""
        a_results = [r for r in self.results if r["variant"] == "A"]
        b_results = [r for r in self.results if r["variant"] == "B"]

        return {
            "A": {
                "avg_latency": np.mean([r["latency"] for r in a_results]),
                "count": len(a_results)
            },
            "B": {
                "avg_latency": np.mean([r["latency"] for r in b_results]),
                "count": len(b_results)
            }
        }
```

---

## 1.7 生产环境最佳实践

### 1. 错误处理

```python
class RobustLLMClient:
    """健壮的 LLM 客户端"""

    def __init__(self, llm, max_retries: int = 3):
        self.llm = llm
        self.max_retries = max_retries

    async def generate(self, prompt: str, **kwargs):
        """带重试的生成"""

        for attempt in range(self.max_retries):
            try:
                response = await self.llm.generate(prompt, **kwargs)
                return response

            except RateLimitError as e:
                # 速率限制：指数退避
                wait_time = 2 ** attempt
                logger.warning(f"Rate limited, retry in {wait_time}s")
                await asyncio.sleep(wait_time)

            except TimeoutError as e:
                # 超时：重试
                logger.warning(f"Timeout on attempt {attempt+1}")
                if attempt == self.max_retries - 1:
                    raise

            except InvalidRequestError as e:
                # 无效请求：不重试
                logger.error(f"Invalid request: {e}")
                raise

        raise Exception(f"Failed after {self.max_retries} retries")
```

### 2. 监控与日志

```python
class MonitoredLLMClient:
    """带监控的 LLM 客户端"""

    async def generate(self, prompt: str, **kwargs):
        start_time = time.time()

        try:
            response = await self.llm.generate(prompt, **kwargs)

            # 记录成功指标
            self.metrics.record({
                "latency": time.time() - start_time,
                "input_tokens": self.count_tokens(prompt),
                "output_tokens": self.count_tokens(response),
                "model": self.llm.model_name,
                "status": "success"
            })

            return response

        except Exception as e:
            # 记录失败
            self.metrics.record({
                "latency": time.time() - start_time,
                "model": self.llm.model_name,
                "status": "error",
                "error_type": type(e).__name__
            })
            raise
```

### 3. 成本控制

```python
class CostControlledClient:
    """成本控制的 LLM 客户端"""

    def __init__(self, llm, budget_per_day: float):
        self.llm = llm
        self.budget_per_day = budget_per_day
        self.today_cost = 0
        self.last_reset = date.today()

    async def generate(self, prompt: str, **kwargs):
        # 检查预算
        self.check_budget()

        # 估算成本
        estimated_cost = self.estimate_cost(prompt, kwargs.get("max_tokens", 1000))

        if self.today_cost + estimated_cost > self.budget_per_day:
            raise BudgetExceededError(
                f"Daily budget ${self.budget_per_day} exceeded"
            )

        # 生成
        response = await self.llm.generate(prompt, **kwargs)

        # 更新成本
        actual_cost = self.calculate_cost(prompt, response)
        self.today_cost += actual_cost

        return response

    def check_budget(self):
        """重置每日预算"""
        if date.today() > self.last_reset:
            self.today_cost = 0
            self.last_reset = date.today()
```

---

## 本章小结

### 核心要点回顾

**1. LLM 能力边界**
- 擅长：模式识别、文本生成、代码生成
- 不擅长：精确计算、实时信息、长期记忆
- 核心：基于训练数据的模式匹配和序列生成

**2. 幻觉问题**
- 类型：事实性、逻辑性、引用性幻觉
- 应对：工具调用、RAG、Self-Consistency、外部验证

**3. Prompt Engineering**
- 明确性：详细的任务描述和要求
- 结构化：清晰的输入输出格式
- 示例驱动：Few-Shot Learning 提升准确率
- 约束条件：格式、内容、风格的明确要求

**4. 模型选择**
- 根据任务复杂度选择合适的模型
- 模型分层降低成本
- Prompt 缓存和批量处理优化

**5. 上下文管理**
- 滑动窗口：保留最近消息
- 摘要压缩：压缩历史对话
- 相关性过滤：只保留相关信息

**6. 质量保证**
- 单元测试：固定输入输出
- LLM 评估：用 LLM 评估 LLM
- A/B 测试：在线对比不同模型

**7. 生产最佳实践**
- 错误处理：重试、退避、降级
- 监控日志：性能、成本、错误追踪
- 成本控制：预算管理、成本估算

### 关键洞察

> **成功的 Agent 系统建立在对 LLM 能力边界的深刻理解之上。不是让 LLM 做所有事情，而是让它做它擅长的事情，其余交给传统工程方法。**

### 下一章预告

下一章我们将进入 **Prompt Engineering 与结构化输出**：学习如何把人的意图整理成模型可执行、系统可验证的任务协议。

---

## 参考资料

1. **Language Models are Few-Shot Learners** - Brown et al., GPT-3 论文
2. **Chain-of-Thought Prompting** - Wei et al., 2022
3. **Retrieval-Augmented Generation** - Lewis et al., 2020
4. **Constitutional AI** - Anthropic, 2022
5. **OpenAI Best Practices** - https://platform.openai.com/docs/guides/prompt-engineering
6. **Anthropic Prompt Engineering Guide** - https://docs.anthropic.com/claude/docs/prompt-engineering
