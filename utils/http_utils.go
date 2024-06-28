package utils

import (
	"errors"
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

var ErrNoRetries = errors.New("no se realizaron reintentos")

// Función para enviar datos a la API con reintentos y análisis de mensajes de error
func SendWithRetries(req *http.Request, client *http.Client) (*http.Response, int, string, error) {
	const maxRetries = 1
	var originalErrorMessage string
	var StatusCode int
	retried := false

	for i := 0; i < maxRetries; i++ {

		reqClone := cloneRequest(req) // Clonar la solicitud original

		resp, err := client.Do(reqClone)
		if err != nil {

			originalErrorMessage = err.Error()
			// Error de comunicación, como red o timeout
			if isNetworkError(err) {
				log.Printf("Intento %d: Error de red o timeout: %v\n", i+1, err)
				retried = true
			} else {
				log.Printf("Intento %d: Error al enviar la solicitud HTTP: %v\n", i+1, err)
				retried = true
			}
		} else {

			StatusCode = resp.StatusCode
			// Verificar si la respuesta indica un error
			if StatusCode >= 400 {
				// Leer el cuerpo de la respuesta si está disponible
				body, readErr := ioutil.ReadAll(resp.Body)
				if readErr != nil {
					log.Printf("Error al leer el cuerpo de la respuesta: %v\n", readErr)
				} else {
					log.Printf("Mensaje de la respuesta: %s\n", string(body))

					if i == 0 {
						originalErrorMessage = string(body)
					}

					// Verificar si el cuerpo de la respuesta coincide con algún patrón de error esperado
					for status, pattern := range errorPatterns {
						if StatusCode == status && strings.Contains(string(body), pattern) {
							log.Printf("Coincidencia encontrada - Patrón: %s\n", pattern)
							log.Printf("Error del servidor recuperable detectado. Reintentando...")
							retried = true
							time.Sleep(2 * time.Second) // Esperar antes de intentar nuevamente
							continue
						}
					}
				}
			} else {
				// La respuesta es exitosa, retornarla
				return resp, StatusCode, "", nil
			}
		}

		time.Sleep(2 * time.Second) // Esperar antes de intentar nuevamente
	}
	if retried {
		// Si se realizó un reintento exitoso, no se excedió el número máximo de reintentos
		return nil, StatusCode, originalErrorMessage, fmt.Errorf("se excedió el número máximo de reintentos")
	}
	return nil, StatusCode, originalErrorMessage, ErrNoRetries
}

func isNetworkError(err error) bool {
	netErr, ok := err.(net.Error)
	return ok && (netErr.Timeout() || netErr.Temporary())
}

func cloneRequest(req *http.Request) *http.Request {
	reqClone := req.Clone(req.Context()) // Clonar la solicitud original
	reqClone.Header = make(http.Header)  // Crear un nuevo encabezado para la solicitud clonada
	// Copiar todos los encabezados de la solicitud original a la solicitud clonada
	for key, values := range req.Header {
		reqClone.Header[key] = values
	}
	return reqClone
}
