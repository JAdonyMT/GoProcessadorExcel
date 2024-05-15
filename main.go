package main

import (
	"GoProcesadorExcel/routes"
	"context"
	"log"
	"os"

	"github.com/go-redis/redis/v8"
	"github.com/joho/godotenv"
)

var rdb *redis.Client

func main() {

	if err := godotenv.Load(); err != nil {
		log.Fatal("Error al cargar archivo .env")
	}

	redisAddr := os.Getenv("REDIS_ADR")
	redisPsw := os.Getenv("REDIS_PSW")
	// Configurar la conexión a Redis
	rdb = redis.NewClient(&redis.Options{
		Addr:     redisAddr,
		Password: redisPsw,
		DB:       0,
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
