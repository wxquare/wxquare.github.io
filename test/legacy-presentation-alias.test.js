'use strict';

const assert = require('assert').strict;
const fs = require('fs');
const path = require('path');

const repoRoot = path.resolve(__dirname, '..');
const aliasScript = path.join(repoRoot, 'scripts/legacy-presentation-alias.js');
const canonical = path.join(repoRoot, 'source/presentations/k8s-network.pdf');
const legacy = path.join(repoRoot, 'source/pdf/k8s-network.pdf');

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
    const route = registrations[0].fn.call({ base_dir: repoRoot });
    assert.equal(route.path, 'pdf/k8s-network.pdf');
    assert.equal(typeof route.data, 'function');
    assert.deepEqual(route.data(), { source: canonical });
    assert.deepEqual(opened, [canonical]);
    assert.equal(opened.includes(legacy), false);
  } finally {
    fs.createReadStream = previousCreateReadStream;
    delete require.cache[aliasScript];
    if (previousHexo === undefined) delete global.hexo;
    else global.hexo = previousHexo;
  }
}

run();
