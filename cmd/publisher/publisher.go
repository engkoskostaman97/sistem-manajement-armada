package main

import (
    "encoding/json"
    "fmt"
    "math/rand"
    "os"
    "time"

    mqtt "github.com/eclipse/paho.mqtt.golang"
)

func main() {
    broker := os.Getenv("MQTT_BROKER")
    if broker == "" { broker = "tcp://127.0.0.1:1883" }
    opts := mqtt.NewClientOptions().AddBroker(broker).SetClientID("fleet-mock-publisher")
    c := mqtt.NewClient(opts)
    if token := c.Connect(); token.Wait() && token.Error() != nil { panic(token.Error()) }

    vehicle := "B1234XYZ"
    if len(os.Args) > 1 { vehicle = os.Args[1] }
    topic := fmt.Sprintf("/fleet/vehicle/%s/location", vehicle)

    rand.Seed(time.Now().Unix())
    for {
        msg := map[string]interface{}{
            "vehicle_id": vehicle,
            "latitude":   -6.2088 + (rand.Float64()-0.5)/100.0,
            "longitude":  106.8456 + (rand.Float64()-0.5)/100.0,
            "timestamp":  time.Now().Unix(),
        }
        b, _ := json.Marshal(msg)
        c.Publish(topic, 0, false, b)
        fmt.Println("published", string(b))
        time.Sleep(2 * time.Second)
    }
}
