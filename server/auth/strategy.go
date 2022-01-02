package auth

import (
	"errors"
	"net/http"
	"regexp"
	"strings"
)

func AllowAllAuthorizationStrategy(user User, r *http.Request) error {
	return nil
}

func PolicyAuthorizationStrategy(apiPrefix string) func(user User, r *http.Request) error {
	return func(user User, r *http.Request) error {
		ep := strings.TrimPrefix((*r).RequestURI, apiPrefix)
		rP := regexp.MustCompile("^(?P<PATH>[^?]*)\\??.*$")
		matches := rP.FindStringSubmatch(ep)
		if len(matches) < 2 || len(strings.TrimSpace(matches[1])) == 0 {
			return errors.New("matches < 2 or matches 1 is empty")
		}
		path := strings.TrimSpace(matches[1])
		allowPolicyMatched := false
		for _, p := range user.Permissions {
			for _, deny := range p.Denys {
				for _, a := range deny.Actions {
					m, err := regexp.MatchString(a, path)
					if m || err != nil {
						return errors.New("Denied by policy")
					}
				}
			}

			for _, allow := range p.Allows {
				for _, a := range allow.Actions {
					m, err := regexp.MatchString(a, path)
					if m && err == nil {
						allowPolicyMatched = true
					}
				}
			}
		}
		if allowPolicyMatched {
			return nil
		} else {
			return errors.New("No match in policy")
		}
	}
}
