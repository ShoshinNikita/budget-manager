package models

import (
	"time"

	"github.com/ShoshinNikita/budget-manager/internal/db"
)

// SearchSpendsReq is used to search for spends, all fields are optional
//
// nolint:maligned
type SearchSpendsReq struct {
	Request

	// Title can be in any case. Search will be performed by lowercased value
	Title string `json:"title,omitempty"`
	// Notes can be in any case. Search will be performed by lowercased value
	Notes string `json:"notes,omitempty"`

	// TitleExactly defines should we search exactly for the given title
	TitleExactly bool `json:"title_exactly,omitempty"`
	// NotesExactly defines should we search exactly for the given notes
	NotesExactly bool `json:"notes_exactly,omitempty"`

	// After must be in the RFC3339 format (https://tools.ietf.org/html/rfc3339#section-5.8)
	After time.Time `json:"after,omitempty" format:"date"`
	// Before must be in the RFC3339 format (https://tools.ietf.org/html/rfc3339#section-5.8)
	Before time.Time `json:"before,omitempty" format:"date"`

	MinCost float64 `json:"min_cost,omitempty"`
	MaxCost float64 `json:"max_cost,omitempty"`

	// WithoutType is used to search for Spends without Spend Type. TypeIDs are ignored when it is true
	WithoutType bool `json:"without_type,omitempty"`
	// TypeIDs is a list of Spend Type ids to search for
	TypeIDs []uint `json:"type_ids,omitempty"`

	// Sort specify field to sort by. Available options: title, cost, date (default)
	Sort string `json:"sort,omitempty"`
	// Order specify sort order. Available options: asc (default), desc
	Order string `json:"order,omitempty"`
}

type SearchSpendsResp struct {
	Response

	Spends []*db.Spend `json:"spends"`
}
