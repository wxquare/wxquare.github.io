'use strict';

const assert = require('assert').strict;
const fs = require('fs');
const path = require('path');

const root = path.resolve(__dirname, '..');
const exists = (relative) => fs.existsSync(path.join(root, relative));

assert.equal(exists('source/library/index.md'), true);
assert.equal(exists('source/library/books'), true);
assert.equal(exists('source/library/slides'), true);
assert.equal(exists('source/library/presentations'), true);
assert.equal(exists('source/library/other'), true);
assert.equal(exists('source/presentations'), false);
assert.equal(exists('source/library/tutorials'), false);
