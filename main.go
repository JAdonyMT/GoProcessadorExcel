package main

import (
	"GoProcesadorExcel/routes"
	"context"
	"log"

	"github.com/go-redis/redis/v8"
)

var rdb *redis.Client

func main() {

	// Configurar la conexión a Redis
	rdb = redis.NewClient(&redis.Options{
		Addr:     "localhost:6379", // Dirección de tu instancia de Redis
		Password: "",               // Contraseña (si es necesario)
		DB:       0,                // Número de la base de datos
	})

	// Verificar la conexión a Redis
	pong, err := rdb.Ping(context.Background()).Result()
	if err != nil {
		log.Fatalf("Error al conectar con Redis: %v", err)
	}
	log.Printf("Conexión a Redis establecida: %s", pong)

	r := routes.SetupRouter(rdb)
	r.Run(":8080")
}
