'use strict';

const assert = require('node:assert/strict');
const fs = require('node:fs');
const path = require('node:path');
const { test } = require('node:test');

const repoRoot = path.resolve(__dirname, '..');
const registryPath = path.join(repoRoot, '.agents/config/post-categories.json');

const registryConsumers = [
  '.agents/commands/stats.md',
  '.agents/skills/new-post/SKILL.md',
  '.agents/skills/organize-posts/SKILL.md'
];

test('post category registry covers every canonical post directory', () => {
  assert.ok(fs.existsSync(registryPath), 'post category registry must exist');

  const registry = JSON.parse(fs.readFileSync(registryPath, 'utf8'));
  const slugs = registry.categories.map((category) => category.slug);

  assert.deepEqual(slugs, ['AI', 'system-design', 'fundamentals', 'other']);

  for (const category of registry.categories) {
    assert.ok(category.label, `${category.slug} must define a display label`);
    assert.ok(
      Array.isArray(category.frontMatterLabels),
      `${category.slug} must define Front Matter labels`
    );
    assert.ok(
      category.frontMatterLabels.includes(category.label),
      `${category.slug} default label must be an allowed Front Matter label`
    );
    assert.ok(category.directory, `${category.slug} must define a directory`);
    assert.ok(
      fs.existsSync(path.join(repoRoot, category.directory)),
      `${category.slug} directory must exist`
    );
  }
});

test('category-aware workflow entries use the shared category registry', () => {
  for (const file of registryConsumers) {
    const content = fs.readFileSync(path.join(repoRoot, file), 'utf8');

    assert.match(content, /\.agents\/config\/post-categories\.json/, `${file} must reference the registry`);
  }
});
