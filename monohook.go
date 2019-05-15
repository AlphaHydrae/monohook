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
	quiet   bool
	reader  *io.PipeReader
}

var httpRequestHeaderToEnvVarRegexp = regexp.MustCompile("-")

func main() {

	var auth string
	var buffer uint64
	var concurrency uint64
	var cwd string
	var forwardRequestBody bool
	var forwardRequestHeaders bool
	var forwardRequestURL bool
	var port uint16

	quiet := false
	errHandler := func(code int, message string) {
		utils.Fail(code, quiet, message)
	}

	flag.CommandLine.SetOutput(os.Stdout)
	utils.BoolOption(&quiet, "quiet", "q", "QUIET", false, "Do not print anything except the command's standard output and error streams (default false)", errHandler)

	utils.StringOption(&auth, "authorization", "a", "AUTHORIZATION", "", "Authentication token that must be sent as a Bearer token in the 'Authorization' header or as the 'authorization' URL query parameter")
	utils.Uint64Option(&buffer, "buffer", "b", "BUFFER", 10, "Maximum number of requests to queue before refusing subsequent ones until the queue is freed (zero for infinite)", errHandler)
	utils.Uint64Option(&concurrency, "concurrency", "c", "CONCURRENCY", 1, "Maximum number of times the command should be executed in parallel (zero for infinite concurrency)", errHandler)
	utils.BoolOption(&forwardRequestBody, "forward-request-body", "B", "FORWARD_REQUEST_BODY", false, "Whether to forward each HTTP request's body to the the command's standard input", errHandler)
	utils.BoolOption(&forwardRequestHeaders, "forward-request-headers", "H", "FORWARD_REQUEST_HEADERS", false, "Whether to forward each HTTP request's headers to the the command as environment variables (e.g. Content-Type becomes $MONOHOOK_REQUEST_HEADER_CONTENT_TYPE)", errHandler)
	utils.BoolOption(&forwardRequestURL, "forward-request-url", "U", "FORWARD_REQUEST_URL", false, "Whether to forward each HTTP request's URL to the the command as the $MONOHOOK_REQUEST_URL environment variable", errHandler)
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
			utils.Fail(2, quiet, "could not find command \"%s\"", os.Args[terminator+1])
		}

		execArgs = os.Args[terminator+2 : len(os.Args)]
	} else {
		extra = flag.Args()
	}

	if len(extra) != 0 {
		utils.Fail(1, quiet, "no argument expected before the terminator (put -- before the command to execute)")
	} else if execCommand == "" {
		utils.Fail(1, quiet, "no command to execute was provided")
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
		opts.quiet = quiet

		// Forward HTTP request body to the command's standard input.
		var pw *io.PipeWriter
		if forwardRequestBody {
			reader, writer := io.Pipe()
			opts.reader = reader
			pw = writer
		}

		var env []string
		env = append(env, os.Environ()...)

		// Forward HTTP request headers to the command as environment variables.
		if forwardRequestHeaders {
			for headerName, headerVals := range r.Header {
				env = append(env, utils.EnvPrefix+"REQUEST_HEADER_"+strings.ToUpper(httpRequestHeaderToEnvVarRegexp.ReplaceAllString(headerName, "_"))+"="+strings.Join(headerVals, ","))
			}
		}

		// Forward the HTTP request URL to the command as an environment variable
		if forwardRequestURL {
			env = append(env, utils.EnvPrefix+"REQUEST_URL="+r.URL.String())
		}

		opts.env = env

		// Refuse extra requests if buffer is full.
		select {
		case execCh <- opts:
			w.WriteHeader(202)
		default:
			w.WriteHeader(429)
		}

		if pw != nil {
			io.Copy(pw, r.Body)
			pw.Close()
		}
	})

	s := &http.Server{
		Addr: ":" + strconv.FormatUint(uint64(port), 10),
	}

	go worker(concurrency, quiet, execCh)

	s.ListenAndServe()
}

func worker(concurrency uint64, quiet bool, execChannel chan *commandOptions) {
	utils.Print(quiet, "Execution worker started\n")

	n := uint64(0)
	wait := &sync.WaitGroup{}

	for opts := range execChannel {
		if concurrency == 0 {
			go execCommand(opts, nil)
		} else {

			n++
			wait.Add(1)
			go execCommand(opts, wait)

			// Wait for queue to clear if concurrency is limited.
			if n >= concurrency {
				wait.Wait()
				n -= concurrency
				wait = &sync.WaitGroup{}
			}
		}
	}
}

func execCommand(opts *commandOptions, waitGroup *sync.WaitGroup) {

	utils.Print(opts.quiet, color.YellowString("Executing %s\n"), opts.command)

	cmd := exec.Command(opts.command, opts.args...)
	cmd.Dir = opts.cwd
	cmd.Env = opts.env
	cmd.Stderr = os.Stdout
	cmd.Stdout = os.Stdout

	if opts.reader != nil {
		cmd.Stdin = opts.reader
	}

	err := cmd.Run()
	if err != nil {
		utils.Print(opts.quiet, color.RedString("Command %s error: %s\n"), opts.command, err)
	} else {
		utils.Print(opts.quiet, color.GreenString("Successfully executed %s\n"), opts.command)
	}

	if opts.reader != nil {
		opts.reader.Close()
	}

	if waitGroup != nil {
		waitGroup.Done()
	}
}
