package web

import (
	"fmt"
	"net/http"
	"time"

	"github.com/sirupsen/logrus"

	"github.com/ShoshinNikita/budget-manager/internal/pkg/request_id"
)

// GET / - redirects to the current month page
//
func (s Server) indexHandler(w http.ResponseWriter, r *http.Request) {
	log := request_id.FromContextToLogger(r.Context(), s.log)

	year, month, _ := time.Now().Date()
	log = log.WithFields(logrus.Fields{"year": year, "month": int(month)})

	log.Debug("redirect to the current month")
	url := fmt.Sprintf("/overview/%d/%d", year, month)
	http.Redirect(w, r, url, http.StatusSeeOther)
}
