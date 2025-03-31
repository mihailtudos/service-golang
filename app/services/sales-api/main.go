package main

import (
	"fmt"
	"os"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var (
	build   = "develop"
	service = "SALES-API"
)

func main() {
	log, err := initLogger(service)
	if err != nil {
		fmt.Println("error constructing the logger: ", err)
		os.Exit(1)
	}

	defer log.Sync()

	if err := run(log); err != nil {
		log.Errorw("startup", "ERROR", err)
	}
}

func run(*zap.SugaredLogger) error {

	return nil
}

func initLogger(service string) (*zap.SugaredLogger, error) {
	config := zap.NewProductionConfig()
	config.OutputPaths = []string{"stdout"}
	config.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	config.DisableStacktrace = true
	config.InitialFields = map[string]interface{}{
		"service": service,
	}

	log, err := config.Build()
	if err != nil {
		return nil, err
	}

	return log.Sugar(), nil
}
