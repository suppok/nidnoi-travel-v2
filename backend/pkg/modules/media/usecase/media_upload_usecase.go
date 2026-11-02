package usecase

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"path"
	"regexp"
	"slices"
	"strings"

	"go.uber.org/zap"

	"backend/pkg/constants/errorcode"
	cerrors "backend/pkg/core/errors"
	"backend/pkg/core/validation"
	"backend/pkg/modules/media/commons"
	"backend/pkg/modules/media/models"
	"backend/pkg/modules/shared/client/objectstorage"
)

var unsafeNameChars = regexp.MustCompile(`[^A-Za-z0-9._-]+`)

func (u *mediaUseCase) UploadValidate(request *models.MediaUploadRequest) error {
	purposes := make([]any, 0, len(purposeRules))
	for p := range purposeRules {
		purposes = append(purposes, p)
	}
	errs := validation.Errors{}
	if err := validation.Validate(request.Purpose, validation.Required, validation.In(purposes...)); err != nil {
		errs["purpose"] = err
	}
	if len(request.Data) == 0 {
		errs["file"] = errors.New("file is required")
	}
	if len(errs) > 0 {
		return errs
	}
	return nil
}

// Upload checks the file by its content, stores it in R2 under a content-addressed key and returns
// its public URL. The file is stored as uploaded.
// TODO: re-encode images to WebP (strips EXIF) once an encoder is chosen — see backend/CLAUDE.md.
func (u *mediaUseCase) Upload(ctx context.Context, request *models.MediaUploadRequest) (*models.MediaUploadResponse, error) {
	if err := u.UploadValidate(request); err != nil {
		return nil, cerrors.Validation(err)
	}
	rule := purposeRules[request.Purpose]

	if len(request.Data) > u.coreRegistry.Config.Media.MaxUploadMB*1024*1024 {
		return nil, cerrors.New(http.StatusRequestEntityTooLarge, errorcode.MediaTooLarge)
	}

	mime := commons.DetectType(request.Data)
	if !slices.Contains(rule.mimes, mime) {
		return nil, cerrors.New(http.StatusBadRequest, errorcode.MediaTypeNotAllowed)
	}

	sum := sha256.Sum256(request.Data)
	hash := hex.EncodeToString(sum[:])[:24]
	ext := commons.Extension(mime)
	response := &models.MediaUploadResponse{Size: len(request.Data), Mime: mime}
	put := objectstorage.PutInput{Body: request.Data, ContentType: mime}

	if rule.isImage() {
		w, h, ok := commons.ImageSize(mime, request.Data)
		if !ok {
			return nil, cerrors.New(http.StatusBadRequest, errorcode.MediaTypeNotAllowed)
		}
		if w*h > u.coreRegistry.Config.Media.MaxMegapixels*1_000_000 {
			return nil, cerrors.New(http.StatusRequestEntityTooLarge, errorcode.MediaTooLarge)
		}
		if w < rule.minWidth || h < rule.minHeight {
			e := cerrors.New(http.StatusBadRequest, errorcode.ImageTooSmall)
			e.Message = fmt.Sprintf("The image must be at least %d×%d pixels (got %d×%d)", rule.minWidth, rule.minHeight, w, h)
			return nil, e
		}
		response.Width, response.Height = w, h
		put.Key = rule.keyPrefix + hash + ext
	} else {
		name := DownloadName(request.FileName, ext)
		safe := ASCIIName(name, ext)
		response.FileName = name
		// The hash folder keeps keys unique while the last segment stays a readable name.
		put.Key = rule.keyPrefix + hash + "/" + safe
		disposition := "attachment"
		if mime == commons.MimePDF {
			disposition = "inline" // opens in the browser's PDF viewer
		}
		put.ContentDisposition = fmt.Sprintf(`%s; filename="%s"; filename*=UTF-8''%s`, disposition, safe, url.PathEscape(name))
	}

	if err := u.shareRegistry.ClientRegistry.ObjectStorage.Put(ctx, put); err != nil {
		if errors.Is(err, objectstorage.ErrNotConfigured) {
			u.coreRegistry.Logger.Warn("media upload rejected: R2 is not configured")
			return nil, cerrors.New(http.StatusServiceUnavailable, errorcode.StorageUnavailable)
		}
		return nil, cerrors.Wrap(err, "upload media")
	}

	u.coreRegistry.Logger.Info("media uploaded",
		zap.String("purpose", request.Purpose), zap.String("key", put.Key), zap.Int("size", len(request.Data)),
		zap.String("by", request.UserRequestInfo.UserID))

	response.Key = put.Key
	response.URL = u.shareRegistry.ClientRegistry.ObjectStorage.PublicURL(put.Key)
	return response, nil
}

// DownloadName keeps the client's base name (Thai is fine) with the extension of the detected type.
func DownloadName(clientName, ext string) string {
	base := strings.TrimSpace(path.Base(strings.ReplaceAll(clientName, `\`, "/")))
	base = strings.TrimSuffix(base, path.Ext(base))
	base = strings.Map(func(r rune) rune {
		if r < 0x20 || r == '"' || r == '/' || r == 0x7f {
			return -1
		}
		return r
	}, base)
	if runes := []rune(base); len(runes) > 120 {
		base = string(runes[:120])
	}
	base = strings.TrimSpace(base)
	if base == "" || base == "." {
		base = "itinerary"
	}
	return base + ext
}

// ASCIIName is the storage-key and legacy filename form: ASCII letters, digits, ".", "_" and "-".
func ASCIIName(name, ext string) string {
	base := strings.Trim(unsafeNameChars.ReplaceAllString(strings.TrimSuffix(name, ext), "-"), "-.")
	if base == "" {
		base = "itinerary"
	}
	return base + ext
}
