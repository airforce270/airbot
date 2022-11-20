package basecommand

import (
	"fmt"
	"testing"

	"github.com/airforce270/airbot/base"
)

func TestCommandUsage(t *testing.T) {
	prefix := "$"
	tests := []struct {
		desc  string
		input Command
		want  string
	}{
		{
			desc: "single required arg",
			input: Command{
				Name: "mycommand",
				Args: []Argument{
					{
						Name:     "arg1",
						Required: true,
						Usage:    "",
					},
				},
			},
			want: prefix + "mycommand <arg1>",
		},
		{
			desc: "multiple args, one required",
			input: Command{
				Name: "mycommand",
				Args: []Argument{
					{
						Name:     "arg1",
						Required: true,
						Usage:    "arg1usage",
					},
					{
						Name:     "optionalarg",
						Required: false,
						Usage:    "",
					},
				},
			},
			want: prefix + "mycommand <arg1usage> [optionalarg]",
		},
	}

	for _, tc := range tests {
		t.Run(tc.desc, func(t *testing.T) {
			if got := tc.input.Usage(prefix); got != tc.want {
				t.Errorf("Usage() = %s, want %s", got, tc.want)
			}
		})
	}
}

func TestArgumentUsageForDocString(t *testing.T) {
	tests := []struct {
		desc  string
		input Argument
		want  string
	}{
		{
			desc: "no usage",
			input: Argument{
				Name:     "myarg",
				Required: true,
				Usage:    "",
			},
			want: "myarg",
		},
		{
			desc: "with usage",
			input: Argument{
				Name:     "myarg",
				Required: true,
				Usage:    "myargusage",
			},
			want: "myargusage",
		},
	}

	for _, tc := range tests {
		t.Run(tc.desc, func(t *testing.T) {
			if got := tc.input.UsageForDocString(); got != tc.want {
				t.Errorf("UsageForDocString() = %s, want %s", got, tc.want)
			}
		})
	}
}

func TestFirstArgOrUsername(t *testing.T) {
	tests := []struct {
		args []string
		msg  *base.IncomingMessage
		want string
	}{
		{
			args: nil,
			msg: &base.IncomingMessage{
				Message: base.Message{
					User:    "user1",
					Channel: "channel1",
				},
			},
			want: "user1",
		},
		{
			args: []string{"someone"},
			msg: &base.IncomingMessage{
				Message: base.Message{
					User:    "user1",
					Channel: "channel1",
				},
			},
			want: "someone",
		},
		{
			args: []string{"someone", "someoneelse"},
			msg: &base.IncomingMessage{
				Message: base.Message{
					User:    "user1",
					Channel: "channel1",
				},
			},
			want: "someone",
		},
	}

	for _, tc := range tests {
		t.Run(fmt.Sprintf("len(args)=%d", len(tc.args)), func(t *testing.T) {
			if got := FirstArgOrUsername(tc.args, tc.msg); got != tc.want {
				t.Errorf("FirstArgOrUsername() = %s, want %s", got, tc.want)
			}
		})
	}
}

func TestFirstArgOrChannel(t *testing.T) {
	tests := []struct {
		args []string
		msg  *base.IncomingMessage
		want string
	}{
		{
			args: nil,
			msg: &base.IncomingMessage{
				Message: base.Message{
					User:    "user1",
					Channel: "channel1",
				},
			},
			want: "channel1",
		},
		{
			args: []string{"someone"},
			msg: &base.IncomingMessage{
				Message: base.Message{
					User:    "user1",
					Channel: "channel1",
				},
			},
			want: "someone",
		},
		{
			args: []string{"someone", "someoneelse"},
			msg: &base.IncomingMessage{
				Message: base.Message{
					User:    "user1",
					Channel: "channel1",
				},
			},
			want: "someone",
		},
	}

	for _, tc := range tests {
		t.Run(fmt.Sprintf("len(args)=%d", len(tc.args)), func(t *testing.T) {
			if got := FirstArgOrChannel(tc.args, tc.msg); got != tc.want {
				t.Errorf("FirstArgOrChannel() = %s, want %s", got, tc.want)
			}
		})
	}
}
