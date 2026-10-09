'use strict';

const assert = require('node:assert/strict');
const fs = require('node:fs');
const path = require('node:path');
const { test } = require('node:test');

const repoRoot = path.resolve(__dirname, '..');
const bookRoot = path.join(
  repoRoot,
  'books',
  'reliable-system-design',
  'src'
);
const archiveRoot = path.join(repoRoot, 'books', 'reliable-system-design', 'archive', 'questionbanks');
const questionBankFiles = [
  'system-design-questionbank.md',
  'interview-basic-question-bank.md',
  'system-design-interview-50.md',
];

function read(relativePath) {
  return fs.readFileSync(path.join(bookRoot, relativePath), 'utf8');
}

function readArchive(filename) {
  return fs.readFileSync(path.join(archiveRoot, filename), 'utf8');
}

test('question banks are archived outside the active manuscript', () => {
  const summary = read('SUMMARY.md');
  const introduction = read('README.md');

  for (const filename of questionBankFiles) {
    assert.equal(fs.existsSync(path.join(bookRoot, 'appendix', filename)), false,
      `active question-bank source remains: ${filename}`);
    assert.equal(fs.existsSync(path.join(archiveRoot, filename)), true,
      `archived question bank missing: ${filename}`);
    assert.ok(!summary.includes(filename));
    assert.ok(!introduction.includes(filename));
  }
  assert.ok(!summary.includes('附录 D'));
  assert.ok(!introduction.includes('附录面试题库'));
});

test('old question-bank URLs redirect to the book homepage', () => {
  const config = fs.readFileSync(path.join(repoRoot, 'books', 'reliable-system-design', 'book.toml'), 'utf8');
  assert.ok(config.includes('[output.html.redirect]'));
  for (const filename of questionBankFiles) {
    assert.ok(config.includes(`"/appendix/${filename.replace('.md', '.html')}" = "../index.html"`));
  }
});

test('archived system-design bank retains every legacy question ID', () => {
  const summary = read('SUMMARY.md');
  assert.ok(fs.existsSync(path.join(archiveRoot, 'system-design-questionbank.md')));
  const questionBank = readArchive('system-design-questionbank.md');

  assert.doesNotMatch(summary, /# 第四部分：系统设计题库/);
  assert.match(summary, /# 第二部分：电商系统设计实战/);

  const coverageIndex = questionBank
    .split('## 历史详细参考材料', 1)[0]
    .match(/## 迁移覆盖索引[\s\S]*/)?.[0] || '';
  const coverageIds = coverageIndex.match(
    /\|\sQ-(?:GEN|ECOM)-[A-Z]+-\d+\s\|/g
  ) || [];

  assert.equal(coverageIds.length, 220);
  assert.match(questionBank, /## 迁移覆盖索引/);
  assert.match(questionBank, /## 历史详细参考材料/);
  assert.match(questionBank, /## 范围、约束与容量/);
  assert.match(questionBank, /## 状态、一致性与恢复/);
  assert.match(questionBank, /## 综合设计案例/);
  assert.ok(questionBank.includes('## 训练材料与阅读入口'));
  assert.ok(questionBank.includes('<a id="原有训练方法保留与覆盖审计"></a>'));
  assert.match(questionBank, /### 系统设计面试在考什么/);
  assert.match(questionBank, /### 回答与白板的展开顺序/);
  assert.match(questionBank, /### 容量估算口径/);
  assert.match(questionBank, /### 存储、中间件与可靠性的答题边界/);
});

test('question types have requirements appropriate to their learning objective', () => {
  assert.ok(fs.existsSync(path.join(archiveRoot, 'system-design-questionbank.md')));
  const questionBank = readArchive('system-design-questionbank.md');

  assert.match(questionBank, /核心案例/);
  assert.match(questionBank, /约束变体/);
  assert.match(questionBank, /快问快答/);
  assert.match(questionBank, /关键决策/);
  assert.match(questionBank, /故障注入/);
  assert.match(questionBank, /适用边界/);
  assert.match(questionBank, /30 分钟：通用后端系统设计/);
  assert.match(questionBank, /45 分钟：电商专项系统设计/);
  assert.match(questionBank, /60 分钟：高级系统设计答辩/);
});

test('active chapters no longer depend on archived question banks', () => {
  const activeFiles = [
    'README.md',
    'SUMMARY.md',
    ...fs.readdirSync(path.join(bookRoot, 'part01'))
      .filter((file) => file.endsWith('.md')).map((file) => `part01/${file}`),
    'part02/10-ecommerce-overview.md',
    'part02/11-product-center-supply-lifecycle.md',
    'part02/12-inventory-system.md',
    'part02/13-marketing-pricing-system.md',
    'part02/14-ecommerce-customer-lifecycle.md',
    'part02/15-ecommerce-order-fulfillment.md',
  ];

  for (const file of activeFiles) {
    const content = read(file);
    assert.doesNotMatch(content, /第 38 章的相应核心案例/);
    assert.doesNotMatch(content, /\.\.\/part04\//);
    for (const filename of questionBankFiles) {
      assert.ok(!content.includes(filename), `${file} still links to ${filename}`);
    }
  }
});

test('message-delivery answer distinguishes Kafka transactions from database effects', () => {
  assert.ok(fs.existsSync(path.join(archiveRoot, 'system-design-questionbank.md')));
  const answer = readArchive('system-design-questionbank.md')
    .split('### Q-ECOM-ARCH-013：')[1]
    .split('### Q-ECOM-ARCH-014：')[0];

  assert.ok(answer.includes('sendOffsetsToTransaction'));
  assert.ok(answer.includes('唯一处理记录与业务更新在同一事务'));
  assert.ok(answer.includes('数据库提交后、offset 确认前崩溃'));
  assert.ok(answer.includes('`abortTransaction()` 不会回滚外部数据库'));
  assert.ok(!answer.includes('db.update(...)'));
  assert.ok(!answer.includes('acks=all              // 所有副本确认'));
});

test('hot-inventory answer establishes authoritative reservations before promises', () => {
  assert.ok(fs.existsSync(path.join(archiveRoot, 'system-design-questionbank.md')));
  const answer = readArchive('system-design-questionbank.md')
    .split('### Q-ECOM-SUPPLY-019：')[1]
    .split('### Q-ECOM-SUPPLY-020：')[0];

  assert.ok(answer.includes('Redis 准入 + 权威预占'));
  assert.ok(answer.includes('只有拿到已提交的 reservation_id'));
  assert.ok(answer.includes('数据库还是 100 件，Redis 已扣减 20 件'));
  assert.ok(answer.includes('未知写入先查询'));
  assert.ok(!answer.includes('采用**Redis原子操作+异步同步DB**'));
});

test('search recovery belongs to chapter 14 and transaction ADR links back to it', () => {
  const discovery = read('part02/14-ecommerce-customer-lifecycle.md');
  const order = read('part02/15-ecommerce-order-fulfillment.md');

  assert.ok(discovery.includes('<a id="搜索读链路的故障矩阵"></a>'));
  assert.ok(discovery.includes('| 重建索引失败 |'));
  assert.ok(!order.includes('| 重建索引失败 |'));
  assert.ok(order.includes('14-ecommerce-customer-lifecycle.md#搜索读链路的故障矩阵'));
  assert.ok(order.includes('15.6.3 ADR-02：订单保存不可变交易快照'));
  assert.ok(!order.includes('**坚决不选 2PC / XA**'));
  assert.ok(!order.includes('**只有金融级资管'));
});
