package models

type AddMonthlyPaymentReq struct {
	Request

	MonthID uint `json:"month_id" validate:"required" example:"1"`

	Title  string  `json:"title" validate:"required" example:"Rent"`
	TypeID uint    `json:"type_id"`
	Notes  string  `json:"notes"`
	Cost   float64 `json:"cost" validate:"required" example:"1500"`
}

func (req AddMonthlyPaymentReq) Check() error {
	if req.MonthID == 0 {
		return emptyOrZeroFieldError("month_id")
	}
	if req.Title == "" {
		return emptyFieldError("title")
	}
	// Skip Type
	// Skip Notes
	if req.Cost <= 0 {
		return notPositiveFieldError("cost")
	}
	return nil
}

type AddMonthlyPaymentResp struct {
	Response

	ID uint `json:"id"`
}

type EditMonthlyPaymentReq struct {
	Request

	ID     uint     `json:"id" validate:"required" example:"1"`
	Title  *string  `json:"title"`
	TypeID *uint    `json:"type_id"`
	Notes  *string  `json:"notes"`
	Cost   *float64 `json:"cost"`
}

func (req EditMonthlyPaymentReq) Check() error {
	if req.ID == 0 {
		return emptyOrZeroFieldError("id")
	}
	if req.Title != nil && *req.Title == "" {
		return emptyFieldError("title")
	}
	// Skip Type
	// Skip Notes
	if req.Cost != nil && *req.Cost <= 0 {
		return notPositiveFieldError("cost")
	}
	return nil
}

type RemoveMonthlyPaymentReq struct {
	Request

	ID uint `json:"id" validate:"required" example:"1"`
}

func (req RemoveMonthlyPaymentReq) Check() error {
	if req.ID == 0 {
		return emptyOrZeroFieldError("id")
	}
	return nil
}
