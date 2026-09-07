'use strict';

const assert = require('assert').strict;
const path = require('path');
const test = require('node:test');

const aliasScript = path.join(__dirname, '..', 'scripts/legacy-post-alias.js');

const expectedAliases = [
  ['/2025/05/15/system-design/14-system-reliability/', 'system-design/07-system-reliability-engineering'],
  ['/2025/06/25/system-design/08-system-design-interview/', 'system-design/08-system-design-interview'],
  ['/2026/04/02/01-claude-code-practices/', 'AI/01-claude-code-practices'],
  ['/2026/04/03/00-vibe-coding-vs-spec-coding/', 'AI/00-vibe-coding-vs-spec-coding'],
  ['/2026/04/03/02-agent-system-design-guid/', 'AI/02-agent-system-design-guid'],
  ['/2026/04/03/03-dod-agent-design/', 'AI/03-dod-agent-design'],
  ['/2026/04/07/system-design/21-ecommerce-product-center/', 'system-design/21-ecommerce-product-center'],
  ['/2026/04/07/system-design/26-ecommerce-order-system/', 'system-design/26-ecommerce-order-system'],
  ['/system-design/00-system-design-overview/', 'system-design/00-system-design-overview'],
  ['/system-design/02-middleware-redis/', 'system-design/02-middleware-redis'],
  ['/system-design/03-middleware-kafka/', 'system-design/03-middleware-kafka'],
  ['/system-design/04-middleware-elasticsearch/', 'system-design/04-middleware-elasticsearch'],
  ['/system-design/07-system-reliability-engineering/', 'system-design/07-system-reliability-engineering'],
  ['/system-design/13-e-commerce/', 'system-design/20-ecommerce-overview'],
  ['/system-design/18-inventory-system-design/', 'system-design/22-ecommerce-inventory'],
  ['/system-design/20-ecommerce-overview/', 'system-design/20-ecommerce-overview'],
  ['/system-design/21-ecommerce-product-center/', 'system-design/21-ecommerce-product-center'],
  ['/system-design/22-ecommerce-inventory/', 'system-design/22-ecommerce-inventory'],
  ['/system-design/23-ecommerce-marketing-system/', 'system-design/23-ecommerce-marketing-system'],
  ['/system-design/24-ecommerce-pricing-engine/', 'system-design/24-ecommerce-pricing-engine'],
  ['/system-design/25-ecommerce-pricing-ddd/', 'system-design/25-ecommerce-pricing-ddd'],
  ['/system-design/26-ecommerce-order-system/', 'system-design/26-ecommerce-order-system'],
  ['/system-design/27-ecommerce-payment-system/', 'system-design/27-ecommerce-payment-system'],
  ['/system-design/28-ecommerce-listing/', 'system-design/28-ecommerce-listing'],
  ['/system-design/29-ecommerce-b-side-ops/', 'system-design/29-ecommerce-b-side-ops'],
  ['/system-design/30-ecommerce-product-lifecycle-management/', 'system-design/30-ecommerce-product-lifecycle-management'],
  ['/system-design/31-ecommerce-search-discovery/', 'system-design/31-ecommerce-search-discovery'],
  ['/system-design/32-ecommerce-cart-checkout/', 'system-design/32-ecommerce-cart-checkout'],
  ['/system-design/34-ecommerce-long-transactions/', 'system-design/34-ecommerce-long-transactions'],
  ['/system-design/41-acc-clean-arch-ddd-cqrs/', 'system-design/41-acc-clean-arch-ddd-cqrs'],
  ['/system-design/42-acc-clean-code/', 'system-design/42-acc-clean-code'],
  ['/system-design/43-acc-ddd-notes/', 'system-design/43-acc-ddd-notes'],
  ['/system-design/44-acc-code-review/', 'system-design/44-acc-code-review']
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

    const posts = expectedAliases.map(([, slug]) => ({
      slug,
      path: `2026/04/01/${slug}/`
    }));
    const context = {
      config: mockConfig(),
      locals: { get: (name) => name === 'posts' ? posts : undefined }
    };
    const generated = registrations[0].fn.call(context);

    assert.equal(generated.length, expectedAliases.length);
    assert.deepEqual(generated.map((item) => item.path), expectedAliases.map(([route]) => `${route.slice(1)}index.html`));

    const redirect = generated.find((item) => item.path === 'system-design/13-e-commerce/index.html');
    assert.match(redirect.data, /rel="canonical"/);
    assert.match(redirect.data, /\/2026\/04\/01\/system-design\/20-ecommerce-overview\//);
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
