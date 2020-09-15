package utils

import (
	"encoding/json"
	"net/http"

	"github.com/sirupsen/logrus"
)

// EncodeResponse encodes response. It process error if needed
func EncodeResponse(w http.ResponseWriter, r *http.Request, log logrus.FieldLogger, resp interface{}) (ok bool) {
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		ProcessError(r.Context(), log, w, "couldn't encode response", http.StatusInternalServerError, err)
		return false
	}
	return true
}
