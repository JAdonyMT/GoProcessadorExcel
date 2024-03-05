package routes

import (
	"GoProcesadorExcel/controllers"

	"github.com/gin-gonic/gin"
)

func SetupRouter() *gin.Engine {
	r := gin.Default()

	r.POST("/convert", controllers.HandleExcelConversion)

	return r
}
