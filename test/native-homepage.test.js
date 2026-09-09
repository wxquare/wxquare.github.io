const assert = require('node:assert/strict');
const fs = require('node:fs');
const path = require('node:path');
const test = require('node:test');

const homepage = fs.readFileSync(path.join(__dirname, '..', 'source/home/index.md'), 'utf8');

test('native homepage contains the curated reader entry points', () => {
  assert.match(homepage, /^layout: page$/m);
  assert.match(homepage, /^permalink: \/$/m);
  const links = [
    '/ai-book/',
    '/system-design-primer/',
    '/archives/',
    '/categories/',
    '/2026/04/07/system-design/08-system-design-interview/',
    '/2026/04/03/AI/00-vibe-coding-vs-spec-coding/',
    '/2026/04/03/system-design/43-acc-ddd-notes/',
    '/2026/04/03/AI/02-agent-system-design-guid/',
    '/2026/05/08/other/ai-content-to-video-open-source-workflow/',
    '/2026/06/09/system-design/34-ecommerce-long-transactions/',
    '/2026/04/16/system-design/33-ecommerce-price-calendar/',
    '/2026/04/10/system-design/30-ecommerce-product-lifecycle-management/',
    '/2026/04/07/AI/06-harness-engineering/',
    '/2026/04/05/AI/04-karpathy-evolving-knowledge-base/'
  ];
  for (const link of links) assert.ok(homepage.includes(link), `missing ${link}`);
});
