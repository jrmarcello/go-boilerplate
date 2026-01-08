package person

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestListFilter_Normalize(t *testing.T) {
	tests := []struct {
		name          string
		input         ListFilter
		expectedPage  int
		expectedLimit int
	}{
		{
			name:          "zero values - should use defaults",
			input:         ListFilter{Page: 0, Limit: 0},
			expectedPage:  1,
			expectedLimit: DefaultLimit,
		},
		{
			name:          "negative page - should normalize to 1",
			input:         ListFilter{Page: -5, Limit: 10},
			expectedPage:  1,
			expectedLimit: 10,
		},
		{
			name:          "negative limit - should use default",
			input:         ListFilter{Page: 1, Limit: -10},
			expectedPage:  1,
			expectedLimit: DefaultLimit,
		},
		{
			name:          "limit exceeds max - should cap to max",
			input:         ListFilter{Page: 1, Limit: 500},
			expectedPage:  1,
			expectedLimit: MaxLimit,
		},
		{
			name:          "valid values - should keep as is",
			input:         ListFilter{Page: 5, Limit: 50},
			expectedPage:  5,
			expectedLimit: 50,
		},
		{
			name:          "limit at max boundary - should keep",
			input:         ListFilter{Page: 1, Limit: MaxLimit},
			expectedPage:  1,
			expectedLimit: MaxLimit,
		},
		{
			name:          "limit at min valid - should keep",
			input:         ListFilter{Page: 1, Limit: 1},
			expectedPage:  1,
			expectedLimit: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			filter := tt.input
			filter.Normalize()

			assert.Equal(t, tt.expectedPage, filter.Page)
			assert.Equal(t, tt.expectedLimit, filter.Limit)
		})
	}
}

func TestListFilter_Offset(t *testing.T) {
	tests := []struct {
		name           string
		page           int
		limit          int
		expectedOffset int
	}{
		{
			name:           "first page",
			page:           1,
			limit:          20,
			expectedOffset: 0,
		},
		{
			name:           "second page",
			page:           2,
			limit:          20,
			expectedOffset: 20,
		},
		{
			name:           "third page with custom limit",
			page:           3,
			limit:          50,
			expectedOffset: 100,
		},
		{
			name:           "page 10 with limit 10",
			page:           10,
			limit:          10,
			expectedOffset: 90,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			filter := ListFilter{Page: tt.page, Limit: tt.limit}
			assert.Equal(t, tt.expectedOffset, filter.Offset())
		})
	}
}

func TestListResult_TotalPages(t *testing.T) {
	tests := []struct {
		name               string
		total              int
		limit              int
		expectedTotalPages int
	}{
		{
			name:               "no results",
			total:              0,
			limit:              20,
			expectedTotalPages: 0,
		},
		{
			name:               "less than one page",
			total:              5,
			limit:              20,
			expectedTotalPages: 1,
		},
		{
			name:               "exactly one page",
			total:              20,
			limit:              20,
			expectedTotalPages: 1,
		},
		{
			name:               "one page plus one item",
			total:              21,
			limit:              20,
			expectedTotalPages: 2,
		},
		{
			name:               "exactly two pages",
			total:              40,
			limit:              20,
			expectedTotalPages: 2,
		},
		{
			name:               "100 items with limit 15",
			total:              100,
			limit:              15,
			expectedTotalPages: 7, // 100/15 = 6.66 -> 7
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ListResult{Total: tt.total, Limit: tt.limit}
			assert.Equal(t, tt.expectedTotalPages, result.TotalPages())
		})
	}
}

func TestListResult_HasNextPage(t *testing.T) {
	tests := []struct {
		name     string
		page     int
		total    int
		limit    int
		expected bool
	}{
		{
			name:     "no results - no next page",
			page:     1,
			total:    0,
			limit:    20,
			expected: false,
		},
		{
			name:     "first page with more pages",
			page:     1,
			total:    50,
			limit:    20,
			expected: true,
		},
		{
			name:     "last page",
			page:     3,
			total:    50,
			limit:    20,
			expected: false,
		},
		{
			name:     "middle page",
			page:     2,
			total:    100,
			limit:    20,
			expected: true,
		},
		{
			name:     "single page - no next",
			page:     1,
			total:    15,
			limit:    20,
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ListResult{Page: tt.page, Total: tt.total, Limit: tt.limit}
			assert.Equal(t, tt.expected, result.HasNextPage())
		})
	}
}

func TestListResult_HasPrevPage(t *testing.T) {
	tests := []struct {
		name     string
		page     int
		expected bool
	}{
		{
			name:     "first page - no prev",
			page:     1,
			expected: false,
		},
		{
			name:     "second page - has prev",
			page:     2,
			expected: true,
		},
		{
			name:     "page 10 - has prev",
			page:     10,
			expected: true,
		},
		{
			name:     "page 0 (invalid) - no prev",
			page:     0,
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ListResult{Page: tt.page}
			assert.Equal(t, tt.expected, result.HasPrevPage())
		})
	}
}

func TestConstants(t *testing.T) {
	// Verify constants are set correctly
	assert.Equal(t, 20, DefaultLimit)
	assert.Equal(t, 100, MaxLimit)
}
