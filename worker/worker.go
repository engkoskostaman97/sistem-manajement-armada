package main

import (
    "log"
    "os"
    "os/signal"
    "syscall"

    "github.com/streadway/amqp"
)

func main() {
    url := os.Getenv("RABBITMQ_URL")
    conn, err := amqp.Dial(url)
    if err != nil { log.Fatal(err) }
    ch, err := conn.Channel()
    if err != nil { log.Fatal(err) }
    defer conn.Close(); defer ch.Close()

    if err := ch.ExchangeDeclare("fleet.events", "fanout", true, false, false, false, nil); err != nil { log.Fatal(err) }
    q, err := ch.QueueDeclare("geofence_alerts", true, false, false, false, nil)
    if err != nil { log.Fatal(err) }
    if err := ch.QueueBind(q.Name, "", "fleet.events", false, nil); err != nil { log.Fatal(err) }

    msgs, err := ch.Consume(q.Name, "", true, false, false, false, nil)
    if err != nil { log.Fatal(err) }

    go func() {
        for d := range msgs {
            log.Printf("[worker] received: %s\n", d.Body)
        }
    }()

    log.Println("[worker] waiting for messages")
    c := make(chan os.Signal, 1)
    signal.Notify(c, os.Interrupt, syscall.SIGTERM)
    <-c
    log.Println("shutting worker")
}
