package main

import (
	"alog/Config"
	"alog/Logger"
	"errors"
	"flag"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

type request struct {
	Message string `json:"message"`
}

func main() {
	mode := getApplicationMode()
	port := getListenPort()
	flag.Parse()

	gin.SetMode(*mode)
	engine := gin.Default()
	engine.Use(getCors())
	engine.POST("/info", infoHandler)
	engine.POST("/warning", warningHandler)
	engine.POST("/error", errorHandler)
	engine.NoRoute(noRoute)
	Logger.GetLogger().Error(http.ListenAndServe(fmt.Sprintf(":%s", *port), engine), false)
}

func errorHandler(c *gin.Context) {
	var req request
	if err := c.Bind(&req); err != nil {
		c.SecureJSON(getResponse400(err))
		return
	}
	Logger.GetLogger().Error(errors.New(req.Message), false)
	c.SecureJSON(getResponse201())
	return
}

func warningHandler(c *gin.Context) {
	var req request
	if err := c.Bind(&req); err != nil {
		c.SecureJSON(getResponse400(err))
		return
	}
	Logger.GetLogger().Warning(req.Message)
	c.SecureJSON(getResponse201())
	return
}

func infoHandler(c *gin.Context) {
	var req request
	if err := c.Bind(&req); err != nil {
		c.SecureJSON(getResponse400(err))
		return
	}
	Logger.GetLogger().Info(req.Message)
	c.SecureJSON(getResponse201())
	return
}

func getCors() gin.HandlerFunc {
	return cors.New(cors.Config{
		AllowMethods: []string{
			http.MethodPost,
		},
		AllowHeaders: []string{
			"Origin",
			"Content-Length",
			"Content-Type",
			"Authorization",
		},
		AllowCredentials: false,
		AllowAllOrigins:  true,
		MaxAge:           12 * time.Hour,
	})
}

func noRoute(c *gin.Context) {
	c.AbortWithStatusJSON(getResponse404())
	return
}

func getResponse404() (int, gin.H) {
	return http.StatusNotFound, gin.H{
		"code":    http.StatusNotFound,
		"message": http.StatusText(http.StatusNotFound),
	}
}

func getResponse400(err error) (int, gin.H) {
	return http.StatusBadRequest, gin.H{
		"code":    http.StatusBadRequest,
		"message": http.StatusText(http.StatusBadRequest),
		"error":   err.Error(),
	}
}

func getResponse201() (int, gin.H) {
	return http.StatusCreated, gin.H{
		"code":    http.StatusCreated,
		"message": http.StatusText(http.StatusCreated),
	}
}

func getListenPort() *string {
	port := "80"
	if envPort := Config.GetEnvStr("PORT"); envPort != "" {
		port = envPort
	}
	return flag.String("port", port, "Example: -port=8080")
}

func getApplicationMode() *string {
	mode := gin.ReleaseMode
	if envMode := Config.GetEnvStr("APP_MODE"); envMode != "" {
		mode = envMode
	}
	return flag.String("mode", mode, "Example: -mode=debug")
}
