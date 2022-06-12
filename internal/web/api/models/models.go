// Package models contains models of requests and responses
package models

// All requests implement the Request interface
type Request interface {
	request()
}

// BaseRequest is a base request model that implements Request interface.
// It must be nested into all requests
type BaseRequest struct{}

func (BaseRequest) request() {}

// All responses implement the Response interface
type Response interface {
	SetBaseResponse(BaseResponse)
}

// BaseResponse is a base response model that implements Response interface.
// It must be nested into all responses
type BaseResponse struct {
	RequestID string `json:"request_id"`
	Success   bool   `json:"success"`
	// Error is specified only when success if false
	Error string `json:"error,omitempty"`
}

func (r *BaseResponse) SetBaseResponse(newResp BaseResponse) {
	*r = newResp
}
