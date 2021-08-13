package api

import (
	"context"
	"net/http"

	"github.com/pkg/errors"

	"github.com/ShoshinNikita/budget-manager/internal/db"
	"github.com/ShoshinNikita/budget-manager/internal/logger"
	"github.com/ShoshinNikita/budget-manager/internal/pkg/reqid"
	"github.com/ShoshinNikita/budget-manager/internal/web/api/models"
	"github.com/ShoshinNikita/budget-manager/internal/web/utils"
)

type SpendTypesHandlers struct {
	db  SpendTypesDB
	log logger.Logger
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
	types, err := h.db.GetSpendTypes(ctx)
	if err != nil {
		utils.EncodeInternalError(ctx, w, log, "couldn't get Spend Types", err)
		return
	}

	resp := &models.GetSpendTypesResp{
		SpendTypes: types,
	}
	utils.Encode(ctx, w, log, utils.EncodeResponse(resp))
}

// @Summary Create Spend Type
// @Tags Spend Types
// @Router /api/spend-types [post]
// @Accept json
// @Param body body models.AddSpendTypeReq true "New Spend Type"
// @Produce json
// @Success 201 {object} models.AddSpendTypeResp
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
	log = log.WithRequest(req)

	// Process
	args := db.AddSpendTypeArgs{
		Name:     req.Name,
		ParentID: req.ParentID,
	}
	id, err := h.db.AddSpendType(ctx, args)
	if err != nil {
		utils.EncodeInternalError(ctx, w, log, "couldn't add Spend Type", err)
		return
	}
	log = log.WithField("id", id)
	log.Debug("Spend Type was successfully added")

	resp := &models.AddSpendTypeResp{
		ID: id,
	}
	utils.Encode(ctx, w, log, utils.EncodeResponse(resp), utils.EncodeStatusCode(http.StatusCreated))
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
	log = log.WithRequest(req)

	if req.ParentID != nil && *req.ParentID != 0 {
		allSpendTypes, err := h.db.GetSpendTypes(ctx)
		if err != nil {
			utils.EncodeInternalError(ctx, w, log, "couldn't get all Spend Types to check for a cycle", err)
			return
		}

		hasCycle, err := checkSpendTypeForCycle(allSpendTypes, req.ID, *req.ParentID)
		if hasCycle {
			err = errors.New("Spend Type with new parent type will have a cycle")
		} else if err != nil {
			err = errors.Wrap(err, "check for a cycle failed")
		}
		if err != nil {
			utils.EncodeError(ctx, w, log, err, http.StatusBadRequest)
			return
		}
	}

	// Process
	args := db.EditSpendTypeArgs{
		ID:       req.ID,
		Name:     req.Name,
		ParentID: req.ParentID,
	}
	err := h.db.EditSpendType(ctx, args)
	if err != nil {
		switch {
		case errors.Is(err, db.ErrSpendTypeNotExist):
			utils.EncodeError(ctx, w, log, err, http.StatusNotFound)
		default:
			utils.EncodeInternalError(ctx, w, log, "couldn't edit Spend Type", err)
		}
		return
	}
	log.Debug("Spend Type was successfully edited")

	utils.Encode(ctx, w, log)
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
	log = log.WithRequest(req)

	// Process
	err := h.db.RemoveSpendType(ctx, req.ID)
	if err != nil {
		switch {
		case errors.Is(err, db.ErrSpendTypeNotExist):
			utils.EncodeError(ctx, w, log, err, http.StatusNotFound)
		default:
			utils.EncodeInternalError(ctx, w, log, "couldn't remove Spend Type", err)
		}
		return
	}
	log.Debug("Spend Type was successfully removed")

	utils.Encode(ctx, w, log)
}
