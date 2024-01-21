package v1

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"karma_s3/internal/api/v1/downloading"
	"karma_s3/internal/api/v1/sender"
	"karma_s3/internal/api/v1/uploading"
)

func NewRouter(
	engine *gin.Engine,
	uploadAPI upload.Uploader,
	downloader downloading.Downlaoder,
	s sender.Sender,
) http.Handler {
	g := engine.Group("/api/v1")
	upload.AddHandlers(g, uploadAPI)
	sender.AddHandlers(g, s)
	downloading.AddHandlers(g, downloader)
	return engine
}
