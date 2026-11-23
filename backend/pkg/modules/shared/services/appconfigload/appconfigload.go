// Package appconfigload reads typed app config values (codes in pkg/constants/appconfigcode).
package appconfigload

import (
	"context"
	"net/http"

	"backend/pkg/constants/appconfigcode/core"
	"backend/pkg/constants/errorcode"
	cerrors "backend/pkg/core/errors"
	"backend/pkg/repository"
)

// GetAppConfigByOption loads and decodes the active row of code. A missing or inactive row is
// APP_CONFIG_NOT_FOUND when required, else (nil, nil). Undecodable JSON is INTERNAL.
func GetAppConfigByOption[T any](ctx context.Context, repo repository.AppConfigRepository,
	code *core.AppConfigCode[T], required bool) (*T, error) {
	row, err := repo.FindByCode(ctx, code.Code())
	if err != nil {
		return nil, cerrors.Wrap(err, "find app config "+code.Code())
	}
	if row == nil || !row.Status {
		if required {
			e := cerrors.New(http.StatusServiceUnavailable, errorcode.AppConfigNotFound)
			e.Message += ": " + code.Code()
			return nil, e
		}
		return nil, nil
	}
	v, err := code.ToValue(row.Value)
	if err != nil {
		return nil, cerrors.Wrap(err, "decode app config "+code.Code())
	}
	return &v, nil
}

// GetAppConfig loads a required config: APP_CONFIG_NOT_FOUND when its row is missing or inactive.
func GetAppConfig[T any](ctx context.Context, repo repository.AppConfigRepository,
	code *core.AppConfigCode[T]) (*T, error) {
	return GetAppConfigByOption(ctx, repo, code, true)
}
