package main

import (
	"github.com/osamaesmail/go-post-api/cmd"
	"github.com/osamaesmail/go-post-api/internal/logger"
)

func main() {
	err := cmd.ExecuteServer()
	if err != nil {
		logger.Log().Fatal().Err(err).Msg("failed to run server")
	}
}
