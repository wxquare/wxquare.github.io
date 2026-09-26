'use strict';

const assert = require('node:assert/strict');
const fs = require('node:fs');
const path = require('node:path');
const { test } = require('node:test');

const repoRoot = path.resolve(__dirname, '..');
const postsRoot = path.join(repoRoot, 'source/_posts');

function collectMarkdownFiles(directory) {
  const files = [];
  for (const name of fs.readdirSync(directory)) {
    const file = path.join(directory, name);
    if (fs.statSync(file).isDirectory()) files.push(...collectMarkdownFiles(file));
    else if (name.endsWith('.md')) files.push(file);
  }
  return files;
}

function withoutFencedCode(content) {
  return content.replace(/```[\s\S]*?```/g, '').replace(/~~~[\s\S]*?~~~/g, '');
}

function citationErrors(file) {
  const original = fs.readFileSync(file, 'utf8');
  const content = withoutFencedCode(original);
  const errors = [];
  const referenceHeading = /^(?:##|###|####)\s+(?:Sources|参考\s*$|.*(?:参考资料|参考资源|参考学习|参考阅读|延伸阅读|学习资料))\s*$/;
  const hasReferenceSection = content.split('\n').some((line) => referenceHeading.test(line)) || /^\*\*参考资料\*\*\s*[：:]/m.test(content);
  const anchors = [...content.matchAll(/<a\s+id=["'](ref-\d+)["']\s*><\/a>/gi)].map((match) => match[1]);
  const anchorSet = new Set(anchors);

  if (new Set(anchors).size !== anchors.length) errors.push('duplicate reference anchor');

  for (const match of content.matchAll(/\[R\d+\]/g)) {
    errors.push(`legacy citation ${match[0]}`);
  }

  if (hasReferenceSection) {
    for (const match of content.matchAll(/【\d+】/g)) {
      errors.push(`legacy citation ${match[0]}`);
    }
  }

  for (const match of content.matchAll(/\[\[(\d+)\]\]\(#(ref-\d+)\)/g)) {
    if (!anchorSet.has(match[2])) errors.push(`missing anchor #${match[2]} for [${match[1]}]`);
  }

  if (anchorSet.size > 0 || hasReferenceSection) {
    for (const match of content.matchAll(/(?<![A-Za-z0-9_\[])\[(\d+)\](?!\]\(#ref-\d+\)|[A-Za-z0-9_,])/g)) {
      errors.push(`unlinked numeric citation ${match[0]}`);
    }
  }

  for (const match of content.matchAll(/<a\s+id=["'](ref-\d+)["']\s*><\/a>[^\n]*(?:\[(\d+)\]|^(?:[-*]\s*)?(\d+)\.)/gim)) {
    const number = match[2] || match[3];
    if (match[1] !== `ref-${number}`) errors.push(`anchor ${match[1]} does not match reference ${number}`);
  }

  if (hasReferenceSection) {
    const lines = content.split('\n');
    for (let index = 0; index < lines.length; index += 1) {
      if (!referenceHeading.test(lines[index])) continue;
      const headingLevel = lines[index].match(/^(#+)/)[1].length;
      for (let next = index + 1; next < lines.length; next += 1) {
        const nextHeading = lines[next].match(/^(#+)\s+/);
        if (nextHeading && nextHeading[1].length <= headingLevel) break;
        if (/^\s*(?:\d+\.|[-*]\s+\d+\.)/.test(lines[next]) && !/<a\s+id=["']ref-\d+["']/.test(lines[next - 1] || '')) {
          errors.push(`reference entry missing anchor near line ${next + 1}`);
        }
      }
    }
  }

  return errors;
}

test('blog citations use numeric linked references with matching anchors', () => {
  const errors = collectMarkdownFiles(postsRoot).flatMap((file) =>
    citationErrors(file).map((error) => `${path.relative(repoRoot, file)}: ${error}`)
  );
  assert.deepEqual(errors, [], errors.join('\n'));
});

module.exports = { citationErrors, collectMarkdownFiles, withoutFencedCode };
