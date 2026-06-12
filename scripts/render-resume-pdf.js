const fs = require('fs');
const path = require('path');
const { marked } = require('marked');

const input = process.argv[2] || 'source/about/resume-backend-architect.md';
const outputDir = process.argv[3] || 'outputs';

const root = process.cwd();
const inputPath = path.resolve(root, input);
const outDir = path.resolve(root, outputDir);
const base = path.basename(inputPath, path.extname(inputPath));
const htmlPath = path.join(outDir, `${base}.html`);

const md = fs.readFileSync(inputPath, 'utf8');

marked.setOptions({
  breaks: true,
  gfm: true,
});

const body = marked.parse(md);

const html = `<!doctype html>
<html lang="zh-CN">
<head>
  <meta charset="utf-8">
  <meta name="viewport" content="width=device-width, initial-scale=1">
  <title>${base}</title>
  <style>
    @page {
      size: A4;
      margin: 8mm 8.5mm;
    }

    * {
      box-sizing: border-box;
    }

    html,
    body {
      margin: 0;
      padding: 0;
      background: #fff;
      color: #111827;
      font-family: -apple-system, BlinkMacSystemFont, "PingFang SC", "Hiragino Sans GB", "Microsoft YaHei", "Noto Sans CJK SC", Arial, sans-serif;
      font-size: 8.9pt;
      line-height: 1.26;
    }

    body {
      width: 193mm;
      min-height: 281mm;
      margin: 0 auto;
    }

    .resume {
      width: 100%;
    }

    h1 {
      margin: 0 0 2.2mm;
      color: #0f172a;
      font-size: 18pt;
      line-height: 1;
      text-align: center;
      letter-spacing: 0;
    }

    h1 + p {
      margin: 0 0 3mm;
      padding-bottom: 2.2mm;
      border-bottom: 1px solid #cbd5e1;
      color: #334155;
      font-size: 8.8pt;
      line-height: 1.25;
      text-align: center;
      white-space: normal;
    }

    h2 {
      margin: 3.2mm 0 1.6mm;
      padding-bottom: 0.8mm;
      border-bottom: 1px solid #e2e8f0;
      color: #0f172a;
      font-size: 10.7pt;
      line-height: 1.1;
    }

    h3 {
      margin: 2.1mm 0 1.4mm;
      color: #111827;
      font-size: 9.2pt;
      line-height: 1.2;
    }

    p {
      margin: 0 0 1.5mm;
    }

    ul {
      margin: 0.6mm 0 1.8mm 0;
      padding-left: 4.8mm;
    }

    li {
      margin: 0 0 0.9mm;
      padding-left: 0.3mm;
    }

    strong {
      color: #0f172a;
      font-weight: 700;
    }

    hr {
      display: none;
    }

    a {
      color: inherit;
      text-decoration: none;
    }

    .resume > h2:first-of-type + p {
      color: #1f2937;
      font-size: 8.45pt;
      line-height: 1.24;
    }

    .resume > h2:nth-of-type(2) + h3 {
      margin-top: 1.3mm;
    }

    @media print {
      html,
      body {
        width: auto;
        min-height: auto;
      }

      .resume {
        break-after: avoid;
      }
    }
  </style>
</head>
<body>
  <main class="resume">
${body}
  </main>
</body>
</html>
`;

fs.mkdirSync(outDir, { recursive: true });
fs.writeFileSync(htmlPath, html);
console.log(htmlPath);
