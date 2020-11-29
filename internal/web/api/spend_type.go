package api

import (
	"context"
	"net/http"
	"strings"

	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"

	"github.com/ShoshinNikita/budget-manager/internal/db"
	reqid "github.com/ShoshinNikita/budget-manager/internal/pkg/request_id"
	"github.com/ShoshinNikita/budget-manager/internal/web/api/models"
	"github.com/ShoshinNikita/budget-manager/internal/web/utils"
)

type SpendTypesHandlers struct {
	db  SpendTypesDB
	log logrus.FieldLogger
}

type SpendTypesDB interface {
	GetSpendTypes(ctx context.Context) ([]db.SpendType, error)
	AddSpendType(ctx context.Context, args db.AddSpendTypeArgs) (id uint, err error)
	EditSpendType(ctx context.Context, args db.EditSpendTypeArgs) error
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
	log := reqid.FromContextToLogger(ctx, h.log)

	// Process
	log.Debug("return all Spend Types")
	types, err := h.db.GetSpendTypes(ctx)
	if err != nil {
		utils.ProcessInternalError(ctx, log, w, "couldn't get Spend Types", err)
		return
	}

	// Encode
	resp := models.GetSpendTypesResp{
		Response: models.Response{
			RequestID: reqid.FromContext(ctx).ToString(),
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
	log := reqid.FromContextToLogger(ctx, h.log)

	// Decode
	req := &models.AddSpendTypeReq{}
	if ok := utils.DecodeRequest(w, r, log, req); !ok {
		return
	}

	log = log.WithFields(logrus.Fields{"name": req.Name, "parent_id": req.ParentID})

	// Process
	log.Debug("add Spend Type")
	args := db.AddSpendTypeArgs{
		Name:     strings.TrimSpace(req.Name),
		ParentID: req.ParentID,
	}
	id, err := h.db.AddSpendType(ctx, args)
	if err != nil {
		utils.ProcessInternalError(ctx, log, w, "couldn't add Spend Type", err)
		return
	}
	log = log.WithField("id", id)
	log.Info("Spend Type was successfully added")

	// Encode
	resp := models.AddSpendTypeResp{
		Response: models.Response{
			RequestID: reqid.FromContext(ctx).ToString(),
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
	log := reqid.FromContextToLogger(ctx, h.log)

	// Decode
	req := &models.EditSpendTypeReq{}
	if ok := utils.DecodeRequest(w, r, log, req); !ok {
		return
	}

	log = log.WithFields(logrus.Fields{"id": req.ID, "name": req.Name})

	if req.ParentID != nil && *req.ParentID != 0 {
		hasCycle, err := h.checkSpendTypeForCycle(ctx, req.ID, *req.ParentID)
		if err != nil {
			utils.ProcessInternalError(ctx, log, w, "couldn't check Spend Type for a cycle", err)
			return
		}
		if hasCycle {
			utils.ProcessError(ctx, w, "Spend Type with new parent type will have a cycle", http.StatusBadRequest)
			return
		}
	}

	// Process
	log.Debug("edit Spend Type")

	args := db.EditSpendTypeArgs{
		ID:       req.ID,
		Name:     trimSpacePointer(req.Name),
		ParentID: req.ParentID,
	}
	err := h.db.EditSpendType(ctx, args)
	if err != nil {
		switch {
		case errors.Is(err, db.ErrSpendTypeNotExist):
			utils.ProcessError(ctx, w, err.Error(), http.StatusNotFound)
		default:
			utils.ProcessInternalError(ctx, log, w, "couldn't edit Spend Type", err)
		}
		return
	}
	log.Info("Spend Type was successfully edited")

	// Encode
	resp := models.Response{
		RequestID: reqid.FromContext(ctx).ToString(),
		Success:   true,
	}
	utils.EncodeResponse(w, r, log, resp)
}

func (h SpendTypesHandlers) checkSpendTypeForCycle(ctx context.Context, originalID, newParentID uint) (bool, error) {
	spendTypes, err := h.db.GetSpendTypes(ctx)
	if err != nil {
		return false, errors.Wrap(err, "couldn't get Spend Types")
	}

	return checkSpendTypeForCycle(spendTypes, originalID, newParentID)
}

func checkSpendTypeForCycle(spendTypesSlice []db.SpendType, originalID, newParentID uint) (hasCycle bool, _ error) {
	spendTypes := make(map[uint]db.SpendType, len(spendTypesSlice))
	for _, t := range spendTypesSlice {
		spendTypes[t.ID] = t
	}

	parentType := spendTypes[newParentID]
	// Max depth is 15
	for i := 0; i < 15; i++ {
		if parentType.ID == 0 {
			// Unexpected error
			return false, errors.New("invalid Spend Type")
		}
		if parentType.ID == originalID {
			// Has cycle
			return true, nil
		}
		if parentType.ParentID == 0 {
			// No more parents
			return false, nil
		}

		parentType = spendTypes[parentType.ParentID]
	}

	return false, errors.New("Spend Type has too many parents or already has a cycle")
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
	log := reqid.FromContextToLogger(ctx, h.log)

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
		switch {
		case errors.Is(err, db.ErrSpendTypeNotExist):
			utils.ProcessError(ctx, w, err.Error(), http.StatusNotFound)
		default:
			utils.ProcessInternalError(ctx, log, w, "couldn't remove Spend Type", err)
		}
		return
	}
	log.Info("Spend Type was successfully removed")

	// Encode
	resp := models.Response{
		RequestID: reqid.FromContext(ctx).ToString(),
		Success:   true,
	}
	utils.EncodeResponse(w, r, log, resp)
}
