package routes

import (
	"GoProcesadorExcel/controllers"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
)

func SetupRouter(rdb *redis.Client) *gin.Engine {
	r := gin.Default()

	// CORS configuration
	config := cors.DefaultConfig()
	config.AllowAllOrigins = true
	config.AllowMethods = []string{"GET", "POST"}
	r.Use(cors.New(config))

	r.POST("/convert", func(c *gin.Context) {
		controllers.HandleExcelConversion(c, rdb)
	})

	status := r.Group("/status")
	{
		status.GET("/lotes", func(c *gin.Context) {
			controllers.HandleStatusConsulta(c, rdb)
		})

		status.GET("/iddte", func(c *gin.Context) {
			controllers.HandleStatusIddte(c, rdb)
		})

		status.GET("/iddte/:id", func(c *gin.Context) {
			controllers.HandleUniqueStatusIddte(c, rdb)
		})
	}

	return r
}
