'use strict';

function escapeAttribute(value) {
  return value
    .replaceAll('&', '&amp;')
    .replaceAll('"', '&quot;')
    .replaceAll('<', '&lt;')
    .replaceAll('>', '&gt;');
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

function registerLegacyBookAlias(hexoContext) {
  hexoContext.extend.generator.register('legacy-book-alias', () => [{
    path: 'ecommerce-book/index.html',
    data: redirectHtml('/booklist/')
  }]);
}

if (typeof hexo !== 'undefined') registerLegacyBookAlias(hexo);

module.exports = { redirectHtml, registerLegacyBookAlias };
