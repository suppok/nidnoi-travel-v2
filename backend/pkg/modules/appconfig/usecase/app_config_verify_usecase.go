package usecase

import (
	"context"
	"net/http"
	"slices"
	"strings"

	"backend/pkg/constants/appconfigcode"
	"backend/pkg/constants/errorcode"
	"backend/pkg/constants/role"
	cerrors "backend/pkg/core/errors"
	"backend/pkg/modules/appconfig/models"
)

const (
	verifyNotFound = "NOT_FOUND"
	verifyInvalid  = "INVALID"
)

// VerifyCheckPermissionData limits verify to the super admin; Casbin can't, because admins hold "*".
func (u *appConfigUseCase) VerifyCheckPermissionData(request *models.AppConfigVerifyRequest) error {
	if request.UserRequestInfo == nil || !slices.Contains(request.UserRequestInfo.Roles, role.SuperAdmin) {
		return cerrors.New(http.StatusForbidden, errorcode.SuperAdminRequired)
	}
	return nil
}

// Verify checks every registered code: a required code without an active row is NOT_FOUND, and
// any active row whose value doesn't decode or fails its rules is INVALID.
func (u *appConfigUseCase) Verify(ctx context.Context, request *models.AppConfigVerifyRequest) (*models.AppConfigVerifyResponse, error) {
	if err := u.VerifyCheckPermissionData(request); err != nil {
		return nil, err
	}

	rows, err := u.repositoryRegistry.AppConfigRepository.List(ctx)
	if err != nil {
		return nil, cerrors.Wrap(err, "list app configs")
	}
	values := make(map[string]string, len(rows))
	for _, row := range rows {
		if row.Status {
			values[row.Code] = row.Value
		}
	}

	issues := models.AppConfigVerifyResponse{}
	for code, cfg := range appconfigcode.GetAppConfigMap() {
		value, ok := values[code]
		if !ok {
			if cfg.Required {
				issues = append(issues, models.AppConfigVerifyIssue{Code: code, Type: verifyNotFound, Message: "no active row"})
			}
			continue
		}
		if err := cfg.AppConfig.Validate(value); err != nil {
			issues = append(issues, models.AppConfigVerifyIssue{Code: code, Type: verifyInvalid, Message: err.Error()})
		}
	}
	slices.SortFunc(issues, func(a, b models.AppConfigVerifyIssue) int { return strings.Compare(a.Code, b.Code) })
	return &issues, nil
}
