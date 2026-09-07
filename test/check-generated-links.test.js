'use strict';

const assert = require('assert').strict;
const fs = require('fs');
const os = require('os');
const path = require('path');
const test = require('node:test');

const { checkGeneratedLinks } = require('../scripts/check-generated-links');

test('generated-link checker resolves local links and reports missing targets', () => {
  const publicDir = fs.mkdtempSync(path.join(os.tmpdir(), 'hexo-link-check-'));

  try {
    fs.mkdirSync(path.join(publicDir, 'docs'), { recursive: true });
    fs.mkdirSync(path.join(publicDir, 'nested'), { recursive: true });
    fs.writeFileSync(path.join(publicDir, 'index.html'), [
      '<a href="/docs/">docs</a>',
      '<a href="/missing/">missing</a>',
      '<a href="https://example.com/external">external</a>',
      '<a href="#section">fragment</a>'
    ].join('\n'));
    fs.writeFileSync(path.join(publicDir, 'docs/index.html'), [
      '<a href="../">home</a>',
      '<a href="/nested/page.html">nested</a>'
    ].join('\n'));
    fs.writeFileSync(path.join(publicDir, 'nested/page.html'), [
      '<a href="../docs/#section">docs section</a>',
      '<a href="missing.html">missing relative</a>'
    ].join('\n'));

    const result = checkGeneratedLinks(publicDir);

    assert.equal(result.checked, 6);
    assert.deepEqual(result.failures, [
      {
        file: 'index.html',
        href: '/missing/',
        target: 'missing/index.html'
      },
      {
        file: 'nested/page.html',
        href: 'missing.html',
        target: 'nested/missing.html'
      }
    ]);
  } finally {
    fs.rmSync(publicDir, { recursive: true, force: true });
  }
});
