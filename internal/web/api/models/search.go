package models

import (
	"time"

	"github.com/ShoshinNikita/budget-manager/internal/db"
)

// SearchSpendsReq is used to search for spends
//
//nolint:maligned
type SearchSpendsReq struct {
	Request

	// Title can be in any case. Search will be performed by lowercased value
	Title string `json:"title"`
	// Notes can be in any case. Search will be performed by lowercased value
	Notes string `json:"notes"`

	// TitleExactly defines should we search exactly for the given title
	TitleExactly bool `json:"title_exactly" default:"false"`
	// NotesExactly defines should we search exactly for the given notes
	NotesExactly bool `json:"notes_exactly" default:"false"`

	// After must be in the RFC3339 format (https://tools.ietf.org/html/rfc3339#section-5.8)
	After time.Time `json:"after" format:"date"`
	// Before must be in the RFC3339 format (https://tools.ietf.org/html/rfc3339#section-5.8)
	Before time.Time `json:"before" format:"date"`

	MinCost float64 `json:"min_cost"`
	MaxCost float64 `json:"max_cost"`

	// WithoutType is used to search for Spends without Spend Type. TypeIDs are ignored when it is true
	WithoutType bool `json:"without_type" default:"false"`
	// TypeIDs is a list of Spend Type ids to search for
	TypeIDs []uint `json:"type_ids"`

	// Sort specify field to sort by
	Sort string `json:"sort" enums:"title,cost,date" default:"date"`
	// Order specify sort order
	Order string `json:"order" enums:"asc,desc" default:"asc"`
}

type SearchSpendsResp struct {
	Response

	Spends []*db.Spend `json:"spends"`
}
