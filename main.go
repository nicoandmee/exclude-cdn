package main

import (
	"bufio"
	"fmt"
	"github.com/projectdiscovery/cdncheck"
	"log"
	"net"
	"net/url"
	"os"
	"strings"
)

func isURL(candidate string) bool {
	return strings.Contains(candidate, "://")
}

func extractHost(rawurl string) string {
	u, err := url.Parse(rawurl)
	if err != nil {
		log.Fatal(err)
	}

	host, _, err := net.SplitHostPort(u.Host)
	if err != nil {
		return u.Host
	}
	return host
}



func CDNFilter() func(string) bool {
	client, err := cdncheck.NewWithCache()
	if err != nil {
		log.Fatal(err)
	}

	return func(line string) bool {
		host := line
		if isURL(line) {
			host = extractHost(line)
		}

		ip := net.ParseIP(host)
		ips := []net.IP{}
		if ip != nil {
			ips = append(ips, ip)
		} else {
			ips = append(ips, resolveName(host)...)
		}
		for _, ip := range ips {
			found, err := client.Check(ip)
			if found && err == nil {
				return true
			}
		}
		return false
	}
}


func resolveName(name string) []net.IP {
	validIPs := []net.IP{}
	ips, err := net.LookupHost(name)
	if err != nil {
		return validIPs
	}

	for _, ip := range ips {
		parsedIP := net.ParseIP(ip)
		if parsedIP.To4() == nil {
			continue
		}
		validIPs = append(validIPs, parsedIP)
	}
	return validIPs
}

func main() {
	scanner := bufio.NewScanner(os.Stdin)
	filter := CDNFilter()
	for scanner.Scan() {
		line := scanner.Text()
		if !filter(line) {
			fmt.Println(line)
		}
	}
}