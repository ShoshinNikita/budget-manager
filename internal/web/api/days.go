import (
	"github.com/sirupsen/logrus"
)

type DaysHandlers struct {
	db  DaysDB
	log logrus.FieldLogger
}

type DaysDB interface {
}
