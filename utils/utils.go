// Package utils contains utility functions that don't make sense to have their own namespaced packages.
package utils

// Chunk breaks a slice of items into slices of a max length.
func Chunk[T any](items []T, chunkSize int) (chunks [][]T) {
	for chunkSize < len(items) {
		items, chunks = items[chunkSize:], append(chunks, items[0:chunkSize:chunkSize])
	}
	return append(chunks, items)
}
