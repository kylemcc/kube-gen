// kube-gen is a
package main

import (
	"errors"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	kubegen "github.com/kylemcc/kube-gen"
)

type stringSlice []string

var (
	// flags
	host         string
	kubeconfig   string
	types        stringSlice
	watch        bool
	preCmd       string
	postCmd      string
	logCmdOutput bool
	overwrite    bool
	wait         string
	interval     int
	quiet        bool
	showVersion  bool

	// build info
	version   string
	buildTime string
	revision  string

	flags = flag.NewFlagSet(os.Args[0], flag.ExitOnError)
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
	flags.PrintDefaults()

	fmt.Printf(`
Arguments:
  template: path or URL of the template file to render, or - to read from STDIN
  output: (Optional) path to write the rendered content. If not specified,
          rendered content is printed to STDOUT. By default, this file will
          be overwritten if it exists. Use -overwrite=false to return an
          error instead
`)
}

func parseFlags() {
	flags.StringVar(&host, "host", "", "If not set will use kubeconfig. If using proxy - set it to http://localhost:8001")
	if home := homeDir(); home != "" {
		flags.StringVar(&kubeconfig, "kubeconfig", filepath.Join(home, ".kube", "config"), "(optional) absolute path to the kubeconfig file")
	} else {
		flags.StringVar(&kubeconfig, "kubeconfig", "", "absolute path to the kubeconfig file")
	}
	flags.Var(&types, "type", "types of resources to pull [pods, services, endpoints] - May be specified multiple times. "+
		"If not specified, all types will be returned")
	flags.BoolVar(&showVersion, "version", false, "display version information")
	flags.BoolVar(&watch, "watch", false, "watch for new events")
	flags.StringVar(&preCmd, "pre-cmd", "", "command to run before template generation")
	flags.StringVar(&postCmd, "post-cmd", "", "command to run after template generation in complete")
	flags.BoolVar(&logCmdOutput, "log-cmd", true, "log the output of the pre/post commands")
	flags.BoolVar(&overwrite, "overwrite", true, "overwrite the output file if it exists")
	flags.StringVar(&wait, "wait", "", "<minimum>[:<maximum>] - the minimum and optional maximum time to wait after an event fires."+
		"E.g.: 500ms:5s")
	flags.IntVar(&interval, "interval", 0, "")
	flags.BoolVar(&quiet, "quiet", false, "when set to true, nothing is logged")

	flags.Usage = usage
	flags.Parse(os.Args[1:])
}

func homeDir() string {
	if h := os.Getenv("HOME"); h != "" {
		return h
	}
	return os.Getenv("USERPROFILE") // windows
}

func printVersion() {
	fmt.Printf(`version:  %s
built at: %s
revision: %s
runtime:  %s
`, version, buildTime, revision, runtime.Version())
}

func parseWait(w string) (min time.Duration, max time.Duration, err error) {
	w = strings.TrimSpace(w)
	if len(w) == 0 {
		return 0, 0, nil
	} else if w[0] == ':' {
		return 0, 0, errors.New("minimum is required")
	}
	parts := strings.Split(w, ":")
	if min, err = time.ParseDuration(parts[0]); err != nil {
		return
	}
	if len(parts) > 1 && len(parts[1]) > 0 {
		max, err = time.ParseDuration(parts[1])
		if err == nil && max < min {
			err = errors.New("max must be greater than or equal to min")
		}
	}
	return
}

func tmplFromStdin() ([]byte, error) {
	return ioutil.ReadAll(os.Stdin)
}

func main() {
	parseFlags()

	if quiet {
		log.SetOutput(ioutil.Discard)
	}

	if showVersion {
		printVersion()
		return
	}

	if narg := flags.NArg(); narg < 1 || narg > 2 {
		flags.Usage()
		os.Exit(1)
	}

	minWait, maxWait, err := parseWait(wait)
	if err != nil {
		log.Fatalf("invalid wait value: %v", err)
	}

	var tmplStr string
	if flags.Arg(0) == "-" {
		log.Printf("reading template from stdin")
		if s, err := tmplFromStdin(); err != nil {
			log.Fatalf("error reading from stdin: %v", err)
		} else {
			tmplStr = strings.TrimSpace(string(s))
		}
	}
	if flags.Arg(1) == "" {
		log.Printf("writing output to stdout")
	}

	conf := kubegen.Config{
		Host:           host,
		Kubeconfig:     kubeconfig,
		TemplateString: tmplStr,
		TemplatePath:   flags.Arg(0),
		Output:         flags.Arg(1),
		Overwrite:      overwrite,
		Watch:          watch,
		PreCmd:         preCmd,
		PostCmd:        postCmd,
		ResourceTypes:  types,
		MinWait:        minWait,
		MaxWait:        maxWait,
		Interval:       interval,
	}

	gen, err := kubegen.NewGenerator(conf)
	if err != nil {
		log.Fatalf("error initializing generator: %v", err)
	}

	if err := gen.Generate(); err != nil {
		log.Fatalf("error generating output: %v", err)
	}
}
