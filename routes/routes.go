package routes

import (
	"GoProcesadorExcel/controllers"

	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
)

func SetupRouter(rdb *redis.Client) *gin.Engine {
	r := gin.Default()

	r.POST("/convert", func(c *gin.Context) {
		controllers.HandleExcelConversion(c, rdb)
	})

	r.GET("/lotesStatus", func(c *gin.Context) {
		controllers.HandleStatusConsulta(c, rdb)
	})

	return r
}
