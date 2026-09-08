'use strict';

const assert = require('node:assert/strict');
const fs = require('node:fs');
const path = require('node:path');
const test = require('node:test');

const root = path.resolve(__dirname, '..');

test('legacy ecommerce-book page guides readers to both current books', () => {
  const legacyPage = fs.readFileSync(path.join(root, 'source/ecommerce-book/index.md'), 'utf8');
  const menuConfig = fs.readFileSync(path.join(root, 'themes/next/_config.yml'), 'utf8');
  const bookList = fs.readFileSync(path.join(root, 'source/booklist/index.md'), 'utf8');

  for (const content of [legacyPage, bookList]) {
    assert.match(content, /\/system-design-primer\//);
    assert.match(content, /\/ai-book\//);
  }
  assert.match(menuConfig, /^  book: \/booklist\/ \|\| fa fa-book$/m);
  assert.doesNotMatch(bookList, /\/ecommerce-book\//);
});
