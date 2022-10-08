package echo

import (
	"testing"
	"time"

	"airbot/message"

	"github.com/google/go-cmp/cmp"
)

func TestTriHard(t *testing.T) {
	tests := []struct {
		desc  string
		input *message.IncomingMessage
		want  []*message.Message
	}{
		{
			desc: "handled message",
			input: &message.IncomingMessage{
				Message: message.Message{
					Text:    "$TriHard",
					User:    "someone",
					Channel: "somechannel",
					Time:    time.Date(2020, 5, 15, 10, 7, 0, 0, time.UTC),
				},
				Prefix: "$",
			},
			want: []*message.Message{
				{
					Text:    "TriHard 7",
					Channel: "somechannel",
				},
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.desc, func(t *testing.T) {
			got, err := triHard(tc.input)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if diff := cmp.Diff(tc.want, got); diff != "" {
				t.Errorf("triHard() diff (-want +got):\n%s", diff)
			}
		})
	}
}
