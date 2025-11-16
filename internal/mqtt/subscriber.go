package mqtt

import (
    "context"
    "encoding/json"
    "log"
    "math"
    "strconv"
    "time"

    mqtt "github.com/eclipse/paho.mqtt.golang"
    "transjakarta-fleet/internal/db"
    "transjakarta-fleet/internal/models"
    "transjakarta-fleet/internal/rabbitmq"
)

func deg2rad(d float64) float64 { return d * math.Pi / 180 }
func haversine(lat1, lon1, lat2, lon2 float64) float64 {
    R := 6371000.0 // meters
    dLat := deg2rad(lat2 - lat1)
    dLon := deg2rad(lon2 - lon1)
    a := math.Sin(dLat/2)*math.Sin(dLat/2) + math.Cos(deg2rad(lat1))*math.Cos(deg2rad(lat2))*math.Sin(dLon/2)*math.Sin(dLon/2)
    c := 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))
    return R * c
}

func ProcessLocationMessage(payload []byte, dbconn *db.DB, rmq rabbitmq.PublisherInterface, geoLat, geoLon, geoRad float64) {
	var msg models.LocationMessage
	if err := json.Unmarshal(payload, &msg); err != nil { log.Println("invalid payload", err); return }
	if msg.VehicleID == "" { log.Println("missing vehicle id"); return }
	if msg.Latitude < -90 || msg.Latitude > 90 { log.Println("invalid latitude"); return }
	if msg.Longitude < -180 || msg.Longitude > 180 { log.Println("invalid longitude"); return }
	if msg.Timestamp == 0 { msg.Timestamp = time.Now().Unix() }

	ctx := context.Background()
	if err := dbconn.InsertLocation(ctx, msg); err != nil { log.Println("db insert err", err) }

	d := haversine(geoLat, geoLon, msg.Latitude, msg.Longitude)
	if d <= geoRad {
		e := models.GeofenceEvent{VehicleID: msg.VehicleID, Event: "geofence_entry", Latitude: msg.Latitude, Longitude: msg.Longitude, Timestamp: msg.Timestamp}
		if err := rmq.PublishGeofenceEvent(e); err != nil { log.Println("rmq publish err", err) }
	}
}

func NewSubscriber(broker string, dbconn *db.DB, rmq rabbitmq.PublisherInterface, geoLatStr, geoLonStr, geoRadStr string) (mqtt.Client, error) {
	opts := mqtt.NewClientOptions().AddBroker(broker).SetClientID("fleet-backend-subscriber")
	client := mqtt.NewClient(opts)
	if token := client.Connect(); token.Wait() && token.Error() != nil { return nil, token.Error() }

	var geoLat, geoLon, geoRad float64
	if geoLatStr != "" {
		if lat, err := strconv.ParseFloat(geoLatStr, 64); err == nil {
			geoLat = lat
		} else {
			log.Printf("invalid GEOFENCE_LAT: %s, using 0", geoLatStr)
		}
	}
	if geoLonStr != "" {
		if lon, err := strconv.ParseFloat(geoLonStr, 64); err == nil {
			geoLon = lon
		} else {
			log.Printf("invalid GEOFENCE_LON: %s, using 0", geoLonStr)
		}
	}
	if geoRadStr != "" {
		if rad, err := strconv.ParseFloat(geoRadStr, 64); err == nil {
			geoRad = rad
		} else {
			log.Printf("invalid GEOFENCE_RADIUS_M: %s, using 0", geoRadStr)
		}
	}

	topic := "/fleet/vehicle/+/location"
	handler := func(c mqtt.Client, m mqtt.Message) {
		ProcessLocationMessage(m.Payload(), dbconn, rmq, geoLat, geoLon, geoRad)
	}

	if token := client.Subscribe(topic, 1, handler); token.Wait() && token.Error() != nil { return nil, token.Error() }
	return client, nil
}
