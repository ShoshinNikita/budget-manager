package api

import "github.com/sirupsen/logrus"

type SpendsHandlers struct {
	db  SpendsDB
	log logrus.FieldLogger
}

type SpendsDB interface {
}
