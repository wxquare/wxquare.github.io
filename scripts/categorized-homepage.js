'use strict';

const fs = require('node:fs');
const path = require('node:path');

const HOMEPAGE_INDEX_VIEW = `{% extends '_layout.njk' %}
{% import '_macro/sidebar.njk' as sidebar_template with context %}

{% block title %}{{ title }}{% if theme.index_with_subtitle and subtitle %} - {{ subtitle }}{% endif %}{% endblock %}
{% block class %}index posts-expand categorized-homepage{% endblock %}
{% block content %}{{ page.homepageMarkup | safe }}{% endblock %}
{% block sidebar %}{{ sidebar_template.render(false) }}{% endblock %}
`;

function asArray(collection) {
  if (Array.isArray(collection)) return collection;
  if (collection && typeof collection.toArray === 'function') return collection.toArray();
  return [];
}

function categoryNames(post) {
  return asArray(post.categories).map(category => category.name).filter(Boolean);
}

function escapeHtml(value) {
  return String(value)
    .replaceAll('&', '&amp;')
    .replaceAll('<', '&lt;')
    .replaceAll('>', '&gt;')
    .replaceAll('"', '&quot;')
    .replaceAll("'", '&#39;');
}

function buildHomepageGroups(posts, registry, limit = 5, onUnmapped = () => {}, categories = []) {
  const archivePaths = new Map(asArray(categories).map(category => [category.name, category.path]));
  const groups = registry.categories.map(category => ({
    slug: category.slug,
    label: category.label,
    labels: new Set(category.frontMatterLabels),
    archivePath: archivePaths.get(category.label),
    posts: []
  }));

  for (const post of asArray(posts)) {
    const names = categoryNames(post);
    const group = groups.find(candidate => names.some(name => candidate.labels.has(name)));
    if (!group) {
      if (names.length) onUnmapped(post, names);
      continue;
    }

    const date = new Date(post.date);
    group.posts.push({
      title: post.title,
      path: post.path,
      date: date.toISOString().slice(0, 10),
      tags: asArray(post.tags).map(tag => tag.name).filter(Boolean),
      timestamp: date.valueOf()
    });
  }

  return groups.map(group => ({
    ...group,
    posts: group.posts.sort((left, right) => right.timestamp - left.timestamp).slice(0, limit)
  }));
}

function renderHomepageMarkup(groups) {
  const bookLinks = [
    ['AI Agent 工程实践', '/ai-book/'],
    ['系统设计与架构实战', '/system-design-architecture-book/']
  ];
  const books = bookLinks.map(([label, href]) =>
    `<a class="categorized-home__book" href="${escapeHtml(href)}">${escapeHtml(label)}</a>`).join('');
  const sections = groups.map(group => {
    const rows = group.posts.length
      ? group.posts.map(post => `<li class="categorized-home__post"><time datetime="${escapeHtml(post.date)}">${escapeHtml(post.date)}</time><a href="/${escapeHtml(post.path)}">${escapeHtml(post.title)}</a><span>${post.tags.map(tag => `<span class="categorized-home__tag">${escapeHtml(tag)}</span>`).join('')}</span></li>`).join('')
      : '<li class="categorized-home__empty">暂无已发布文章。</li>';
    const archive = group.archivePath
      ? `<a href="/${escapeHtml(group.archivePath)}">查看全部</a>`
      : '<span class="categorized-home__archive-missing">暂无归档</span>';
    return `<section class="categorized-home__section"><div class="categorized-home__section-heading"><h2>${escapeHtml(group.label)}</h2>${archive}</div><ul>${rows}</ul></section>`;
  }).join('');

  return `<section class="categorized-home"><header class="categorized-home__intro"><h1>博客文章</h1><p>按主题浏览最新文章。</p><nav>${books}</nav></header>${sections}</section>`;
}

function registerCategorizedHomepage(hexoContext) {
  const registryPath = path.join(hexoContext.base_dir, '.agents/config/post-categories.json');
  const registry = JSON.parse(fs.readFileSync(registryPath, 'utf8'));

  hexoContext.extend.filter.register('before_generate', () => {
    hexoContext.theme.setView('categorized-homepage.njk', HOMEPAGE_INDEX_VIEW);
  });
  hexoContext.extend.generator.register('index', function categorizedHomepage(locals) {
    const groups = buildHomepageGroups(locals.posts, registry, 5, (post, names) => {
      hexoContext.log.warn(`Homepage omitted unmapped post ${post.source || post.path}: ${names.join(', ')}`);
    }, locals.categories);

    return [{
      path: 'index.html',
      layout: 'categorized-homepage',
      data: {
        __index: true,
        posts: locals.posts,
        homepageMarkup: renderHomepageMarkup(groups)
      }
    }];
  });
}

if (typeof hexo !== 'undefined') registerCategorizedHomepage(hexo);

module.exports = {
  asArray,
  categoryNames,
  buildHomepageGroups,
  escapeHtml,
  renderHomepageMarkup,
  registerCategorizedHomepage
};
