'use strict';

const assert = require('node:assert/strict');
const test = require('node:test');

const {
  buildHomepageGroups,
  renderHomepageMarkup
} = require('../scripts/categorized-homepage');

const registry = {
  categories: [
    { slug: 'AI', label: 'AI 与 Agent', frontMatterLabels: ['AI 与 Agent', 'AI'] },
    { slug: 'system-design', label: '系统设计基础', frontMatterLabels: ['系统设计基础'] },
    { slug: 'fundamentals', label: '计算机基础', frontMatterLabels: ['计算机基础'] },
    { slug: 'other', label: 'other', frontMatterLabels: ['other'] }
  ]
};

const collection = (...items) => ({ toArray: () => items });
const categoryArchives = collection(
  { name: 'AI 与 Agent', path: 'categories/AI-与-Agent/' },
  { name: '系统设计基础', path: 'categories/系统设计基础/' },
  { name: '计算机基础', path: 'categories/计算机基础/' },
  { name: 'other', path: 'categories/other/' }
);
const post = (title, date, category, path) => ({
  title,
  date: new Date(date),
  path,
  categories: collection({ name: category }),
  tags: collection({ name: 'agent' })
});

test('buildHomepageGroups keeps registry order, sorts posts, and limits every group', () => {
  const posts = [
    post('older AI', '2026-01-01', 'AI', 'older-ai/'),
    post('newer AI', '2026-02-01', 'AI 与 Agent', 'newer-ai/'),
    post('system post', '2026-03-01', '系统设计基础', 'system/'),
    ...Array.from({ length: 5 }, (_, index) =>
      post(`extra AI ${index}`, `2026-04-0${index + 1}`, 'AI', `extra-${index}/`))
  ];

  const groups = buildHomepageGroups(posts, registry, 5, () => {}, categoryArchives);

  assert.deepEqual(groups.map(group => group.slug), ['AI', 'system-design', 'fundamentals', 'other']);
  assert.equal(groups[0].posts.length, 5);
  assert.equal(groups[0].posts[0].title, 'extra AI 4');
  assert.equal(groups[1].posts[0].title, 'system post');
  assert.equal(groups[2].posts.length, 0);
});

test('renderHomepageMarkup escapes untrusted post text and emits archive and book links', () => {
  const groups = [{
    slug: 'AI',
    label: 'AI <Agent>',
    archivePath: 'categories/AI-与-Agent/',
    posts: [{
      title: '<script>alert(1)</script>',
      path: '2026/01/01/test/?x=1&y=2',
      date: '2026-01-01',
      tags: ['a&b']
    }]
  }];

  const html = renderHomepageMarkup(groups);

  assert.match(html, /AI &lt;Agent&gt;/);
  assert.match(html, /&lt;script&gt;alert\(1\)&lt;\/script&gt;/);
  assert.match(html, /a&amp;b/);
  assert.match(html, /\/categories\/AI-与-Agent\//);
  assert.match(html, /\/ai-book\//);
  assert.match(html, /\/system-design-architecture-book\//);
});

test('buildHomepageGroups reports an unmapped categorized post', () => {
  const warnings = [];
  buildHomepageGroups(
    [post('unknown', '2026-01-01', '未注册分类', 'unknown/')],
    registry,
    5,
    (candidate, names) => warnings.push([candidate.title, names]),
    categoryArchives
  );

  assert.deepEqual(warnings, [['unknown', ['未注册分类']]]);
});
