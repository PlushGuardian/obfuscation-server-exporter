package config

import (
	"fmt"
	"regexp"

	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

func AliasFlags(v *viper.Viper, fs *pflag.FlagSet, rules map[string]string) error {
	type compiledRule struct {
		re   *regexp.Regexp
		repl string
	}
	compiled := make([]compiledRule, 0, len(rules))

	for pat, sub := range rules {
		re, err := regexp.Compile(pat)
		if err != nil {
			return fmt.Errorf("invalid alias regex pattern %q: %w", pat, err)
		}
		compiled = append(compiled, compiledRule{re, sub})
	}

	var bindErr error
	fs.VisitAll(func(f *pflag.Flag) {
		if bindErr != nil {
			return
		}

		name := f.Name
		for _, r := range compiled {
			name = r.re.ReplaceAllString(name, r.repl)
		}

		if name != f.Name {
			if err := v.BindPFlag(name, f); err != nil {
				bindErr = fmt.Errorf("error binding alias %q for flag %q: %w", name, f.Name, err)
			}
		}
	})

	return bindErr
}
