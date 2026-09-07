'use strict';

const assert = require('node:assert/strict');
const fs = require('node:fs');
const path = require('node:path');
const { test } = require('node:test');

const repoRoot = path.resolve(__dirname, '..');
const packageJson = JSON.parse(
  fs.readFileSync(path.join(repoRoot, 'package.json'), 'utf8')
);
const workflowDir = path.join(repoRoot, '.github/workflows');

test('package exposes canonical Hexo and mdBook build entrypoints', () => {
  assert.equal(packageJson.scripts.build, 'hexo generate');
  assert.equal(packageJson.scripts['build:ai-book'], 'mdbook build books/ai-book');
  assert.equal(
    packageJson.scripts['build:system-design-book'],
    'mdbook build books/system-design-architecture-book'
  );
  assert.equal(
    packageJson.scripts['build:books'],
    'npm run build:ai-book && npm run build:system-design-book'
  );
  assert.equal(packageJson.scripts['stage:books'], 'node tools/stage-books.js');
  assert.equal(
    packageJson.scripts['server:site'],
    'npm run clean && npm run build && npm run build:books && npm run stage:books && hexo server -p 3000'
  );
  assert.equal(
    packageJson.scripts['server:system-design-architecture-book'],
    'npm run server:site'
  );
  assert.equal(packageJson.scripts['build:system-design-architecture-book'], undefined);
});

test('one GitHub Actions workflow builds and publishes all site areas', () => {
  const workflowFiles = fs
    .readdirSync(workflowDir)
    .filter((file) => file.endsWith('.yml') || file.endsWith('.yaml'))
    .sort();

  assert.deepEqual(workflowFiles, ['deploy-site.yml']);

  const workflow = fs.readFileSync(path.join(workflowDir, 'deploy-site.yml'), 'utf8');
  assert.match(workflow, /mdbook-version:\s*'0\.5\.2'/);
  assert.match(workflow, /npm run build\s*$/m);
  assert.match(workflow, /npm run build:books/);
  assert.match(workflow, /npm run stage:books/);
  assert.doesNotMatch(workflow, /\bmdbook build\b/);
  assert.match(workflow, /publish_dir:\s*\.\/public/);
  assert.doesNotMatch(workflow, /destination_dir:/);
  assert.doesNotMatch(workflow, /keep_files:/);
  assert.doesNotMatch(workflow, /cp\s+-R|cp\s+-r/);
  assert.match(workflow, /'scripts\/\*\*'/);
  assert.match(workflow, /'tools\/\*\*'/);

  const buildIndex = workflow.indexOf('run: npm run clean && npm run build');
  const booksIndex = workflow.indexOf('run: npm run build:books');
  const stageIndex = workflow.indexOf('run: npm run stage:books');
  const publishIndex = workflow.indexOf('uses: peaceiris/actions-gh-pages@v3');
  assert.ok(buildIndex < booksIndex, 'Hexo must build before the books');
  assert.ok(booksIndex < stageIndex, 'books must build before staging');
  assert.ok(stageIndex < publishIndex, 'staging must happen before publishing');
  assert.equal((workflow.match(/peaceiris\/actions-gh-pages@v3/g) || []).length, 1);
});
