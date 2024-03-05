package controllers

import (
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	"github.com/gin-gonic/gin"
)

func HandleExcelConversion(c *gin.Context) {
	// authToken = c.GetHeader("Authorization")
	// if authToken == "" {
	// 	c.JSON(http.StatusUnauthorized, gin.H{"error": "Se requiere un token de autorización"})
	// 	return
	// }

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
