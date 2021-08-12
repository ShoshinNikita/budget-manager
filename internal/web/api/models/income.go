package models

type AddIncomeReq struct {
	BaseRequest

	MonthID uint    `json:"month_id" validate:"required" example:"1"`
	Title   string  `json:"title" validate:"required" example:"Salary"`
	Notes   string  `json:"notes"`
	Income  float64 `json:"income" validate:"required" example:"10000"`
}

func (req *AddIncomeReq) SanitizeAndCheck() error {
	sanitizeString(&req.Title)
	sanitizeString(&req.Notes)

	if req.MonthID == 0 {
		return emptyOrZeroFieldError("month_id")
	}
	if req.Title == "" {
		return emptyFieldError("title")
	}
	// Skip Notes
	if req.Income <= 0 {
		return notPositiveFieldError("income")
	}
	return nil
}

type AddIncomeResp struct {
	BaseResponse

	ID uint `json:"id"`
}

type EditIncomeReq struct {
	BaseRequest

	ID     uint     `json:"id" validate:"required" example:"1"`
	Title  *string  `json:"title"`
	Notes  *string  `json:"notes"`
	Income *float64 `json:"income"`
}

func (req *EditIncomeReq) SanitizeAndCheck() error {
	sanitizeString(req.Title)
	sanitizeString(req.Notes)

	if req.ID == 0 {
		return emptyOrZeroFieldError("id")
	}
	if req.Title != nil && *req.Title == "" {
		return emptyFieldError("title")
	}
	// Skip Notes
	if req.Income != nil && *req.Income <= 0 {
		return notPositiveFieldError("income")
	}
	return nil
}

type RemoveIncomeReq struct {
	BaseRequest

	ID uint `json:"id" validate:"required" example:"1"`
}

func (req *RemoveIncomeReq) SanitizeAndCheck() error {
	if req.ID == 0 {
		return emptyOrZeroFieldError("id")
	}
	return nil
}
