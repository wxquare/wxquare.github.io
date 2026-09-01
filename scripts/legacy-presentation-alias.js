'use strict';

const fs = require('fs');
const path = require('path');

hexo.extend.generator.register('legacy-presentation-alias', function legacyPresentationAlias() {
  const canonical = path.join(this.base_dir, 'source/presentations/k8s-network.pdf');
  return {
    path: 'pdf/k8s-network.pdf',
    data: () => fs.createReadStream(canonical)
  };
});
