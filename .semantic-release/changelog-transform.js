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

  commit.typeLabel = typeLabelMap[commit.type] || commit.type;

  if (commit.body) {
    commit.descriptionLines = commit.body
      .split('\n')
      .filter(line => line.trim() !== '');
  } else {
    commit.descriptionLines = [];
  }

  const allowedTypes = Object.keys(typeLabelMap);
  if (!allowedTypes.includes(commit.type)) {
    return null;
  }

  return commit;
};
