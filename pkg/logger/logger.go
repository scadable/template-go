package logger

import (
	"go.uber.org/zap"
)

var log *zap.Logger

func Init() {
	var err error
	log, err = zap.NewProduction()
	if err != nil {
		panic("cannot initialize zap logger: " + err.Error())
	}
}

func Sync() {
	_ = log.Sync()
}

func getLogger() *zap.Logger {
	return log
}
