package generator

import (
	"bufio"
	"compress/gzip"
	"encoding/csv"
	"errors"
	"fmt"
	"io"
	"net/http"
	"proxy-rules/trie"
	"strings"
	"unicode"

	"golang.org/x/text/encoding/charmap"
)

const defaultDumpURL = "https://raw.githubusercontent.com/zapret-info/z-i/master/dump.csv.gz"
const defaultNXDomainURL = "https://raw.githubusercontent.com/zapret-info/z-i/master/nxdomain.txt"

type AntiZapret struct {
	dumpURL     string
	nxDomainURL string
}

func NewAntiZapret() *AntiZapret {
	return &AntiZapret{
		dumpURL:     defaultDumpURL,
		nxDomainURL: defaultNXDomainURL,
	}
}

type Filter struct {
	nxDomains           map[string]struct{}
	forbiddenSubstrings []string
	forbiddenTLDs       map[string]struct{}
	commonTLDs          map[string]struct{}
}

func NewFilter(nxDomains map[string]struct{}) *Filter {
	filter := Filter{
		nxDomains: nxDomains,
		forbiddenSubstrings: []string{
			"casino",
			"poker",
			"bet",
			"vulkan",
			"vegas",
			"rezka",
			"slot",
			"eldorado",
			"vavada",
			"fortuna",
			"vulcan",
			"prostitut",
			"777",
			"1x",
			"1w",
			"azino",
			"xn--",
			"gay",
			"porn",
			"kinovod",
			"sex",
			"seks",
		},
		forbiddenTLDs: map[string]struct{}{
			"win":        {},
			"sex":        {},
			"ua":         {},
			"kz":         {},
			"рф":         {},
			"fun":        {},
			"cash":       {},
			"quest":      {},
			"rest":       {},
			"best":       {},
			"market":     {},
			"adult":      {},
			"africa":     {},
			"agency":     {},
			"apartments": {},
			"army":       {},
			"audio":      {},
			"autos":      {},
			"baby":       {},
			"band":       {},
		},
	}
	var commonTLDs = []string{"com", "org", "net", "info", "gov", "edu", "co", "ai", "io", "me", "cc"}
	filter.commonTLDs = make(map[string]struct{})
	for _, commonTLD := range commonTLDs {
		filter.commonTLDs[commonTLD] = struct{}{}
	}
	return &filter
}

func (f *Filter) Check(domain string) bool {
	if _, ok := f.nxDomains[domain]; ok {
		return true
	}
	dotIndex := strings.LastIndex(domain, ".")
	tld := domain[dotIndex+1:]
	if _, ok := f.commonTLDs[tld]; !ok {
		return true
	}
	if isNumeric(tld) {
		return true
	}
	for _, forbiddenSubstring := range f.forbiddenSubstrings {
		if strings.Contains(domain, forbiddenSubstring) {
			return true
		}
	}
	return false
}

func isNumeric(s string) bool {
	if s == "" {
		return false
	}
	for _, char := range s {
		if !unicode.IsDigit(char) {
			return false
		}
	}
	return true
}

func (az *AntiZapret) Generate() (*Rules, error) {
	nxDomains, err := getNXDomains(az.nxDomainURL)
	if err != nil {
		return nil, err
	}

	resp, err := http.Get(az.dumpURL)
	if err != nil {
		return nil, fmt.Errorf("get dump file: %w", err)
	}
	defer resp.Body.Close()

	gzipReader, err := gzip.NewReader(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("decode dump file: %w", err)
	}
	defer gzipReader.Close()

	reader := csv.NewReader(charmap.Windows1251.NewDecoder().Reader(gzipReader))
	reader.Comma = ';'
	reader.FieldsPerRecord = -1

	_, err = reader.Read()
	if err != nil {
		return nil, fmt.Errorf("read dump file, line %v: %w", 1, err)
	}

	filter := NewFilter(nxDomains)
	domains := make(map[string]struct{})
	tlds := make(map[string]struct{})

	for {
		rec, err := reader.Read()
		line, _ := reader.FieldPos(0)

		if err != nil {
			if errors.Is(err, io.EOF) {
				break
			}
			return nil, fmt.Errorf("read dump file, line %v: %w", line, err)
		}

		if len(rec) < 2 {
			return nil, fmt.Errorf("unexpected number of fields (%v), line %v", len(rec), line)
		}

		wildcard := rec[1]
		if wildcard == "" {
			continue
		}

		dotCount, i := 0, len(wildcard)-1
		for i >= 0 {
			if wildcard[i] == '.' {
				dotCount++
				if dotCount == 2 {
					break
				}
			}
			i--
		}

		domain := wildcard[i+1:]

		if filter.Check(domain) {
			continue
		}

		dotIndex := strings.LastIndex(domain, ".")
		tld := domain[dotIndex+1:]
		tlds[tld] = struct{}{}

		domains[domain] = struct{}{}
	}

	fmt.Println(tlds)

	return &Rules{
		DomainSuffixes: getSuffixes(domains),
	}, nil
}

func getNXDomains(nxDomainURL string) (map[string]struct{}, error) {
	resp, err := http.Get(nxDomainURL)
	if err != nil {
		return nil, fmt.Errorf("get nxdomain file: %w", err)
	}
	defer resp.Body.Close()
	nxDomains := make(map[string]struct{})
	scanner := bufio.NewScanner(resp.Body)
	for scanner.Scan() {
		nxDomains[scanner.Text()] = struct{}{}
	}
	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("read nxdomain file: %w", err)
	}
	return nxDomains, nil
}

func getSuffixes(domains map[string]struct{}) []string {
	t := trie.NewTrie()
	for domain := range domains {
		t.Insert(reverseString(domain))
	}
	suffixes := t.GetAllFirstLevelPrefixes()
	for i := range suffixes {
		suffixes[i] = reverseString(suffixes[i])
	}
	return suffixes
}

func reverseString(input string) string {
	runes := []rune(input)
	for l, r := 0, len(runes)-1; l < r; l, r = l+1, r-1 {
		runes[l], runes[r] = runes[r], runes[l]
	}
	return string(runes)
}
