'use strict';

const assert = require('node:assert/strict');
const fs = require('node:fs');
const path = require('node:path');
const { test } = require('node:test');

const repoRoot = path.resolve(__dirname, '..');

test('repository declares the supported Node.js and npm runtimes', () => {
  const packageJson = JSON.parse(
    fs.readFileSync(path.join(repoRoot, 'package.json'), 'utf8')
  );
  const nvmrc = fs.readFileSync(path.join(repoRoot, '.nvmrc'), 'utf8').trim();

  assert.deepEqual(packageJson.engines, {
    node: '>=20 <23',
    npm: '>=10'
  });
  assert.equal(nvmrc, '20');
});
