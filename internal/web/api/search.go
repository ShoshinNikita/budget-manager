package api

import (
	"context"
	"net/http"
	"strings"

	"github.com/sirupsen/logrus"

	"github.com/ShoshinNikita/budget-manager/internal/db"
	"github.com/ShoshinNikita/budget-manager/internal/pkg/money"
	reqid "github.com/ShoshinNikita/budget-manager/internal/pkg/request_id"
	"github.com/ShoshinNikita/budget-manager/internal/web/api/models"
	"github.com/ShoshinNikita/budget-manager/internal/web/utils"
)

type SearchHandlers struct {
	db  SearchDB
	log logrus.FieldLogger
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

	log = log.WithFields(logrus.Fields{
		"title": req.Title, "title_exactly": req.TitleExactly,
		"notes": req.Notes, "notes_exactly": req.NotesExactly,
		"after": req.After, "before": req.Before, "type_ids": req.TypeIDs,
		"min_cost": req.MinCost, "max_cost": req.MaxCost,
		"sort": req.Sort, "order": req.Order,
	})

	// Process
	log.Debug("search for Spends")
	args := db.SearchSpendsArgs{
		Title:        strings.ToLower(strings.TrimSpace(req.Title)),
		Notes:        strings.ToLower(strings.TrimSpace(req.Notes)),
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
		utils.ProcessInternalError(ctx, log, w, "couldn't search for Spends", err)
		return
	}
	log.WithField("spend_number", len(spends)).Debug("finish Spend search")

	// Encode
	resp := models.SearchSpendsResp{
		Response: models.Response{
			RequestID: reqid.FromContext(ctx).ToString(),
			Success:   true,
		},
		Spends: spends,
	}
	utils.EncodeResponse(w, r, log, resp)
}
