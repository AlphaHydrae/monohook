package utils

import (
	"fmt"
	"os"
	"regexp"
	"strconv"

	"github.com/fatih/color"
	flag "github.com/spf13/pflag"
)

var falseRegexp = regexp.MustCompile("^(?:0|n|no|f|false)$")
var trueRegexp = regexp.MustCompile("^(?:1|y|yes|t|true)$")

// EnvPrefix is the prefix common to environment variables for monohook.
const EnvPrefix = "MONOHOOK_"

// ErrorHandler can be called if an error occurs while parsing a command line
// option.
type ErrorHandler func(code int, message string)

// Fail prints a message in red to the standard error stream (as long as the
// `quiet` option is false) and exits the process with a non-zero code.
func Fail(code int, quiet bool, format string, values ...interface{}) {
	if !quiet {
		Print(quiet, fmt.Sprintf(color.RedString("Error: "+format+"\n"), values...))
	}

	os.Exit(code)
}

// Print prints a message to the standard error stream (as long as the `quiet`
// option is false).
func Print(quiet bool, format string, values ...interface{}) {
	if !quiet {
		fmt.Fprintf(os.Stderr, format, values...)
	}
}

// BoolOption parses and returns the value of a boolean option, either from
// command line flags or from an environment variable.
func BoolOption(value *bool, longOpt string, shortOpt string, envVarName string, defaultVal bool, desc string, fail ErrorHandler) {

	defaultFlagVal := defaultVal

	envVar := EnvPrefix + envVarName
	envValue := os.Getenv(envVar)
	if trueRegexp.MatchString(envValue) {
		defaultFlagVal = true
	} else if falseRegexp.MatchString(envValue) {
		defaultFlagVal = false
	} else if envValue != "" {
		fail(1, fmt.Sprintf("environment variable $%s must be one of: 0, n, no, f, false, 1, y, yes, t, true", envVar))
	}

	flag.BoolVarP(value, longOpt, shortOpt, defaultFlagVal, desc)
}

// StringOption parses and returns the value of a free-form text option, either
// from command line flags or from an environment variable.
func StringOption(value *string, longOpt string, shortOpt string, envVarName string, defaultVal string, desc string) {

	defaultFlagVal := defaultVal

	envVar := EnvPrefix + envVarName
	envValue := os.Getenv(envVar)
	if envValue != "" {
		defaultFlagVal = envValue
	}

	flag.StringVarP(value, longOpt, shortOpt, defaultFlagVal, desc)
}

// Uint64Option parses and returns the value of an unsigned integer option,
// either from command line flags or from an environment variable.
func Uint64Option(value *uint64, longOpt string, shortOpt string, envVarName string, defaultVal uint64, desc string, fail ErrorHandler) {

	defaultFlagVal := defaultVal

	envVar := EnvPrefix + envVarName
	envValue := os.Getenv(envVar)
	if envValue != "" {

		parsed, err := strconv.ParseUint(envValue, 10, 64)
		if err != nil {
			fail(1, fmt.Sprintf("environment variable $%s must be an integer greater than or equal to zero", envVar))
		}

		defaultFlagVal = parsed
	}

	flag.Uint64VarP(value, longOpt, shortOpt, defaultFlagVal, desc)
}
