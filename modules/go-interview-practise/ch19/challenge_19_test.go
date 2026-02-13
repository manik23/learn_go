package ch19

import (
	"reflect"
	"testing"
)

func TestFindMax(t *testing.T) {
	tests := []struct {
		name    string
		numbers []int
		want    int
	}{
		{"Empty slice", []int{}, 0},
		{"Single element", []int{42}, 42},
		{"Multiple elements, positive only", []int{3, 1, 4, 1, 5, 9, 2, 6}, 9},
		{"Multiple elements, with negative", []int{-3, -1, -4, -1, -5, -9, -2, -6}, -1},
		{"Multiple elements, mixed signs", []int{-3, 1, -4, 1, -5, 9, -2, 6}, 9},
		{"Duplicate max values", []int{1, 9, 3, 9, 5}, 9},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := FindMax(tt.numbers); got != tt.want {
				t.Errorf("FindMax(%v) = %v, want %v", tt.numbers, got, tt.want)
			}
		})
	}
}

func TestRemoveDuplicates(t *testing.T) {
	tests := []struct {
		name    string
		numbers []int
		want    []int
	}{
		{"Empty slice", []int{}, []int{}},
		{"No duplicates", []int{1, 2, 3, 4, 5}, []int{1, 2, 3, 4, 5}},
		{"With duplicates", []int{3, 1, 4, 1, 5, 9, 2, 6}, []int{3, 1, 4, 5, 9, 2, 6}},
		{"All duplicates", []int{1, 1, 1, 1, 1}, []int{1}},
		{"Adjacent duplicates", []int{1, 1, 2, 2, 3, 3}, []int{1, 2, 3}},
		{"Non-adjacent duplicates", []int{1, 2, 3, 1, 2, 3}, []int{1, 2, 3}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := RemoveDuplicates(tt.numbers)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("RemoveDuplicates(%v) = %v, want %v", tt.numbers, got, tt.want)
			}
		})
	}
}

func TestReverseSlice(t *testing.T) {
	tests := []struct {
		name  string
		slice []int
		want  []int
	}{
		{"Empty slice", []int{}, []int{}},
		{"Single element", []int{42}, []int{42}},
		{"Even number of elements", []int{1, 2, 3, 4}, []int{4, 3, 2, 1}},
		{"Odd number of elements", []int{1, 2, 3, 4, 5}, []int{5, 4, 3, 2, 1}},
		{"With duplicates", []int{1, 2, 2, 3, 1}, []int{1, 3, 2, 2, 1}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ReverseSlice(tt.slice)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ReverseSlice(%v) = %v, want %v", tt.slice, got, tt.want)
			}

			// Ensure original slice wasn't modified
			if len(tt.slice) > 0 && len(got) > 0 {
				// Change the first element of the result
				got[0] = -999
				// Ensure the original slice is unaffected
				if len(tt.slice) > 0 && tt.slice[0] == -999 {
					t.Errorf("ReverseSlice modified the original slice")
				}
			}
		})
	}
}

func TestFilterEven(t *testing.T) {
	tests := []struct {
		name    string
		numbers []int
		want    []int
	}{
		{"Empty slice", []int{}, []int{}},
		{"No even numbers", []int{1, 3, 5, 7, 9}, []int{}},
		{"Only even numbers", []int{2, 4, 6, 8, 10}, []int{2, 4, 6, 8, 10}},
		{"Mixed numbers", []int{1, 2, 3, 4, 5, 6}, []int{2, 4, 6}},
		{"Negative numbers", []int{-1, -2, -3, -4}, []int{-2, -4}},
		{"Zero included", []int{0, 1, 2, 3}, []int{0, 2}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := FilterEven(tt.numbers)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("FilterEven(%v) = %v, want %v", tt.numbers, got, tt.want)
			}
		})
	}
}
