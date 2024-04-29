package controllers

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
	"time"
)

func TestScriptCall(t *testing.T) {
	// Llamar al script de Python con algunos argumentos
	cmd := exec.Command("python", "../utils/excelProcessor.py", "../utils/factura.xlsx", "01", "2")

	// Capturar la salida estándar y la salida de error del proceso
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	// Ejecutar el comando
	err := cmd.Run()
	if err != nil {
		// Si hay un error, imprimir la salida de error
		t.Fatalf("Error al ejecutar el script: %v\nSalida de error: %s", err, stderr.String())
	}

	// Verificar si se generaron archivos CSV y JSON
	// csvExists := fileExists("ruta_del_archivo.csv")
	jsonExists := fileExists("factura.json")

	// Si no se generaron los archivos, se considera un error en la conversión
	if !jsonExists {
		t.Error("Error en la conversión: no se generaron los archivos CSV y JSON esperados")
	}

	// Limpiar archivos generados
	err = deleteGeneratedFiles()
	if err != nil {
		t.Fatalf("Error al eliminar archivos generados: %v", err)
	}
}

func fileExists(filename string) bool {
	// Verificar si el archivo existe
	_, err := os.Stat(filename)
	return !os.IsNotExist(err)
}

func deleteGeneratedFiles() error {
	// Directorio donde se esperan los archivos CSV y JSON
	dir := "./"

	// Buscar archivos CSV y JSON en el directorio
	files, err := filepath.Glob(filepath.Join(dir, "*.csv"))
	if err != nil {
		return err
	}
	// Eliminar archivos CSV
	for _, file := range files {
		if err := os.Remove(file); err != nil {
			return err
		}
	}

	files, err = filepath.Glob(filepath.Join(dir, "*.json"))
	if err != nil {
		return err
	}
	time.Sleep(2 * time.Second)
	// Eliminar archivos JSON
	for _, file := range files {
		if err := os.Remove(file); err != nil {
			return err
		}
	}

	return nil
}
