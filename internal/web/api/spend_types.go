package api

import "github.com/sirupsen/logrus"

type SpendTypesHandlers struct {
	db  SpendTypesDB
	log logrus.FieldLogger
}

type SpendTypesDB interface {
}
