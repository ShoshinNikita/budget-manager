package api

import (
	"context"
	"net/http"
	"strings"

	"github.com/sirupsen/logrus"

	"github.com/ShoshinNikita/budget-manager/internal/db"
	"github.com/ShoshinNikita/budget-manager/internal/pkg/request_id"
	"github.com/ShoshinNikita/budget-manager/internal/web/models"
	"github.com/ShoshinNikita/budget-manager/internal/web/utils"
)

type SpendTypesHandlers struct {
	db  SpendTypesDB
	log logrus.FieldLogger
}

type SpendTypesDB interface {
	GetSpendTypes(ctx context.Context) ([]*db.SpendType, error)
	AddSpendType(ctx context.Context, name string) (id uint, err error)
	EditSpendType(ctx context.Context, id uint, newName string) error
	RemoveSpendType(ctx context.Context, id uint) error
}

// @Summary Get All Spend Types
// @Tags Spend Types
// @Router /api/spend-types [get]
// @Produce json
// @Success 200 {object} models.GetSpendTypesResp
// @Failure 500 {object} models.Response "Internal error"
//
func (h SpendTypesHandlers) GetSpendTypes(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	log := request_id.FromContextToLogger(ctx, h.log)

	// Process
	log.Debug("return all Spend Types")
	types, err := h.db.GetSpendTypes(ctx)
	if err != nil {
		utils.ProcessError(ctx, log, w, "couldn't get Spend Types", http.StatusInternalServerError, err)
		return
	}

	// Encode
	resp := models.GetSpendTypesResp{
		Response: models.Response{
			RequestID: request_id.FromContext(ctx).ToString(),
			Success:   true,
		},
		SpendTypes: types,
	}
	utils.EncodeResponse(w, r, log, resp)
}

// @Summary Create Spend Type
// @Tags Spend Types
// @Router /api/spend-types [post]
// @Accept json
// @Param body body models.AddSpendTypeReq true "New Spend Type"
// @Produce json
// @Success 200 {object} models.AddSpendTypeResp
// @Failure 400 {object} models.Response "Invalid request"
// @Failure 500 {object} models.Response "Internal error"
//
func (h SpendTypesHandlers) AddSpendType(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	log := request_id.FromContextToLogger(ctx, h.log)

	// Decode
	req := &models.AddSpendTypeReq{}
	if ok := utils.DecodeRequest(w, r, log, req); !ok {
		return
	}

	log = log.WithField("name", req.Name)

	// Process
	log.Debug("add Spend Type")
	id, err := h.db.AddSpendType(ctx, strings.TrimSpace(req.Name))
	if err != nil {
		utils.ProcessError(ctx, log, w, "couldn't add Spend Type", http.StatusInternalServerError, err)
		return
	}
	log = log.WithField("id", id)
	log.Info("Spend Type was successfully added")

	// Encode
	resp := models.AddSpendTypeResp{
		Response: models.Response{
			RequestID: request_id.FromContext(ctx).ToString(),
			Success:   true,
		},
		ID: id,
	}
	utils.EncodeResponse(w, r, log, resp)
}

// @Summary Edit Spend Type
// @Tags Spend Types
// @Router /api/spend-types [put]
// @Accept json
// @Param body body models.EditSpendTypeReq true "Updated Spend Type"
// @Produce json
// @Success 200 {object} models.Response
// @Failure 400 {object} models.Response "Invalid request"
// @Failure 404 {object} models.Response "Spend Type doesn't exist"
// @Failure 500 {object} models.Response "Internal error"
//
func (h SpendTypesHandlers) EditSpendType(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	log := request_id.FromContextToLogger(ctx, h.log)

	// Decode
	req := &models.EditSpendTypeReq{}
	if ok := utils.DecodeRequest(w, r, log, req); !ok {
		return
	}

	log = log.WithFields(logrus.Fields{"id": req.ID, "name": req.Name})

	// Process
	log.Debug("edit Spend Type")
	err := h.db.EditSpendType(ctx, req.ID, strings.TrimSpace(req.Name))
	if err != nil {
		switch err {
		case db.ErrSpendTypeNotExist:
			utils.ProcessError(ctx, log, w, err.Error(), http.StatusNotFound, err)
		default:
			utils.ProcessError(ctx, log, w, "couldn't edit Spend Type", http.StatusInternalServerError, err)
		}
		return
	}
	log.Info("Spend Type was successfully edited")

	// Encode
	resp := models.Response{
		RequestID: request_id.FromContext(ctx).ToString(),
		Success:   true,
	}
	utils.EncodeResponse(w, r, log, resp)
}

// @Summary Remove Spend Type
// @Tags Spend Types
// @Router /api/spend-types [delete]
// @Accept json
// @Param body body models.RemoveSpendTypeReq true "Spend Type id"
// @Produce json
// @Success 200 {object} models.Response
// @Failure 400 {object} models.Response "Invalid request"
// @Failure 404 {object} models.Response "Spend Type doesn't exist"
// @Failure 500 {object} models.Response "Internal error"
//
func (h SpendTypesHandlers) RemoveSpendType(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	log := request_id.FromContextToLogger(ctx, h.log)

	// Decode
	req := &models.RemoveSpendTypeReq{}
	if ok := utils.DecodeRequest(w, r, log, req); !ok {
		return
	}

	log = log.WithField("id", req.ID)

	// Process
	log.Debug("remove Spend Type")
	err := h.db.RemoveSpendType(ctx, req.ID)
	if err != nil {
		switch err {
		case db.ErrSpendTypeNotExist:
			utils.ProcessError(ctx, log, w, err.Error(), http.StatusNotFound, err)
		default:
			utils.ProcessError(ctx, log, w, "couldn't remove Spend Type", http.StatusInternalServerError, err)
		}
		return
	}
	log.Info("Spend Type was successfully removed")

	// Encode
	resp := models.Response{
		RequestID: request_id.FromContext(ctx).ToString(),
		Success:   true,
	}
	utils.EncodeResponse(w, r, log, resp)
}
