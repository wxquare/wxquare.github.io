'use strict';

const assert = require('assert').strict;
const fs = require('fs');
const path = require('path');
const test = require('node:test');

const repoRoot = path.resolve(__dirname, '..');
const postsRoot = path.join(repoRoot, 'source/_posts');

function collectMarkdownFiles(directory) {
  const files = [];

  for (const name of fs.readdirSync(directory)) {
    const file = path.join(directory, name);
    const stat = fs.statSync(file);

    if (stat.isDirectory()) files.push(...collectMarkdownFiles(file));
    else if (name.endsWith('.md')) files.push(file);
  }

  return files;
}

function withoutFencedCode(content) {
  return content.replace(/```[\s\S]*?```/g, '').replace(/~~~[\s\S]*?~~~/g, '');
}

function postSlugs(files) {
  return new Set(files.map((file) =>
    path.relative(postsRoot, file).replaceAll(path.sep, '/').replace(/\.md$/, '')));
}

function sourcePostBasenames(files) {
  return new Set(files.map((file) => path.basename(file)));
}

function findInvalidPostLinks(files, slugs, basenames) {
  const invalid = [];
  const hardcodedRoute = /\]\((\/(?:20\d{2}\/\d{2}\/\d{2}\/)?(?:AI|system-design|fundamentals)\/[^\s)<>]+)\)/g;
  const postTag = /{%\s*post_link\s+([^\s%]+)[^%]*%}/g;
  const relativeMarkdown = /\]\(((?:\.\.\/|\.\/)+[^\s)<>]+\.md(?:#[^\s)<>]+)?)\)/g;

  for (const file of files) {
    const content = withoutFencedCode(fs.readFileSync(file, 'utf8'));
    let match;

    while ((match = postTag.exec(content))) {
      if (!slugs.has(match[1].split('#')[0])) {
        invalid.push(`${path.relative(repoRoot, file)}: unknown post_link slug ${match[1]}`);
      }
    }

    while ((match = hardcodedRoute.exec(content))) {
      invalid.push(`${path.relative(repoRoot, file)}: hardcoded article route ${match[1]}`);
    }

    while ((match = relativeMarkdown.exec(content))) {
      const basename = path.basename(match[1].split('#')[0]);
      if (basenames.has(basename) || /^\d{1,3}-/.test(basename)) {
        invalid.push(`${path.relative(repoRoot, file)}: source-relative article link ${match[1]}`);
      }
    }
  }

  return invalid;
}

test('article references use valid post_link slugs instead of route-shaped links', () => {
  const files = collectMarkdownFiles(postsRoot);
  const invalid = findInvalidPostLinks(files, postSlugs(files), sourcePostBasenames(files));

  assert.deepEqual(invalid, [], invalid.join('\n'));
});

module.exports = {
  collectMarkdownFiles,
  findInvalidPostLinks,
  postSlugs,
  sourcePostBasenames,
  withoutFencedCode
};
