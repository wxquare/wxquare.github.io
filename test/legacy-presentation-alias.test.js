'use strict';

const assert = require('assert').strict;
const fs = require('fs');
const path = require('path');

const repoRoot = path.resolve(__dirname, '..');
const aliasScript = path.join(repoRoot, 'scripts/legacy-presentation-alias.js');
const presentationDir = path.join(repoRoot, 'source/library/presentations');
const canonical = path.join(presentationDir, 'k8s-network.pdf');
const legacy = path.join(repoRoot, 'source/pdf/k8s-network.pdf');
const presentationFiles = [
  'ddia-reading-share-2020.pdf',
  'ddia-reading-share-2022.pptx',
  'k8s-network.pdf'
];

function run() {
  const registrations = [];
  const opened = [];
  const previousHexo = global.hexo;
  const previousCreateReadStream = fs.createReadStream;

  try {
    fs.createReadStream = (file) => {
      const resolved = path.resolve(file);
      opened.push(resolved);
      return { source: resolved };
    };
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
    assert.equal(registrations[0].name, 'legacy-presentation-alias');
    const routes = registrations[0].fn.call({ base_dir: repoRoot });
    assert.equal(Array.isArray(routes), true);
    assert.deepEqual(routes.map((route) => route.path), [
      'presentations/index.html',
      'presentations/ddia-reading-share-2020.pdf',
      'presentations/ddia-reading-share-2022.pptx',
      'presentations/k8s-network.pdf',
      'pdf/k8s-network.pdf'
    ]);

    const indexRoute = routes[0];
    assert.match(indexRoute.data(), /\/library\//);

    const expectedSources = presentationFiles.map((filename) => path.join(presentationDir, filename));
    routes.slice(1).forEach((route) => {
      assert.equal(typeof route.data, 'function');
      assert.deepEqual(route.data(), { source: expectedSources[route.path.endsWith('k8s-network.pdf') ? 2 : route.path.includes('2022') ? 1 : 0] });
    });

    assert.deepEqual(opened, [...expectedSources, canonical]);
    assert.equal(opened.includes(legacy), false);
  } finally {
    fs.createReadStream = previousCreateReadStream;
    delete require.cache[aliasScript];
    if (previousHexo === undefined) delete global.hexo;
    else global.hexo = previousHexo;
  }
}

run();
