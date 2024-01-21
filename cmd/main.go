package main

import (
	"github.com/gin-gonic/gin"
	"karma_s3/internal/config"
	redis2 "karma_s3/internal/repository/metadata"
	sender2 "karma_s3/internal/usecase/data_service"
	downloader2 "karma_s3/internal/usecase/downloader"
	metadataManager "karma_s3/internal/usecase/metadata_service"
	"karma_s3/internal/usecase/uploader"
	logger "karma_s3/pkg/logger"

	v1 "karma_s3/internal/api/v1"
)

func main() {
	logger := logger.NewLogger()
	redis, err := redis2.NewClient()
	if err != nil {
		panic(err)
	}
	metadataStore := metadataManager.NewMetadataManager(logger, redis)
	sender, err := sender2.NewSender(logger, config.SenderConfig{})
	uploaderSrv := uploader.NewService(logger, sender, metadataStore)
	downloader := downloader2.NewService(logger, sender, metadataStore)
	if err != nil {
		panic(err)
	}
	r := gin.Default()
	v1.NewRouter(r, uploaderSrv, downloader, sender)
	if err = r.Run(); err != nil {
		panic(err)
	}
}
