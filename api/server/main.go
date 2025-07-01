package main

import (
	"github.com/gin-gonic/gin"
)

func uploadArticle(c *gin.Context)      {}
func handleInteraction(c *gin.Context)  {}
func getRecommendations(c *gin.Context) {}

func main() {
	r := gin.Default()
	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "OK"})
	})

	r.POST("/api/v1/upload", uploadArticle)
	r.POST("/api/v1/interact", handleInteraction)
	r.GET("/api/v1/recommend", getRecommendations)

	r.Run(":8080")
}
