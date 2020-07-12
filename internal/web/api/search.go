import (
	"github.com/sirupsen/logrus"
)

type SearchHandlers struct {
	db  SearchDB
	log logrus.FieldLogger
}

type SearchDB interface {
}
