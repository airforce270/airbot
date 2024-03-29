// Package argument contains command parameter/argument types.
package arg

import (
	"fmt"
	"log"
	"regexp"
	"slices"
	"strconv"
	"strings"
)

var (
	intPrefixPattern      = regexp.MustCompile("^([0-9]+)(.*)")
	stringPrefixPattern   = regexp.MustCompile(`^(\S+)(.*)`)
	booleanPrefixPattern  = regexp.MustCompile(fmt.Sprintf("^(%s)(.*)", strings.Join(booleanStrs, "|")))
	usernamePrefixPattern = regexp.MustCompile(`^@?(\S+)(.*)`)
	variadicPrefixPattern = regexp.MustCompile("(.+)")

	trueStrs    = []string{"on", "true", "enabled"}
	falseStrs   = []string{"off", "false", "disabled"}
	booleanStrs = append(trueStrs, falseStrs...)
)

// Param represents a single parameter to a command.
type Param struct {
	// Name is the name of the param.
	Name string
	// Type is the type of the param.
	// Default: String.
	Type Type
	// Required is whether the param is required.
	// Optional params should come last.
	// It is currently undefined behavior to have multiple optional params,
	// and to have any optional params with variadic args.
	Required bool
	// Usage is an optional human-readable string describing the param.
	// This is only used for the usage string, i.e. if Name:"myparam" Usage:"something",
	// the usage string will say $command <something> rather than $command <myparam>
	Usage string
}

// Arg represents a parsed argument from a message.
type Arg struct {
	// Present is whether the arg is present and a value is set.
	// The caller should check this field before accessing a value.
	Present bool
	// Type is the type of the arg.
	// This field indicates which value should be read from.
	Type Type
	// StringValue is the value of the arg, if it is a string.
	StringValue string
	// BoolValue is the value of the arg, if it is a bool.
	BoolValue bool
	// IntValue is the value of the arg, if it is an int.
	IntValue int64
}

// Parse parses the param it defines from the given message.
// It then returns the parsed arg and the remaining message text after parsing.
func (a Param) Parse(msg string) (Arg, string) {
	var pattern *regexp.Regexp
	switch a.Type {
	case Int:
		pattern = intPrefixPattern
	case Boolean:
		pattern = booleanPrefixPattern
	case Username:
		pattern = usernamePrefixPattern
	case Variadic:
		pattern = variadicPrefixPattern
	case String:
		pattern = stringPrefixPattern
	default:
		log.Printf("Warning: parsing unhandled message type (%q)! Treating as string.", a.Type)
		pattern = stringPrefixPattern
	}

	matches := pattern.FindStringSubmatch(msg)
	if len(matches) < 2 {
		return Arg{Present: false}, msg
	}

	match := matches[1]
	rest := ""
	if len(matches) > 2 {
		rest = strings.TrimSpace(matches[2])
	}

	switch a.Type {
	case Int:
		i, err := strconv.ParseInt(match, 10, 64)
		if err != nil {
			return Arg{Present: false}, msg
		}
		return Arg{
			Present:  true,
			Type:     a.Type,
			IntValue: i,
		}, rest
	case Boolean:
		if !slices.Contains(booleanStrs, match) {
			return Arg{Present: false}, msg
		}
		return Arg{
			Present:   true,
			Type:      a.Type,
			BoolValue: slices.Contains(trueStrs, match),
		}, rest
	default:
		return Arg{
			Present:     true,
			Type:        a.Type,
			StringValue: match,
		}, rest
	}
}

// UsageForDocString returns the usage information for putting in a usage docstring,
// i.e. for a command named "mycommand" with a param named "myparam" with param usage of "myparamusage"
// $command <myparamusage>
func (p Param) UsageForDocString() string {
	if p.Usage != "" {
		return p.Usage
	}
	if p.Type == Boolean {
		return "on|off"
	}
	return p.Name
}

// Type is the type of a parameter.
// It affects both how the parame will be parsed to an arg, and the type it will be parsed to.
type Type uint8

const (
	// String is an parameter accepting any string, stopping at whitespace.
	// Type of value: string.
	String Type = iota + 1
	// Int is an parameter accepting a single integer.
	// Type of value: int.
	Int
	// Boolean is an parameter accepting either true or false.
	// True values: "on", "true", "enabled"
	// False values: "off", "false", "disabled"
	// Type of value: bool.
	Boolean
	// Username is an parameter accepting usernames,
	// which are strings that can take the format username or @username.
	// Type of value: string.
	Username
	// Variadic represents variadic parameters.
	// i.e.: this type will consume the rest of the remaining parameters,
	// and should only be used as the terminal parameter in a list.
	// Type of value: string.
	Variadic
)
