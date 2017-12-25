package helpers

import (
	"regexp"
	"strings"
)

func GetDomainRegexp(domains []string) (*regexp.Regexp, error) {
	domainList := strings.Join(domains, "|")
	return regexp.Compile("https://(?:" + domainList + ")(?:[[:graph:]]{2,})?/(?:[@])?([-a-zA-Z0-9.]{2,256})/([-a-zA-Z0-9@:%_+.~?&=]{2,256})")
}
