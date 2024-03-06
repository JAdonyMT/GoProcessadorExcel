package controllers

import (
	"context"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
)

func HandleStatusIddte(c *gin.Context, rdb *redis.Client) {
	// Obtener todos los hashes que comienzan con "Lote_"
	keys, err := rdb.Keys(context.Background(), "Lote_*").Result()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error al obtener los estados de los archivos"})
		return
	}

	// Crear un mapa para almacenar los resultados
	historial := make(map[string]map[string]string)

	// Recorrer cada clave encontrada
	for _, key := range keys {
		// Obtener el hash correspondiente
		estados, err := rdb.HGetAll(context.Background(), key).Result()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error al obtener los estados de los archivos"})
			return
		}
		// Agregar el hash al mapa de historial
		historial[key] = estados
	}

	// Devolver el historial como respuesta JSON
	response := gin.H{"historial_iddtes": historial}
	c.JSON(http.StatusOK, response)
}
