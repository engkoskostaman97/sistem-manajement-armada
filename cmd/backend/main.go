package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/joho/godotenv"

	"transjakarta-fleet/internal/api"
	"transjakarta-fleet/internal/db"
	"transjakarta-fleet/internal/mqtt"
	"transjakarta-fleet/internal/rabbitmq"
)

func main() {

	if err := godotenv.Load(); err != nil {
		log.Println("WARNING: .env file not found, using system environment variables")
	}

	log.Println("Starting backend...")

	pg := os.Getenv("POSTGRES_DSN")
	log.Printf("POSTGRES_DSN: %s", pg)

	mqttBroker := os.Getenv("MQTT_BROKER")
	if mqttBroker == "" {
		mqttBroker = "tcp://localhost:1883"
	}
	log.Printf("MQTT_BROKER: %s", mqttBroker)

	rabbitURL := os.Getenv("RABBITMQ_URL")
	log.Printf("RABBITMQ_URL: %s", rabbitURL)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	log.Printf("PORT: %s", port)

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	log.Println("Connecting to database...")
	dbconn, err := db.New(pg)
	if err != nil {
		log.Fatalf("DB error: %v", err)
	}
	defer dbconn.Close()
	log.Println("Database connected")

	log.Println("Connecting to RabbitMQ...")
	rmq, err := rabbitmq.NewPublisher(rabbitURL)
	if err != nil {
		log.Fatalf("RabbitMQ error: %v", err)
	}
	defer rmq.Close()
	log.Println("RabbitMQ connected")

	geoLat := os.Getenv("GEOFENCE_LAT")
	geoLon := os.Getenv("GEOFENCE_LON")
	geoRad := os.Getenv("GEOFENCE_RADIUS_M")
	log.Printf("Geofence: lat=%s, lon=%s, rad=%s", geoLat, geoLon, geoRad)

	log.Println("Connecting to MQTT...")
	mqttClient, err := mqtt.NewSubscriber(mqttBroker, dbconn, rmq, geoLat, geoLon, geoRad)
	if err != nil {
		log.Fatalf("MQTT error: %v", err)
	}
	defer mqttClient.Disconnect(250)
	log.Println("MQTT connected")

	srv := api.NewServer(dbconn)
	log.Printf("Starting server on port %s...", port)
	go func() {
		if err := srv.Run(":" + port); err != nil {
			log.Fatalf("Server error: %v", err)
		}
	}()
	log.Println("Server started")

	<-ctx.Done()
	log.Println("Shutting down...")
}
