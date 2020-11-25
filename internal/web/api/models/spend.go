package models

type AddSpendReq struct {
	Request

	DayID uint `json:"day_id" validate:"required" example:"1"`

	Title  string  `json:"title" validate:"required" example:"Food"`
	TypeID uint    `json:"type_id"`
	Notes  string  `json:"notes"`
	Cost   float64 `json:"cost" validate:"required" example:"30"`
}

func (req AddSpendReq) Check() error {
	if req.DayID == 0 {
		return emptyOrZeroFieldError("day_id")
	}
	if req.Title == "" {
		return emptyFieldError("title")
	}
	// Skip Type
	// Skip Notes
	if req.Cost < 0 {
		return negativeFieldError("cost")
	}
	return nil
}

type AddSpendResp struct {
	Response

	ID uint `json:"id"`
}

type EditSpendReq struct {
	Request

	ID     uint     `json:"id" validate:"required" example:"1"`
	Title  *string  `json:"title"`
	TypeID *uint    `json:"type_id"`
	Notes  *string  `json:"notes"`
	Cost   *float64 `json:"cost"`
}

func (req EditSpendReq) Check() error {
	if req.ID == 0 {
		return emptyOrZeroFieldError("id")
	}
	if req.Title != nil && *req.Title == "" {
		return emptyFieldError("title")
	}
	// Skip Type
	// Skip Notes
	if req.Cost != nil && *req.Cost < 0 {
		return negativeFieldError("cost")
	}
	return nil
}

type RemoveSpendReq struct {
	Request

	ID uint `json:"id" validate:"required" example:"1"`
}

func (req RemoveSpendReq) Check() error {
	if req.ID == 0 {
		return emptyOrZeroFieldError("id")
	}
	return nil
}
