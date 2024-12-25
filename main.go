package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"path"
	"proxy-rules/generator"
)

const (
	noRussiaURL = "https://raw.githubusercontent.com/dartraiden/no-russia-hosts/refs/heads/master/hosts.txt"

	gubernievDomainsURL = "https://raw.githubusercontent.com/GubernievS/AntiZapret-VPN/refs/heads/main/setup/root/antizapret/download/include-hosts.txt"

	gubernievIPsURL = "https://raw.githubusercontent.com/GubernievS/AntiZapret-VPN/refs/heads/main/setup/root/antizapret/download/include-ips.txt"
)

func main() {
	antiZapret := generator.NewAntiZapret()
	rules, err := antiZapret.Generate()
	if err != nil {
		log.Fatalf("cannot generate AntiZapret rules: %v", err)
	}

	outputDir := "output"
	mustWriteFile := func(filename, suffix string, items []string) {
		if err := writeFile(path.Join(outputDir, filename), suffix, items); err != nil {
			log.Fatal(err.Error())
		}
	}

	//mustWriteFile("antizapret_domains.lst", "DOMAIN", rules.Domains)
	mustWriteFile("antizapret_domain_suffixes.lst", "DOMAIN_SUFFIX", rules.DomainSuffixes)

	{
		simple := generator.NewSimple(noRussiaURL)
		rules, err := simple.Generate()
		if err != nil {
			log.Fatalf("cannot generate no russia rules: %v", err)
		}
		mustWriteFile("no_russia_domain_suffixes.lst", "DOMAIN_SUFFIX", rules.DomainSuffixes)
	}

	{
		simple := generator.NewSimple(gubernievDomainsURL)
		rules, err := simple.Generate()
		if err != nil {
			log.Fatalf("cannot generate guberniev domain rules: %v", err)
		}
		mustWriteFile("guberniev_domain_suffixes.lst", "DOMAIN_SUFFIX", rules.DomainSuffixes)
	}

	{
		ip := generator.NewIP(gubernievIPsURL)
		rules, err := ip.Generate()
		if err != nil {
			log.Fatalf("cannot generate guberniev ip rules: %v", err)
		}
		mustWriteFile("guberniev_ips.lst", "IP-CIDR", rules.IPNets)
	}
}

func writeFile(filepath, suffix string, items []string) error {
	file, err := os.Create(filepath)
	if err != nil {
		return fmt.Errorf("cannot create file %v, %v", filepath, err)
	}

	writer := bufio.NewWriter(file)

	for _, item := range items {
		_, err := writer.WriteString(fmt.Sprintf("%s,%s\n", suffix, item))
		if err != nil {
			return fmt.Errorf("cannot write string, %v", err)
		}
	}

	err = writer.Flush()
	if err != nil {
		return fmt.Errorf("cannot write data to file %v, %v", filepath, err)
	}

	err = file.Close()
	if err != nil {
		return fmt.Errorf("cannot close file %v, %v", filepath, err)
	}

	return nil
}
