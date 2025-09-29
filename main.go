package main

import (
	"net/http"

	"lazy-lagoon/routes"

	"github.com/gin-gonic/gin"
)

/*
	Main function to start the server
*/
func main() {
	router := gin.Default()

	err := initSentry(router)
	if err != nil {
		panic(err)
	}

	router.POST("/lazy-lagoon/paginate", routes.Paginate)

	router.POST("/lazy-lagoon/transform", routes.Transform)

	router.GET("/lazy-lagoon/healthz/ready", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "ok"})
	})

	router.GET("/lazy-lagoon/healthz/live", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "ok"})
	})

	if err := router.Run(":4040"); err != nil {
		panic(err)
	}
}
