package utils

import (
	"fmt"
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestChunk(t *testing.T) {
	tests := []struct {
		input []int
		size  int
		want  [][]int
	}{
		{
			input: []int{1, 2, 3},
			size:  5,
			want: [][]int{
				{1, 2, 3},
			},
		},
		{
			input: []int{1, 2, 3, 4, 5},
			size:  5,
			want: [][]int{
				{1, 2, 3, 4, 5},
			},
		},
		{
			input: []int{1, 2, 3, 4, 5, 6},
			size:  5,
			want: [][]int{
				{1, 2, 3, 4, 5},
				{6},
			},
		},
		{
			input: []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15},
			size:  5,
			want: [][]int{
				{1, 2, 3, 4, 5},
				{6, 7, 8, 9, 10},
				{11, 12, 13, 14, 15},
			},
		},
	}

	for _, tc := range tests {
		t.Run(fmt.Sprintf("%d", len(tc.input)), func(t *testing.T) {
			got := Chunk(tc.input, tc.size)
			if diff := cmp.Diff(tc.want, got); diff != "" {
				t.Errorf("Chunk() diff (-want +got):\n%s", diff)
			}
		})
	}
}
