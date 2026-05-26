// Copyright (c) 2015-present Mattermost, Inc. All Rights Reserved.
// See LICENSE.txt for license information.

const config = {
    presets: [
        ['@babel/preset-env', {
            targets: {
                chrome: 66,
                firefox: 60,
                edge: 42,
                safari: 12,
            },
            modules: false,
            corejs: 3,
            debug: false,
            useBuiltIns: 'usage',
            shippedProposals: true,
        }],
        ['@babel/preset-react', {
            runtime: 'automatic',
        }],
        ['@babel/preset-typescript', {
            allExtensions: true,
            isTSX: true,
        }],
        ['@emotion/babel-preset-css-prop'],
    ],
    plugins: [
        '@babel/plugin-syntax-dynamic-import',
        'babel-plugin-typescript-to-proptypes',
    ],
};

// Jest needs module transformation
config.env = {
    test: {
        presets: [
            ['@babel/preset-env', {...config.presets[0][1], modules: 'auto'}],
            ...config.presets.slice(1),
        ],
        plugins: config.plugins,
    },
};

module.exports = config;
