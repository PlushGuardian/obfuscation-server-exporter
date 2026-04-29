// {
//   "branches": ["main"],
//   "plugins": [
//     "@semantic-release/commit-analyzer",
//     [
//       "@semantic-release/release-notes-generator",
//       {
//         "preset": "conventionalcommits",
//         "presetConfig": {
//           "types": [
//             { "type": "feat",     "section": "◈ Features",      "hidden": false },
//             { "type": "fix",      "section": "⚒ Fixes",         "hidden": false },
//             { "type": "perf",     "section": "⏵ Performance",   "hidden": false },
//             { "type": "docs",     "section": "✍ Documentation", "hidden": false },
//             { "type": "refactor", "section": "↻ Refactoring",   "hidden": false },
//             { "type": "test",     "section": "⚑ Tests",         "hidden": false },
//             { "type": "build",    "section": "⚙ Build",         "hidden": false },
//             { "type": "ci",       "section": "⚡ CI",            "hidden": false },
//             { "type": "chore",    "section": "☑ Chores",        "hidden": false }
//           ]
//         },
//         "writerOpts": {
//           "transform": "./.semantic-release/changelog-transform.js",
//           "commitPartial": "./.semantic-release/commit-template.hbs"
//         }
//       }
//     ],
//     "@semantic-release/changelog",
//     [
//       "@semantic-release/git",
//       {
//         "assets": ["CHANGELOG.md"]
//       }
//     ],
//     "@semantic-release/github"
//   ]
// }


// .releaserc.js
const fs = require('fs');
const path = require('path');

const commitPartial = fs.readFileSync(
  path.resolve(__dirname, '.semantic-release', 'commit-template.hbs'),
  'utf8'
);

module.exports = {
  plugins: [
    ['@semantic-release/release-notes-generator', {
      preset: 'conventionalcommits',
      presetConfig: {
        types: [
          { "type": "feat",     "section": "◈ Features",      "hidden": false },
          { "type": "fix",      "section": "⚒ Fixes",         "hidden": false },
          { "type": "perf",     "section": "⏵ Performance",   "hidden": false },
          { "type": "docs",     "section": "✍ Documentation", "hidden": false },
          { "type": "refactor", "section": "↻ Refactoring",   "hidden": false },
          { "type": "test",     "section": "⚑ Tests",         "hidden": false },
          { "type": "build",    "section": "⚙ Build",         "hidden": false },
          { "type": "ci",       "section": "⚡ CI",            "hidden": false },
          { "type": "chore",    "section": "☑ Chores",        "hidden": false }
        ]
      },
      writerOpts: {
        transform: './.semantic-release/changelog-transform.js',
        commitPartial: commitPartial
      }
    }],
    "@semantic-release/commit-analyzer",
    "@semantic-release/changelog",
    [
      "@semantic-release/git",
      {
        "assets": ["CHANGELOG.md"]
      }
    ],
    "@semantic-release/github"
  ]
};
