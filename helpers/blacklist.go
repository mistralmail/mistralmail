package helpers

import (
	"bufio"
	"net/http"
	"sort"
	"strings"
)

// Interface for handling blaclists
// it is meant to be replaced by your own implementation
type Blacklist interface {
	// CheckIp will return true if the IP is blacklisted and false if the IP was not found in a blacklist
	CheckIp(ip string) bool
}

// This blacklist implementation will download NiX Spam's blacklist
// and load it into memory
type Nixspam struct {
	IpList []string
}

func NewNixspam() (*Nixspam, error) {

	ns := Nixspam{}
	ns.IpList = make([]string, 0)

	resp, err := http.Get("http://www.dnsbl.manitu.net/download/nixspam-ip.dump")
	if err != nil {
		return nil, err
	}

	scanner := bufio.NewScanner(resp.Body)
	scanner.Split(bufio.ScanLines)

	for scanner.Scan() {
		rawLine := scanner.Text()
		line := strings.Split(rawLine, " ")
		if len(line) < 2 {
			continue
		}
		ns.IpList = append(ns.IpList, strings.TrimPrefix(strings.TrimSpace(line[1]), "..."))

	}

	sort.Strings(ns.IpList)

	return &ns, nil

}

func (ns *Nixspam) CheckIp(ip string) bool {
	index := sort.SearchStrings(ns.IpList, ip)
	if index == len(ns.IpList) {
		return false
	}

	return ns.IpList[index] == ip
}
