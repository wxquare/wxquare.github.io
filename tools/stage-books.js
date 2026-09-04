'use strict';

// Repository CLI utility: copies generated mdBook output into Hexo's public tree.
// It solves the deployment layout problem without editing generated files by hand.
// Invoke it with `npm run stage:books`; it is not a Hexo generator plugin.

const fs = require('node:fs');
const path = require('node:path');

const BOOK_OUTPUTS = [
  ['books/ai-book/book', 'public/ai-book'],
  ['books/system-design-architecture-book/book', 'public/system-design-architecture-book']
];

function copyDirectory(sourceDir, destinationDir) {
  fs.mkdirSync(destinationDir, { recursive: true });

  for (const entry of fs.readdirSync(sourceDir, { withFileTypes: true })) {
    const sourcePath = path.join(sourceDir, entry.name);
    const destinationPath = path.join(destinationDir, entry.name);

    if (entry.isDirectory()) {
      copyDirectory(sourcePath, destinationPath);
    } else if (entry.isSymbolicLink()) {
      fs.symlinkSync(fs.readlinkSync(sourcePath), destinationPath);
    } else {
      fs.copyFileSync(sourcePath, destinationPath);
    }
  }
}

function stageBooks(repoRoot = path.resolve(__dirname, '..')) {
  for (const [sourceRelative, destinationRelative] of BOOK_OUTPUTS) {
    const destinationDir = path.join(repoRoot, destinationRelative);
    fs.rmSync(destinationDir, { recursive: true, force: true });
    copyDirectory(
      path.join(repoRoot, sourceRelative),
      destinationDir
    );
  }
}

if (require.main === module) {
  stageBooks();
}

module.exports = { BOOK_OUTPUTS, stageBooks };
