#!/usr/bin/env node
'use strict';

const fs = require('node:fs');
const path = require('node:path');

const NUMBERED_PREFIX =
  /^(?:(?:第)?[一二三四五六七八九十百]+[、.．)]\s*|第(?:\d+|[一二三四五六七八九十百]+)[章节部分]\s*|\d+(?:\.\d+)*(?:[、．)]\s*|\.(?!\d)\s*|\s+(?=\S))|[IVXLCDM]+[、.．)]\s*|(?:原则|实践|案例)\s*\d+\s*[:：])/i;
const EXPLICIT_STEP =
  /^(?:(?:第)?(?:[一二三四五六七八九十百]+|\d+)步|步骤\s*\d+|Step\s+\d+)(?:\s*[:：.、-]|\s|$)/i;
const SEQUENCE_PARENT = /(?:步骤|流程|工作流|执行顺序|操作顺序)/;
const QUESTION_ENDING = /[?？](?:[)\]`*_]*)$/;
const ORDERED_SECTION = /^(?:案例|原则|实践)\s*[一二三四五六七八九十百\d]+\s*[:：]/;

function scanPostMarkdown(markdown) {
  const lines = markdown.split(/\r?\n/);
  const headings = [];
  let inFrontMatter = lines[0]?.trim() === '---';
  let fenceCharacter = '';
  let fenceLength = 0;
  let fenceStartLine = 0;

  for (let index = 0; index < lines.length; index += 1) {
    const line = lines[index];

    if (index === 0 && inFrontMatter) continue;
    if (inFrontMatter) {
      if (line.trim() === '---') inFrontMatter = false;
      continue;
    }

    const fence = line.match(/^[ \t]*(`{3,}|~{3,})(.*)$/);
    if (fence) {
      const character = fence[1][0];
      const length = fence[1].length;
      if (!fenceCharacter) {
        fenceCharacter = character;
        fenceLength = length;
        fenceStartLine = index + 1;
      } else if (
        character === fenceCharacter &&
        length >= fenceLength &&
        /^\s*$/.test(fence[2])
      ) {
        fenceCharacter = '';
        fenceLength = 0;
        fenceStartLine = 0;
      }
      continue;
    }
    if (fenceCharacter) continue;

    const heading = line.match(/^ {0,3}(#{1,6})[ \t]+(.+?)\s*#*\s*$/);
    if (!heading) continue;

    headings.push({
      level: heading[1].length,
      text: heading[2].trim(),
      line: index + 1
    });
  }

  return { headings, unclosedFenceLine: fenceStartLine };
}

function parsePostHeadings(markdown) {
  return scanPostMarkdown(markdown).headings;
}

function validatePostHeadings(markdown) {
  const { headings, unclosedFenceLine } = scanPostMarkdown(markdown);
  const errors = [];
  const headingStack = [];
  let previousLevel = 0;

  if (unclosedFenceLine > 0) {
    errors.push(`第 ${unclosedFenceLine} 行：代码围栏未闭合`);
  }

  if (headings.length > 0 && headings[0].level !== 2) {
    errors.push(`第 ${headings[0].line} 行：首个正文标题必须使用 H2`);
  }

  for (const heading of headings) {
    const { level, text, line } = heading;
    if (level === 1) errors.push(`第 ${line} 行：正文不得使用 H1，文章标题由 Front Matter 的 title 提供`);
    if (level > 4) errors.push(`第 ${line} 行：正文标题最多使用 H4，当前为 H${level}`);
    if (previousLevel > 0 && level > previousLevel + 1) {
      errors.push(`第 ${line} 行：标题层级不能跳级（H${previousLevel} 后不能直接使用 H${level}）`);
    }

    while (headingStack.length > 0 && headingStack.at(-1).level >= level) {
      headingStack.pop();
    }

    const parent = headingStack.at(-1);
    const hasOrdinal = NUMBERED_PREFIX.test(text);
    const isOrderedQuestion = hasOrdinal && QUESTION_ENDING.test(text);
    const isOrderedStep =
      hasOrdinal &&
      (EXPLICIT_STEP.test(text) || SEQUENCE_PARENT.test(parent?.text || '') || ORDERED_SECTION.test(text));

    if (hasOrdinal && !isOrderedQuestion && !isOrderedStep) {
      errors.push(`第 ${line} 行：章节标题不得带手工序号“${text}”`);
    }

    headingStack.push(heading);
    previousLevel = level;
  }

  return errors;
}

function checkFiles(files) {
  let errorCount = 0;
  for (const filename of files) {
    const file = path.resolve(filename);
    const errors = validatePostHeadings(fs.readFileSync(file, 'utf8'));
    for (const error of errors) {
      console.error(`${filename}: ${error}`);
      errorCount += 1;
    }
  }
  return errorCount;
}

if (require.main === module) {
  const files = process.argv.slice(2);
  if (files.length === 0) {
    console.error('用法：node tools/check-post-headings.js <文章.md> [...]');
    process.exitCode = 2;
  } else if (checkFiles(files) > 0) {
    process.exitCode = 1;
  }
}

module.exports = { parsePostHeadings, validatePostHeadings };
