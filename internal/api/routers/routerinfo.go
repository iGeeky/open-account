package routers

import (
	"open-account/internal/api/service"

	"github.com/gin-gonic/gin"
)

const (
	HTTP_GET    = 1
	HTTP_POST   = 2
	HTTP_PUT    = 3
	HTTP_DELETE = 4
	HTTP_OPTION = 10

	TokenNone = service.TokenNone
	TokenUser = service.TokenUser
)

// RouterInfo Router信息.
type RouterInfo struct {
	Op         int
	URL        string
	CheckSign  bool
	CheckToken int
	Handler    gin.HandlerFunc
}
