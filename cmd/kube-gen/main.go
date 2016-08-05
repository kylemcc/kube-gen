// kube-gen is a
package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"strings"

	kubegen "github.com/kylemcc/kube-gen"
)

type stringSlice []string

var (
	// flags
	host        string
	types       stringSlice
	watch       bool
	notifyCmd   string
	overwrite   bool
	showVersion bool

	// build info
	version   string
	buildTime string
	revision  string
)

// implement flag.Value interface
func (s *stringSlice) String() string {
	return strings.Join(*s, ", ")
}

func (s *stringSlice) Set(v string) error {
	*s = append(*s, v)
	return nil
}

func usage() {
	fmt.Printf(`Usage: kube-gen [options] <template> [<output>]

Render templates using Kubernetes metadata and events

Options:
`)
	flag.PrintDefaults()

	fmt.Printf(`
Arguments:
  template: path of the template file to render, or - to read from STDIN
  output: (Optional) path to write the rendered content. If not specified,
          rendered content is printed to STDOUT. By default, this file will
          be overwritten if it exists. Use -overwrite=false to return an
          error instead

Environment Variables:
  KUBERNETES_SERVICE_HOST
  KUBERNETES_SERVICE_PORT

Examples:
`)
}

func parseFlags() {
	flag.StringVar(&host, "host", "http://localhost:8001", "")
	flag.Var(&types, "type", "types of resources to pull [pods, services, endpoints] - May be specified multiple times. "+
		"If not specified, all types will be returned")
	flag.BoolVar(&showVersion, "version", false, "display version information")
	flag.BoolVar(&watch, "watch", false, "watch for new events")
	flag.StringVar(&notifyCmd, "notify", "", "command to run after template generation in complete")
	flag.BoolVar(&overwrite, "overwrite", true, "overwrite the output file if it exists")

	flag.Usage = usage
	flag.Parse()
}

func printVersion() {
	fmt.Printf(`kube-gen v%s
built at: %s
revision: %s
`, version, buildTime, revision)
}

func main() {
	parseFlags()

	if showVersion {
		printVersion()
		return
	}

	if narg := flag.NArg(); narg < 1 || narg > 2 {
		flag.Usage()
		os.Exit(1)
	}

	conf := kubegen.Config{
		Host:          host,
		Template:      flag.Arg(0),
		Output:        flag.Arg(1),
		Overwrite:     overwrite,
		Watch:         watch,
		NotifyCmd:     notifyCmd,
		ResourceTypes: types,
	}

	gen, err := kubegen.NewGenerator(conf)
	if err != nil {
		log.Fatalf("error initializing generator: %v", err)
	}

	if err := gen.Generate(); err != nil {
		log.Fatalf("error generating output: %v", err)
	}
}
