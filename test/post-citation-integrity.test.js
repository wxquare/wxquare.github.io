'use strict';

const assert = require('node:assert/strict');
const fs = require('node:fs');
const path = require('node:path');
const { test } = require('node:test');
const { marked } = require('marked');
const { parseFragment } = require('parse5');

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

function citationErrors(file, original = fs.readFileSync(file, 'utf8')) {
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
    let prose = content;
    marked.walkTokens(marked.lexer(content), (token) => {
      if (token.type === 'codespan') prose = prose.replaceAll(token.raw, '');
    });
    for (const match of prose.matchAll(/(?<![A-Za-z0-9_\[])\[(\d+)\](?!\]\(#ref-\d+\)|[A-Za-z0-9_,])/g)) {
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
        if (
          /^\s*(?:\d+\.|[-*]\s+\d+\.)/.test(lines[next]) &&
          !/<a\s+id=["']ref-\d+["']/.test(lines[next]) &&
          !/<a\s+id=["']ref-\d+["']/.test(lines[next - 1] || '')
        ) {
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

test('inline code array literals are not mistaken for numeric citations', () => {
  const content = [
    'Array examples: `[1]`, `[2]`, and `[3]`. Source: [[1]](#ref-1).',
    '',
    '## 参考资料',
    '',
    '1. <a id="ref-1"></a> [Example](https://example.com/)'
  ].join('\n');
  assert.deepEqual(citationErrors('inline-code-example', content), []);
});

test('Superpowers references render as one ordered list with each anchor inside its item', () => {
  const file = path.join(postsRoot, 'AI/03-superpowers-source-workflow-practice.md');
  const content = withoutFencedCode(fs.readFileSync(file, 'utf8'));
  const sectionStart = content.indexOf('## 参考资料');
  const sectionEnd = content.indexOf('\n## ', sectionStart + 1);
  const section = content.slice(sectionStart, sectionEnd === -1 ? undefined : sectionEnd);
  const html = marked.parse(section);
  const lists = html.match(/<ol>/g) || [];
  const items = [...html.matchAll(/<li>([\s\S]*?)<\/li>/g)];
  const anchoredItems = items.filter((match) => /<a id="ref-\d+"><\/a>/.test(match[1]));

  assert.equal(lists.length, 1, 'all 22 references should share one ordered list');
  assert.equal(items.length, 22, 'the reference list should contain 22 items');
  assert.equal(anchoredItems.length, 22, 'each reference anchor should be inside its corresponding list item');
});

test('Vibe Coding references render as one ordered list with matching anchors', () => {
  const file = path.join(postsRoot, 'AI/01-vibe-coding-vs-spec-coding.md');
  const content = withoutFencedCode(fs.readFileSync(file, 'utf8'));
  const sectionStart = content.indexOf('## 参考资料');
  const sectionEnd = content.indexOf('\n---', sectionStart + 1);
  const section = content.slice(sectionStart, sectionEnd === -1 ? undefined : sectionEnd);
  const html = marked.parse(section);
  const lists = html.match(/<ol(?:\s[^>]*)?>/g) || [];
  const items = [...html.matchAll(/<li>([\s\S]*?)<\/li>/g)];
  const anchorIds = items.flatMap((match) =>
    [...match[1].matchAll(/<a id="(ref-\d+)"><\/a>/g)].map((anchor) => anchor[1])
  );

  assert.equal(lists.length, 1, 'all six references should share one ordered list');
  assert.equal(items.length, 6, 'the reference list should contain six items');
  assert.deepEqual(anchorIds, Array.from({ length: 6 }, (_, index) => `ref-${index + 1}`));
});

function textContent(node) {
  return node.nodeName === '#text' ? node.value : (node.childNodes || []).map(textContent).join('');
}

function visit(node, callback) {
  callback(node);
  for (const child of node.childNodes || []) visit(child, callback);
}

test('all reference anchors belong to numbered entries with readable links', () => {
  const errors = [];
  for (const file of collectMarkdownFiles(postsRoot)) {
    const content = withoutFencedCode(fs.readFileSync(file, 'utf8'));
    const fragment = parseFragment(marked.parse(content));
    visit(fragment, (node) => {
      const id = node.attrs?.find((attribute) => attribute.name === 'id')?.value;
      if (node.tagName !== 'a' || !/^ref-\d+$/.test(id || '')) return;
      let item = node.parentNode;
      while (item && item.tagName !== 'li') item = item.parentNode;
      const prefix = `${path.relative(repoRoot, file)} #${id}`;
      if (!item) {
        errors.push(`${prefix}: anchor is outside its list item`);
        return;
      }
      const list = item.parentNode;
      const start = Number(list.attrs?.find((attribute) => attribute.name === 'start')?.value || 1);
      const index = list.childNodes.filter((child) => child.tagName === 'li').indexOf(item);
      if (list.tagName !== 'ol' || start + index !== Number(id.slice(4))) {
        errors.push(`${prefix}: list number does not match reference anchor`);
      }
      if (/https?:\/\//.test(textContent(item))) errors.push(`${prefix}: visible raw URL`);
      visit(item, (child) => {
        if (child.tagName === 'br') errors.push(`${prefix}: unnecessary forced line break`);
      });
    });
  }
  assert.deepEqual(errors, [], errors.join('\n'));
});

test('reference list styling is scoped and left aligned', () => {
  const styles = fs.readFileSync(path.join(repoRoot, 'source/_data/styles.styl'), 'utf8');
  assert.match(styles, /\.post-body ol:has\(a\[id\^=['"]ref-['"]\]\)/);
  assert.match(styles, /\.post-body ul:has\(a\[id\^=['"]ref-['"]\]\)/);
  const referenceStyles = styles.slice(styles.indexOf('.post-body ol:has'));
  assert.match(referenceStyles, /text-align:\s*left/);
  assert.match(referenceStyles, /overflow-wrap:\s*anywhere/);
});

test('recommended reading and chapter-only sources are included in reference normalization', () => {
  for (const [relative, expectedCount] of [
    ['fundamentals/07-mysql-database-practice.md', 13],
    ['fundamentals/10-elasticsearch-search-practice.md', 16],
    ['AI/07-tensorflow-model-optimization.md', 36]
  ]) {
    const content = fs.readFileSync(path.join(postsRoot, relative), 'utf8');
    const anchors = [...content.matchAll(/<a id="ref-\d+"><\/a>/g)];
    assert.equal(anchors.length, expectedCount, `${relative}: incomplete reference coverage`);
  }
});

module.exports = { citationErrors, collectMarkdownFiles, withoutFencedCode };
