package controllers

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"

	// "fmt"
	// "net/http"
	"strings"
	"time"
	// "github.com/dgrijalva/jwt-go"
	// "github.com/gin-gonic/gin"
)

func ValidateToken(tokenString string) error {
	tokenInvalido := errors.New("token Invalido")

	// Lógica para verificar el token
	if tokenString == "" {
		fmt.Println("token vacio")
		return errors.New("token vacío")
	}

	// Extrae el token del encabezado
	tokenString = strings.Replace(tokenString, "Bearer ", "", 1)

	// Divide el token en partes (header, payload, firma)
	parts := strings.Split(tokenString, ".")
	if len(parts) != 3 {
		fmt.Println("token invalido")
		return tokenInvalido
	}

	// Decodifica y analiza la sección de encabezado del token
	headerData, err := base64.RawURLEncoding.DecodeString(parts[0])
	if err != nil {
		fmt.Println("token invalido")
		return tokenInvalido
	}

	var header map[string]interface{}
	if err := json.Unmarshal(headerData, &header); err != nil {
		fmt.Println("token invalido")
		return tokenInvalido
	}

	// Verificaciones específicas del encabezado
	// Puedes agregar más validaciones según tus requisitos
	if typ, ok := header["typ"].(string); !ok || typ != "JWT" {
		fmt.Println("token invalido")
		return tokenInvalido
	}

	if alg, ok := header["alg"].(string); !ok || alg != "HS256" {
		fmt.Println("token invalido")
		return tokenInvalido
	}

	// Decodifica y analiza la sección de payload del token
	payloadData, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		fmt.Println("token invalido")
		return tokenInvalido
	}

	var payload map[string]interface{}
	if err := json.Unmarshal(payloadData, &payload); err != nil {
		fmt.Println("token invalido")
		return tokenInvalido
	}

	// Verificaciones específicas del payload
	// Verifica la expiración del token
	expirationTime, ok := payload["exp"].(float64)
	if !ok {
		fmt.Println("token invalido")
		return tokenInvalido
	}

	expiration := time.Unix(int64(expirationTime), 0)
	if expiration.Before(time.Now()) {
		fmt.Println("token Expirado")
		return errors.New("el token ha expirado")
	}

	// Si el token está presente y válido
	return nil
}
