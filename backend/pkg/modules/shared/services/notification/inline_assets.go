package notification

import (
	"path"
	"regexp"
	"slices"
	"strings"
)

// Inline images are attached to the email (with a content id) rather than linked, so mail clients
// show them without "load remote images". The HTML references them as <img src="cid:<id>">.

// DefaultContactCardID is the built-in "ติดต่อทีมงานบริษัทนิดหน่อยทราเวล" card (assets/contact-card.png),
// used until an admin uploads one.
const DefaultContactCardID = "contact-card.png"

// ContactCardKeyPrefix is where uploaded contact cards live in object storage. It must match the
// email-contact-card purpose in pkg/modules/media/usecase.
const ContactCardKeyPrefix = "email/contact-cards/"

// uploadedCardIDPrefix marks a content id that names an uploaded card: "contact-card-<file>" is the
// object ContactCardKeyPrefix + <file>. Uploaded keys are content-addressed and never reused, so a
// queued email keeps the card it was queued with, and the image can be cached forever.
const uploadedCardIDPrefix = "contact-card-"

var safeFile = regexp.MustCompile(`^[a-f0-9]{8,64}\.(png|jpg|jpeg)$`)

// ContactCardID returns the content id of the card at imageURL, a URL of an uploaded card under
// publicBaseURL. Anything else (empty, foreign URL) gives the built-in card and ok = false.
func ContactCardID(imageURL, publicBaseURL string) (id string, ok bool) {
	if imageURL == "" || publicBaseURL == "" {
		return DefaultContactCardID, false
	}
	file, found := strings.CutPrefix(imageURL, strings.TrimRight(publicBaseURL, "/")+"/"+ContactCardKeyPrefix)
	if !found || !safeFile.MatchString(file) {
		return DefaultContactCardID, false
	}
	return uploadedCardIDPrefix + file, true
}

// UploadedCardKey returns the object key of an uploaded card's content id ("" when id isn't one).
func UploadedCardKey(id string) string {
	file, found := strings.CutPrefix(id, uploadedCardIDPrefix)
	if !found || !safeFile.MatchString(file) {
		return ""
	}
	return ContactCardKeyPrefix + file
}

// DefaultContactCard is the built-in card image (PNG).
func DefaultContactCard() []byte {
	data, _ := assetFS.ReadFile("assets/contact-card.png")
	return data
}

// ImageContentType of a file name by its extension (PNG or JPEG).
func ImageContentType(name string) string {
	switch strings.ToLower(path.Ext(name)) {
	case ".jpg", ".jpeg":
		return "image/jpeg"
	default:
		return "image/png"
	}
}

var cidRef = regexp.MustCompile(`cid:([A-Za-z0-9._-]+)`)

// ContentIDs returns the content ids an HTML body references, each once, in order.
func ContentIDs(html string) []string {
	var ids []string
	for _, m := range cidRef.FindAllStringSubmatch(html, -1) {
		if !slices.Contains(ids, m[1]) {
			ids = append(ids, m[1])
		}
	}
	return ids
}
