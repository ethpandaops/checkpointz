package proxy

import (
	"fmt"
	"regexp"
)

type Matcher struct {
	rules string
}

func NewAllowMatcher(rules []string) Matcher {
	r := ""
	for i, path := range rules {
		r += fmt.Sprintf("(%s)", path)
		if i < len(rules)-1 {
			r += "|"
		}
	}

	return Matcher{
		rules: r,
	}
}

func (m *Matcher) Matches(value string) bool {
	match, _ := regexp.MatchString(m.rules, value)
	return match
}
