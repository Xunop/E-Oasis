package v1

import (
	"encoding/json"
	"net/http"

	"github.com/Xunop/e-oasis/http/request"
	"github.com/Xunop/e-oasis/http/response"
	"github.com/Xunop/e-oasis/log"
	"github.com/Xunop/e-oasis/validator"
	"github.com/Xunop/e-oasis/model"
	"go.uber.org/zap"
)

func (h *Handler) SetGeneralSettings(w http.ResponseWriter, r *http.Request) {
	if request.GetUserRole(r) != model.RoleHost && request.GetUserRole(r) != model.RoleAdmin {
		log.Error("Unauthorized request by", zap.String("role", request.GetUserRole(r).String()),
			zap.String("username", request.GetUsername(r)))
		response.Unauthorized(w, r)
		return
	}

	var settings model.SystemSettingGeneral
	if err := json.NewDecoder(r.Body).Decode(&settings); err != nil {
		log.Error("Failed to decode request body", zap.Error(err))
		response.BadRequest(w, r, err)
		return
	}

	if err := validator.ValidateGeneralSettings(&settings); err != nil {
		log.Error("Failed to validate general settings", zap.Error(err))
		response.BadRequest(w, r, err)
		return
	}

	newSettings, err := h.store.UpsetGeneralSettings(&settings);
	if err != nil {
		log.Error("Failed to set general settings", zap.Error(err))
		response.ServerError(w, r, err)
		return
	}

	response.OK(w, r, newSettings)
}
