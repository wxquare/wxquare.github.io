'use strict';

const assert = require('node:assert/strict');
const { spawnSync } = require('node:child_process');
const fs = require('node:fs');
const path = require('node:path');
const { test } = require('node:test');

const repoRoot = path.resolve(__dirname, '..');
const bookRoot = path.join(repoRoot, 'books', 'reliable-system-design');
const sourceRoot = path.join(bookRoot, 'src');
const checker = path.join(repoRoot, 'tools', 'check-reliable-system-design.py');
const agents = fs.readFileSync(path.join(repoRoot, 'AGENTS.md'), 'utf8');

function read(relativePath) {
  return fs.readFileSync(path.join(sourceRoot, relativePath), 'utf8');
}

function chapterEntries(markdown) {
  return [...markdown.matchAll(/\[第\s*(\d+)\s*章[^\]]*\]\((part0[12]\/[^)]+\.md)\)/g)]
    .map((match) => ({ number: Number(match[1]), path: match[2] }));
}

test('active navigation exposes chapters 1 through 14 exactly once', () => {
  const summaryEntries = chapterEntries(read('SUMMARY.md'));
  const readmeEntries = chapterEntries(read('README.md'));

  assert.deepEqual(
    summaryEntries.map((entry) => entry.number),
    Array.from({ length: 14 }, (_, index) => index + 1)
  );
  assert.deepEqual(readmeEntries, summaryEntries);
  assert.equal(
    summaryEntries.find((entry) => entry.number === 13).path,
    'part02/13-marketing-pricing-system.md'
  );
  assert.doesNotMatch(read('SUMMARY.md'), /第三部分/);
});

test('reliable-system-design checker passes the repository source', () => {
  const result = spawnSync('python3', [checker], {
    cwd: repoRoot,
    encoding: 'utf8',
  });

  assert.equal(result.status, 0, `${result.stdout}\n${result.stderr}`);
});

test('AGENTS.md declares the system-design source boundary and generated output', () => {
  assert.match(agents, /books\/reliable-system-design\/src\//);
  assert.match(agents, /books\/reliable-system-design\/book\.toml/);
  assert.match(agents, /books\/reliable-system-design\/images\//);
  assert.match(agents, /books\/\*\/book\//);
  assert.match(agents, /当前结构不存在第三部分或第 15 章/);
  assert.doesNotMatch(agents, /第三部分实战章节/);
});
