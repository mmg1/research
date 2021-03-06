package sources

import (
	"bufio"
	"errors"

	"github.com/subfinder/research/core"
)

// DNSDbDotCom is a source to process subdomains from http://www.dnsdb.org/f/
type DNSDbDotCom struct{}

// ProcessDomain takes a given base domain and attempts to enumerate subdomains.
func (source *DNSDbDotCom) ProcessDomain(domain string) <-chan *core.Result {
	results := make(chan *core.Result)
	go func(domain string, results chan *core.Result) {
		defer close(results)

		domainExtractor, err := core.NewSubdomainExtractor(domain)
		if err != nil {
			results <- core.NewResult("dnsdbdotcom", nil, err)
			return
		}

		uniqFilter := map[string]bool{}

		resp, err := core.HTTPClient.Get("http://www.dnsdb.org/f/" + domain + ".dnsdb.org/")
		if err != nil {
			results <- core.NewResult("dnsdbdotcom", nil, err)
			return
		}
		defer resp.Body.Close()

		if resp.StatusCode != 200 {
			results <- core.NewResult("dnsdbdotcom", nil, errors.New(resp.Status))
			return
		}

		scanner := bufio.NewScanner(resp.Body)

		for scanner.Scan() {
			for _, str := range domainExtractor.FindAllString(scanner.Text(), -1) {
				_, found := uniqFilter[str]
				if !found {
					uniqFilter[str] = true
					results <- core.NewResult("dnsdbdotcom", str, nil)
				}
			}
		}

	}(domain, results)
	return results
}
