package service

import "testing"

func TestXPMultiplier(t *testing.T) {
	cases := []struct {
		streak uint
		want   float64
	}{
		{1, 1},
		{2, 1},
		{3, 1.5},
		{7, 2},
		{14, 2.5},
		{30, 3},
	}

	for _, tc := range cases {
		if got := xpMultiplier(tc.streak); got != tc.want {
			t.Fatalf("xpMultiplier(%d) = %v, want %v", tc.streak, got, tc.want)
		}
	}
}

func TestCalculateLevel(t *testing.T) {
	cases := []struct {
		xp   uint
		want uint
	}{
		{0, 1},
		{99, 1},
		{100, 2},
		{399, 2},
		{400, 3},
		{900, 4},
	}

	for _, tc := range cases {
		if got := calculateLevel(tc.xp); got != tc.want {
			t.Fatalf("calculateLevel(%d) = %d, want %d", tc.xp, got, tc.want)
		}
	}
}

func TestClampCredit(t *testing.T) {
	if got := clampCredit(230); got != 200 {
		t.Fatalf("clampCredit upper bound = %d, want 200", got)
	}
	if got := clampCredit(-5); got != 0 {
		t.Fatalf("clampCredit lower bound = %d, want 0", got)
	}
	if got := clampCredit(120); got != 120 {
		t.Fatalf("clampCredit normal value = %d, want 120", got)
	}
}

func TestProgressPercent(t *testing.T) {
	if got := progressPercent(3, 10); got != 30 {
		t.Fatalf("progressPercent(3, 10) = %v, want 30", got)
	}
	if got := progressPercent(12, 10); got != 100 {
		t.Fatalf("progressPercent caps at 100, got %v", got)
	}
	if got := progressPercent(1, 0); got != 0 {
		t.Fatalf("progressPercent with zero target = %v, want 0", got)
	}
}

