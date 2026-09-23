const test = require('node:test');
const assert = require('node:assert/strict');
const fs = require('node:fs');
const path = require('node:path');

const postsDir = path.join(__dirname, '..', 'source', '_posts', 'fundamentals');

function postFiles() {
  return fs.readdirSync(postsDir)
    .filter((name) => name.endsWith('.md'))
    .sort()
    .map((name) => ({ name, path: path.join(postsDir, name) }));
}

function frontMatter(source) {
  const match = source.match(/^---\n([\s\S]*?)\n---\n/);
  assert.ok(match, 'missing Front Matter');
  return match[1];
}

test('fundamentals posts follow the naming and metadata contract', () => {
  for (const post of postFiles()) {
    assert.match(post.name, /^\d{2}-[a-z0-9]+(?:-[a-z0-9]+)*\.md$/, post.name);
    const source = fs.readFileSync(post.path, 'utf8');
    const metadata = frontMatter(source);

    for (const field of ['title', 'date', 'categories', 'tags']) {
      assert.match(metadata, new RegExp(`^${field}:`, 'm'), `${post.name}: missing ${field}`);
    }
    assert.match(metadata, /^date:\s*\d{4}-\d{2}-\d{2}$/m, `${post.name}: invalid date`);
    assert.match(metadata, /^tags:\n(?:\s+- .+\n){2,}/m, `${post.name}: needs at least two tags`);
  }
});

test('fundamentals posts use fenced-code languages and markdown links', () => {
  for (const post of postFiles()) {
    const source = fs.readFileSync(post.path, 'utf8');
    let inFence = false;
    for (const line of source.split('\n')) {
      const fence = line.match(/^\s*(```+|~~~+)\s*(.*)$/);
      if (!fence) continue;
      if (inFence) {
        inFence = false;
        continue;
      }
      assert.notEqual(fence[2].trim(), '', `${post.name}: found an unlabeled code fence`);
      inFence = true;
    }
    assert.doesNotMatch(source, /^https?:\/\/\S+$/m, `${post.name}: found a bare URL`);
  }
});
