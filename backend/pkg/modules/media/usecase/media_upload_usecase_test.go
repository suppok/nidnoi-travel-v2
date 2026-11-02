package usecase

import "testing"

func TestDownloadAndASCIIName(t *testing.T) {
	cases := []struct{ in, ext, download, ascii string }{
		{"Tokyo Osaka 5D3N.PDF", ".pdf", "Tokyo Osaka 5D3N.pdf", "Tokyo-Osaka-5D3N.pdf"},
		{`C:\fakepath\โปรแกรม ญี่ปุ่น.docx`, ".docx", "โปรแกรม ญี่ปุ่น.docx", "itinerary.docx"},
		{"", ".pdf", "itinerary.pdf", "itinerary.pdf"},
		{`a"b/../c.pdf`, ".pdf", "c.pdf", "c.pdf"},
	}
	for _, c := range cases {
		got := DownloadName(c.in, c.ext)
		if got != c.download {
			t.Errorf("DownloadName(%q) = %q, want %q", c.in, got, c.download)
		}
		if a := ASCIIName(got, c.ext); a != c.ascii {
			t.Errorf("ASCIIName(%q) = %q, want %q", got, a, c.ascii)
		}
	}
}
