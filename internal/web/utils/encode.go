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
		ProcessInternalError(r.Context(), log, w, "couldn't encode response", err)
		return false
	}
	return true
}
