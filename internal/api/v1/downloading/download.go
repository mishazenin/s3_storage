package downloading

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"

	"karma_s3/internal/usecase/downloader"
)

type Handler struct {
	srv Downlaoder
}

func AddHandlers(
	g *gin.RouterGroup,
	srv Downlaoder,
) {
	h := &Handler{srv: srv}
	g.GET("/download_object", h.download)
}

type query struct {
	ObjName string
}

func (h *Handler) download(c *gin.Context) {
	var p query
	if err := c.ShouldBindQuery(&p); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": http.StatusText(http.StatusBadRequest)})
		return
	}
	obj, err := h.srv.DownloadObject(c, p.ObjName)
	if err != nil {
		if errors.Is(err, downloader.ErrObjectNotExist) {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": http.StatusText(http.StatusInternalServerError)})
		return
	}
	c.JSON(http.StatusOK, obj)

}
