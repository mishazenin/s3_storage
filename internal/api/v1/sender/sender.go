package sender

import (
	"net"
	"net/http"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	srv Sender
}

func AddHandlers(
	g *gin.RouterGroup,
	srv Sender,
) {
	h := &Handler{srv: srv}
	g.GET("/add_servers", h.addServers)
}

type query struct {
	IP net.IP
}

func (h *Handler) addServers(c *gin.Context) {
	var q query
	if err := c.ShouldBindQuery(&q); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": http.StatusText(http.StatusBadRequest)})
		return
	}
	if err := h.srv.AddNewServer(c, q.IP); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": http.StatusText(http.StatusInternalServerError)})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": "successfully added"})

}
