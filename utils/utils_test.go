package utils_test

import (
	"strconv"
	"testing"

	"github.com/airforce270/airbot/utils"

	"github.com/google/go-cmp/cmp"
)

func TestChunk(t *testing.T) {
	t.Parallel()
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
		tc := tc
		t.Run(strconv.Itoa(len(tc.input)), func(t *testing.T) {
			t.Parallel()
			got := utils.Chunk(tc.input, tc.size)
			if diff := cmp.Diff(tc.want, got); diff != "" {
				t.Errorf("Chunk() diff (-want +got):\n%s", diff)
			}
		})
	}
}
