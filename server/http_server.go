package server

import (
	"fmt"

	"backend/common/httputil"
	"third/gin"
)

func StartHttpServer(listen string) {
	fmt.Println("StartServer")
	router := gin.New()
	router.Use(httputil.GinLogger())

	user_router := router.Group("/loggather")
	{
		user_router.POST("/report", ReportLogHandle)
	}

	router.Run(listen)
}
