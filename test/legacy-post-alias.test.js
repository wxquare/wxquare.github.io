'use strict';

const assert = require('assert').strict;
const path = require('path');
const test = require('node:test');

const aliasScript = path.join(__dirname, '..', 'scripts/legacy-post-alias.js');

const expectedAliases = [
  ['/2025/05/15/system-design/14-system-reliability/', 'system-design/07-system-reliability-engineering'],
  ['/2026/04/03/00-vibe-coding-vs-spec-coding/', 'AI/00-vibe-coding-vs-spec-coding'],
  ['/2026/04/05/AI/04-karpathy-evolving-knowledge-base/', 'AI/02-karpathy-evolving-knowledge-base'],
  ['/2020/08/13/AI/初始OpenCL及在的移动端的一些测试数据/', 'AI/03-opencl-mobile-performance-testing'],
  ['/2026/09/22/AI/tensorflow-model-optimization/', 'AI/04-tensorflow-model-optimization'],
  ['/2026/09/22/AI/tvm-operator-optimization-practice/', 'AI/05-tvm-operator-optimization-practice'],
  ['/2020/08/13/AI/video-object-tracking/', 'AI/06-video-object-tracking'],
  ['/2024/03/01/1-os-fundamentals/', 'fundamentals/01-operating-system-fundamentals'],
  ['/2024/03/02/2-network-fundamentals/', 'fundamentals/02-computer-network-fundamentals'],
  ['/2024/03/03/3-bash-shell/', 'fundamentals/03-bash-shell-practice'],
  ['/2024/03/04/4-python-practice/', 'fundamentals/04-python-practice'],
  ['/2024/03/05/5-cpp-practice/', 'fundamentals/05-cpp-practice'],
  ['/2024/03/06/6-golang-practice/', 'fundamentals/06-go-practice'],
  ['/2024/03/04/01-middleware-mysql/', 'fundamentals/07-mysql-database'],
  ['/2024/03/06/02-middleware-redis/', 'fundamentals/08-redis'],
  ['/2024/03/10/03-middleware-kafka/', 'fundamentals/09-kafka'],
  ['/2024/03/07/04-middleware-elasticsearch/', 'fundamentals/10-elasticsearch'],
  ['/2024/12/20/05-infrastructure-k8s-docker/', 'fundamentals/11-docker-kubernetes'],
  ['/fundamentals/1-os-fundamentals/', 'fundamentals/01-operating-system-fundamentals'],
  ['/fundamentals/2-network-fundamentals/', 'fundamentals/02-computer-network-fundamentals'],
  ['/fundamentals/3-bash-shell/', 'fundamentals/03-bash-shell-practice'],
  ['/fundamentals/4-python-practice/', 'fundamentals/04-python-practice'],
  ['/fundamentals/5-cpp-practice/', 'fundamentals/05-cpp-practice'],
  ['/fundamentals/6-golang-practice/', 'fundamentals/06-go-practice'],
  ['/fundamentals/01-middleware-mysql/', 'fundamentals/07-mysql-database'],
  ['/fundamentals/02-middleware-redis/', 'fundamentals/08-redis'],
  ['/fundamentals/03-middleware-kafka/', 'fundamentals/09-kafka'],
  ['/fundamentals/04-middleware-elasticsearch/', 'fundamentals/10-elasticsearch'],
  ['/fundamentals/05-infrastructure-k8s-docker/', 'fundamentals/11-docker-kubernetes'],
  ['/system-design/02-middleware-redis/', 'fundamentals/08-redis'],
  ['/system-design/03-middleware-kafka/', 'fundamentals/09-kafka'],
  ['/system-design/04-middleware-elasticsearch/', 'fundamentals/10-elasticsearch'],
  ['/system-design/07-system-reliability-engineering/', 'system-design/07-system-reliability-engineering'],
  ['/system-design/41-acc-clean-arch-ddd-cqrs/', 'system-design/41-acc-clean-arch-ddd-cqrs'],
  ['/system-design/42-acc-clean-code/', 'system-design/42-acc-clean-code'],
  ['/system-design/44-acc-code-review/', 'system-design/44-acc-code-review']
];

const expectedUnavailableAliases = [
  ['/2026/04/02/01-claude-code-practices/', '/archives/'],
  ['/2026/04/03/02-agent-system-design-guid/', '/archives/'],
  ['/2026/04/03/03-dod-agent-design/', '/archives/']
];

const expectedMergedDddAliases = [
  ['/system-design/25-ecommerce-pricing-ddd/', '/2026/09/23/system-design/45-ddd-principles-and-pricing-practice/'],
  ['/system-design/43-acc-ddd-notes/', '/2026/09/23/system-design/45-ddd-principles-and-pricing-practice/']
];

const expectedBookAliases = [
  ['/system-design/00-system-design-overview/', '/books/system-design-primer/'],
  ['/2025/06/25/system-design/08-system-design-interview/', '/books/system-design-primer/appendix/system-design-interview-50.html'],
  ['/system-design/34-ecommerce-long-transactions/', '/books/system-design-primer/part01/04-large-transaction-orchestration.html'],
  ['/2026/06/09/system-design/34-ecommerce-long-transactions/', '/books/system-design-primer/part01/04-large-transaction-orchestration.html'],
  ['/system-design/30-ecommerce-product-lifecycle-management/', '/books/system-design-primer/part02/11-product-center-supply-lifecycle.html'],
  ['/2026/04/10/system-design/30-ecommerce-product-lifecycle-management/', '/books/system-design-primer/part02/11-product-center-supply-lifecycle.html'],
  ['/system-design/29-ecommerce-b-side-ops/', '/books/system-design-primer/part02/11-product-center-supply-lifecycle.html'],
  ['/2025/09/04/system-design/29-ecommerce-b-side-ops/', '/books/system-design-primer/part02/11-product-center-supply-lifecycle.html'],
  ['/system-design/28-ecommerce-listing/', '/books/system-design-primer/part02/11-product-center-supply-lifecycle.html'],
  ['/2025/08/21/system-design/28-ecommerce-listing/', '/books/system-design-primer/part02/11-product-center-supply-lifecycle.html'],
  ['/2026/04/07/system-design/21-ecommerce-product-center/', '/books/system-design-primer/part03/02-product-center.html'],
  ['/2026/04/07/system-design/26-ecommerce-order-system/', '/books/system-design-primer/part03/09-order-system.html'],
  ['/system-design/13-e-commerce/', '/books/system-design-primer/part03/01-ecommerce-overview.html'],
  ['/system-design/18-inventory-system-design/', '/books/system-design-primer/part03/04-inventory-system.html'],
  ['/system-design/20-ecommerce-overview/', '/books/system-design-primer/part03/01-ecommerce-overview.html'],
  ['/system-design/21-ecommerce-product-center/', '/books/system-design-primer/part03/02-product-center.html'],
  ['/system-design/22-ecommerce-inventory/', '/books/system-design-primer/part03/04-inventory-system.html'],
  ['/system-design/23-ecommerce-marketing-system/', '/books/system-design-primer/part03/05-marketing-system.html'],
  ['/system-design/24-ecommerce-pricing-engine/', '/books/system-design-primer/part03/06-pricing-system.html'],
  ['/system-design/26-ecommerce-order-system/', '/books/system-design-primer/part03/09-order-system.html'],
  ['/system-design/27-ecommerce-payment-system/', '/books/system-design-primer/part03/10-payment-system.html'],
  ['/system-design/31-ecommerce-search-discovery/', '/books/system-design-primer/part03/07-search-discovery.html'],
  ['/system-design/32-ecommerce-cart-checkout/', '/books/system-design-primer/part03/08-cart-checkout.html']
];

const expectedArchivedTvmAliases = [
  ['/2020/08/13/AI/tvm/TVM-GEMM-CPU/', '/2026/09/22/AI/05-tvm-operator-optimization-practice/'],
  ['/2020/08/13/AI/tvm/TVM-Graph-optimization/', '/2026/09/22/AI/05-tvm-operator-optimization-practice/'],
  ['/2020/08/13/AI/tvm/TVM-code-generation/', '/2026/09/22/AI/05-tvm-operator-optimization-practice/'],
  ['/2020/08/13/AI/tvm/TVM-hello/', '/2026/09/22/AI/05-tvm-operator-optimization-practice/'],
  ['/2020/08/13/AI/tvm/TVM-quantization/', '/2026/09/22/AI/05-tvm-operator-optimization-practice/'],
  ['/2020/08/13/AI/tvm/TVM-tutorial/', '/2026/09/22/AI/05-tvm-operator-optimization-practice/']
];

const expectedArchivedTrackingAliases = [
  ['/2020/08/13/AI/computer-vision/video-object-tracking/', '/2020/08/13/AI/06-video-object-tracking/'],
  ['/2020/08/13/AI/computer-vision/初始OpenCL及在的移动端的一些测试数据/', '/2020/08/13/AI/03-opencl-mobile-performance-testing/'],
  ['/2020/08/13/AI/computer-vision/DaSiamRPN/', '/2020/08/13/AI/06-video-object-tracking/'],
  ['/2020/08/13/AI/computer-vision/visp-template-tracker/', '/2020/08/13/AI/06-video-object-tracking/']
];

function mockConfig() {
  return {
    url: 'https://example.test',
    root: '/',
    relative_link: false,
    pretty_urls: { trailing_index: true, trailing_html: true }
  };
}

test('legacy post alias generator emits all compatibility redirects', () => {
  const registrations = [];
  const previousHexo = global.hexo;

  try {
    global.hexo = {
      extend: {
        generator: {
          register(name, fn) {
            registrations.push({ name, fn });
          }
        }
      }
    };

    delete require.cache[aliasScript];
    require(aliasScript);

    assert.equal(registrations.length, 1);
    assert.equal(registrations[0].name, 'legacy-post-alias');

    const currentPostPaths = new Map([
      ['AI/02-karpathy-evolving-knowledge-base', '2026/04/05/AI/02-karpathy-evolving-knowledge-base/'],
      ['AI/03-opencl-mobile-performance-testing', '2020/08/13/AI/03-opencl-mobile-performance-testing/'],
      ['AI/04-tensorflow-model-optimization', '2026/09/22/AI/04-tensorflow-model-optimization/'],
      ['AI/05-tvm-operator-optimization-practice', '2026/09/22/AI/05-tvm-operator-optimization-practice/'],
      ['AI/06-video-object-tracking', '2020/08/13/AI/06-video-object-tracking/']
    ]);
    const posts = expectedAliases.map(([, slug]) => ({
      slug,
      path: currentPostPaths.get(slug) || `2026/04/01/${slug}/`
    }));
    const context = {
      config: mockConfig(),
      locals: { get: (name) => name === 'posts' ? posts : undefined }
    };
    const generated = registrations[0].fn.call(context);

    assert.equal(
      generated.length,
      expectedAliases.length + expectedMergedDddAliases.length + expectedUnavailableAliases.length + expectedBookAliases.length + expectedArchivedTvmAliases.length + expectedArchivedTrackingAliases.length
    );
    assert.deepEqual(
      generated.map((item) => item.path).sort(),
      [...expectedAliases, ...expectedMergedDddAliases, ...expectedUnavailableAliases, ...expectedBookAliases, ...expectedArchivedTvmAliases, ...expectedArchivedTrackingAliases]
        .map(([route]) => `${route.slice(1)}index.html`)
        .sort()
    );

    for (const [route, target] of expectedMergedDddAliases) {
      const redirect = generated.find((item) => item.path === `${route.slice(1)}index.html`);
      assert.ok(redirect, `missing redirect for ${route}`);
      assert.match(redirect.data, new RegExp(target.replaceAll('/', '\\/')));
    }

    const redirect = generated.find((item) => item.path === 'system-design/13-e-commerce/index.html');
    assert.match(redirect.data, /rel="canonical"/);
    assert.match(redirect.data, /\/books\/system-design-primer\/part03\/01-ecommerce-overview\.html/);
    assert.match(redirect.data, /location\.replace/);
  } finally {
    delete require.cache[aliasScript];
    if (previousHexo === undefined) delete global.hexo;
    else global.hexo = previousHexo;
  }
});

test('legacy post alias generator fails when a target post is missing', () => {
  const registrations = [];
  const previousHexo = global.hexo;

  try {
    global.hexo = {
      extend: {
        generator: {
          register(name, fn) {
            registrations.push({ name, fn });
          }
        }
      }
    };

    delete require.cache[aliasScript];
    require(aliasScript);

    const posts = expectedAliases.filter(([, slug]) => slug !== 'system-design/07-system-reliability-engineering').map(([, slug]) => ({
      slug,
      path: `2026/04/01/${slug}/`
    }));
    const context = {
      config: mockConfig(),
      locals: { get: () => posts }
    };

    assert.throws(
      () => registrations[0].fn.call(context),
      /Missing post for legacy alias: system-design\/07-system-reliability-engineering/
    );
  } finally {
    delete require.cache[aliasScript];
    if (previousHexo === undefined) delete global.hexo;
    else global.hexo = previousHexo;
  }
});
