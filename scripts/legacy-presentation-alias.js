'use strict';

const fs = require('fs');
const path = require('path');

hexo.extend.generator.register('legacy-presentation-alias', function legacyPresentationAlias() {
  const presentationDir = path.join(this.base_dir, 'source/library/presentations');
  const presentationFiles = [
    'ddia-reading-share-2020.pdf',
    'ddia-reading-share-2022.pptx',
    'k8s-network.pdf'
  ];

  const aliases = [
    {
      path: 'presentations/index.html',
      data: () => '<!doctype html><meta http-equiv="refresh" content="0; url=/library/">'
    },
    ...presentationFiles.map((filename) => ({
      path: `presentations/${filename}`,
      data: () => fs.createReadStream(path.join(presentationDir, filename))
    }))
  ];

  return [
    ...aliases,
    {
      path: 'pdf/k8s-network.pdf',
      data: () => fs.createReadStream(path.join(presentationDir, 'k8s-network.pdf'))
    }
  ];
});
