package utils

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"

	"github.com/go-redis/redis/v8"
)

var apiMap = map[string]string{
	"01": "/dte/fc",
	"03": "/dte/ccf",
	"11": "/dte/fex",
	"05": "/dte/nc",
}

func ProcesarArchivoJSON(rutaEntrada string, tipoDte string, authToken string, rdb *redis.Client, correlativo int) {
	dteApi, ok := apiMap[tipoDte]
	if !ok {
		log.Printf("Tipo de DTE no válido: %s\n", tipoDte)
		return
	}

	apiURL := os.Getenv("LOCALHOST_API")
	api := apiURL + dteApi

	nombreLote := fmt.Sprintf("Lote_%03d", correlativo)

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

	// Enviar cada estructura a la API y registrar su estado en Redis
	for id, estructura := range estructuras {
		contenidoJSON, err := json.Marshal(estructura)
		if err != nil {
			log.Printf("Error al convertir la estructura a JSON: %v\n", err)
			continue
		}

		// Crear la solicitud HTTP
		req, err := http.NewRequest("POST", api, bytes.NewBuffer(contenidoJSON))
		if err != nil {
			log.Printf("Error al crear la solicitud HTTP: %v\n", err)
			// Guardar el error en Redis
			guardarEstadoEnRedis(rdb, nombreLote, "IDDTE-"+id, err.Error())
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
			// Guardar el error en Redis
			guardarEstadoEnRedis(rdb, nombreLote, "IDDTE-"+id, err.Error())
			continue
		}
		defer respuesta.Body.Close()

		// Leer el cuerpo de la respuesta
		cuerpoRespuesta, err := ioutil.ReadAll(respuesta.Body)
		if err != nil {
			log.Printf("Error al leer la respuesta de la API: %v\n", err)
			// Guardar el error en Redis
			guardarEstadoEnRedis(rdb, nombreLote, "IDDTE-"+id, err.Error())
			continue
		}

		// Obtener el estado de la respuesta
		estadoRespuesta := fmt.Sprintf("Códe: %d %s, Menssage: %s", respuesta.StatusCode, respuesta.Status, string(cuerpoRespuesta))

		// Registrar el estado del IDDTE en Redis
		guardarEstadoEnRedis(rdb, nombreLote, "IDDTE-"+id, estadoRespuesta)

		// Crear el archivo de registro
		logFileName := "log.txt"
		logFile, err := os.OpenFile(logFileName, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			log.Printf("Error al abrir o crear el archivo de registro %s: %v\n", logFileName, err)
			return
		}
		defer logFile.Close()

		// Escribir en el archivo de registro
		logEntry := fmt.Sprintf("IDDTE: %s - Código de estado de la respuesta: %d %s\n", "IDDTE-"+id, respuesta.StatusCode, respuesta.Status)
		logEntry += fmt.Sprintf("IDDTE: %s - Mensaje de la respuesta: %s\n", "IDDTE-"+id, string(cuerpoRespuesta))
		if _, err := logFile.WriteString(logEntry); err != nil {
			log.Printf("Error al escribir en el archivo de registro: %v\n", err)
			continue
		}
	}
}

func guardarEstadoEnRedis(rdb *redis.Client, nombreLote string, id string, estado string) {
	// Guardar el estado en el hash del lote correspondiente
	if err := rdb.HSet(context.Background(), nombreLote, id, estado).Err(); err != nil {
		log.Printf("Error al guardar el estado en Redis para IDDTE %s del lote %s: %v\n", id, nombreLote, err)
	}
}
