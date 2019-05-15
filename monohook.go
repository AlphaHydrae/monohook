// The monohook command runs a single HTTP webhook endpoint that executes a command.
package main

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
	"sync"

	"github.com/alphahydrae/monohook/utils"
	"github.com/fatih/color"
	flag "github.com/spf13/pflag"
)

const usageHeader = `%s exposes a single HTTP webhook endpoint that executes a command.

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
	env     []string
	reader  *io.PipeReader
}

var httpRequestHeaderToEnvVarRegexp = regexp.MustCompile("-")

func main() {

	var auth string
	var buffer uint64
	var concurrency uint64
	var cwd string
	var port uint16

	quiet := false
	errHandler := func(code int, message string) {
		fail(code, quiet, message)
	}

	flag.CommandLine.SetOutput(os.Stdout)
	utils.BoolOption(&quiet, "quiet", "q", "QUIET", false, "Do not print anything (default false)", errHandler)

	utils.StringOption(&auth, "authorization", "a", "AUTHORIZATION", "", "Authentication token that must be sent as a Bearer token in the 'Authorization' header or as the 'authorization' URL query parameter")
	utils.Uint64Option(&buffer, "buffer", "b", "BUFFER", 10, "Maximum number of requests to queue before refusing subsequent ones until the queue is freed (zero for infinite)", errHandler)
	utils.Uint64Option(&concurrency, "concurrency", "c", "CONCURRENCY", 1, "Maximum number of times the command should be executed in parallel (zero for infinite concurrency)", errHandler)
	utils.StringOption(&cwd, "cwd", "C", "CWD", "", "Working directory in which to run the command")
	utils.Uint16Option(&port, "port", "p", "PORT", 5000, "Port on which to listen to", errHandler)

	flag.Usage = func() {
		fmt.Printf(usageHeader, os.Args[0], os.Args[0])
		flag.PrintDefaults()
		fmt.Print(usageFooter)
	}

	flag.Parse()

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
		execCommand, err = exec.LookPath(os.Args[terminator+1])
		if err != nil {
			fail(2, quiet, "could not find command \"%s\"", os.Args[terminator+1])
		}

		execArgs = os.Args[terminator+2 : len(os.Args)]
	} else {
		extra = flag.Args()
	}

	if len(extra) != 0 {
		fail(1, quiet, "no argument expected before the terminator (put -- before the command to execute)")
	} else if execCommand == "" {
		fail(1, quiet, "no command to execute was provided")
	}

	execCh := make(chan *commandOptions, buffer)

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {

		// Refuse non-POST requests
		if strings.ToUpper(r.Method) != "POST" {
			w.WriteHeader(405)
			return
		}

		// Refuse request if unauthorized.
		if !utils.Authorized(auth, r) {
			w.WriteHeader(401)
			return
		}

		opts := &commandOptions{}
		opts.command = execCommand
		opts.args = execArgs
		opts.cwd = cwd

		var env []string
		env = append(env, os.Environ()...)
		for headerName, headerVals := range r.Header {
			env = append(env, utils.EnvPrefix+"REQUEST_HEADER_"+strings.ToUpper(httpRequestHeaderToEnvVarRegexp.ReplaceAllString(headerName, "_"))+"="+strings.Join(headerVals, ","))
		}

		opts.env = env

		pr, pw := io.Pipe()
		opts.reader = pr

		// Refuse extra requests if buffer is full.
		select {
		case execCh <- opts:
			w.WriteHeader(202)
		default:
			w.WriteHeader(429)
		}

		io.Copy(pw, r.Body)
		pw.Close()
	})

	s := &http.Server{
		Addr: ":" + strconv.FormatUint(uint64(port), 10),
	}

	go worker(concurrency, execCh)

	s.ListenAndServe()
}

func worker(concurrency uint64, execChannel chan *commandOptions) {
	fmt.Fprintf(os.Stderr, "Execution worker started\n")

	n := uint64(0)
	wait := &sync.WaitGroup{}

	for opts := range execChannel {

		if concurrency >= 1 {
			n++
			wait.Add(1)
		}

		go execCommand(opts, wait)

		// Wait for queue to clear if concurrency is limited.
		if concurrency >= 1 && n >= concurrency {
			wait.Wait()
			n -= concurrency
		}
	}
}

func execCommand(opts *commandOptions, waitGroup *sync.WaitGroup) {

	fmt.Fprintf(os.Stderr, color.YellowString("Executing %s\n"), opts.command)

	cmd := exec.Command(opts.command, opts.args...)
	cmd.Env = opts.env
	cmd.Stderr = os.Stdout
	cmd.Stdout = os.Stdout
	cmd.Stdin = opts.reader

	err := cmd.Run()
	if err != nil {
		fmt.Fprintf(os.Stderr, color.RedString("Command %s error: %s\n"), opts.command, err)
	} else {
		fmt.Fprintf(os.Stderr, color.GreenString("Successfully executed %s\n"), opts.command)
	}

	opts.reader.Close()

	waitGroup.Done()
}

func fail(code int, quiet bool, format string, values ...interface{}) {
	if !quiet {
		fmt.Fprintf(os.Stderr, color.RedString("Error: "+format+"\n"), values...)
	}

	os.Exit(code)
}
