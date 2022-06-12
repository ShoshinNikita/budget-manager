package utils

import (
	"github.com/ShoshinNikita/budget-manager/v2/internal/logger"
)

func LogInternalError(log logger.Logger, respMsg string, internalErr error) {
	log.WithFields(logger.Fields{"msg": respMsg, "error": internalErr}).Error("request error")
}
