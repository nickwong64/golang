// routes.go

package main

import (
	"github.com/gin-gonic/gin"
)

func initializeRoutes() {

	// Handle the index routes
	router.GET("/", showIndexPage)

}

func showIndexPage(c *gin.Context) {

	servers := read_servers()

	render(c, gin.H{
		"title":   "LIS Test",
		"servers": servers}, "index.html")

}
