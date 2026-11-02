package commons

import (
	"archive/zip"
	"bytes"
	"image"
	"image/jpeg"
	"image/png"
	"testing"
)

func zipOf(t *testing.T, files map[string]string) []byte {
	t.Helper()
	var buf bytes.Buffer
	zw := zip.NewWriter(&buf)
	for name, body := range files {
		w, err := zw.Create(name)
		if err != nil {
			t.Fatal(err)
		}
		_, _ = w.Write([]byte(body))
	}
	if err := zw.Close(); err != nil {
		t.Fatal(err)
	}
	return buf.Bytes()
}

func TestDetectTypeAndSize(t *testing.T) {
	img := image.NewRGBA(image.Rect(0, 0, 1300, 700))

	var jpg bytes.Buffer
	if err := jpeg.Encode(&jpg, img, nil); err != nil {
		t.Fatal(err)
	}
	var pngBuf bytes.Buffer
	if err := png.Encode(&pngBuf, img); err != nil {
		t.Fatal(err)
	}

	for name, data := range map[string][]byte{MimeJPEG: jpg.Bytes(), MimePNG: pngBuf.Bytes()} {
		if got := DetectType(data); got != name {
			t.Fatalf("DetectType = %q, want %q", got, name)
		}
		w, h, ok := ImageSize(name, data)
		if !ok || w != 1300 || h != 700 {
			t.Fatalf("%s ImageSize = %d×%d %v", name, w, h, ok)
		}
	}
}

func TestWebPSize(t *testing.T) {
	// VP8X header with a 2000×1000 canvas.
	b := make([]byte, 30)
	copy(b, "RIFF")
	copy(b[8:], "WEBPVP8X")
	w, h := 2000-1, 1000-1
	b[24], b[25], b[26] = byte(w), byte(w>>8), byte(w>>16)
	b[27], b[28], b[29] = byte(h), byte(h>>8), byte(h>>16)
	if gw, gh, ok := webpSize(b); !ok || gw != 2000 || gh != 1000 {
		t.Fatalf("VP8X = %d×%d %v", gw, gh, ok)
	}

	// VP8L header with 640×480.
	l := make([]byte, 30)
	copy(l, "RIFF")
	copy(l[8:], "WEBPVP8L")
	l[20] = 0x2f
	bits := uint32(640-1) | uint32(480-1)<<14
	l[21], l[22], l[23], l[24] = byte(bits), byte(bits>>8), byte(bits>>16), byte(bits>>24)
	if gw, gh, ok := webpSize(l); !ok || gw != 640 || gh != 480 {
		t.Fatalf("VP8L = %d×%d %v", gw, gh, ok)
	}
}

func TestDetectDOCX(t *testing.T) {
	docx := zipOf(t, map[string]string{
		"[Content_Types].xml": `<Types><Override ContentType="application/vnd.openxmlformats-officedocument.wordprocessingml.document.main+xml"/></Types>`,
		"word/document.xml":   "<w:document/>",
	})
	if got := DetectType(docx); got != MimeDOCX {
		t.Fatalf("docx detected as %q", got)
	}

	docm := zipOf(t, map[string]string{
		"[Content_Types].xml": `<Types><Override ContentType="application/vnd.ms-word.document.macroEnabled.main+xml"/></Types>`,
		"word/document.xml":   "<w:document/>",
	})
	if got := DetectType(docm); got != "" {
		t.Fatalf("docm detected as %q, want rejected", got)
	}

	if got := DetectType(zipOf(t, map[string]string{"a.txt": "x"})); got != "" {
		t.Fatalf("plain zip detected as %q", got)
	}
}

func TestDetectPDFAndUnknown(t *testing.T) {
	if got := DetectType([]byte("%PDF-1.7\n...")); got != MimePDF {
		t.Fatalf("pdf detected as %q", got)
	}
	if got := DetectType([]byte("<svg xmlns='http://www.w3.org/2000/svg'/>")); got != "" {
		t.Fatalf("svg detected as %q", got)
	}
}
