module.exports = {
  branches: ['main'],
  plugins: [
    '@semantic-release/commit-analyzer',
    ['@semantic-release/release-notes-generator', {
      writerOpts: {
        // 1. Transform function: Intercepts and modifies the commit data before rendering
        transform: (commit, context) => {
          // Map your custom headings
          const typeMapping = {
            feat: '◈ Features',
            fix: '⚒ Fixes',
            perf: '⏵ Performance',
            docs: '✍ Documentation',
            refactor: '↻ Refactoring',
            test: '⚑ Tests',
            build: '⚙ Build',
            ci: '⚡ CI',
            chore: '☑ Chores'
          };

          // Assign the new heading, or discard if it's not in the list
          if (commit.type && typeMapping[commit.type]) {
            commit.type = typeMapping[commit.type];
          } else {
            return; // Skip this commit in the changelog
          }

          // Clean up scope
          if (commit.scope === '*') {
            commit.scope = '';
          }

          // Generate a short hash
          if (typeof commit.hash === 'string') {
            commit.shortHash = commit.hash.substring(0, 7);
          }

          // 2. Format the body: split by newline, trim, remove empty lines, add tab
          if (commit.body) {
            commit.customBody = commit.body
              .split(/\r?\n/)                     // Handle both Windows and Unix newlines
              .map(line => line.trim())           // Remove trailing/leading spaces
              .filter(line => line !== '')        // Remove empty lines
              .map(line => `\t${line}`)           // Prepend a tab to each line
              .join('\n');                        // Put it back together
          }

          return commit;
        },

        // 3. Custom Template: Defines how each list item looks in Markdown
        // Triple braces {{{customBody}}} are used so Handlebars doesn't escape the tabs/newlines
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
