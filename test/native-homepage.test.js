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
  assert.equal(packageJson.dependencies['hexo-generator-index'], undefined);
  assert.doesNotMatch(config, /^index_generator:/m);
  const links = [
    '/archive/AI/07-ai-agent-development-interview-50.html',
    '/2026/09/11/other/08-agent-electronic-commerce-research-report/',
    '/ai-book/',
    '/system-design-primer/',
    '/archives/',
    '/categories/',
    '/books/system-design-primer/appendix/system-design-interview-50.html',
    '/2026/04/03/AI/00-vibe-coding-vs-spec-coding/',
    '/2026/09/23/system-design/45-ddd-principles-and-pricing-practice/',
    '/ai-book/part2/01-agent-architecture.html',
    '/2026/05/08/other/ai-content-to-video-open-source-workflow/',
    '/2026/04/05/AI/02-karpathy-evolving-knowledge-base/'
  ];
  for (const link of links) assert.ok(homepage.includes(link), `missing ${link}`);
});

test('homepage distinguishes public projects from private exploration', () => {
  for (const link of [
    'https://github.com/wxquare/leetcode-primer',
    'https://github.com/wxquare/wxquare.github.io',
    'https://github.com/wxquare/leetcode-primer/blob/master/README.md'
  ]) assert.ok(homepage.includes(link), `missing public link ${link}`);

  for (const [name, repository] of [
    ['SkillForge', 'skillforge'],
    ['Investment Assistant', 'investment-assistant'],
    ['wxquare-private', 'wxquare-private']
  ]) {
    assert.ok(homepage.includes(name));
    assert.ok(!homepage.includes(`https://github.com/wxquare/${repository}`));
  }
});
