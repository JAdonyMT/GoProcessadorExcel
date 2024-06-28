package controllers

import (
	"GoProcesadorExcel/authentication"
	"context"
	"fmt"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"unicode"

	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
	"github.com/iancoleman/orderedmap"
)

func HandleStatusIddte(c *gin.Context, rdb *redis.Client) {

	token := c.GetHeader("Authorization")

	// Validar el token
	empid, err := authentication.ValidateToken(token)
	if err != nil {
		// Manejar el error, por ejemplo, enviar una respuesta de error al cliente
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	// Obtener todas las claves que coinciden con el patrón "Lote_*"
	keys, err := rdb.Keys(context.Background(), empid+"_Lote_*").Result()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error al obtener las claves de los archivos"})
		return
	}

	// Crear un mapa para almacenar los resultados
	historial := make(map[string]*orderedmap.OrderedMap)

	// Filtrar las claves para quedarnos solo con aquellas que corresponden a hashes
	hashKeys := make([]string, 0)
	for _, key := range keys {
		keyType, err := rdb.Type(context.Background(), key).Result()
		if err != nil {
			// Manejar el error
			continue
		}
		if keyType == "hash" {
			hashKeys = append(hashKeys, key)
		}
	}

	// Aplicar HGetAll solo a las claves que corresponden a hashes
	for _, key := range hashKeys {
		estados, err := rdb.HGetAll(context.Background(), key).Result()
		if err != nil {
			// Manejar el error
			continue
		}

		// Crear un mapa ordenado para almacenar los estados ordenados
		estadosOrdenados := orderedmap.New()

		// Crear un slice para almacenar las claves ordenadas
		var claves []string
		for clave := range estados {
			claves = append(claves, clave)
		}

		// Función para obtener el valor numérico de una clave
		getNumero := func(clave string) int {
			numeroStr := strings.TrimLeftFunc(clave, func(r rune) bool {
				return !unicode.IsDigit(r) // Eliminar los caracteres no numéricos al principio
			})
			numero, _ := strconv.Atoi(numeroStr)
			return numero
		}

		// Ordenar las claves basadas en sus valores numéricos
		sort.Slice(claves, func(i, j int) bool {
			return getNumero(claves[i]) < getNumero(claves[j])
		})

		// Insertar los estados en el mapa ordenado
		for _, clave := range claves {
			estadosOrdenados.Set(clave, estados[clave])
		}

		lote := strings.TrimPrefix(key, empid+"_")
		// Agregar el mapa ordenado al historial
		historial[lote] = estadosOrdenados
	}

	// Devolver el historial como respuesta JSON
	response := gin.H{"historial_iddtes": historial}
	c.JSON(http.StatusOK, response)
}

// HandleUniqueStatusIddte maneja la solicitud para obtener estados IDDTE únicos
func HandleUniqueStatusIddte(c *gin.Context, rdb *redis.Client) {

	token := c.GetHeader("Authorization")

	// Validar el token
	empid, err := authentication.ValidateToken(token)
	if err != nil {
		// Manejar el error, por ejemplo, enviar una respuesta de error al cliente
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	// Obtener el correlativo del lote de los parámetros de la solicitud
	correlativo := c.Param("id")

	// Construir el nombre del lote usando el correlativo
	nombreLote := fmt.Sprintf("%s_Lote_%s", empid, correlativo)

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

	// Crear un mapa ordenado para almacenar los estados ordenados
	estadosOrdenados := orderedmap.New()

	// Crear un slice para almacenar las claves ordenadas
	var claves []string
	for clave := range estados {
		claves = append(claves, clave)
	}

	// Función para obtener el valor numérico de una clave
	getNumero := func(clave string) int {
		numeroStr := strings.TrimLeftFunc(clave, func(r rune) bool {
			return !unicode.IsDigit(r) // Eliminar los caracteres no numéricos al principio
		})
		numero, _ := strconv.Atoi(numeroStr)
		return numero
	}

	// Ordenar las claves basadas en sus valores numéricos
	sort.Slice(claves, func(i, j int) bool {
		return getNumero(claves[i]) < getNumero(claves[j])
	})

	// Insertar los estados en el mapa ordenado
	for _, clave := range claves {
		estadosOrdenados.Set(clave, estados[clave])
	}

	lote := strings.TrimPrefix(nombreLote, empid+"_")

	response := gin.H{lote: estadosOrdenados}
	// Devolver los estados del lote como respuesta JSON
	c.JSON(http.StatusOK, response)
}
