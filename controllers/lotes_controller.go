package controllers

import (
	"context"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
)

func HandleStatusConsulta(c *gin.Context, rdb *redis.Client) {

	token := c.GetHeader("Authorization")

	// Validar el token
	if err := ValidateToken(token); err != nil {
		// Manejar el error, por ejemplo, enviar una respuesta de error al cliente
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	// Obtener todos los estados de los archivos guardados en Redis
	estados, err := rdb.Keys(context.Background(), "*").Result()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error al obtener los estados de los archivos"})
		return
	}

	// Filtrar los archivos por extensión ".xlsx"
	xlsxFiles := make(map[string]string)
	for _, estado := range estados {
		if strings.HasSuffix(estado, ".xlsx") {
			// Obtener el estado del archivo y agregarlo al mapa
			status, _ := rdb.Get(context.Background(), estado).Result()
			xlsxFiles[estado] = status
		}
	}

	response := gin.H{"historial_lotes": xlsxFiles}
	// Devolver los estados filtrados como respuesta JSON
	c.JSON(http.StatusOK, response)
}
