'use strict';

const assert = require('node:assert/strict');
const fs = require('node:fs');
const path = require('node:path');
const test = require('node:test');

const root = path.resolve(__dirname, '..');
const booklist = fs.readFileSync(
  path.join(root, 'source/booklist/index.md'),
  'utf8'
);

test('booklist presents curated routes and verified destinations', () => {
  assert.match(booklist, /把零散经验沉淀成可复用的知识，让未来的自己少走一遍已经走过的弯路。/);

  for (const heading of [
    '互联网系统设计',
    'AI 与 Agent',
    '思维、商业与人文'
  ]) {
    assert.match(booklist, new RegExp(`^## ${heading}$`, 'm'));
  }

  assert.match(booklist, /^## 正在维护的书籍$/m);
  assert.match(booklist, /\]\(\/system-design-primer\/\)/);
  assert.match(booklist, /\]\(\/ai-book\/\)/);
  assert.doesNotMatch(booklist, /\]\(\/library\/\)/);
  assert.match(booklist, /https:\/\/www\.oreilly\.com\//);
  assert.match(booklist, /AI-Agents-in-Depth-zh-CN\.pdf/);
  assert.doesNotMatch(booklist, /github\.com\/backstudy\/bookrack/);
});

test('booklist stays within the native NexT reading style', () => {
  assert.doesNotMatch(booklist, /<style[\s>]/i);
  assert.doesNotMatch(booklist, /class="booklist/);
  assert.doesNotMatch(booklist, /CURATED READING/);
  assert.match(booklist, /^## 互联网系统设计$/m);
});
