package api

import (
	"context"
	"net/http"
	"strings"

	"github.com/ShoshinNikita/budget-manager/internal/db"
	"github.com/ShoshinNikita/budget-manager/internal/logger"
	"github.com/ShoshinNikita/budget-manager/internal/pkg/money"
	"github.com/ShoshinNikita/budget-manager/internal/pkg/reqid"
	"github.com/ShoshinNikita/budget-manager/internal/web/api/models"
	"github.com/ShoshinNikita/budget-manager/internal/web/utils"
)

type SearchHandlers struct {
	db  SearchDB
	log logger.Logger
}

type SearchDB interface {
	SearchSpends(ctx context.Context, args db.SearchSpendsArgs) ([]db.Spend, error)
}

// @Summary Search Spends
// @Tags Search
// @Router /api/search/spends [get]
// @Param params query models.SearchSpendsReq true "Search args"
// @Produce json
// @Success 200 {object} models.SearchSpendsResp
// @Failure 400 {object} models.Response "Invalid request"
// @Failure 500 {object} models.Response "Internal error"
//
func (h SearchHandlers) SearchSpends(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	log := reqid.FromContextToLogger(ctx, h.log)

	// Decode
	req := &models.SearchSpendsReq{}
	if ok := utils.DecodeRequest(w, r, log, req); !ok {
		return
	}
	log = log.WithRequest(req)

	// Process
	args := db.SearchSpendsArgs{
		Title:        strings.ToLower(req.Title),
		Notes:        strings.ToLower(req.Notes),
		TitleExactly: req.TitleExactly,
		NotesExactly: req.NotesExactly,
		After:        req.After,
		Before:       req.Before,
		MinCost:      money.FromFloat(req.MinCost),
		MaxCost:      money.FromFloat(req.MaxCost),
		TypeIDs:      req.TypeIDs,
	}
	switch req.Sort {
	case "title":
		args.Sort = db.SortSpendsByTitle
	case "cost":
		args.Sort = db.SortSpendsByCost
	default:
		args.Sort = db.SortSpendsByDate
	}
	switch req.Order {
	case "desc":
		args.Order = db.OrderByDesc
	default:
		args.Order = db.OrderByAsc
	}

	spends, err := h.db.SearchSpends(ctx, args)
	if err != nil {
		utils.EncodeInternalError(ctx, w, log, "couldn't search for Spends", err)
		return
	}
	log.WithField("spend_number", len(spends)).Debug("finish Spend search")

	resp := &models.SearchSpendsResp{
		Spends: spends,
	}
	utils.Encode(ctx, w, log, utils.EncodeResponse(resp))
}
