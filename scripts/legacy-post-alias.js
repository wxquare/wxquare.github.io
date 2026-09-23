'use strict';

const { url_for } = require('hexo-util');

const aliases = [
  { path: '/2025/05/15/system-design/14-system-reliability/', slug: 'system-design/07-system-reliability-engineering' },
  { path: '/2025/06/25/system-design/08-system-design-interview/', target: '/books/system-design-primer/appendix/system-design-interview-50.html' },
  { path: '/2026/04/02/01-claude-code-practices/', target: '/archives/' },
  { path: '/2026/04/03/00-vibe-coding-vs-spec-coding/', slug: 'AI/00-vibe-coding-vs-spec-coding' },
  { path: '/2026/04/03/02-agent-system-design-guid/', target: '/archives/' },
  { path: '/2026/04/03/03-dod-agent-design/', target: '/archives/' },
  { path: '/2026/04/05/AI/04-karpathy-evolving-knowledge-base/', slug: 'AI/02-karpathy-evolving-knowledge-base' },
  { path: '/2020/08/13/AI/初始OpenCL及在的移动端的一些测试数据/', slug: 'AI/03-opencl-mobile-performance-testing' },
  { path: '/2026/09/22/AI/tensorflow-model-optimization/', slug: 'AI/04-tensorflow-model-optimization' },
  { path: '/2026/09/22/AI/tvm-operator-optimization-practice/', slug: 'AI/05-tvm-operator-optimization-practice' },
  { path: '/2020/08/13/AI/video-object-tracking/', slug: 'AI/06-video-object-tracking' },
  { path: '/2026/04/07/system-design/21-ecommerce-product-center/', target: '/books/system-design-primer/part03/02-product-center.html' },
  { path: '/2026/04/07/system-design/26-ecommerce-order-system/', target: '/books/system-design-primer/part03/09-order-system.html' },
  { path: '/system-design/00-system-design-overview/', target: '/books/system-design-primer/' },
  { path: '/2024/03/01/1-os-fundamentals/', slug: 'fundamentals/01-operating-system-fundamentals' },
  { path: '/2024/03/02/2-network-fundamentals/', slug: 'fundamentals/02-computer-network-fundamentals' },
  { path: '/2024/03/03/3-bash-shell/', slug: 'fundamentals/03-bash-shell-practice' },
  { path: '/2024/03/04/4-python-practice/', slug: 'fundamentals/04-python-practice' },
  { path: '/2024/03/05/5-cpp-practice/', slug: 'fundamentals/05-cpp-practice' },
  { path: '/2024/03/06/6-golang-practice/', slug: 'fundamentals/06-go-practice' },
  { path: '/2024/03/04/01-middleware-mysql/', slug: 'fundamentals/07-mysql-database' },
  { path: '/2024/03/06/02-middleware-redis/', slug: 'fundamentals/08-redis' },
  { path: '/2024/03/10/03-middleware-kafka/', slug: 'fundamentals/09-kafka' },
  { path: '/2024/03/07/04-middleware-elasticsearch/', slug: 'fundamentals/10-elasticsearch' },
  { path: '/2024/12/20/05-infrastructure-k8s-docker/', slug: 'fundamentals/11-docker-kubernetes' },
  { path: '/fundamentals/1-os-fundamentals/', slug: 'fundamentals/01-operating-system-fundamentals' },
  { path: '/fundamentals/2-network-fundamentals/', slug: 'fundamentals/02-computer-network-fundamentals' },
  { path: '/fundamentals/3-bash-shell/', slug: 'fundamentals/03-bash-shell-practice' },
  { path: '/fundamentals/4-python-practice/', slug: 'fundamentals/04-python-practice' },
  { path: '/fundamentals/5-cpp-practice/', slug: 'fundamentals/05-cpp-practice' },
  { path: '/fundamentals/6-golang-practice/', slug: 'fundamentals/06-go-practice' },
  { path: '/fundamentals/01-middleware-mysql/', slug: 'fundamentals/07-mysql-database' },
  { path: '/fundamentals/02-middleware-redis/', slug: 'fundamentals/08-redis' },
  { path: '/fundamentals/03-middleware-kafka/', slug: 'fundamentals/09-kafka' },
  { path: '/fundamentals/04-middleware-elasticsearch/', slug: 'fundamentals/10-elasticsearch' },
  { path: '/fundamentals/05-infrastructure-k8s-docker/', slug: 'fundamentals/11-docker-kubernetes' },
  { path: '/system-design/02-middleware-redis/', slug: 'fundamentals/08-redis' },
  { path: '/system-design/03-middleware-kafka/', slug: 'fundamentals/09-kafka' },
  { path: '/system-design/04-middleware-elasticsearch/', slug: 'fundamentals/10-elasticsearch' },
  { path: '/system-design/07-system-reliability-engineering/', slug: 'system-design/07-system-reliability-engineering' },
  { path: '/system-design/13-e-commerce/', target: '/books/system-design-primer/part03/01-ecommerce-overview.html' },
  { path: '/system-design/18-inventory-system-design/', target: '/books/system-design-primer/part03/04-inventory-system.html' },
  { path: '/system-design/20-ecommerce-overview/', target: '/books/system-design-primer/part03/01-ecommerce-overview.html' },
  { path: '/system-design/21-ecommerce-product-center/', target: '/books/system-design-primer/part03/02-product-center.html' },
  { path: '/system-design/22-ecommerce-inventory/', target: '/books/system-design-primer/part03/04-inventory-system.html' },
  { path: '/system-design/23-ecommerce-marketing-system/', target: '/books/system-design-primer/part03/05-marketing-system.html' },
  { path: '/system-design/24-ecommerce-pricing-engine/', target: '/books/system-design-primer/part03/06-pricing-system.html' },
  { path: '/system-design/25-ecommerce-pricing-ddd/', target: '/2026/09/23/system-design/45-ddd-principles-and-pricing-practice/' },
  { path: '/system-design/26-ecommerce-order-system/', target: '/books/system-design-primer/part03/09-order-system.html' },
  { path: '/system-design/27-ecommerce-payment-system/', target: '/books/system-design-primer/part03/10-payment-system.html' },
  { path: '/system-design/28-ecommerce-listing/', target: '/books/system-design-primer/part02/11-product-center-supply-lifecycle.html' },
  { path: '/2025/08/21/system-design/28-ecommerce-listing/', target: '/books/system-design-primer/part02/11-product-center-supply-lifecycle.html' },
  { path: '/system-design/29-ecommerce-b-side-ops/', target: '/books/system-design-primer/part02/11-product-center-supply-lifecycle.html' },
  { path: '/2025/09/04/system-design/29-ecommerce-b-side-ops/', target: '/books/system-design-primer/part02/11-product-center-supply-lifecycle.html' },
  { path: '/system-design/30-ecommerce-product-lifecycle-management/', target: '/books/system-design-primer/part02/11-product-center-supply-lifecycle.html' },
  { path: '/2026/04/10/system-design/30-ecommerce-product-lifecycle-management/', target: '/books/system-design-primer/part02/11-product-center-supply-lifecycle.html' },
  { path: '/system-design/31-ecommerce-search-discovery/', target: '/books/system-design-primer/part03/07-search-discovery.html' },
  { path: '/system-design/32-ecommerce-cart-checkout/', target: '/books/system-design-primer/part03/08-cart-checkout.html' },
  { path: '/system-design/34-ecommerce-long-transactions/', target: '/books/system-design-primer/part01/04-large-transaction-orchestration.html' },
  { path: '/2026/06/09/system-design/34-ecommerce-long-transactions/', target: '/books/system-design-primer/part01/04-large-transaction-orchestration.html' },
  { path: '/system-design/41-acc-clean-arch-ddd-cqrs/', slug: 'system-design/41-acc-clean-arch-ddd-cqrs' },
  { path: '/system-design/42-acc-clean-code/', slug: 'system-design/42-acc-clean-code' },
  { path: '/system-design/43-acc-ddd-notes/', target: '/2026/09/23/system-design/45-ddd-principles-and-pricing-practice/' },
  { path: '/system-design/44-acc-code-review/', slug: 'system-design/44-acc-code-review' }
  ,{ path: '/2020/08/13/AI/tvm/TVM-GEMM-CPU/', target: '/2026/09/22/AI/05-tvm-operator-optimization-practice/' }
  ,{ path: '/2020/08/13/AI/tvm/TVM-Graph-optimization/', target: '/2026/09/22/AI/05-tvm-operator-optimization-practice/' }
  ,{ path: '/2020/08/13/AI/tvm/TVM-code-generation/', target: '/2026/09/22/AI/05-tvm-operator-optimization-practice/' }
  ,{ path: '/2020/08/13/AI/tvm/TVM-hello/', target: '/2026/09/22/AI/05-tvm-operator-optimization-practice/' }
  ,{ path: '/2020/08/13/AI/tvm/TVM-quantization/', target: '/2026/09/22/AI/05-tvm-operator-optimization-practice/' }
  ,{ path: '/2020/08/13/AI/tvm/TVM-tutorial/', target: '/2026/09/22/AI/05-tvm-operator-optimization-practice/' }
  ,{ path: '/2020/08/13/AI/computer-vision/video-object-tracking/', slug: 'AI/06-video-object-tracking' }
  ,{ path: '/2020/08/13/AI/computer-vision/初始OpenCL及在的移动端的一些测试数据/', slug: 'AI/03-opencl-mobile-performance-testing' }
  ,{ path: '/2020/08/13/AI/computer-vision/DaSiamRPN/', slug: 'AI/06-video-object-tracking' }
  ,{ path: '/2020/08/13/AI/computer-vision/visp-template-tracker/', slug: 'AI/06-video-object-tracking' }
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
