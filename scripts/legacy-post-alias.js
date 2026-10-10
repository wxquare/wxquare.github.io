'use strict';

const { url_for } = require('hexo-util');

const aliases = [
  { path: '/2025/06/25/system-design/08-system-design-interview/', target: '/reliable-system-design/' },
  { path: '/2026/04/02/01-claude-code-practices/', target: '/archives/' },
  { path: '/2026/04/03/02-agent-system-design-guid/', target: '/archives/' },
  { path: '/2026/04/03/03-dod-agent-design/', target: '/archives/' },
  { path: '/2026/09/24/AI/2026-09-24-superpowers-source-workflow-practice/', slug: 'AI/03-superpowers-source-workflow-practice' },
  { path: '/2026/04/07/system-design/21-ecommerce-product-center/', target: '/reliable-system-design/part03/02-product-center.html' },
  { path: '/2026/04/07/system-design/26-ecommerce-order-system/', target: '/reliable-system-design/part03/09-order-system.html' },
  { path: '/system-design/00-system-design-overview/', target: '/reliable-system-design/' },
  { path: '/2024/03/01/1-os-fundamentals/', slug: 'fundamentals/01-operating-system-fundamentals' },
  { path: '/2024/03/02/2-network-fundamentals/', slug: 'fundamentals/02-computer-network-fundamentals' },
  { path: '/2024/03/03/3-bash-shell/', slug: 'fundamentals/03-bash-shell-practice' },
  { path: '/2024/03/04/4-python-practice/', slug: 'fundamentals/04-python-practice' },
  { path: '/2024/03/05/5-cpp-practice/', slug: 'fundamentals/05-cpp-practice' },
  { path: '/2024/03/06/6-golang-practice/', slug: 'fundamentals/06-go-practice' },
  { path: '/fundamentals/1-os-fundamentals/', slug: 'fundamentals/01-operating-system-fundamentals' },
  { path: '/fundamentals/2-network-fundamentals/', slug: 'fundamentals/02-computer-network-fundamentals' },
  { path: '/fundamentals/3-bash-shell/', slug: 'fundamentals/03-bash-shell-practice' },
  { path: '/fundamentals/4-python-practice/', slug: 'fundamentals/04-python-practice' },
  { path: '/fundamentals/5-cpp-practice/', slug: 'fundamentals/05-cpp-practice' },
  { path: '/fundamentals/6-golang-practice/', slug: 'fundamentals/06-go-practice' },
  { path: '/system-design/13-e-commerce/', target: '/reliable-system-design/part03/01-ecommerce-overview.html' },
  { path: '/system-design/18-inventory-system-design/', target: '/reliable-system-design/part03/04-inventory-system.html' },
  { path: '/system-design/20-ecommerce-overview/', target: '/reliable-system-design/part03/01-ecommerce-overview.html' },
  { path: '/system-design/21-ecommerce-product-center/', target: '/reliable-system-design/part03/02-product-center.html' },
  { path: '/system-design/22-ecommerce-inventory/', target: '/reliable-system-design/part03/04-inventory-system.html' },
  { path: '/system-design/23-ecommerce-marketing-system/', target: '/reliable-system-design/part03/05-marketing-system.html' },
  { path: '/system-design/24-ecommerce-pricing-engine/', target: '/reliable-system-design/part03/06-pricing-system.html' },
  { path: '/system-design/26-ecommerce-order-system/', target: '/reliable-system-design/part03/09-order-system.html' },
  { path: '/system-design/27-ecommerce-payment-system/', target: '/reliable-system-design/part03/10-payment-system.html' },
  { path: '/system-design/28-ecommerce-listing/', target: '/reliable-system-design/part02/11-product-center-supply-lifecycle.html' },
  { path: '/2025/08/21/system-design/28-ecommerce-listing/', target: '/reliable-system-design/part02/11-product-center-supply-lifecycle.html' },
  { path: '/system-design/29-ecommerce-b-side-ops/', target: '/reliable-system-design/part02/11-product-center-supply-lifecycle.html' },
  { path: '/2025/09/04/system-design/29-ecommerce-b-side-ops/', target: '/reliable-system-design/part02/11-product-center-supply-lifecycle.html' },
  { path: '/system-design/30-ecommerce-product-lifecycle-management/', target: '/reliable-system-design/part02/11-product-center-supply-lifecycle.html' },
  { path: '/2026/04/10/system-design/30-ecommerce-product-lifecycle-management/', target: '/reliable-system-design/part02/11-product-center-supply-lifecycle.html' },
  { path: '/system-design/31-ecommerce-search-discovery/', target: '/reliable-system-design/part03/07-search-discovery.html' },
  { path: '/system-design/32-ecommerce-cart-checkout/', target: '/reliable-system-design/part03/08-cart-checkout.html' },
  { path: '/system-design/34-ecommerce-long-transactions/', target: '/reliable-system-design/part01/04-large-transaction-orchestration.html' },
  { path: '/2026/06/09/system-design/34-ecommerce-long-transactions/', target: '/reliable-system-design/part01/04-large-transaction-orchestration.html' }
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
      const post = alias.slug && posts.find((candidate) => candidate.slug === alias.slug);
      if (alias.slug && !post) throw new Error(`Missing post for legacy alias: ${alias.slug}`);

      const targetUrl = alias.target || url_for.call(this, post.path);
      return {
        path: `${alias.path.replace(/^\/+/, '')}index.html`,
        data: redirectHtml(targetUrl)
      };
    });
  });
}

if (typeof hexo !== 'undefined') registerLegacyPostAlias(hexo);

module.exports = { aliases, asArray, redirectHtml, registerLegacyPostAlias };
