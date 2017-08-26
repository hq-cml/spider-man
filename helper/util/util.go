package util

import (
	"errors"
	"regexp"
	"strings"
)

/*
 * 分析域名，得到主host
 * 比如：www.baidu.com => baidu.com
 */
func GetPrimaryDomain(host string) (string, error) {
	regexpForIp := regexp.MustCompile(`((?:(?:25[0-5]|2[0-4]\d|[01]?\d?\d)\.){3}(?:25[0-5]|2[0-4]\d|[01]?\d?\d))`)

	regexpForDomains := []*regexp.Regexp{
		// *.xx or *.xxx.xx
		regexp.MustCompile(`\.(com|com\.\w{2})$`),
		regexp.MustCompile(`\.(gov|gov\.\w{2})$`),
		regexp.MustCompile(`\.(net|net\.\w{2})$`),
		regexp.MustCompile(`\.(org|org\.\w{2})$`),
		// *.xx
		regexp.MustCompile(`\.me$`),
		regexp.MustCompile(`\.biz$`),
		regexp.MustCompile(`\.info$`),
		regexp.MustCompile(`\.name$`),
		regexp.MustCompile(`\.mobi$`),
		regexp.MustCompile(`\.so$`),
		regexp.MustCompile(`\.asia$`),
		regexp.MustCompile(`\.tel$`),
		regexp.MustCompile(`\.tv$`),
		regexp.MustCompile(`\.cc$`),
		regexp.MustCompile(`\.co$`),
		regexp.MustCompile(`\.\w{2}$`),
	}

	host = strings.TrimSpace(host)
	if host == "" {
		return "", errors.New("The host is empty!")
	}
	if regexpForIp.MatchString(host) {
		return host, nil
	}
	var suffixIndex int
	for _, re := range regexpForDomains {
		pos := re.FindStringIndex(host)
		if pos != nil {
			suffixIndex = pos[0]
			break
		}
	}
	if suffixIndex > 0 {
		var pdIndex int
		firstPart := host[:suffixIndex]
		index := strings.LastIndex(firstPart, ".")
		if index < 0 {
			pdIndex = 0
		} else {
			pdIndex = index + 1
		}
		return host[pdIndex:], nil
	} else {
		return "", errors.New("Unrecognized host!")
	}
}
