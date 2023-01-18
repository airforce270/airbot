package arg

import (
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestParse(t *testing.T) {
	tests := []struct {
		desc       string
		inputArg   Param
		inputMsg   string
		wantParsed Arg
		wantLeft   string
	}{
		{
			desc: "int arg, present",
			inputArg: Param{
				Name: "intcommand",
				Type: Int,
			},
			inputMsg: "123 something else",
			wantParsed: Arg{
				IsPresent: true,
				Value:     int(123),
			},
			wantLeft: "something else",
		},
		{
			desc: "int arg, not present",
			inputArg: Param{
				Name: "intcommand",
				Type: Int,
			},
			inputMsg: "something else",
			wantParsed: Arg{
				IsPresent: false,
			},
			wantLeft: "something else",
		},
		{
			desc: "int arg, not first",
			inputArg: Param{
				Name: "intcommand",
				Type: Int,
			},
			inputMsg: "something 123 something else",
			wantParsed: Arg{
				IsPresent: false,
			},
			wantLeft: "something 123 something else",
		},
		{
			desc: "int arg, empty",
			inputArg: Param{
				Name: "intcommand",
				Type: Int,
			},
			inputMsg: "",
			wantParsed: Arg{
				IsPresent: false,
			},
			wantLeft: "",
		},
		{
			desc: "string arg, present",
			inputArg: Param{
				Name: "stringcommand",
				Type: String,
			},
			inputMsg: "something something else",
			wantParsed: Arg{
				IsPresent: true,
				Value:     "something",
			},
			wantLeft: "something else",
		},
		{
			desc: "string arg, empty",
			inputArg: Param{
				Name: "stringcommand",
				Type: String,
			},
			inputMsg: "",
			wantParsed: Arg{
				IsPresent: false,
			},
			wantLeft: "",
		},
		{
			desc: "boolean arg, on",
			inputArg: Param{
				Name: "booleancommand",
				Type: Boolean,
			},
			inputMsg: "on blah",
			wantParsed: Arg{
				IsPresent: true,
				Value:     true,
			},
			wantLeft: "blah",
		},
		{
			desc: "boolean arg, true",
			inputArg: Param{
				Name: "booleancommand",
				Type: Boolean,
			},
			inputMsg: "true blah",
			wantParsed: Arg{
				IsPresent: true,
				Value:     true,
			},
			wantLeft: "blah",
		},
		{
			desc: "boolean arg, enabled",
			inputArg: Param{
				Name: "booleancommand",
				Type: Boolean,
			},
			inputMsg: "enabled blah",
			wantParsed: Arg{
				IsPresent: true,
				Value:     true,
			},
			wantLeft: "blah",
		},
		{
			desc: "boolean arg, off",
			inputArg: Param{
				Name: "booleancommand",
				Type: Boolean,
			},
			inputMsg: "off blah",
			wantParsed: Arg{
				Value:     false,
				IsPresent: true,
			},
			wantLeft: "blah",
		},
		{
			desc: "boolean arg, false",
			inputArg: Param{
				Name: "booleancommand",
				Type: Boolean,
			},
			inputMsg: "false blah",
			wantParsed: Arg{
				Value:     false,
				IsPresent: true,
			},
			wantLeft: "blah",
		},
		{
			desc: "boolean arg, disabled",
			inputArg: Param{
				Name: "booleancommand",
				Type: Boolean,
			},
			inputMsg: "disabled blah",
			wantParsed: Arg{
				Value:     false,
				IsPresent: true,
			},
			wantLeft: "blah",
		},
		{
			desc: "boolean arg, not present",
			inputArg: Param{
				Name: "booleancommand",
				Type: Boolean,
			},
			inputMsg: "something else",
			wantParsed: Arg{
				IsPresent: false,
			},
			wantLeft: "something else",
		},
		{
			desc: "boolean arg, not first",
			inputArg: Param{
				Name: "booleancommand",
				Type: Boolean,
			},
			inputMsg: "something else on",
			wantParsed: Arg{
				IsPresent: false,
			},
			wantLeft: "something else on",
		},
		{
			desc: "boolean arg, empty",
			inputArg: Param{
				Name: "booleancommand",
				Type: Boolean,
			},
			inputMsg: "",
			wantParsed: Arg{
				IsPresent: false,
			},
			wantLeft: "",
		},
		{
			desc: "username arg with at-symbol",
			inputArg: Param{
				Name: "usernamecommand",
				Type: Username,
			},
			inputMsg: "@someuser something else",
			wantParsed: Arg{
				IsPresent: true,
				Value:     "someuser",
			},
			wantLeft: "something else",
		},
		{
			desc: "username arg without at-symbol",
			inputArg: Param{
				Name: "usernamecommand",
				Type: Username,
			},
			inputMsg: "someuser something else",
			wantParsed: Arg{
				IsPresent: true,
				Value:     "someuser",
			},
			wantLeft: "something else",
		},
		{
			desc: "username arg, empty",
			inputArg: Param{
				Name: "usernamecommand",
				Type: Username,
			},
			inputMsg: "",
			wantParsed: Arg{
				IsPresent: false,
			},
			wantLeft: "",
		},
		{
			desc: "variadic arg, present",
			inputArg: Param{
				Name: "variadiccommand",
				Type: Variadic,
			},
			inputMsg: "blah blah 123 something",
			wantParsed: Arg{
				IsPresent: true,
				Value:     "blah blah 123 something",
			},
			wantLeft: "",
		},
		{
			desc: "variadic arg, empty",
			inputArg: Param{
				Name: "variadiccommand",
				Type: Variadic,
			},
			inputMsg: "",
			wantParsed: Arg{
				IsPresent: false,
			},
			wantLeft: "",
		},
	}

	for _, tc := range tests {
		t.Run(tc.desc, func(t *testing.T) {
			gotParsed, gotLeft := tc.inputArg.Parse(tc.inputMsg)

			if diff := cmp.Diff(tc.wantParsed, gotParsed); diff != "" {
				t.Errorf("Arg.Parse() diff (-want +got):\n%s", diff)
			}
			if gotLeft != tc.wantLeft {
				t.Errorf("Arg.Parse() left = %q, want %q", gotLeft, tc.wantLeft)
			}
		})
	}
}
