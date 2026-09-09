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
  for (const heading of [
    '系统设计与数据系统',
    '软件设计与代码质量',
    '编程语言与基础',
    'AI 与 Agent',
    '思维、商业与人文'
  ]) {
    assert.match(booklist, new RegExp(`<h2[^>]*>${heading}</h2>`));
  }

  assert.match(booklist, /href="\/system-design-primer\/"/);
  assert.match(booklist, /href="\/ai-book\/"/);
  assert.match(booklist, /href="\/library\/"/);
  assert.match(booklist, /https:\/\/www\.oreilly\.com\//);
  assert.doesNotMatch(booklist, /github\.com\/backstudy\/bookrack/);
});

test('booklist defines scoped responsive card styles', () => {
  assert.match(booklist, /\.booklist\s*\{/);
  assert.match(booklist, /\.booklist__grid\s*\{/);
  assert.match(booklist, /@media\s*\(max-width:\s*640px\)/);
});
