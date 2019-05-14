// The monohook command runs a single HTTP webhook endpoint that executes a command.
package main

import (
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"strconv"
	"sync"
	"time"

	"github.com/buildkite/interpolate"
	"github.com/fatih/color"
	flag "github.com/spf13/pflag"
)

const usageHeader = `%s runs a single HTTP webhook endpoint that executes a command.

Usage:
  %s [OPTION...] [--] [EXEC...]

Options:
`

const usageFooter = `
Examples:
  Update a file when the hook is triggered:
    monohook -- touch hooked.txt
  Deploy an application when the hook is triggered:
    monohook -a letmein -- deploy-stuff.sh
`

type commandOptions struct {
	command string
	args    []string
	cwd     string
}

func main() {

	var authString string
	var bufferString string
	var concurrencyString string
	var cwdString string
	var quiet bool
	var portString string

	flag.CommandLine.SetOutput(os.Stdout)

	flag.StringVarP(&authString, "authorization", "a", "", "Bearer token that must be sent in the Authorization header to authenticate")
	flag.StringVarP(&bufferString, "buffer", "b", "10", "Maximum number of requests to queue before refusing subsequent ones until the queue is freed (zero for infinite)")
	flag.StringVarP(&concurrencyString, "concurrency", "c", "1", "Maximum number of times the command should be executed in parallel (zero for infinite concurrency)")
	flag.StringVarP(&cwdString, "cwd", "C", "", "Working directory in which to run the command")
	flag.BoolVarP(&quiet, "quiet", "q", false, "Do not print anything (default false)")
	flag.StringVarP(&portString, "port", "p", "5000", "Port on which to listen to")

	flag.Usage = func() {
		fmt.Printf(usageHeader, os.Args[0], os.Args[0])
		flag.PrintDefaults()
		fmt.Print(usageFooter)
	}

	flag.Parse()

	auth := parseStringOption(authString, "authorization", quiet)
	buffer := parseUint64Option(bufferString, "buffer", quiet)
	concurrency := parseUint64Option(concurrencyString, "concurrency", quiet)
	cwd := parseStringOption(cwdString, "cwd", quiet)
	port := parseUint64Option(portString, "port", quiet)

	if port > 65535 {
		fail(1, quiet, "port number must be smaller than or equal to 65535")
	}

	terminator := -1
	for i := 0; i < len(os.Args); i++ {
		if os.Args[i] == "--" {
			terminator = i
			break
		}
	}

	var extra []string
	var execCommand string
	var execArgs []string
	if terminator >= 0 && terminator < len(os.Args)-1 {
		extra = flag.Args()[0 : len(flag.Args())-(len(os.Args)-terminator-1)]

		var err error
		execCommand, err = exec.LookPath(parseStringOption(os.Args[terminator+1], "command", quiet))
		if err != nil {
			fail(3, quiet, "could not find command \"%s\"", os.Args[terminator+1])
		}

		execArgs = os.Args[terminator+2 : len(os.Args)]

		for k := range execArgs {
			execArgs[k] = parseStringOption(execArgs[k], "command argument "+strconv.FormatInt(int64(k), 10), quiet)
		}
	} else {
		extra = flag.Args()
	}

	if len(extra) != 0 {
		fail(1, quiet, "no argument expected before the terminator")
	} else if execCommand == "" {
		fail(1, quiet, "no command to execute was provided")
	}

	opts := &commandOptions{}
	opts.command = execCommand
	opts.args = execArgs
	opts.cwd = cwd

	execCh := make(chan *commandOptions, buffer)

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {

		if auth != "" {
			// Refuse request if unauthorized.
			header := r.Header.Get("Authorization")
			if header == "" || header[0:7] != "Bearer " || header[7:] != auth {
				w.WriteHeader(403)
				return
			}
		}

		// Refuse extra requests if buffer is full.
		select {
		case execCh <- opts:
			w.WriteHeader(202)
		default:
			w.WriteHeader(429)
		}
	})

	s := &http.Server{
		Addr:           ":" + strconv.FormatUint(port, 10),
		Handler:        nil,
		ReadTimeout:    10 * time.Second,
		WriteTimeout:   10 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}

	go worker(concurrency, execCh)

	s.ListenAndServe()
}

func interpolateValue(value string) (string, error) {
	return interpolate.Interpolate(interpolate.NewSliceEnv(os.Environ()), value)
}

func parseStringOption(value string, name string, quiet bool) string {

	interpolated, err := interpolateValue(value)
	if err != nil {
		fail(2, quiet, "%s could not be interpolated", name)
	}

	return interpolated
}

func parseUint64Option(value string, name string, quiet bool) uint64 {

	interpolated, err := interpolate.Interpolate(interpolate.NewSliceEnv(os.Environ()), value)
	if err != nil {
		fail(2, quiet, "the \"%s\" option could not be interpolated", name)
	}

	parsed, err := strconv.ParseUint(interpolated, 10, 64)
	if err != nil {
		fail(1, quiet, "the \"%s\" option must be an unsigned 64-bit integer", name)
	}

	return parsed
}

func worker(concurrency uint64, execChannel chan *commandOptions) {
	fmt.Fprintf(os.Stderr, "Execution worker started\n")

	n := uint64(0)
	wait := &sync.WaitGroup{}

	for job := range execChannel {

		if concurrency >= 1 {
			n++
			wait.Add(1)
		}

		go execCommand(job, wait)

		// Wait for queue to clear if concurrency is limited.
		if concurrency >= 1 && n >= concurrency {
			wait.Wait()
			n -= concurrency
		}
	}
}

func execCommand(opts *commandOptions, waitGroup *sync.WaitGroup) {

	fmt.Fprintf(os.Stderr, "Executing %s\n", opts.command)

	cmd := exec.Command(opts.command, opts.args...)
	cmd.Env = os.Environ()
	cmd.Stderr = os.Stdout
	cmd.Stdout = os.Stdout

	err := cmd.Run()
	if err != nil {
		fmt.Fprintf(os.Stderr, color.RedString("Command %s error: %s\n"), opts.command, err)
	} else {
		fmt.Fprintf(os.Stderr, color.GreenString("Successfully executed %s\n"), opts.command)
	}

	waitGroup.Done()
}

func fail(code int, quiet bool, format string, values ...interface{}) {
	if !quiet {
		fmt.Fprintf(os.Stderr, color.RedString("Error: "+format+"\n"), values...)
	}

	os.Exit(code)
}
