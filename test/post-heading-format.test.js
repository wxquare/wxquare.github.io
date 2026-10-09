'use strict';

const assert = require('node:assert/strict');
const fs = require('node:fs');
const path = require('node:path');
const { test } = require('node:test');

const repoRoot = path.resolve(__dirname, '..');
const postsRoot = path.join(repoRoot, 'source', '_posts');
const { parsePostHeadings, validatePostHeadings } = require('../tools/check-post-headings');

function collectMarkdownFiles(directory) {
  const files = [];
  for (const name of fs.readdirSync(directory)) {
    const file = path.join(directory, name);
    if (fs.statSync(file).isDirectory()) files.push(...collectMarkdownFiles(file));
    else if (name.endsWith('.md')) files.push(file);
  }
  return files;
}

test('post heading parser ignores front matter and fenced code', () => {
  const source = [
    '---',
    'title: Example',
    '# metadata example',
    '---',
    '',
    '## Main section',
    '',
    '```markdown',
    '# code example',
    '### 1. numbered sample',
    '```',
    '',
    '### Subsection',
    ''
  ].join('\n');

  assert.deepEqual(
    parsePostHeadings(source).map(({ level, text, line }) => ({ level, text, line })),
    [
      { level: 2, text: 'Main section', line: 6 },
      { level: 3, text: 'Subsection', line: 13 }
    ]
  );
});

test('post heading validator accepts semantic hierarchy and meaningful sequences', () => {
  const source = [
    '---',
    'title: Example',
    '---',
    '',
    '## Main section',
    '### A subsection',
    '#### Step 1: Run the check',
    '### Question 1. What does it do?',
    '#### Common questions'
  ].join('\n');

  assert.deepEqual(validatePostHeadings(source), []);
});

test('post heading validator rejects body H1 and headings deeper than H4', () => {
  const source = '# Repeated title\n\n## Main section\n\n##### Too deep\n';
  const errors = validatePostHeadings(source);

  assert.ok(errors.some((error) => error.includes('H1')));
  assert.ok(errors.some((error) => error.includes('H4')));
});

test('post heading validator rejects skipped heading levels', () => {
  const errors = validatePostHeadings('## Main section\n\n#### Skipped subsection\n');

  assert.ok(errors.some((error) => error.includes('跳级')));
});

test('post heading validator requires the first body heading to be H2', () => {
  const errors = validatePostHeadings('### Orphaned subsection\n');

  assert.ok(errors.some((error) => error.includes('首个正文标题必须使用 H2')));
});

test('post heading validator rejects an unclosed code fence', () => {
  const errors = validatePostHeadings('## Main section\n\n```bash\n### Hidden heading\n');

  assert.ok(errors.some((error) => error.includes('代码围栏未闭合')));
});

test('post heading validator detects code fences nested in list indentation', () => {
  const errors = validatePostHeadings('- example\n    ```bash\n    printf test\n');

  assert.ok(errors.some((error) => error.includes('代码围栏未闭合')));
});

test('post heading validator rejects manual section numbering', () => {
  const source = [
    '## 一、第一章',
    '### 1.1 Nested section',
    '#### 2、Another numbered section',
    '### Section 3: semantic number is not a prefix'
  ].join('\n');
  const errors = validatePostHeadings(source);

  assert.equal(errors.filter((error) => error.includes('手工序号')).length, 3);
});

test('all blog posts follow the heading format', () => {
  const errors = collectMarkdownFiles(postsRoot).flatMap((file) =>
    validatePostHeadings(fs.readFileSync(file, 'utf8')).map((error) =>
      `${path.relative(repoRoot, file)}: ${error}`
    )
  );

  assert.deepEqual(errors, [], errors.join('\n'));
});
