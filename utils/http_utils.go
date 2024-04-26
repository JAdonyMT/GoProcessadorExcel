package utils

import (
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"strings"
	"time"
)

var errorPatterns = map[int]string{
	http.StatusInternalServerError: "Error al Generar DTE",
	//Agregar mas patrones de errores que pueden ser reprocesados
}

// Función para enviar datos a la API con reintentos y análisis de mensajes de error
func SendWithRetries(req *http.Request, client *http.Client) (*http.Response, error) {
	const maxRetries = 2
	for i := 0; i < maxRetries; i++ {
		resp, err := client.Do(req)
		if err == nil {
			return resp, nil
		}

		// Verificar si el error es de red o timeout
		if isNetworkError(err) {
			log.Printf("Intento %d: Error de red o timeout: %v\n", i+1, err)
		} else {
			log.Printf("Intento %d: Error al enviar la solicitud HTTP: %v\n", i+1, err)
		}

		// Leer el cuerpo de la respuesta si está disponible
		if resp != nil && resp.Body != nil {
			body, readErr := ioutil.ReadAll(resp.Body)
			if readErr != nil {
				log.Printf("Error al leer el cuerpo de la respuesta: %v\n", readErr)
			} else {
				log.Printf("Mensaje de la respuesta: %s\n", string(body))

				// Verificar si el código de respuesta y el mensaje coinciden con algún patrón
				if pattern, ok := errorPatterns[resp.StatusCode]; ok && strings.Contains(string(body), pattern) {
					log.Printf("Mensaje de la respuesta coincide con el patrón del código de estado %d. Reintentando...\n", resp.StatusCode)
					time.Sleep(2 * time.Second) // Esperar antes de intentar nuevamente
					continue
				}
			}
		}

		time.Sleep(2 * time.Second) // Esperar antes de intentar nuevamente
	}
	return nil, fmt.Errorf("se excedió el número máximo de reintentos")
}

func isNetworkError(err error) bool {
	netErr, ok := err.(net.Error)
	return ok && (netErr.Timeout() || netErr.Temporary())
}
