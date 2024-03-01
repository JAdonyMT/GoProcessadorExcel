package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	"github.com/gin-gonic/gin"
)

var authToken string

func handleExcelConversion(c *gin.Context) {

	authToken = c.GetHeader("Authorization")
	if authToken == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Se requiere un token de autorización"})
		return
	}

	// Obtener el archivo Excel del formulario
	file, _, err := c.Request.FormFile("excel")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "No se pudo obtener el archivo Excel"})
		return
	}
	defer file.Close()

	// Crear una carpeta temporal para almacenar los archivos recibidos
	tempDir := "archivos_excel"
	if err := os.MkdirAll(tempDir, 0755); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error al crear carpeta temporal"})
		return
	}

	// Crear el archivo en la carpeta temporal con un nombre único basado en la fecha y hora actual
	fechaHoraActual := time.Now().Format("20060102_150405") // Formato: YYYYMMDD_HHMMSS
	nombreArchivo := "excel_" + fechaHoraActual + ".xlsx"
	tempFile, err := os.Create(filepath.Join(tempDir, nombreArchivo))
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

	// Llamar al script de Python para procesar el archivo Excel
	cmd := exec.Command("python", "excelProcessor.py", tempFile.Name())
	err = cmd.Run()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error al ejecutar script de Python"})
		return
	}

	// Mover archivos JSON y CSV a carpetas específicas
	responseJSONDir := "responseJSON"
	if err := os.MkdirAll(responseJSONDir, 0755); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error al crear carpeta para archivos JSON"})
		return
	}

	csvJSONDir := "csvErrors"
	if err := os.MkdirAll(csvJSONDir, 0755); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error al crear carpeta para archivos CSV"})
		return
	}

	// Obtener los nombres de los archivos generados
	files, err := filepath.Glob("*.json")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error al obtener nombres de archivos JSON"})
		return
	}

	// Mover archivos JSON a la carpeta responseJSON
	for _, f := range files {
		if err := moveFile(f, responseJSONDir); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error al mover archivo JSON"})
			return
		}
	}

	// Obtener los nombres de los archivos generados
	files, err = filepath.Glob("*.csv")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error al obtener nombres de archivos CSV"})
		return
	}

	// Mover archivos CSV a la carpeta csvJSON
	for _, f := range files {
		if err := moveFile(f, csvJSONDir); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error al mover archivo CSV"})
			return
		}
	}

	c.JSON(http.StatusOK, gin.H{"message": "Archivos procesados exitosamente"})
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

func procesarArchivoJSON(rutaEntrada string, urlAPI string, authToken string) {
	// Leer el archivo JSON
	contenido, err := ioutil.ReadFile(rutaEntrada)
	if err != nil {
		log.Printf("Error al leer el archivo JSON %s: %v\n", rutaEntrada, err)
		return
	}

	// Analizar el JSON en una estructura de datos
	var estructuras map[string]interface{}
	err = json.Unmarshal(contenido, &estructuras)
	if err != nil {
		log.Printf("Error al analizar el JSON en %s: %v\n", rutaEntrada, err)
		return
	}

	// Archivo de registro
	logFile, err := os.OpenFile("log.txt", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Printf("Error al abrir o crear el archivo de registro: %v\n", err)
		return
	}
	defer logFile.Close()

	// Enviar cada estructura a la API
	for id, estructura := range estructuras {
		contenidoJSON, err := json.Marshal(estructura)
		if err != nil {
			log.Printf("Error al convertir la estructura a JSON: %v\n", err)
			continue
		}

		// Crear la solicitud HTTP
		req, err := http.NewRequest("POST", urlAPI, bytes.NewBuffer(contenidoJSON))
		if err != nil {
			log.Printf("Error al crear la solicitud HTTP: %v\n", err)
			continue
		}

		// Agregar el encabezado de autorización
		req.Header.Set("Authorization", authToken)
		req.Header.Set("Content-Type", "application/json")

		// Realizar la solicitud HTTP POST a la API
		cliente := &http.Client{}
		respuesta, err := cliente.Do(req)
		if err != nil {
			log.Printf("Error al enviar la solicitud HTTP: %v\n", err)
			continue
		}
		defer respuesta.Body.Close()

		// Leer el cuerpo de la respuesta
		cuerpoRespuesta, err := ioutil.ReadAll(respuesta.Body)
		if err != nil {
			log.Printf("Error al leer la respuesta de la API: %v\n", err)
			continue
		}

		// Escribir en el archivo de registro
		logEntry := fmt.Sprintf("IDDTE: %s - Código de estado de la respuesta: %d %s\n", id, respuesta.StatusCode, respuesta.Status)
		logEntry += fmt.Sprintf("IDDTE: %s - Mensaje de la respuesta: %s\n", id, string(cuerpoRespuesta))
		if _, err := logFile.WriteString(logEntry); err != nil {
			log.Printf("Error al escribir en el archivo de registro: %v\n", err)
			continue
		}
	}
}

// Variables globales para almacenar los archivos procesados
var archivosProcesados = make(map[string]bool)

func main() {
	// Iniciar el servidor HTTP en un goroutine
	go func() {
		r := gin.Default()
		r.POST("/convert", handleExcelConversion)
		r.Run(":8080")
	}()

	// Monitorear la carpeta de entrada en busca de nuevos archivos JSON
	for {
		archivos, err := ioutil.ReadDir("responseJSON")
		if err != nil {
			log.Printf("Error al leer la carpeta responseJSON: %v\n", err)
			time.Sleep(10 * time.Second)
			continue
		}

		for _, archivo := range archivos {
			if archivo.IsDir() || filepath.Ext(archivo.Name()) != ".json" {
				continue
			}

			rutaArchivo := filepath.Join("responseJSON", archivo.Name())
			if !archivosProcesados[archivo.Name()] {
				procesarArchivoJSON(rutaArchivo, "https://test.factured.sv/ApiFEL/api/v1/dte/fc", authToken)
				archivosProcesados[archivo.Name()] = true
			}
		}

		// Esperar un tiempo antes de volver a buscar
		time.Sleep(10 * time.Second)
	}
}
