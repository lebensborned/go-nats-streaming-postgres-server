package streaming

import (
	"encoding/json"
	"fmt"
	"log"

	"github.com/lebensborned/go-nats-streaming-postgres-server/internal/db"
	"github.com/nats-io/stan.go"
)

type StreamingHandler struct {
	stan.Conn
}

// InitStreaming inits a connection to NATS-streaming server
func InitStreaming(client, clusterID, url string) (*StreamingHandler, error) {
	sc, err := stan.Connect(clusterID, client, stan.NatsURL(url),
		stan.Pings(5, 3),
		stan.SetConnectionLostHandler(func(_ stan.Conn, reason error) {
			log.Println("Connection lost, reason: ", reason)
		}))
	if err != nil {
		return nil, fmt.Errorf("cant connect: %v", err)
	}

	log.Println("Connected to streaming")

	scHandler := &StreamingHandler{sc}
	return scHandler, nil
}

// Start init subscribe to queue from channel in nats-streaming
func (s *StreamingHandler) Start(subject, qgroup, durable string, repo db.Repository) error {
	mcb := func(msg *stan.Msg) {
		if err := msg.Ack(); err != nil {
			log.Printf("Subscription error: %v", err)
		}
		var order db.Order
		err := json.Unmarshal(msg.Data, &order)
		if err != nil {
			log.Printf("Invalid data received: %v", err)
			return
		}
		err = repo.Insert(&order)
		if err != nil {
			log.Printf("Error while insert in repo: %v", err)
			return
		}
		log.Println("Got order from nats-streaming: ", order.OrderUid)
	}

	_, err := s.QueueSubscribe(subject,
		qgroup, mcb,
		stan.SetManualAckMode())
	if err != nil {
		return fmt.Errorf("cant subscribe: %v", err)
	}
	return nil
}

// Close connection to nats-streaming
func (s *StreamingHandler) Finish() error {
	err := s.Close()
	if err != nil {
		return fmt.Errorf("cant finish streaming: %v", err)
	}
	log.Println("Finishing the streaming...")
	return nil
}
