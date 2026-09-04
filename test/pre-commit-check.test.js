'use strict';

const assert = require('node:assert/strict');
const { spawnSync } = require('node:child_process');
const fs = require('node:fs');
const os = require('node:os');
const path = require('node:path');
const { test } = require('node:test');

const repoRoot = path.resolve(__dirname, '..');
const sourceScript = path.join(repoRoot, 'bin', 'pre-commit-check.sh');

function runGit(cwd, args) {
  const result = spawnSync('git', args, { cwd, encoding: 'utf8' });
  assert.equal(result.status, 0, `${args.join(' ')} failed:\n${result.stderr}`);
}

test('pre-commit check runs site and book builds for a valid staged Markdown file', () => {
  assert.ok(fs.existsSync(sourceScript), 'bin/pre-commit-check.sh must exist');

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

    const tempScript = path.join(tempRoot, 'bin', 'pre-commit-check.sh');
    fs.mkdirSync(path.dirname(tempScript), { recursive: true });
    fs.copyFileSync(sourceScript, tempScript);
    fs.chmodSync(tempScript, 0o755);

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
