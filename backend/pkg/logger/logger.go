package logger

import (
	"fmt"

	"go.uber.org/zap"
)

var log *zap.SugaredLogger

func Init() {
	logger, _ := zap.NewProduction()
	log = logger.Sugar()
}

func Get() *zap.SugaredLogger {
	if log == nil {
		Init()
	}
	return log
}



func Sync() {
	if log != nil {
		err := log.Sync()
		if err != nil {
			fmt.Println("Error syncing logger:", err)
		}
	}
}