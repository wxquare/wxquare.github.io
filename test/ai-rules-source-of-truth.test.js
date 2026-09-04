'use strict';

const assert = require('node:assert/strict');
const fs = require('node:fs');
const path = require('node:path');
const { test } = require('node:test');

const repoRoot = path.resolve(__dirname, '..');
const sourceOfTruth = path.join(repoRoot, 'AGENTS.md');

const adapters = [
  '.cursorrules',
  '.agents/cursor/rules/ai-directory.md',
  '.agents/cursor/rules/hexo-blog.md',
  'CONTRIBUTING.md'
];

const executionEntries = [
  '.agents/commands/stats.md',
  '.agents/skills/new-post/SKILL.md',
  '.agents/skills/organize-posts/SKILL.md',
  '.agents/skills/review-post/SKILL.md',
  '.agents/skills/link-check/SKILL.md',
  '.agents/skills/generate-summary/SKILL.md'
];

function read(relativePath) {
  return fs.readFileSync(path.join(repoRoot, relativePath), 'utf8');
}

test('AGENTS.md owns shared writing policy', () => {
  const content = read('AGENTS.md');

  for (const policy of [
    '博客主分类规范',
    '教程',
    '标签',
    '图片路径'
  ]) {
    assert.match(content, new RegExp(policy), `AGENTS.md must define ${policy}`);
  }
});

test('non-normative entry points only route tools to AGENTS.md', () => {
  for (const file of adapters) {
    const content = read(file);

    assert.match(content, /AGENTS\.md/, `${file} must link to AGENTS.md`);
    for (const duplicatedPolicy of [
      /AI\/system-design\/fundamentals\/other/,
      /标签.*小写/,
      /图片路径.*相对路径/,
      /绝对路径.*图片/,
      /教程类文章/
    ]) {
      assert.doesNotMatch(content, duplicatedPolicy, `${file} must not copy ${duplicatedPolicy}`);
    }
  }
});

test('post workflow entries keep execution details and reference shared policy', () => {
  for (const file of executionEntries) {
    const content = read(file);

    assert.match(content, /AGENTS\.md/, `${file} must link to AGENTS.md`);
    assert.doesNotMatch(
      content,
      /分类（.*AI\/system-design\/fundamentals\/other）/,
      `${file} must not own the category list`
    );
  }
});
