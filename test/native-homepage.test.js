const assert = require('node:assert/strict');
const fs = require('node:fs');
const path = require('node:path');
const test = require('node:test');

const homepage = fs.readFileSync(path.join(__dirname, '..', 'source/home/index.md'), 'utf8');
const packageJson = JSON.parse(fs.readFileSync(path.join(__dirname, '..', 'package.json'), 'utf8'));
const config = fs.readFileSync(path.join(__dirname, '..', '_config.yml'), 'utf8');

test('native homepage contains the curated reader entry points', () => {
  assert.match(homepage, /^layout: page$/m);
  assert.match(homepage, /^permalink: \/$/m);
  assert.match(homepage, /^toc:\n  enable: false$/m);
  assert.match(homepage, /AI Engineering：大模型与智能体系统工程/);
  assert.doesNotMatch(homepage, /AI Engineering 阅读入口/);
  assert.doesNotMatch(homepage, /AI Agent 系统设计内容已整合至书稿/);
  assert.equal(packageJson.dependencies['hexo-generator-index'], undefined);
  assert.doesNotMatch(config, /^index_generator:/m);
  const links = [
    '/ai-book/',
    '/reliable-system-design/',
    '/archives/',
    '/categories/'
  ];
  for (const link of links) assert.ok(homepage.includes(link), `missing ${link}`);
  for (const filename of [
    'system-design-interview-50.html',
    'system-design-questionbank.html',
    'interview-basic-question-bank.html',
  ]) {
    assert.ok(!homepage.includes(`/reliable-system-design/appendix/${filename}`));
  }
});

test('homepage omits interview and question-bank entry points', () => {
  assert.doesNotMatch(homepage, /面试准备与题库|LeetCode 500 精选题单|AI 与 Agent 开发高频 50 题/);
  assert.doesNotMatch(homepage, /leetcode-primer|ai-agent-development-interview-50\.html/);
});

test('featured posts inherit article titles through canonical post links', () => {
  const references = [...homepage.matchAll(/{%\s*post_link\s+(\S+)\s*%}/g)]
    .map((match) => match[1]);
  assert.ok(references.length > 0, 'featured posts must use post_link without title overrides');
  assert.equal(
    [...homepage.matchAll(/{%\s*post_link\b[^%]*%}/g)].length,
    references.length,
    'featured post links must not override article titles'
  );
  assert.equal(new Set(references).size, references.length, 'featured posts must be unique');
  for (const slug of references) {
    const filename = path.join(__dirname, '..', 'source/_posts', `${slug}.md`);
    assert.ok(fs.existsSync(filename), `missing canonical post: ${slug}`);
  }
  assert.doesNotMatch(homepage, /\]\(\/20\d{2}\/\d{2}\/\d{2}\//);
  for (const slug of [
    'AI/01-vibe-coding-vs-spec-coding',
    'system-design/04-ddd-pricing-practice',
    'other/03-ecommerce-agent-research-report'
  ]) {
    assert.ok(references.includes(slug), `missing featured post: ${slug}`);
  }
});

test('featured topics include delivery review and AI model experiments', () => {
  for (const slug of [
    'system-design/05-architecture-code-review-checklist',
    'AI/05-video-object-tracking-practice',
    'AI/07-tensorflow-model-optimization',
    'AI/08-tvm-operator-optimization-practice'
  ]) {
    assert.ok(homepage.includes(`{% post_link ${slug} %}`), `missing featured post: ${slug}`);
  }
});

test('featured blogs precede manuscripts and omit the removed topic groups', () => {
  const featuredIndex = homepage.indexOf('## 精选博客');
  const manuscriptsIndex = homepage.indexOf('## 持续编写的书稿');
  assert.ok(featuredIndex >= 0 && manuscriptsIndex > featuredIndex);
  for (const heading of [
    '### 数据系统与生产排障',
    '### 知识管理与内容生产',
    '### 模型优化与视觉算法'
  ]) {
    assert.ok(!homepage.includes(heading), `removed section: ${heading}`);
  }
  for (const slug of [
    'fundamentals/08-redis-principles-and-practice',
    'fundamentals/09-kafka-messaging-practice',
    'AI/04-karpathy-evolving-knowledge-base',
    'other/02-ai-video-production-workflow'
  ]) {
    assert.ok(!homepage.includes(`{% post_link ${slug} %}`), `removed featured post: ${slug}`);
  }
});
