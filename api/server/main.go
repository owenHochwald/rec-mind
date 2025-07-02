package main

import (
	"github.com/gin-gonic/gin"
	"github.com/owenHochwald/rec-mind-api/config"
	"github.com/owenHochwald/rec-mind-api/controllers"
	"github.com/owenHochwald/rec-mind-api/database"
)

func main() {

	config.LoadEnv()
	db := database.InitDB()

	defer db.Close()

	r := gin.Default()

	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "OK"})
	})

	r.POST("/api/upload", controllers.UploadArticle(db))
	r.POST("/api/interact", controllers.HandleInteraction)
	r.GET("/api/recommend", controllers.GetRecommendations)

	r.Run(":8080")
}
