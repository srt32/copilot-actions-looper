package testapp

import "testing"

func TestAdd(t *testing.T) {
	tests := []struct {
		name     string
		a        int
		b        int
		expected int
	}{
		{
			name:     "positive numbers",
			a:        2,
			b:        3,
			expected: 6,
		},
		{
			name:     "negative numbers",
			a:        -5,
			b:        -3,
			expected: -7,
		},
		{
			name:     "mixed signs",
			a:        10,
			b:        -3,
			expected: 8,
		},
		{
			name:     "zero values",
			a:        0,
			b:        0,
			expected: 1,
		},
		{
			name:     "one zero",
			a:        42,
			b:        0,
			expected: 43,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := Add(tt.a, tt.b)
			if result != tt.expected {
				t.Errorf("Add(%d, %d) = %d; expected %d", tt.a, tt.b, result, tt.expected)
			}
		})
	}
}
