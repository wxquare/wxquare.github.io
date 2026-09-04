'use strict';

const fs = require('fs');
const path = require('path');
const { URL } = require('url');

const SITE_ORIGIN = 'https://hexo-generated-site.invalid';
const HREF_PATTERN = /<a\b[^>]*\bhref\s*=\s*(?:"([^"]*)"|'([^']*)'|([^\s>]+))/gi;

function collectHtmlFiles(directory) {
  const files = [];

  for (const name of fs.readdirSync(directory)) {
    const file = path.join(directory, name);
    const stat = fs.statSync(file);

    if (stat.isDirectory()) files.push(...collectHtmlFiles(file));
    else if (stat.isFile() && name.endsWith('.html')) files.push(file);
  }

  return files;
}

function localTarget(publicDir, htmlFile, href) {
  const currentPath = `/${path.relative(publicDir, htmlFile).replaceAll(path.sep, '/')}`;
  const url = new URL(href, `${SITE_ORIGIN}${currentPath}`);
  const pathname = decodeURIComponent(url.pathname);
  const relativePath = pathname.replace(/^\/+/, '');
  const candidates = pathname.endsWith('/')
    ? [`${relativePath}index.html`]
    : [relativePath || 'index.html', `${relativePath}/index.html`, `${relativePath}.html`];
  const existing = candidates.find((candidate) => {
    const candidatePath = path.join(publicDir, ...candidate.split('/'));
    return fs.existsSync(candidatePath) && fs.statSync(candidatePath).isFile();
  });

  return {
    target: candidates[0],
    exists: Boolean(existing)
  };
}

function isIgnoredHref(href) {
  return !href || href.startsWith('#') || href.startsWith('//') ||
    /^(?:mailto|tel|javascript|data):/i.test(href) || /^https?:/i.test(href);
}

function isCheckableHref(href) {
  if (isIgnoredHref(href)) return false;

  const pathname = new URL(href, SITE_ORIGIN).pathname;
  const lastSegment = pathname.split('/').pop();
  return !lastSegment || !lastSegment.includes('.') || /\.html?$/i.test(lastSegment);
}

function isArticleHref(href) {
  if (!isCheckableHref(href)) return false;

  const pathname = new URL(href, SITE_ORIGIN).pathname;
  return /^\/(?:20\d{2}\/\d{2}\/\d{2}\/(?:AI|system-design|fundamentals|other)\/|system-design\/)/.test(pathname);
}

function checkGeneratedLinks(publicDir, options = {}) {
  const filter = options.filter || isCheckableHref;
  const files = collectHtmlFiles(publicDir);
  const failures = [];
  let checked = 0;

  for (const file of files) {
    const html = fs.readFileSync(file, 'utf8');
    let match;

    HREF_PATTERN.lastIndex = 0;
    while ((match = HREF_PATTERN.exec(html))) {
      const href = (match[1] ?? match[2] ?? match[3]).trim();
      if (!filter(href, file)) continue;

      checked += 1;
      const resolved = localTarget(publicDir, file, href);
      if (!resolved.exists) {
        failures.push({
          file: path.relative(publicDir, file).replaceAll(path.sep, '/'),
          href,
          target: resolved.target
        });
      }
    }
  }

  return { checked, failures };
}

if (require.main === module) {
  const publicDir = path.resolve(process.argv[2] || 'public');
  const result = checkGeneratedLinks(publicDir, { filter: isArticleHref });

  if (result.failures.length > 0) {
    console.error(`Found ${result.failures.length} broken local link(s) in ${publicDir}:`);
    for (const failure of result.failures) {
      console.error(`- ${failure.file}: ${failure.href} -> ${failure.target}`);
    }
    process.exitCode = 1;
  } else {
    console.log(`Checked ${result.checked} local article link(s); no broken targets found.`);
  }
}

module.exports = { checkGeneratedLinks, collectHtmlFiles, isArticleHref, isCheckableHref, localTarget };
