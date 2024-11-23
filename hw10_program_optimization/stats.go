package hw10programoptimization

import (
	"errors"
	"io"
	"regexp"
	"strings"

	//nolint:depguard
	jsoniter "github.com/json-iterator/go"
)

type User struct {
	Email string `json:"email"`
}

type DomainStat map[string]int

var json = jsoniter.ConfigCompatibleWithStandardLibrary

func GetDomainStat(r io.Reader, domain string) (DomainStat, error) {
	result := make(DomainStat)
	domainRegexp := regexp.MustCompile(`@(.+\.)?` + regexp.QuoteMeta(domain) + `$`)
	decoder := json.NewDecoder(r)

	for {
		var user User
		if err := decoder.Decode(&user); errors.Is(err, io.EOF) {
			break
		} else if err != nil {
			return nil, err
		}

		matches := domainRegexp.FindStringSubmatch(user.Email)
		if len(matches) != 2 {
			continue
		}

		domainPart := matches[1]
		parts := strings.Split(domainPart, ".")
		if len(parts) < 2 {
			continue
		}

		secondLevelDomain := strings.ToLower(parts[len(parts)-2])
		result[secondLevelDomain+"."+domain]++
	}

	return result, nil
}
