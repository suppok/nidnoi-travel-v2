package emailoutbox

import (
	"context"
	"sync"

	"backend/pkg/modules/shared/client/mailer"
	"backend/pkg/modules/shared/client/objectstorage"
	"backend/pkg/modules/shared/services/notification"
)

// imageCacheSize bounds the in-memory image cache. Only a few cards are ever in use (the current one
// and, briefly, the one before it); each is a few hundred KB.
const imageCacheSize = 4

// inlineImages turns the cid: references of an email into attachments: the built-in card from the
// binary, uploaded cards from object storage. Uploaded keys are content-addressed and immutable, so a
// cached copy never goes stale and needs no expiry: the sender downloads each card once, not once per
// email.
type inlineImages struct {
	storage objectstorage.ObjectStorage

	mu    sync.Mutex
	cache map[string][]byte // object key → bytes
	order []string          // keys, oldest first
}

func newInlineImages(storage objectstorage.ObjectStorage) *inlineImages {
	return &inlineImages{storage: storage, cache: map[string][]byte{}}
}

// attachments returns one attachment per known content id in html. Unknown ids are skipped (the image
// shows as missing). A card that can't be loaded is a retryable error: the email is not sent without it.
func (c *inlineImages) attachments(ctx context.Context, html string) ([]mailer.Attachment, error) {
	var out []mailer.Attachment
	for _, id := range notification.ContentIDs(html) {
		var content []byte
		switch {
		case id == notification.DefaultContactCardID:
			content = notification.DefaultContactCard()
		case notification.UploadedCardKey(id) != "":
			data, err := c.get(ctx, notification.UploadedCardKey(id))
			if err != nil {
				return nil, &mailer.SendError{Retryable: true, Message: "load inline image " + id}
			}
			content = data
		default:
			continue
		}
		out = append(out, mailer.Attachment{
			Filename:    id,
			ContentType: notification.ImageContentType(id),
			ContentID:   id,
			Content:     content,
		})
	}
	return out, nil
}

func (c *inlineImages) get(ctx context.Context, key string) ([]byte, error) {
	c.mu.Lock()
	data, ok := c.cache[key]
	c.mu.Unlock()
	if ok {
		return data, nil
	}

	data, err := c.storage.Get(ctx, key)
	if err != nil {
		return nil, err
	}

	c.mu.Lock()
	defer c.mu.Unlock()
	if _, ok := c.cache[key]; !ok {
		if len(c.order) >= imageCacheSize {
			delete(c.cache, c.order[0])
			c.order = c.order[1:]
		}
		c.cache[key] = data
		c.order = append(c.order, key)
	}
	return data, nil
}
