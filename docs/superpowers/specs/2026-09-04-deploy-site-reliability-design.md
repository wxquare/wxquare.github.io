# Deploy Site Reliability Design

**Date:** 2026-09-04
**Status:** Approved for implementation by the repository owner

## Context

The GitHub Actions run for `Deploy Site` completed successfully in about 7 minutes 53 seconds, but emitted repeated warnings that its JavaScript actions were running on the deprecated Node 20 runtime. The workflow currently uses `actions/checkout@v4`, `actions/setup-node@v4`, `peaceiris/actions-mdbook@v1`, and `peaceiris/actions-gh-pages@v3`.

The deployment quality gate also has two gaps:

- The workflow builds and publishes the site without running the repository test suite.
- `npm run check:links` invokes the link checker with the article-only filter. A full generated-site scan found 81 broken local targets: 75 malformed mdBook redirect targets, two obsolete book-list routes, and four missing library directory index pages.

The existing `hexo` worktree contains unrelated, uncommitted article, image, and diagram changes. This work must not modify or stage those changes.

## Goals

1. Remove the GitHub Actions Node 20 runtime warnings before Node 20 is removed from hosted runners.
2. Make CI validate the same JavaScript tests that are required by the local pre-commit check.
3. Prevent an older deployment from publishing after a newer push to the same branch.
4. Make the generated-site link check cover all checkable local HTML-style routes.
5. Fix the currently detected broken generated routes without changing the public deployment model.
6. Keep the changes limited to deployment, generated-link validation, book redirect configuration, and the affected navigation pages.

## Non-goals

- Do not migrate from the `master` publishing branch to the official Pages artifact/deploy actions; that would require a repository Settings change and is a separate project.
- Do not update unrelated npm dependencies or redesign the Hexo theme.
- Do not restore, delete, or reformat the existing uncommitted content changes.
- Do not commit generated `public/` or mdBook `book/` output.

## Design

### 1. Workflow runtime and execution order

Keep one sequential deployment job and retain `contents: write`, because the existing `peaceiris/actions-gh-pages` deployment pushes the generated tree to the repository's `master` branch.

Update the action references and build runtime to the current Node 24-compatible versions:

```yaml
concurrency:
  group: deploy-site-${{ github.ref }}
  cancel-in-progress: true

steps:
  - uses: actions/checkout@v7
  - uses: actions/setup-node@v7
    with:
      node-version: 24
      cache: npm
  - run: npm ci
  - run: npm test
  - uses: peaceiris/actions-mdbook@v2
    with:
      mdbook-version: '0.5.2'
  - run: npm run clean && npm run build
  - run: npm run build:books
  - run: npm run stage:books
  - run: npm run check:links
  - uses: peaceiris/actions-gh-pages@v4
```

The deployment step remains last, so a test, site build, book build, staging, or link-check failure prevents publication. Concurrency cancellation ensures that a newer push supersedes an older in-progress build for the same ref.

### 2. Generated-link quality gate

Change the command-line entry point in `scripts/check-generated-links.js` to use the existing general `isCheckableHref` filter instead of `isArticleHref`. Keep `isArticleHref` exported because it remains useful for focused unit tests and diagnostics.

The general filter continues to ignore fragments, external URLs, mail/telephone links, scripts, data URLs, and static assets with file extensions that the checker cannot resolve. It checks local routes that resolve to an existing path, `index.html`, or `.html` file. The expected post-fix result is zero failures for the complete generated `public/` tree.

### 3. Broken route fixes

- In `books/system-design-architecture-book/book.toml`, make every redirect destination relative to the generated redirect file's directory by prefixing the destination with `../`. The old aliases are all one directory below the book root, so this preserves the aliases while making browser resolution correct.
- In `source/booklist/index.md`, remove the two obsolete `/system-design-book/` and `/ecommerce-book/` entries. The consolidated system-design book is already linked at `/system-design-architecture-book/`; the AI book remains linked at `/ai-book/`.
- Update the DaSiamRPN local README reference in `source/library/index.md` from the source-only `.md` route to the generated `README.html` route.
- Add minimal generated index pages for `source/library/other/` and `source/library/other/DaSiamRPN/`, so the theme's breadcrumb links resolve without changing the third-party material itself.

### 4. Tests and verification

Update `test/build-entrypoints.test.js` to assert the new action versions, Node 24 runtime, concurrency block, and the presence/order of `npm test`. Add or update focused tests so the CLI link-check behavior is covered by an observable result rather than only by source-text matching.

The final verification sequence is:

```bash
npm test
npm run clean
npm run build
npm run build:books
npm run stage:books
npm run check:links
```

The generated-site check must report zero broken local targets. The final diff must contain only the approved files and must not include `public/`, mdBook output, or any pre-existing unrelated worktree changes.

## Failure handling and rollback

The workflow stays fail-fast through the existing shell command behavior. Because deployment is the final step, failed validation cannot update the publishing branch. If the Node 24 runtime exposes an incompatibility, the implementation commit can be reverted as one unit; no generated artifacts or repository Settings changes are required to roll back.

## Acceptance criteria

- The workflow no longer references the deprecated action majors or Node 20 runtime.
- CI runs `npm test` before any generated output is published.
- Concurrent pushes to `hexo` share a cancellation group.
- The complete generated local-link scan passes with zero failures.
- The consolidated book navigation and DaSiamRPN breadcrumb navigation resolve to existing generated files.
- Existing unrelated worktree changes remain unmodified and unstaged.
