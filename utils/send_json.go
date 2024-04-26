package utils

import (
	"GoProcesadorExcel/authentication"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/go-redis/redis/v8"
)

var apiMap = map[string]string{
	"01":     "/dte/fc",
	"03":     "/dte/ccf",
	"04":     "/dte/nr",
	"05":     "/dte/ncnd",
	"06":     "/dte/ncnd",
	"07":     "/dte/cr",
	"08":     "/dte/cl",
	"09":     "/dte/dcl",
	"11":     "/dte/fex",
	"14":     "/dte/fse",
	"15":     "/dte/cd",
	"cancel": "/dte/cancel",
}

// ProcesarArchivoJSON procesa un archivo JSON enviando sus estructuras a una API y registrando su estado en Redis
func ProcesarArchivoJSON(rutaEntrada string, tipoDte string, authToken string, rdb *redis.Client, correlativo int) {

	empid, _ := authentication.ValidateToken(authToken)

	// Paso 1: Obtener la API correspondiente al tipo de DTE
	dteApi, ok := apiMap[tipoDte]
	if !ok {
		log.Printf("Tipo de DTE no válido: %s\n", tipoDte)
		return
	}

	// Paso 2: Construir la URL de la API
	apiURL := os.Getenv("FACTURED_API")
	api := apiURL + dteApi

	// Paso 3: Generar un nombre de lote único
	nombreLote := fmt.Sprintf("%s_Lote_%03d", empid, correlativo)

	// Paso 4: Leer el archivo JSON
	contenido, err := os.ReadFile(rutaEntrada)
	if err != nil {
		log.Printf("Error al leer el archivo JSON %s: %v\n", rutaEntrada, err)
		return
	}

	// Paso 5: Analizar el JSON en una estructura de datos
	var estructuras map[string]interface{}
	err = json.Unmarshal(contenido, &estructuras)
	if err != nil {
		log.Printf("Error al analizar el JSON en %s: %v\n", rutaEntrada, err)
		return
	}

	// Paso 6: Crear un cliente HTTP para reutilizarlo
	cliente := &http.Client{}

	// Paso 7: Crear un canal para limitar el número de goroutines
	maxGoroutines := 15 // Establece el número máximo de goroutines
	semaforo := make(chan struct{}, maxGoroutines)

	// Paso 8: Utilizar un WaitGroup para esperar a que todas las goroutines terminen
	var wg sync.WaitGroup

	log.Println("Iniciando el envío de las estructuras a la API...")
	// Paso 9: Enviar cada estructura a la API y registrar su estado en Redis
	for id, estructura := range estructuras {
		wg.Add(1) // Incrementar el contador del WaitGroup

		// Añadir una marca al canal
		semaforo <- struct{}{}

		go func(id string, estructura interface{}) {
			defer func() {
				// Eliminar una marca del canal al terminar
				<-semaforo

				// Decrementar el contador del WaitGroup
				wg.Done()
			}()

			log.Printf("Iniciando envío de la estructura %s\n", id)

			// Paso 10: Convertir la estructura a JSON
			contenidoJSON, err := json.Marshal(estructura)
			if err != nil {
				log.Printf("Error al convertir la estructura a JSON: %v\n", err)
				return
			}

			// Paso 11: Crear la solicitud HTTP
			req, err := http.NewRequest("POST", api, bytes.NewBuffer(contenidoJSON))
			if err != nil {
				log.Printf("Error al crear la solicitud HTTP: %v\n", err)
				// Guardar el error en Redis
				guardarEstadoEnRedis(rdb, nombreLote, "IDDTE-"+id, err.Error())
				return
			}

			// Paso 12: Agregar el encabezado de autorización
			req.Header.Set("Authorization", authToken)
			req.Header.Set("Content-Type", "application/json")

			// Paso 13: Realizar la solicitud HTTP POST a la API de forma asíncrona
			respuesta, err := SendWithRetries(req, cliente)
			if err != nil {
				log.Printf("Error al enviar la estructura %s: %v\n", id, err)
				// Guardar el error en Redis
				guardarEstadoEnRedis(rdb, nombreLote, "IDDTE-"+id, err.Error())
				return
			}
			defer respuesta.Body.Close()

			// Paso 14: Leer el cuerpo de la respuesta
			cuerpoRespuesta, err := ioutil.ReadAll(respuesta.Body)
			if err != nil {
				log.Printf("Error al leer la respuesta de la API: %v\n", err)
				// Guardar el error en Redis
				guardarEstadoEnRedis(rdb, nombreLote, "IDDTE-"+id, err.Error())
				return
			}

			// Paso 15: Obtener el estado de la respuesta
			estadoRespuesta := fmt.Sprintf("Código: %d %s, Mensaje: %s", respuesta.StatusCode, respuesta.Status, string(cuerpoRespuesta))

			// Paso 16: Registrar el estado del IDDTE en Redis
			guardarEstadoEnRedis(rdb, nombreLote, "IDDTE-"+id, estadoRespuesta)

			// Paso 17: Crear el archivo de registro
			logFileName := "IDDTElog.txt"
			logFile, err := os.OpenFile(logFileName, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
			if err != nil {
				log.Printf("Error al abrir o crear el archivo de registro %s: %v\n", logFileName, err)
				return
			}
			defer logFile.Close()

			dt := time.Now()

			// Escribir en el archivo de registro
			logEntry := fmt.Sprintf("%s - %s - Código de estado de la respuesta: %d %s\n", dt.Format(time.Stamp), "IDDTE-"+id, respuesta.StatusCode, respuesta.Status)
			logEntry += fmt.Sprintf("%s - %s - Mensaje de la respuesta: %s\n", dt.Format(time.Stamp), "IDDTE-"+id, string(cuerpoRespuesta))
			logEntry += ("\n<------------------------------------------------------------->\n")
			if _, err := logFile.WriteString(logEntry); err != nil {
				log.Printf("Error al escribir en el archivo de registro: %v\n", err)
				return
			}
		}(id, estructura)
	}

	// Paso 18: Esperar a que todas las goroutines terminen
	wg.Wait()

	log.Println("Envío de las estructuras completado.")

	// fmt.Println("Documentos JSON enviados con éxito")
}

func guardarEstadoEnRedis(rdb *redis.Client, nombreLote string, id string, estado string) {
	// Guardar el estado en el hash del lote correspondiente
	if err := rdb.HSet(context.Background(), nombreLote, id, estado).Err(); err != nil {
		log.Printf("Error al guardar el estado en Redis para IDDTE %s del lote %s: %v\n", id, nombreLote, err)
	} else {
		// Establecer un tiempo de expiración para la clave específica dentro del hash
		// expiration := 3 * 30 * 24 * time.Hour // 3 meses en horas,en vez de time.Minute
		err = rdb.Expire(context.Background(), nombreLote, 24*time.Hour).Err()
		if err != nil {
			log.Printf("Error al establecer el tiempo de expiración en Redis para IDDTE %s del lote %s: %v\n", id, nombreLote, err)
		}
	}
}
