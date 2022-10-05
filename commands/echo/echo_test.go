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
		input *message.Message
		want  *message.Message
	}{
		{
			desc: "handled message",
			input: &message.Message{
				Text:    "TriHard any homies",
				User:    "someone",
				Channel: "somechannel",
				Time:    time.Date(2020, 5, 15, 10, 7, 0, 0, time.UTC),
			},
			want: &message.Message{
				Text:    "TriHard 7",
				Channel: "somechannel",
			},
		},
		{
			desc: "unhandled but close message",
			input: &message.Message{
				Text:    "TriHard 7",
				User:    "someone",
				Channel: "somechannel",
				Time:    time.Date(2020, 5, 15, 10, 7, 0, 0, time.UTC),
			},
			want: nil,
		},
		{
			desc: "unhandled message",
			input: &message.Message{
				Text:    "PeepoGlad :rose:",
				User:    "someone",
				Channel: "somechannel",
				Time:    time.Date(2020, 5, 15, 10, 7, 0, 0, time.UTC),
			},
			want: nil,
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
