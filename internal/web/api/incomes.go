package api

import "github.com/sirupsen/logrus"

type IncomesHandlers struct {
	db  IncomesDB
	log logrus.FieldLogger
}

type IncomesDB interface {
}
