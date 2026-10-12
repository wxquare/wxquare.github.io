'use strict';

const assert = require('node:assert/strict');
const fs = require('node:fs');
const path = require('node:path');
const { test } = require('node:test');
const Hexo = require('hexo');
const frontMatter = require('hexo-front-matter');
const { url_for } = require('hexo-util');

const repoRoot = path.resolve(__dirname, '..');
const postsRoot = path.join(repoRoot, 'source/_posts');

// Explicit published paths follow the current category-local filenames.
const publishedPaths = new Map([
  ['AI/01-vibe-coding-vs-spec-coding', '2026/04/03/AI/01-vibe-coding-vs-spec-coding/'],
  ['AI/02-ai-agent-workflow-practice', '2026/09/23/AI/02-ai-agent-workflow-practice/'],
  ['AI/03-superpowers-source-workflow-practice', '2026/09/24/AI/03-superpowers-source-workflow-practice/'],
  ['AI/04-karpathy-evolving-knowledge-base', '2026/04/05/AI/04-karpathy-evolving-knowledge-base/'],
  ['AI/05-video-object-tracking-practice', '2020/08/13/AI/05-video-object-tracking-practice/'],
  ['AI/06-opencl-mobile-performance-testing', '2020/08/13/AI/06-opencl-mobile-performance-testing/'],
  ['AI/07-tensorflow-model-optimization', '2026/09/22/AI/07-tensorflow-model-optimization/'],
  ['AI/08-tvm-operator-optimization-practice', '2026/09/22/AI/08-tvm-operator-optimization-practice/'],
  ['fundamentals/01-operating-system-fundamentals', '2024/03/01/fundamentals/01-operating-system-fundamentals/'],
  ['fundamentals/02-computer-network-fundamentals', '2024/03/02/fundamentals/02-computer-network-fundamentals/'],
  ['fundamentals/03-bash-shell-practice', '2024/03/03/fundamentals/03-bash-shell-practice/'],
  ['fundamentals/04-python-practice', '2024/03/04/fundamentals/04-python-practice/'],
  ['fundamentals/05-cpp-practice', '2024/03/05/fundamentals/05-cpp-practice/'],
  ['fundamentals/06-go-practice', '2024/03/06/fundamentals/06-go-practice/'],
  ['fundamentals/07-mysql-database-practice', '2024/03/04/fundamentals/07-mysql-database-practice/'],
  ['fundamentals/08-redis-principles-and-practice', '2024/03/06/fundamentals/08-redis-principles-and-practice/'],
  ['fundamentals/09-kafka-messaging-practice', '2024/03/10/fundamentals/09-kafka-messaging-practice/'],
  ['fundamentals/10-elasticsearch-search-practice', '2024/03/07/fundamentals/10-elasticsearch-search-practice/'],
  ['fundamentals/11-docker-kubernetes-practice', '2024/12/20/fundamentals/11-docker-kubernetes-practice/'],
  ['system-design/01-system-design-methodology', '2025/04/01/system-design/01-system-design-methodology/'],
  ['system-design/02-clean-architecture-ddd-cqrs', '2026/04/01/system-design/02-clean-architecture-ddd-cqrs/'],
  ['system-design/03-clean-code-practice', '2026/04/02/system-design/03-clean-code-practice/'],
  ['system-design/04-ddd-pricing-practice', '2026/09/23/system-design/04-ddd-pricing-practice/'],
  ['system-design/05-architecture-code-review-checklist', '2026/04/04/system-design/05-architecture-code-review-checklist/'],
  ['system-design/06-system-reliability-engineering', '2025/05/15/system-design/06-system-reliability-engineering/'],
  ['system-design/07-engineering-diagrams-ai-workflow', '2026/10/10/system-design/07-engineering-diagrams-ai-workflow/'],
  ['system-design/08-ddia-reading-data-systems', '2026/10/12/system-design/08-ddia-reading-data-systems/'],
  ['other/01-hexo-github-pages-blog-setup', '2023/08/13/other/01-hexo-github-pages-blog-setup/'],
  ['other/02-ai-video-production-workflow', '2026/05/08/other/02-ai-video-production-workflow/'],
  ['other/03-ecommerce-agent-research-report', '2026/09/11/other/03-ecommerce-agent-research-report/'],
  ['other/04-future-predictions-career-opportunities', '2026/10/12/other/04-future-predictions-career-opportunities/']
]);

test('post filenames use consecutive category-local reading-order numbers', () => {
  for (const category of fs.readdirSync(postsRoot)) {
    const files = fs.readdirSync(path.join(postsRoot, category)).sort();
    files.forEach((name, index) => {
      assert.match(name, /^\d{2}-[a-z0-9]+(?:-[a-z0-9]+)+\.md$/, `${category}/${name}`);
      assert.equal(Number(name.slice(0, 2)), index + 1, `${category}/${name}: nonconsecutive number`);
    });
  }
});

test('posts resolve to their intended explicit published URLs', async () => {
  const hexo = new Hexo(repoRoot, { silent: true });
  await hexo.init();
  await hexo.load();
  const posts = hexo.locals.get('posts').toArray();

  assert.equal(posts.length, publishedPaths.size);
  assert.equal(new Set(posts.map((post) => post.path)).size, posts.length, 'duplicate published URLs');
  for (const [slug, publishedPath] of publishedPaths) {
    const post = posts.find((candidate) => candidate.slug === slug);
    assert.ok(post, `missing renamed post: ${slug}`);
    const expectedPath = `/${post.date.format('YYYY/MM/DD')}/${slug}/`;
    const metadata = frontMatter.parse(fs.readFileSync(path.join(repoRoot, 'source', post.source), 'utf8'));
    assert.equal(metadata.permalink, expectedPath, `${slug}: explicit permalink does not match filename`);
    assert.equal(url_for.call(hexo, post.path), expectedPath, `${slug}: generated URL does not match filename`);
    assert.equal(url_for.call(hexo, post.path), `/${publishedPath}`, `${slug}: published URL changed`);
  }
});
