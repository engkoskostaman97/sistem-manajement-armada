package rabbitmq

import (
    "encoding/json"
    "errors"

    "github.com/streadway/amqp"
    "transjakarta-fleet/internal/models"
)

type PublisherInterface interface {
	PublishGeofenceEvent(models.GeofenceEvent) error
	Close()
	EnsureQueue(string) error
}

type Publisher struct {
    conn     *amqp.Connection
    ch       *amqp.Channel
    exchange string
}

func NewPublisher(url string) (*Publisher, error) {
    c, err := amqp.Dial(url)
    if err != nil { return nil, err }
    ch, err := c.Channel()
    if err != nil { c.Close(); return nil, err }
    ex := "fleet.events"
    if err := ch.ExchangeDeclare(ex, "fanout", true, false, false, false, nil); err != nil { ch.Close(); c.Close(); return nil, err }
    return &Publisher{conn: c, ch: ch, exchange: ex}, nil
}

func (p *Publisher) Close() {
    if p.ch != nil { p.ch.Close() }
    if p.conn != nil { p.conn.Close() }
}

func (p *Publisher) PublishGeofenceEvent(e models.GeofenceEvent) error {
    b, err := json.Marshal(map[string]interface{}{
        "vehicle_id": e.VehicleID,
        "event":      e.Event,
        "location": map[string]float64{"latitude": e.Latitude, "longitude": e.Longitude},
        "timestamp":  e.Timestamp,
    })
    if err != nil { return err }
    return p.ch.Publish(p.exchange, "", false, false, amqp.Publishing{ContentType: "application/json", Body: b})
}

func (p *Publisher) EnsureQueue(queueName string) error {
    if p.ch == nil { return errors.New("channel closed") }
    _, err := p.ch.QueueDeclare(queueName, true, false, false, false, nil)
    if err != nil { return err }
    return p.ch.QueueBind(queueName, "", p.exchange, false, nil)
}
