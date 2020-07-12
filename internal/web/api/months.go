import (
	"github.com/sirupsen/logrus"
)

type MonthsHandlers struct {
	db  MonthsDB
	log logrus.FieldLogger
}

type MonthsDB interface {
}
