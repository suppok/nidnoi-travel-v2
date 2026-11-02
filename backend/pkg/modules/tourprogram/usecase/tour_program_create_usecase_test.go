package usecase

import "testing"

func TestFormatTourCode(t *testing.T) {
	cases := map[int64]string{1: "NIDN000001", 42: "NIDN000042", 261690: "NIDN261690", 999999: "NIDN999999"}
	for n, want := range cases {
		if got := FormatTourCode("NIDN", n); got != want {
			t.Errorf("FormatTourCode(%d) = %q, want %q", n, got, want)
		}
	}
}
