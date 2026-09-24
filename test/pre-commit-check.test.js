'use strict';

const assert = require('node:assert/strict');
const { spawnSync } = require('node:child_process');
const fs = require('node:fs');
const os = require('node:os');
const path = require('node:path');
const { test } = require('node:test');

const repoRoot = path.resolve(__dirname, '..');
const sourceScript = path.join(repoRoot, 'tools', 'pre-commit-check.sh');

function runGit(cwd, args) {
  const result = spawnSync('git', args, { cwd, encoding: 'utf8' });
  assert.equal(result.status, 0, `${args.join(' ')} failed:\n${result.stderr}`);
}

function copyPreCommitScript(tempRoot) {
  const tempScript = path.join(tempRoot, 'tools', 'pre-commit-check.sh');
  fs.mkdirSync(path.dirname(tempScript), { recursive: true });
  fs.copyFileSync(sourceScript, tempScript);
  fs.chmodSync(tempScript, 0o755);
  return tempScript;
}

test('pre-commit check runs site and book builds for a valid staged Markdown file', () => {
  assert.ok(fs.existsSync(sourceScript), 'tools/pre-commit-check.sh must exist');

  const tempRoot = fs.mkdtempSync(path.join(os.tmpdir(), 'wxquare-precommit-'));

  try {
    runGit(tempRoot, ['init', '-q']);

    const stagedPost = [
      '---',
      'title: Test post',
      'date: 2026-09-03',
      'categories:',
      '  - test',
      'tags:',
      '  - test',
      '---',
      '',
      '```js',
      'console.log("ok");',
      '```',
      ''
    ].join('\n');
    fs.writeFileSync(path.join(tempRoot, 'post.md'), stagedPost);
    runGit(tempRoot, ['add', 'post.md']);

    const tempBin = path.join(tempRoot, 'fake-bin');
    fs.mkdirSync(tempBin);
    const callLog = path.join(tempRoot, 'npm-calls.log');
    const fakeNpm = path.join(tempBin, 'npm');
    fs.writeFileSync(fakeNpm, '#!/bin/sh\nprintf "%s\\n" "$*" >> "$NPM_CALL_LOG"\n');
    fs.chmodSync(fakeNpm, 0o755);

    const tempScript = copyPreCommitScript(tempRoot);

    const result = spawnSync('bash', [tempScript], {
      cwd: tempRoot,
      encoding: 'utf8',
      env: {
        ...process.env,
        NPM_CALL_LOG: callLog,
        PATH: `${tempBin}:${process.env.PATH}`
      }
    });

    assert.equal(result.status, 0, `${result.stdout}\n${result.stderr}`);
    assert.deepEqual(
      fs.readFileSync(callLog, 'utf8').trim().split('\n'),
      ['run clean', 'test', 'run build', 'run build:books', 'run clean']
    );
  } finally {
    fs.rmSync(tempRoot, { recursive: true, force: true });
  }
});

test('pre-commit check does not apply blog Front Matter rules to mdBook Markdown', () => {
  const tempRoot = fs.mkdtempSync(path.join(os.tmpdir(), 'wxquare-precommit-book-'));

  try {
    runGit(tempRoot, ['init', '-q']);
    const stagedBook = path.join(
      tempRoot,
      'books/reliable-system-design/src/part01/01-test.md'
    );
    fs.mkdirSync(path.dirname(stagedBook), { recursive: true });
    fs.writeFileSync(stagedBook, '# 第 1 章\n\n## 1.1 测试\n');
    runGit(tempRoot, ['add', stagedBook]);

    const tempBin = path.join(tempRoot, 'fake-bin');
    fs.mkdirSync(tempBin);
    const npmLog = path.join(tempRoot, 'npm-calls.log');
    const pythonLog = path.join(tempRoot, 'python-calls.log');
    const fakeNpm = path.join(tempBin, 'npm');
    const fakePython = path.join(tempBin, 'python3');
    fs.writeFileSync(fakeNpm, '#!/bin/sh\nprintf "%s\\n" "$*" >> "$NPM_CALL_LOG"\n');
    fs.writeFileSync(
      fakePython,
      '#!/bin/sh\nprintf "%s\\n" "$*" >> "$PYTHON_CALL_LOG"\n'
    );
    fs.chmodSync(fakeNpm, 0o755);
    fs.chmodSync(fakePython, 0o755);

    const result = spawnSync('bash', [copyPreCommitScript(tempRoot)], {
      cwd: tempRoot,
      encoding: 'utf8',
      env: {
        ...process.env,
        NPM_CALL_LOG: npmLog,
        PYTHON_CALL_LOG: pythonLog,
        PATH: `${tempBin}:${process.env.PATH}`
      }
    });

    assert.equal(result.status, 0, `${result.stdout}\n${result.stderr}`);
    assert.doesNotMatch(result.stdout, /缺少 Front Matter/);
    assert.match(result.stdout, /跳过博客 Front Matter 检查/);
    assert.match(
      fs.readFileSync(pythonLog, 'utf8'),
      /tools\/check-reliable-system-design\.py/
    );
    assert.deepEqual(
      fs.readFileSync(npmLog, 'utf8').trim().split('\n'),
      ['run clean', 'test', 'run build', 'run build:books', 'run clean']
    );
  } finally {
    fs.rmSync(tempRoot, { recursive: true, force: true });
  }
});

test('pre-commit check still rejects a blog Markdown file without Front Matter', () => {
  const tempRoot = fs.mkdtempSync(path.join(os.tmpdir(), 'wxquare-precommit-blog-'));

  try {
    runGit(tempRoot, ['init', '-q']);
    const stagedPost = path.join(tempRoot, 'source/_posts/missing-front-matter.md');
    fs.mkdirSync(path.dirname(stagedPost), { recursive: true });
    fs.writeFileSync(stagedPost, '# Blog post without metadata\n');
    runGit(tempRoot, ['add', stagedPost]);

    const result = spawnSync('bash', [copyPreCommitScript(tempRoot)], {
      cwd: tempRoot,
      encoding: 'utf8'
    });

    assert.equal(result.status, 1, `${result.stdout}\n${result.stderr}`);
    assert.match(result.stdout, /缺少 Front Matter/);
  } finally {
    fs.rmSync(tempRoot, { recursive: true, force: true });
  }
});
