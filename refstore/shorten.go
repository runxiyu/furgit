package refstore

import "strings"

type shortenRule struct {
	prefix string
	suffix string
}

//nolint:gochecknoglobals
var shortenRules = [...]shortenRule{
	{prefix: "", suffix: ""},
	{prefix: "refs/", suffix: ""},
	{prefix: "refs/tags/", suffix: ""},
	{prefix: "refs/heads/", suffix: ""},
	{prefix: "refs/remotes/", suffix: ""},
	{prefix: "refs/remotes/", suffix: "/HEAD"},
}

func (rule shortenRule) match(name string) (string, bool) {
	if !strings.HasPrefix(name, rule.prefix) {
		return "", false
	}

	if !strings.HasSuffix(name, rule.suffix) {
		return "", false
	}

	short := strings.TrimPrefix(name, rule.prefix)

	short = strings.TrimSuffix(short, rule.suffix)
	if short == "" {
		return "", false
	}

	if rule.prefix+short+rule.suffix != name {
		return "", false
	}

	return short, true
}

func (rule shortenRule) render(short string) string {
	return rule.prefix + short + rule.suffix
}

// ShortenName returns the shortest unambiguous shorthand for name among all.
//
// all must contain full reference names visible to the shortening scope.
func ShortenName(name string, all []string) string {
	names := make(map[string]struct{}, len(all))
	for _, full := range all {
		if full == "" {
			continue
		}

		names[full] = struct{}{}
	}

	for i := len(shortenRules) - 1; i > 0; i-- {
		short, ok := shortenRules[i].match(name)
		if !ok {
			continue
		}

		ambiguous := false

		for j := range shortenRules {
			if j == i {
				continue
			}

			full := shortenRules[j].render(short)
			if _, found := names[full]; found {
				ambiguous = true

				break
			}
		}

		if !ambiguous {
			return short
		}
	}

	return name
}
