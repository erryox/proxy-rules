package generator

import (
	"bufio"
	"fmt"
	"net/http"
	"strings"
)

type Simple struct {
	domainsURL string
}

func NewSimple(domainsURL string) *Simple {
	return &Simple{
		domainsURL: domainsURL,
	}
}

func (s *Simple) Generate() (*Rules, error) {
	res, err := http.DefaultClient.Get(s.domainsURL)
	if err != nil {
		return nil, fmt.Errorf("cannot get file, %w", err)
	}
	defer res.Body.Close()

	scanner := bufio.NewScanner(res.Body)

	var rules Rules

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		rules.DomainSuffixes = append(rules.DomainSuffixes, line)
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("cannot scan file, %w", err)
	}

	return &rules, nil
}
