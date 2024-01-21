package upload

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	srv Uploader
}

func AddHandlers(
	g *gin.RouterGroup,
	srv Uploader,
) {
	h := &Handler{srv: srv}
	g.GET("/upload_object", h.upload)
}

type query struct {
	Obj []byte
}

func (h *Handler) upload(c *gin.Context) {
	var p query
	if err := c.ShouldBindQuery(&p); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": http.StatusText(http.StatusBadRequest)})
		return
	}
	obj := []byte("the data used as a test text for custom s3 storage")
	// TODO
	objName, err := h.srv.UploadObject(c, obj)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": http.StatusText(http.StatusInternalServerError)})
		return
	}
	c.JSON(http.StatusOK, gin.H{"object name": objName})

}
