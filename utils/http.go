package utils

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/livekit/protocol/logger"
)

func LogHttpRequest(params any) {
	if logger.IsEnabledDebug() {
		logger.Debugw("httpReq", "method", ParentFuncName(), "params", params)
	}
}

func ReturnRsp(c *gin.Context, code int, obj any) {
	if logger.IsEnabledDebug() {
		logger.Debugw("httpRsp", "method", CallerName(1), "obj", obj)
	}

	c.JSON(code, obj)
}

func CreateGinHttp(skipLogPaths []string) *gin.Engine {
	gin.SetMode(gin.ReleaseMode)
	return defaultEngine(skipLogPaths)
}

func CreateDebugGinHttp(skipLogPaths []string) *gin.Engine {
	gin.SetMode(gin.DebugMode)
	return defaultEngine(skipLogPaths)
}

func EnableCORS(engine *gin.Engine) {
	engine.Use(func(c *gin.Context) {
		originSrc := c.Request.Header.Get("Origin")
		origin, err := safeCheck(originSrc)
		if err != nil {
			c.AbortWithStatus(http.StatusNonAuthoritativeInfo)
			return
		}
		if origin != "" {
			c.Writer.Header().Set("Access-Control-Allow-Origin", origin)
			c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS")
			c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Token")
			c.Writer.Header().Set("Access-Control-Expose-Headers", "Access-Control-Allow-Headers, Token")
			c.Writer.Header().Set("Access-Control-Max-Age", "172800")
			c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		}

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		c.Next()
	})
}

func safeCheck(input string) (string, error) {
	// valid, err := regexp.MatchString("[0-9A-Za-z*]+", input)
	// if !valid {
	// 	return input, err
	// }
	return input, nil
}

func defaultEngine(skipPaths []string) *gin.Engine {
	logger := gin.LoggerWithWriter(nil, skipPaths...)

	engine := gin.New()
	engine.Use(logger, gin.Recovery())
	return engine
}
