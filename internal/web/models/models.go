// Package models contains models of requests and responses
package models

// -------------------------------------------------
// Common
// -------------------------------------------------

type Request struct {
}

type Response struct {
	Success bool   `json:"success"`
	Error   string `json:"error,omitempty"`
}
