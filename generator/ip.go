package generator

import (
	"bufio"
	"fmt"
	"net/http"
	"strings"
)

type IP struct {
	ipsURL string
}

func NewIP(ipsURL string) *IP {
	return &IP{
		ipsURL: ipsURL,
	}
}

func (s *IP) Generate() (*Rules, error) {
	res, err := http.DefaultClient.Get(s.ipsURL)
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

		rules.IPNets = append(rules.IPNets, line)
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("cannot scan file, %w", err)
	}

	return &rules, nil
}
