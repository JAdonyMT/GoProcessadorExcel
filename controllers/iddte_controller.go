package controllers

import (
	"context"
	"fmt"
	"net/http"
	"regexp"
	"sort"
	"strconv"

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

func HandleUniqueStatusIddte(c *gin.Context, rdb *redis.Client) {
	// Obtener el correlativo del lote de los parámetros de la solicitud
	correlativo := c.Param("id")

	// Construir el nombre del lote usando el correlativo
	nombreLote := fmt.Sprintf("Lote_%s", correlativo)

	// Obtener los estados de los IDDTEs dentro del lote en Redis
	estados, err := rdb.HGetAll(context.Background(), nombreLote).Result()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error al obtener los estados del lote"})
		return
	}

	// Verificar si no se encontraron estados para el correlativo dado
	if len(estados) == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": fmt.Sprintf("No existe un lote con el correlativo %s", correlativo)})
		return
	}

	// Crear un slice para almacenar las claves
	var claves []string
	for clave := range estados {
		claves = append(claves, clave)
	}

	// Definir una función para obtener el valor numérico de una clave
	getNumero := func(clave string) int {
		numeroStr := regexp.MustCompile(`\d+`).FindString(clave)
		numero, _ := strconv.Atoi(numeroStr)
		return numero
	}

	// Ordenar las claves basadas en sus valores numéricos
	sort.Slice(claves, func(i, j int) bool {
		return getNumero(claves[i]) < getNumero(claves[j])
	})

	// Crear un mapa ordenado para almacenar los estados
	estadosOrdenados := make(map[string]string)
	for _, clave := range claves {
		estadosOrdenados[clave] = estados[clave]
	}

	// Devolver los estados del lote como respuesta JSON
	response := gin.H{nombreLote: estadosOrdenados}
	c.JSON(http.StatusOK, response)
}
