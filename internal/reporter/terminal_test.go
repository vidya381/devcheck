package reporter

import "testing"

func TestSummaryLine(t *testing.T) {
	tests := []struct {
		passed, warned, failed int
		want                   string
	}{
		{4, 0, 0, "  4 passed  0 warnings  0 failed"},
		{3, 1, 1, "  3 passed  1 warning  1 failed"},
		{0, 2, 0, "  0 passed  2 warnings  0 failed"},
	}
	for _, tt := range tests {
		if got := summaryLine(tt.passed, tt.warned, tt.failed); got != tt.want {
			t.Errorf("summaryLine(%d,%d,%d) = %q, want %q", tt.passed, tt.warned, tt.failed, got, tt.want)
		}
	}
}
