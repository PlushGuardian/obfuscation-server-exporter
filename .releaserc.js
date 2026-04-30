module.exports = {
  // Add your existing branches here
  branches: ['main'],

  plugins: [
    '@semantic-release/commit-analyzer',
    ['@semantic-release/release-notes-generator', {
      writerOpts: {
        transform: (originalCommit, context) => {
          // Create a copy so we don't try to mutate an immutable object
          const commit = { ...originalCommit };

          const typeMapping = {
            feat: '◈ Features',
            fix: '⚒ Fixes',
            perf: '⏵ Performance',
            docs: '✍ Documentation',
            refactor: '↻ Refactoring',
            test: '⚑ Tests',
            build: '⚙ Build',
            ci: '⌁ CI',
            chore: '☑ Chores'
          };

          if (commit.type && typeMapping[commit.type]) {
            commit.type = typeMapping[commit.type];
          } else {
            return;
          }

          if (commit.scope === '*') {
            commit.scope = '';
          }

          if (typeof commit.hash === 'string') {
            commit.shortHash = commit.hash.substring(0, 7);
          }

          if (commit.body) {
            commit.customBody = commit.body
              .split(/\r?\n/)
              .map(line => line.trim())
              .filter(line => line !== '')
              .map(line => `\t${line}`)
              .join('\n');
          }

          return commit;
        },

        commitPartial: `* {{#if scope}}**{{scope}}:** {{/if}}{{subject}} {{#if hash}}{{#if @root.linkReferences}}([{{shortHash}}]({{@root.host}}/{{@root.owner}}/{{@root.repository}}/commit/{{hash}})){{else}}({{shortHash}}){{/if}}{{/if}}
{{#if customBody}}
{{{customBody}}}
{{/if}}`
      }
    }],
    '@semantic-release/changelog',
    ['@semantic-release/git', {
      assets: ['CHANGELOG.md'],
      message: 'chore(release): ${nextRelease.version} [skip ci]\n\n${nextRelease.notes}'
    }]
  ]
};
