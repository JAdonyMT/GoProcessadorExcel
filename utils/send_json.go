package utils

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
)

func ProcesarArchivoJSON(rutaEntrada string, urlAPI string, authToken string) {
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
