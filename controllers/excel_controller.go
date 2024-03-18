package controllers

import (
	"GoProcesadorExcel/utils"
	"bytes"
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
)

func HandleExcelConversion(c *gin.Context, rdb *redis.Client) {

	authToken := c.GetHeader("Authorization")

	// Validar el token
	if err := ValidateToken(authToken); err != nil {
		// Manejar el error, por ejemplo, enviar una respuesta de error al cliente
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	tipoDte := c.GetHeader("tipoDte")
	if tipoDte == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Falta el parámetro tipoDte"})
		return
	}

	// Obtener el archivo Excel del formulario
	file, fileHeader, err := c.Request.FormFile("excel")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "No se pudo obtener el archivo Excel"})
		return
	}
	defer file.Close()

	if path.Ext(fileHeader.Filename) != ".xlsx" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "El archivo no es un archivo Excel"})
		return
	}

	// Crear una carpeta temporal para almacenar los archivos recibidos
	tempDir := "data/archivos_excel"
	if err := os.MkdirAll(tempDir, 0755); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error al crear carpeta temporal"})
		return
	}

	// Generar el nombre del archivo con el formato Lote_{correlativo}
	correlativo, err := generateCorrelativo(rdb)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error al generar el correlativo"})
		return
	}
	nombreArchivo := fmt.Sprintf("Lote_%03d.xlsx", correlativo)
	tempFilePath := filepath.Join(tempDir, nombreArchivo)
	tempFile, err := os.Create(tempFilePath)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error al crear archivo temporal"})
		return
	}
	defer tempFile.Close()

	// Escribir el archivo en el sistema de archivos
	_, err = io.Copy(tempFile, file)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error al guardar archivo Excel"})
		return
	}

	// Devolver una respuesta al cliente indicando que el archivo se está procesando
	c.JSON(http.StatusOK, gin.H{"message": "El archivo se está procesando"})

	// Llamar al script de Python para procesar el archivo Excel
	cmd := exec.Command("python", "excelProcessor.py", tempFilePath, tipoDte)

	// Capturar la salida estándar y la salida de error del proceso
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	// Ejecutar el proceso en segundo plano
	go func() {
		err := cmd.Run()
		if err != nil {
			// Si la ejecución del script no fue exitosa, guardar un mensaje de error en Redis
			errMsg := fmt.Sprintf("Error en la conversión: %v. Detalles: %s", err, stderr.String())
			// Guardar el estado con expiración
			err = rdb.Set(context.Background(), nombreArchivo, errMsg, 10*time.Minute).Err()
			if err != nil {
				log.Println("Error al guardar el estado en el historial de Redis:", err)
			}
			return
		}

		successMessage := "Proceso de conversion exitoso"
		// Guardar el estado con expiración
		err = rdb.Set(context.Background(), nombreArchivo, successMessage, 24*time.Hour).Err()
		if err != nil {
			log.Println("Error al guardar el estado en el historial de Redis:", err)
		}
		// Mover archivos JSON y CSV a carpetas específicas
		responseJSONDir := "data/responseJSON"
		if err := os.MkdirAll(responseJSONDir, 0755); err != nil {
			log.Println("Error al crear carpeta para archivos JSON:", err)
			return
		}

		csvJSONDir := "data/csvErrors"
		if err := os.MkdirAll(csvJSONDir, 0755); err != nil {
			log.Println("Error al crear carpeta para archivos CSV:", err)
			return
		}

		// Obtener los nombres de los archivos generados
		files, err := filepath.Glob("*.json")
		if err != nil {
			log.Println("Error al obtener nombres de archivos JSON:", err)
			return
		}

		// Mover archivos JSON a la carpeta responseJSON
		for _, f := range files {
			if err := moveFile(f, responseJSONDir); err != nil {
				log.Println("Error al mover archivo JSON:", err)
				return
			}
		}

		// Obtener los nombres de los archivos generados
		files, err = filepath.Glob("*.csv")
		if err != nil {
			log.Println("Error al obtener nombres de archivos CSV:", err)
			return
		}

		// Mover archivos CSV a la carpeta csvErrors
		for _, f := range files {
			if err := moveFile(f, csvJSONDir); err != nil {
				log.Println("Error al mover archivo CSV:", err)
				return
			}
		}
		// Obtener el nombre del archivo JSON basado en el correlativo
		nombreArchivoJSON := fmt.Sprintf("Lote_%03d.json", correlativo)

		// Construir la ruta completa del archivo JSON
		rutaArchivoJSON := filepath.Join(responseJSONDir, nombreArchivoJSON)

		// Llamar a la función para procesar el archivo JSON recibido y enviarlo a la API
		utils.ProcesarArchivoJSON(rutaArchivoJSON, tipoDte, authToken, rdb, correlativo)
	}()
}

func generateCorrelativo(rdb *redis.Client) (int, error) {
	// Incrementar el contador en Redis
	val, err := rdb.Incr(context.Background(), "contador_lotes").Result()
	if err != nil {
		return 0, err
	}
	return int(val), nil
}

func moveFile(fileName, destDir string) error {
	// Obtener el nombre del archivo sin la ruta
	base := filepath.Base(fileName)

	src := fileName
	dst := filepath.Join(destDir, base)

	err := os.Rename(src, dst)
	if err != nil {
		return err
	}
	return nil
}
