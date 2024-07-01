package controllers

import (
	"GoProcesadorExcel/authentication"
	"context"
	"fmt"
	"net/http"
	"regexp"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
)

func HandleStatusConsulta(c *gin.Context, rdb *redis.Client) {

	token := c.GetHeader("Authorization")

	// Validar el token
	empid, err := authentication.ValidateToken(token)
	if err != nil {
		// Manejar el error, por ejemplo, enviar una respuesta de error al cliente
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	// Obtener todos los estados de los archivos guardados en Redis
	empPrefix := fmt.Sprintf("%s_", empid)
	estados, err := rdb.Keys(context.Background(), empPrefix+"Lote_*").Result()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error al obtener los estados de los archivos"})
		return
	}

	// Filtrar los archivos por extensión ".xlsx"
	xlsxFiles := make(map[string]string)

	patronCancel := regexp.MustCompile(`\.xlsx:cancel$`)
	patron := regexp.MustCompile(`\.xlsx:\d+$`)
	all := regexp.MustCompile(`\.xlsx`)

	for _, estado := range estados {
		if patron.MatchString(estado) || patronCancel.MatchString(estado) || all.MatchString(estado) {
			// Obtener el estado del archivo y agregarlo al mapa
			status, _ := rdb.Get(context.Background(), estado).Result()
			lote := strings.TrimPrefix(estado, empPrefix)

			xlsxFiles[lote] = status
		}
	}

	response := gin.H{"historial_lotes": xlsxFiles}
	// Devolver los estados filtrados como respuesta JSON
	c.JSON(http.StatusOK, response)
}
