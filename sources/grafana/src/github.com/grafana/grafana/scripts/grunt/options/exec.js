module.exports = function(config, grunt) {
  'use strict';

  return {
    eslint: {
      command: 'yarn lint',
    },
    typecheckPackages: {
      command: 'yarn packages:typecheck',
    },
    typecheckRoot: {
      command: 'yarn typecheck',
    },
    jest: 'yarn jest-ci',
    webpack: 'node ./node_modules/webpack/bin/webpack.js --config scripts/webpack/webpack.prod.js',
  };
};
