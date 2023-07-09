package slices_test

import (
	"testing"

	"github.com/romshark/taskhub/slices"

	"github.com/stretchr/testify/require"
)

func TestWithoutIndex(t *testing.T) {
	for _, td := range []struct {
		name   string
		input  []int
		index  int
		expect []int
	}{
		{name: "1", input: []int{0}, index: 0, expect: []int{}},
		{name: "3", input: []int{1, 2, 3}, index: 1, expect: []int{1, 3}},
	} {
		t.Run(td.name, func(t *testing.T) {
			slices.WithoutIndex(td.input, 0)
		})
	}

	t.Run("nil", func(t *testing.T) {
		require.Panics(t, func() {
			slices.WithoutIndex([]int(nil), 0)
		})
	})
	t.Run("out_of_bound", func(t *testing.T) {
		require.Panics(t, func() {
			slices.WithoutIndex([]int{1, 2}, 2)
		})
	})
}

func TestIsSubset(t *testing.T) {
	for _, td := range []struct {
		name   string
		a, b   []int
		expect bool
	}{
		{"all_match_2", []int{1, 2}, []int{1, 2, 3}, true},
		{"all_match_3_duplicate", []int{2, 1, 1}, []int{1, 2, 3}, true},
		{"all_match_1", []int{1}, []int{1, 2, 3}, true},
		{"empty_a", []int{}, []int{1, 2, 3}, true},
		{"nil_a", nil, []int{1, 2, 3}, true},
		{"nil_both", nil, nil, true},

		// Expect false
		{"no_match", []int{1, 5}, []int{1, 2, 3}, false},
		{"nil_b", []int{1}, nil, false},
	} {
		t.Run(td.name, func(t *testing.T) {
			a := slices.IsSubset(td.a, td.b)
			require.Equal(t, td.expect, a)
		})
	}
}

func TestIsSubsetGet(t *testing.T) {
	for _, td := range []struct {
		name   string
		a, b   []int
		expect bool
	}{
		{"all_match_2", []int{1, 2}, []int{1, 2, 3}, true},
		{"all_match_3_duplicate", []int{2, 1, 1}, []int{1, 2, 3}, true},
		{"all_match_1", []int{1}, []int{1, 2, 3}, true},
		{"empty_a", []int{}, []int{1, 2, 3}, true},
		{"nil_a", nil, []int{1, 2, 3}, true},
		{"nil_both", nil, nil, true},

		// Expect false
		{"no_match", []int{1, 5}, []int{1, 2, 3}, false},
		{"nil_b", []int{1}, nil, false},
	} {
		t.Run(td.name, func(t *testing.T) {
			a := slices.IsSubsetGet(td.a, td.b, func(x int) int { return x })
			require.Equal(t, td.expect, a)
		})
	}
}

func TestFilterInPlace(t *testing.T) {
	t.Run("filter_odd", func(t *testing.T) {
		s := []int{1, 2, 3, 4}
		s = slices.FilterInPlace(s, func(x int) bool { return x%2 == 0 })
		require.Equal(t, []int{2, 4}, s)
	})

	t.Run("filter_odd_none", func(t *testing.T) {
		s := []int{2, 4, 8, 16, 32}
		s = slices.FilterInPlace(s, func(x int) bool { return x%2 == 0 })
		require.Equal(t, []int{2, 4, 8, 16, 32}, s)
	})

	t.Run("filter_odd_all", func(t *testing.T) {
		s := []int{1, 3, 5}
		s = slices.FilterInPlace(s, func(x int) bool { return x%2 == 0 })
		require.Len(t, s, 0)
	})

	t.Run("empty", func(t *testing.T) {
		s := slices.FilterInPlace([]int{}, func(x int) bool { return x%2 == 0 })
		require.Equal(t, []int{}, s)
	})

	t.Run("nil", func(t *testing.T) {
		s := slices.FilterInPlace(nil, func(x int) bool { return x%2 == 0 })
		require.Nil(t, s)
	})
}

func TestCopy(t *testing.T) {
	t.Run("non_empty", func(t *testing.T) {
		s := []int{1, 2, 3}
		c := slices.Copy(s)
		require.Equal(t, s, c)
		c[1] = 42
		require.NotEqual(t, s, c)
	})

	t.Run("empty", func(t *testing.T) {
		c := slices.Copy([]int{})
		require.Equal(t, []int{}, c)
	})

	t.Run("nil", func(t *testing.T) {
		c := slices.Copy([]int(nil))
		require.Nil(t, c)
	})
}

func TestSortAndLimit(t *testing.T) {
	sortFn := func(a, b int) bool { return a > b }
	t.Run("non_empty", func(t *testing.T) {
		s := []int{3, 1, 2, 0, 6}
		s = slices.SortAndLimit(s, sortFn, 4)
		require.Equal(t, []int{6, 3, 2, 1}, s)
	})

	t.Run("limit_higher_than_len", func(t *testing.T) {
		s := []int{3, 1, 2, 0, 6}
		s = slices.SortAndLimit(s, sortFn, 42)
		require.Equal(t, []int{6, 3, 2, 1, 0}, s)
	})

	t.Run("limit_equal_len", func(t *testing.T) {
		s := []int{3, 1, 2, 0, 6}
		s = slices.SortAndLimit(s, sortFn, len(s))
		require.Equal(t, []int{6, 3, 2, 1, 0}, s)
	})

	t.Run("only_sort", func(t *testing.T) {
		s := []int{3, 1, 2, 0, 6}
		s = slices.SortAndLimit(s, sortFn, -1)
		require.Equal(t, []int{6, 3, 2, 1, 0}, s)
	})

	t.Run("only_limit", func(t *testing.T) {
		s := []int{3, 1, 2, 0, 6}
		s = slices.SortAndLimit(s, nil, 4)
		require.Equal(t, []int{3, 1, 2, 0}, s)
	})

	t.Run("noop", func(t *testing.T) {
		s := []int{3, 1, 2, 0, 6}
		s = slices.SortAndLimit(s, nil, -1)
		require.Equal(t, []int{3, 1, 2, 0, 6}, s)
	})

	t.Run("nil", func(t *testing.T) {
		s := slices.SortAndLimit([]int(nil), sortFn, 4)
		require.Nil(t, s)
	})

	t.Run("nil_only_sort", func(t *testing.T) {
		s := slices.SortAndLimit([]int(nil), sortFn, -1)
		require.Nil(t, s)
	})

	t.Run("nil_only_limit", func(t *testing.T) {
		s := slices.SortAndLimit([]int(nil), nil, 2)
		require.Nil(t, s)
	})

	t.Run("nil_noop", func(t *testing.T) {
		s := slices.SortAndLimit([]int(nil), nil, -1)
		require.Nil(t, s)
	})
}

func TestContains(t *testing.T) {
	t.Run("contains", func(t *testing.T) {
		s := []int{1, 2, 3}
		a := slices.Contains(s, 2)
		require.True(t, a)
		require.Equal(t, []int{1, 2, 3}, s, "expected unchanged original")
	})

	t.Run("not_contains", func(t *testing.T) {
		s := []int{1, 2, 3}
		a := slices.Contains(s, 4)
		require.False(t, a)
		require.Equal(t, []int{1, 2, 3}, s, "expected unchanged original")
	})

	t.Run("empty", func(t *testing.T) {
		s := []int{}
		a := slices.Contains(s, 1)
		require.False(t, a)
		require.Equal(t, []int{}, s, "expected unchanged original")
	})

	t.Run("nil", func(t *testing.T) {
		a := slices.Contains(nil, 1)
		require.False(t, a)
	})
}

func TestAppendUnique(t *testing.T) {
	t.Run("append", func(t *testing.T) {
		s := []int{1, 2, 3}
		s = slices.AppendUnique(s, 4)
		require.Equal(t, []int{1, 2, 3, 4}, s)
	})

	t.Run("duplicate", func(t *testing.T) {
		s := []int{1, 2, 3}
		s = slices.AppendUnique(s, 2)
		require.Equal(t, []int{1, 2, 3}, s)
	})

	t.Run("nil", func(t *testing.T) {
		s := slices.AppendUnique([]int(nil), 1)
		require.Equal(t, []int{1}, s)
	})
}
