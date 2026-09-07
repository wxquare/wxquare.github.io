# Categorized Blog Homepage Design

**Goal:** Make the site root a topic-oriented blog entry page. Readers should see the latest articles grouped by the four registered blog categories instead of a single reverse-chronological stream.

## Scope

The homepage will show, in this fixed order:

1. AI 与 Agent
2. 系统设计基础
3. 计算机基础
4. other

Each section contains its five most recent published posts. Every item shows the publication date, title, and tags; the section ends with a link to the category archive. A compact book-entry area links to the published AI Agent and system-design books without including book chapters in the blog lists.

The existing category archive, post URLs, pagination paths, theme, and article source directories remain unchanged. The work does not delete or rewrite any articles.

## Design

`source/home/index.md` remains the canonical root page and declares that it needs the categorized homepage content. A small root `scripts/` Hexo plugin prepares that page's content at generation time from Hexo's published post collection.

The plugin uses `.agents/config/post-categories.json` as the category registry. It maps each post's front-matter category to a registry category, excludes drafts and unpublished entries as Hexo does, sorts each group by date descending, and takes five posts per group. It produces semantic HTML in the page content, using each post's generated path and title rather than constructing URLs from filenames. This keeps links correct if permalink rules change.

The generated homepage has three visual units:

- A short introduction and two book links.
- Four category sections, each with a heading, five linked post rows, and a “查看全部” link to the category archive.
- A clear empty-state message only if a registered category has no published posts.

Presentation rules live in `source/_data/styles.styl`, which is the repository's existing source-level styling extension. No files under `themes/next/` are edited. The layout must remain readable on narrow screens: post metadata may wrap below the title, while each category heading and its archive link stay associated.

The existing Hexo index generator is disabled for the root path so it does not compete with `source/home/index.md`. If a chronological all-posts entry is wanted later, it can be introduced deliberately as an archive-style page; it is outside this change.

## Data Flow

```text
source/_posts front matter + post dates
        +
post-categories.json registry
        ↓
homepage Hexo plugin
        ↓
source/home/index.md rendered as /
        ↓
four category sections + book entry links
```

## Error Handling and Compatibility

- A post whose category cannot be mapped to the registry is omitted from the homepage and produces a build warning with its source path. It remains available through its normal post URL.
- Missing category archives do not stop generation; the plugin uses Hexo's configured category URL helper rather than hardcoded encoded paths.
- HTML text derived from titles, tags, and category labels is escaped before rendering.
- No post URL is removed or changed, so redirects are not required.

## Verification

- Add focused Node tests for category mapping, date ordering, five-item limits, escaping, and unmapped-category warnings.
- Run `npm test`.
- Run `npm run clean` and `npm run build`.
- Inspect generated `public/index.html` to confirm all four headings, five-or-fewer posts per group, category links, book links, and no duplicate root output.
- Check the page at desktop and narrow viewport widths with the local Hexo server.

## Non-goals

- Reorganizing or deleting duplicate blog and book content.
- Altering NexT templates or third-party theme files.
- Adding filters, search, article summaries, a separate all-posts archive, or a new CMS-like category editor.
