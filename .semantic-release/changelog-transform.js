// changelog-transform.js
module.exports = function (commit, context) {
  // ── Map commit type → display label ──────────────────
  const typeLabelMap = {
    feat:     '◈ Features',
    fix:      '⚒ Fixes',
    perf:     '⏵ Performance',
    docs:     '✍ Documentation',
    refactor: '↻ Refactoring',
    test:     '⚑ Tests',
    build:    '⚙ Build',
    ci:       '⚡ CI',
    chore:    '☑ Chores',
  };

  if (!Object.keys(typeLabelMap).includes(commit.type)) {
    return null;
  }

  commit.typeLabel = typeLabelMap[commit.type];

  if (commit.body) {
    commit.descriptionLines = commit.body
      .split('\n')
      .filter(line => line.trim() !== '');
  } else {
    commit.descriptionLines = [];
  }

  return commit;
};
