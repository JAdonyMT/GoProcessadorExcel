package controllers

import (
	"GoProcesadorExcel/authentication"
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
	empid, err := authentication.ValidateToken(authToken)

	if err != nil {
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
	correlativo, err := generateCorrelativo(rdb, empid)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error al generar el correlativo"})
		return
	}
	nombreArchivo := fmt.Sprintf("%s_Lote_%03d.xlsx", empid, correlativo)
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
	cmd := exec.Command("python", "./utils/excelProcessor.py", tempFilePath, tipoDte, empid)

	// Capturar la salida estándar y la salida de error del proceso
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout

	dt := time.Now()

	go func() {

		err = cmd.Run()
		if err != nil {
			// Si la ejecución del script no fue exitosa, guardar un mensaje de error en Redis
			errMsg := fmt.Sprintf("Error en la conversión: %v. Detalles: %s", err, stderr.String())

			logEntry := fmt.Sprintf("\n%s - %s_Lote_%03d: Error en la conversión: %v. Detalles: ", dt.Format(time.Stamp), empid, correlativo, err)
			logWrite(logEntry, stdout.String())
			logWrite("", "<===========================================>\n")

			// Guardar el estado con expiración
			err = rdb.Set(context.Background(), nombreArchivo, errMsg, 24*time.Hour).Err()
			if err != nil {
				log.Println("Error al guardar el estado en el historial de Redis:", err)
			}
			return
		}

		successMessage := ""
		if stdout.String() == "" {
			successMessage = fmt.Sprintln("Proceso de conversion exitoso")
			logEntry := fmt.Sprintf("\n%s - %s_Lote: %03d - Proceso de conversión exitoso\n", dt.Format(time.Stamp), empid, correlativo)
			logWrite(logEntry, "")
			logWrite("", "<==========================================>\n")
		} else {
			successMessage = fmt.Sprintf("Proceso de conversion con inconvenientes \n %v", stdout.String())
			logEntry := fmt.Sprintf("\n%s - %s_Lote: %03d - Proceso de conversión con inconvenientes\n", dt.Format(time.Stamp), empid, correlativo)
			logWrite(logEntry, stdout.String())
			logWrite("", "<==========================================>\n")
		}
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
		nombreArchivoJSON := fmt.Sprintf("%s_Lote_%03d.json", empid, correlativo)

		// Construir la ruta completa del archivo JSON
		rutaArchivoJSON := filepath.Join(responseJSONDir, nombreArchivoJSON)

		// Llamar a la función para procesar el archivo JSON recibido y enviarlo a la API
		utils.ProcesarArchivoJSON(rutaArchivoJSON, tipoDte, authToken, rdb, correlativo)

	}()
}

func generateCorrelativo(rdb *redis.Client, empid string) (int, error) {
	// Incrementar el contador en Redis
	val, err := rdb.Incr(context.Background(), empid+"_contador_lotes").Result()
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

func logWrite(logentry string, stdout string) {
	logFileName := "Lotelog.txt"
	logFile, err := os.OpenFile(logFileName, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Printf("Error al abrir o crear el archivo de registro %s: %v\n", logFileName, err)
		return
	}
	defer logFile.Close()

	if _, err := logFile.WriteString(logentry); err != nil {
		log.Printf("Error al escribir en el archivo de registro: %v\n", err)
	}

	if _, err := logFile.WriteString(stdout); err != nil {
		log.Printf("Error al escribir en el archivo de registro: %v\n", err)
	}
}
