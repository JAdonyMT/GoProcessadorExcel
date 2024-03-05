package controllers

import (
	"context"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
)

func HandleStatusConsulta(c *gin.Context, rdb *redis.Client) {
	// Obtener todos los estados de los archivos guardados en Redis
	estados, err := rdb.HGetAll(context.Background(), "historial_archivos").Result()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error al obtener los estados de los archivos"})
		return
	}

	response := gin.H{"historial_lotes": estados}
	// Devolver los estados como respuesta JSON
	c.JSON(http.StatusOK, response)
}
