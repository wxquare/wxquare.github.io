'use strict';

const assert = require('node:assert/strict');
const fs = require('node:fs');
const path = require('node:path');
const test = require('node:test');

const root = path.resolve(__dirname, '..');
const aliasScript = path.join(root, 'scripts/legacy-book-alias.js');

test('canonical booklist page guides readers to both current books', () => {
  const bookList = fs.readFileSync(path.join(root, 'source/booklist/index.md'), 'utf8');
  const menuConfig = fs.readFileSync(path.join(root, 'themes/next/_config.yml'), 'utf8');
  const zhCN = fs.readFileSync(path.join(root, 'themes/next/languages/zh-CN.yml'), 'utf8');

  assert.match(bookList, /\/system-design-primer\//);
  assert.match(bookList, /\/ai-book\//);
  assert.match(menuConfig, /^  book: \/booklist\/ \|\| fa fa-book$/m);
  assert.doesNotMatch(menuConfig, /^  categories: \/categories\/ \|\| fa fa-th$/m);
  assert.match(zhCN, /^  book: Books$/m);
  assert.doesNotMatch(bookList, /\/ecommerce-book\//);
  assert.equal(fs.existsSync(path.join(root, 'source/ecommerce-book/index.md')), false);
});

test('legacy ecommerce-book URL redirects to canonical booklist', () => {
  const registrations = [];
  const previousHexo = global.hexo;

  try {
    global.hexo = {
      extend: {
        generator: {
          register(name, fn) {
            registrations.push({ name, fn });
          }
        }
      }
    };

    delete require.cache[aliasScript];
    require(aliasScript);

    assert.equal(registrations.length, 1);
    assert.equal(registrations[0].name, 'legacy-book-alias');
    const routes = registrations[0].fn();

    assert.deepEqual(routes.map(route => route.path), ['ecommerce-book/index.html']);
    assert.match(routes[0].data, /rel="canonical"/);
    assert.match(routes[0].data, /url=\/booklist\//);
    assert.match(routes[0].data, /location\.replace/);
  } finally {
    delete require.cache[aliasScript];
    if (previousHexo === undefined) delete global.hexo;
    else global.hexo = previousHexo;
  }
});
