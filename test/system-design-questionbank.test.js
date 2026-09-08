'use strict';

const assert = require('node:assert/strict');
const fs = require('node:fs');
const path = require('node:path');
const { test } = require('node:test');

const repoRoot = path.resolve(__dirname, '..');
const bookRoot = path.join(
  repoRoot,
  'books',
  'system-design-primer',
  'src'
);

function read(relativePath) {
  return fs.readFileSync(path.join(bookRoot, relativePath), 'utf8');
}

test('question bank is a single appendix and retains every legacy question ID', () => {
  const summary = read('SUMMARY.md');
  const questionBank = read('appendix/system-design-questionbank.md');

  assert.match(
    summary,
    /附录 D 系统设计题库.*appendix\/system-design-questionbank\.md/
  );
  assert.doesNotMatch(summary, /# 第四部分：系统设计题库/);
  assert.match(summary, /# 第三部分：电商系统设计实战/);

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
  assert.match(questionBank, /## 原有训练方法保留与覆盖审计/);
  assert.match(questionBank, /### 系统设计面试在考什么/);
  assert.match(questionBank, /### 回答与白板的展开顺序/);
  assert.match(questionBank, /### 容量估算口径/);
  assert.match(questionBank, /### 存储、中间件与可靠性的答题边界/);
});

test('question types have requirements appropriate to their learning objective', () => {
  const questionBank = read('appendix/system-design-questionbank.md');

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

test('book links point to the appendix instead of a removed question-bank chapter', () => {
  const partThreeFiles = [
    'part03/01-ecommerce-overview.md',
    'part03/02-product-center.md',
    'part03/03-product-supply-lifecycle-ops.md',
    'part03/04-inventory-system.md',
    'part03/07-search-discovery.md',
    'part03/08-cart-checkout.md',
    'part03/09-order-system.md',
    'part03/10-payment-system.md',
    'part03/11-b2b2c-platform-architecture.md',
  ];

  for (const file of partThreeFiles) {
    const content = read(file);
    assert.doesNotMatch(content, /第 38 章的相应核心案例/);
    assert.doesNotMatch(content, /\.\.\/part04\//);
  }
});
