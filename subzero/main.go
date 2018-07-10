package main

import (
	"fmt"
	"sync"

	"github.com/spf13/cobra"
	core "github.com/subfinder/research/core"
	sources "github.com/subfinder/research/core/sources"
)

var sourcesList = []core.Source{
	&sources.ArchiveIs{},
	&sources.CertSpotter{},
	&sources.CommonCrawlDotOrg{},
	&sources.CrtSh{},
	&sources.FindSubdomainsDotCom{},
	&sources.HackerTarget{},
	&sources.Riddler{},
	&sources.Threatminer{},
	&sources.WaybackArchive{},
}

func Enumerate(domain string) chan *core.Result {
	wg := sync.WaitGroup{}
	results := make(chan *core.Result, len(sourcesList)*4)
	go func(domain string) {
		defer close(results)
		for _, source := range sourcesList {
			wg.Add(1)
			go func(domain string, source core.Source, results chan *core.Result) {
				defer wg.Done()
				for result := range source.ProcessDomain(domain) {
					results <- result
				}
			}(domain, source, results)
		}
		wg.Wait()
	}(domain)
	return results
}

func main() {
	jobs := sync.WaitGroup{}
	var cmdEnumerateVerboseOpt bool

	var cmdEnumerate = &cobra.Command{
		Use:   "enumerate [domains to enumerate]",
		Short: "Enumerate subdomains for the given domains.",
		Args:  cobra.MinimumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			for _, domain := range args {
				jobs.Add(1)
				go func(domain string) {
					defer jobs.Done()
					for result := range Enumerate(domain) {
						if result.Failure != nil && cmdEnumerateVerboseOpt {
							fmt.Println(result.Type, result.Failure)
						} else {
							fmt.Println(result.Type, result.Success)
						}
					}
				}(domain)
			}
		},
	}
	cmdEnumerate.Flags().BoolVar(&cmdEnumerateVerboseOpt, "verbose", false, "Show errors and other available diagnostic information.")

	var rootCmd = &cobra.Command{Use: "subzero"}
	rootCmd.AddCommand(cmdEnumerate)
	rootCmd.Execute()

	jobs.Wait()
}
