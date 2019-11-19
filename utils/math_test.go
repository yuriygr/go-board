package utils

import "testing"

func TestAbs(t *testing.T) {
	got := Abs(-1)
	if got != 1 {
		t.Errorf("Abs(-1) = %d; want 1", got)
	}
}

func TestLimitMaxValue(t *testing.T) {
	testCases := []struct {
		name string
		got  int64
		want int64
	}{
		{"LimitMaxValue(100, 95)", LimitMaxValue(100, 95), 95},
		{"LimitMaxValue(50, 95)", LimitMaxValue(50, 95), 50},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			if tc.got != tc.want {
				t.Errorf("got %d; want %d", tc.got, tc.want)
			}
		})
	}
}

func TestLimitMinValue(t *testing.T) {
	testCases := []struct {
		name string
		got  int64
		want int64
	}{
		{"LimitMinValue(4, 10)", LimitMinValue(4, 10), 10},
		{"LimitMinValue(15, 10)", LimitMinValue(15, 10), 15},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			if tc.got != tc.want {
				t.Errorf("got %d; want %d", tc.got, tc.want)
			}
		})
	}
}
