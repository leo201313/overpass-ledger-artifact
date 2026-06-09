package smartcontract

import (
	"sort"
	"testing"
)

// manualQSort implements a manual quicksort algorithm
func manualQSort(arr []int, left int, right int) {
	i := left
	j := right

	if i >= j {
		return
	}

	pivot := arr[left+(right-left)/2]

	for i <= j {
		for arr[i] < pivot {
			i++
		}
		for arr[j] > pivot {
			j--
		}
		if i <= j {
			arr[i], arr[j] = arr[j], arr[i]
			i++
			j--
		}
	}

	if left < j {
		manualQSort(arr, left, j)
	}
	if i < right {
		manualQSort(arr, i, right)
	}
}

// BenchmarkManualQSort benchmarks the manual quicksort implementation
func BenchmarkManualQSort(b *testing.B) {
	for n := 0; n < b.N; n++ {
		arr := generateReverseOrderSlice(100000)
		manualQSort(arr, 0, len(arr)-1)
	}
}

// BenchmarkSortInts benchmarks the standard library's sort.Ints
func BenchmarkSortInts(b *testing.B) {
	for n := 0; n < b.N; n++ {
		arr := generateReverseOrderSlice(100000)
		sort.Ints(arr)
	}
}

// generateReverseOrderSlice generates a reverse order slice of integers of a given size
func generateReverseOrderSlice(size int) []int {
	slice := make([]int, size)
	for i := range slice {
		slice[i] = size - i
	}
	return slice
}
