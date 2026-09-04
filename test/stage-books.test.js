'use strict';

const assert = require('node:assert/strict');
const fs = require('node:fs');
const os = require('node:os');
const path = require('node:path');
const { test } = require('node:test');
const { stageBooks } = require('../bin/stage-books');

test('stageBooks copies both mdBook outputs under the public site tree', () => {
  const tempRoot = fs.mkdtempSync(path.join(os.tmpdir(), 'wxquare-stage-books-'));

  try {
    const outputs = [
      ['books/ai-book/book', 'public/ai-book', 'AI book'],
      [
        'books/system-design-architecture-book/book',
        'public/system-design-architecture-book',
        'System design book'
      ]
    ];

    for (const [sourceRelative, destinationRelative, content] of outputs) {
      const sourceDir = path.join(tempRoot, sourceRelative);
      fs.mkdirSync(sourceDir, { recursive: true });
      fs.writeFileSync(path.join(sourceDir, 'index.html'), content);
      const destinationDir = path.join(tempRoot, destinationRelative);
      fs.mkdirSync(destinationDir, { recursive: true });
      fs.writeFileSync(path.join(destinationDir, 'stale.html'), 'stale output');
    }

    stageBooks(tempRoot);

    for (const [, destinationRelative, content] of outputs) {
      assert.equal(
        fs.readFileSync(path.join(tempRoot, destinationRelative, 'index.html'), 'utf8'),
        content
      );
      assert.equal(fs.existsSync(path.join(tempRoot, destinationRelative, 'stale.html')), false);
    }
  } finally {
    fs.rmSync(tempRoot, { recursive: true, force: true });
  }
});
