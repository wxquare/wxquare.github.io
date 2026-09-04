'use strict';

const { url_for } = require('hexo-util');

const aliases = [
  { path: '/2025/05/15/system-design/14-system-reliability/', slug: 'system-design/07-system-reliability-engineering' },
  { path: '/2025/06/25/system-design/08-system-design-interview/', slug: 'system-design/08-system-design-interview' },
  { path: '/2026/04/02/01-claude-code-practices/', slug: 'AI/01-claude-code-practices' },
  { path: '/2026/04/03/00-vibe-coding-vs-spec-coding/', slug: 'AI/00-vibe-coding-vs-spec-coding' },
  { path: '/2026/04/03/02-agent-system-design-guid/', slug: 'AI/02-agent-system-design-guid' },
  { path: '/2026/04/03/03-dod-agent-design/', slug: 'AI/03-dod-agent-design' },
  { path: '/2026/04/07/system-design/21-ecommerce-product-center/', slug: 'system-design/21-ecommerce-product-center' },
  { path: '/2026/04/07/system-design/26-ecommerce-order-system/', slug: 'system-design/26-ecommerce-order-system' },
  { path: '/system-design/00-system-design-overview/', slug: 'system-design/00-system-design-overview' },
  { path: '/system-design/02-middleware-redis/', slug: 'system-design/02-middleware-redis' },
  { path: '/system-design/03-middleware-kafka/', slug: 'system-design/03-middleware-kafka' },
  { path: '/system-design/04-middleware-elasticsearch/', slug: 'system-design/04-middleware-elasticsearch' },
  { path: '/system-design/07-system-reliability-engineering/', slug: 'system-design/07-system-reliability-engineering' },
  { path: '/system-design/13-e-commerce/', slug: 'system-design/20-ecommerce-overview' },
  { path: '/system-design/18-inventory-system-design/', slug: 'system-design/22-ecommerce-inventory' },
  { path: '/system-design/20-ecommerce-overview/', slug: 'system-design/20-ecommerce-overview' },
  { path: '/system-design/21-ecommerce-product-center/', slug: 'system-design/21-ecommerce-product-center' },
  { path: '/system-design/22-ecommerce-inventory/', slug: 'system-design/22-ecommerce-inventory' },
  { path: '/system-design/23-ecommerce-marketing-system/', slug: 'system-design/23-ecommerce-marketing-system' },
  { path: '/system-design/24-ecommerce-pricing-engine/', slug: 'system-design/24-ecommerce-pricing-engine' },
  { path: '/system-design/25-ecommerce-pricing-ddd/', slug: 'system-design/25-ecommerce-pricing-ddd' },
  { path: '/system-design/26-ecommerce-order-system/', slug: 'system-design/26-ecommerce-order-system' },
  { path: '/system-design/27-ecommerce-payment-system/', slug: 'system-design/27-ecommerce-payment-system' },
  { path: '/system-design/28-ecommerce-listing/', slug: 'system-design/28-ecommerce-listing' },
  { path: '/system-design/29-ecommerce-b-side-ops/', slug: 'system-design/29-ecommerce-b-side-ops' },
  { path: '/system-design/30-ecommerce-product-lifecycle-management/', slug: 'system-design/30-ecommerce-product-lifecycle-management' },
  { path: '/system-design/31-ecommerce-search-discovery/', slug: 'system-design/31-ecommerce-search-discovery' },
  { path: '/system-design/32-ecommerce-cart-checkout/', slug: 'system-design/32-ecommerce-cart-checkout' },
  { path: '/system-design/34-ecommerce-long-transactions/', slug: 'system-design/34-ecommerce-long-transactions' },
  { path: '/system-design/41-acc-clean-arch-ddd-cqrs/', slug: 'system-design/41-acc-clean-arch-ddd-cqrs' },
  { path: '/system-design/42-acc-clean-code/', slug: 'system-design/42-acc-clean-code' },
  { path: '/system-design/43-acc-ddd-notes/', slug: 'system-design/43-acc-ddd-notes' },
  { path: '/system-design/44-acc-code-review/', slug: 'system-design/44-acc-code-review' }
];

function escapeAttribute(value) {
  return value
    .replaceAll('&', '&amp;')
    .replaceAll('"', '&quot;')
    .replaceAll('<', '&lt;')
    .replaceAll('>', '&gt;');
}

function asArray(collection) {
  if (Array.isArray(collection)) return collection;
  if (collection && typeof collection.toArray === 'function') return collection.toArray();
  return [];
}

function redirectHtml(targetUrl) {
  const escapedUrl = escapeAttribute(targetUrl);
  const javascriptUrl = JSON.stringify(targetUrl);

  return `<!doctype html>
<html lang="zh-CN">
<head>
  <meta charset="utf-8">
  <link rel="canonical" href="${escapedUrl}">
  <meta http-equiv="refresh" content="0; url=${escapedUrl}">
  <title>页面已迁移</title>
</head>
<body>
  <p>页面已迁移到 <a href="${escapedUrl}">${escapedUrl}</a>。</p>
  <script>location.replace(${javascriptUrl});</script>
</body>
</html>
`;
}

function registerLegacyPostAlias(hexoContext) {
  hexoContext.extend.generator.register('legacy-post-alias', function legacyPostAlias() {
    const posts = asArray(this.locals.get('posts'));

    return aliases.map((alias) => {
      const post = posts.find((candidate) => candidate.slug === alias.slug);
      if (!post) throw new Error(`Missing post for legacy alias: ${alias.slug}`);

      const targetUrl = url_for.call(this, post.path);
      return {
        path: `${alias.path.replace(/^\/+/, '')}index.html`,
        data: redirectHtml(targetUrl)
      };
    });
  });
}

if (typeof hexo !== 'undefined') registerLegacyPostAlias(hexo);

module.exports = { aliases, asArray, redirectHtml, registerLegacyPostAlias };
