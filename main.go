package main

import (
	"GoProcesadorExcel/routes"
	"context"
	"log"

	"github.com/go-redis/redis/v8"
	"github.com/joho/godotenv"
)

var rdb *redis.Client

func main() {

	if err := godotenv.Load(); err != nil {
		log.Fatal("Error al cargar archivo .env")
	}

	// redisUrl := os.Getenv("REDIS_URL")
	// redisKey := os.Getenv("REDIS_PSW")
	// Configurar la conexión a Redis
	rdb = redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "",
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
