package utils

import (
	"fmt"
	"os"
	"regexp"
	"strconv"

	flag "github.com/spf13/pflag"
)

const envPrefix = "MONOHOOK_"

var falseRegexp = regexp.MustCompile("^(?:0|n|no|f|false)$")
var trueRegexp = regexp.MustCompile("^(?:1|y|yes|t|true)$")

// ErrorHandler can be called if an error occurs while parsing a command line
// option.
type ErrorHandler func(code int, message string)

// BoolOption parses and returns the value of a boolean option, either from
// command line flags or from an environment variable.
func BoolOption(value *bool, longOpt string, shortOpt string, envVarName string, defaultVal bool, desc string, fail ErrorHandler) {

	defaultFlagVal := defaultVal

	envVar := envPrefix + envVarName
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

	envVar := envPrefix + envVarName
	envValue := os.Getenv(envVar)
	if envValue != "" {
		defaultFlagVal = envValue
	}

	flag.StringVarP(value, longOpt, shortOpt, defaultFlagVal, desc)
}

// Uint16Option parses and returns the value of an unsigned integer option with
// a maximum value of 65535, either from command line flags or from an
// environment variable.
func Uint16Option(value *uint16, longOpt string, shortOpt string, envVarName string, defaultVal uint64, desc string, fail ErrorHandler) {

	var uint64Value uint64

	Uint64Option(&uint64Value, longOpt, shortOpt, envVarName, defaultVal, desc, fail)
	if uint64Value > 65535 {
		fail(1, fmt.Sprintf("option --%s or environment variable $%s must be an integer smaller than or equal to 65535", longOpt, envVarName))
	}

	*value = uint16(uint64Value)
}

// Uint64Option parses and returns the value of an unsigned integer option,
// either from command line flags or from an environment variable.
func Uint64Option(value *uint64, longOpt string, shortOpt string, envVarName string, defaultVal uint64, desc string, fail ErrorHandler) {

	defaultFlagVal := defaultVal

	envVar := envPrefix + envVarName
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
