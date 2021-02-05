package middleware

import (
	"github.com/iGeeky/open-account/configs"
	"github.com/iGeeky/open-account/pkg/baselib/errors"
	"github.com/iGeeky/open-account/pkg/baselib/ginplus"
	"github.com/iGeeky/open-account/pkg/baselib/log"

	"github.com/gin-gonic/gin"
)

// PanicWrapper panic wrapper to ErrServerError
func PanicHandler(c *gin.Context) {
	ctx := ginplus.NewContetPlus(c)
	url := ctx.GetURI()

	defer func() {
		if err := recover(); err != nil {
			log.Errorf("####### request [%s %s] failed! err: %+v", c.Request.Method, url, err)
			// stacktrace := string(debug.Stack())
			// log.Errorf("####### request [%s] failed! err: %v", url, stacktrace)
			apiError, ok := err.(*errors.ApiError)
			if ok {
				if apiError.Errmsg != "" {
					ctx.JsonFailWithMsg(apiError.Reason, apiError.Errmsg)
				} else {
					ctx.JsonFail(apiError.Reason)
				}
			} else {
				if configs.Config.Debug {
					ctx.JsonFailWithMsg(errors.ErrServerError, err)
				} else {
					ctx.JsonFail(errors.ErrServerError)
				}
			}
		}
	}()

	c.Next()
}
