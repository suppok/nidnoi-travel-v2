// Package commons inspects uploaded files by their content (never by the client's file name or
// Content-Type header).
package commons

import (
	"archive/zip"
	"bytes"
	"encoding/binary"
	"image"
	_ "image/jpeg" // register decoders for image.DecodeConfig
	_ "image/png"
	"io"
	"net/http"
	"strings"
)

const (
	MimeJPEG = "image/jpeg"
	MimePNG  = "image/png"
	MimeWebP = "image/webp"
	MimePDF  = "application/pdf"
	MimeDOCX = "application/vnd.openxmlformats-officedocument.wordprocessingml.document"
)

// Extension returns the file extension used in storage keys for a detected type.
func Extension(mime string) string {
	switch mime {
	case MimeJPEG:
		return ".jpg"
	case MimePNG:
		return ".png"
	case MimeWebP:
		return ".webp"
	case MimePDF:
		return ".pdf"
	case MimeDOCX:
		return ".docx"
	}
	return ""
}

// DetectType returns one of the Mime* constants, or "" for anything else.
func DetectType(data []byte) string {
	switch sniffed := http.DetectContentType(data); sniffed {
	case MimeJPEG, MimePNG, MimeWebP, MimePDF:
		return sniffed
	case "application/zip":
		if isDOCX(data) {
			return MimeDOCX
		}
	}
	return ""
}

// isDOCX accepts a Word document (.docx) and rejects macro-enabled files (.docm) and other
// Office Open XML packages.
func isDOCX(data []byte) bool {
	zr, err := zip.NewReader(bytes.NewReader(data), int64(len(data)))
	if err != nil {
		return false
	}
	hasDocument := false
	for _, f := range zr.File {
		switch f.Name {
		case "word/document.xml":
			hasDocument = true
		case "[Content_Types].xml":
			if f.UncompressedSize64 > 1<<20 {
				return false
			}
			rc, err := f.Open()
			if err != nil {
				return false
			}
			types, err := io.ReadAll(io.LimitReader(rc, 1<<20))
			_ = rc.Close()
			if err != nil || !strings.Contains(string(types), "wordprocessingml.document.main+xml") {
				return false
			}
		}
	}
	return hasDocument
}

// ImageSize reads the pixel dimensions from the image header without decoding the pixels.
func ImageSize(mime string, data []byte) (width, height int, ok bool) {
	if mime == MimeWebP {
		return webpSize(data)
	}
	cfg, _, err := image.DecodeConfig(bytes.NewReader(data))
	if err != nil {
		return 0, 0, false
	}
	return cfg.Width, cfg.Height, true
}

// webpSize parses the RIFF header of a lossy (VP8), lossless (VP8L) or extended (VP8X) WebP file.
func webpSize(b []byte) (int, int, bool) {
	if len(b) < 30 || string(b[0:4]) != "RIFF" || string(b[8:12]) != "WEBP" {
		return 0, 0, false
	}
	switch string(b[12:16]) {
	case "VP8X":
		w := 1 + int(uint32(b[24])|uint32(b[25])<<8|uint32(b[26])<<16)
		h := 1 + int(uint32(b[27])|uint32(b[28])<<8|uint32(b[29])<<16)
		return w, h, true
	case "VP8L":
		if b[20] != 0x2f {
			return 0, 0, false
		}
		bits := binary.LittleEndian.Uint32(b[21:25])
		return 1 + int(bits&0x3fff), 1 + int((bits>>14)&0x3fff), true
	case "VP8 ":
		if b[23] != 0x9d || b[24] != 0x01 || b[25] != 0x2a {
			return 0, 0, false
		}
		w := int(binary.LittleEndian.Uint16(b[26:28]) & 0x3fff)
		h := int(binary.LittleEndian.Uint16(b[28:30]) & 0x3fff)
		return w, h, w > 0 && h > 0
	}
	return 0, 0, false
}
