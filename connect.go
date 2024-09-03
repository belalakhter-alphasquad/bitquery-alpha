package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/gorilla/websocket"
)

func Connect(url string, query string) {
	headers := http.Header{}
	headers.Set("Sec-WebSocket-Protocol", "graphql-ws")
	headers.Set("Content-Type", "application/json")

	conn, _, err := websocket.DefaultDialer.Dial(url, headers)
	if err != nil {
		log.Fatal("Failed to connect to WebSocket server:", err)
	}
	defer conn.Close()

	fmt.Println("Connected to WebSocket server")

	initPayload := map[string]interface{}{
		"type": "connection_init",
	}
	err = conn.WriteJSON(initPayload)
	if err != nil {
		log.Fatal("Failed to send connection init:", err)
	}

	_, ackMessage, err := conn.ReadMessage()
	if err != nil {
		log.Fatal("Failed to read connection acknowledgment:", err)
	}
	fmt.Println("Connection Acknowledged:", string(ackMessage))

	subscriptionPayload := map[string]interface{}{
		"type":    "start",
		"id":      "1",
		"payload": map[string]string{"query": query},
	}
	err = conn.WriteJSON(subscriptionPayload)
	if err != nil {
		log.Fatal("Failed to send subscription query:", err)
	}

	for {
		_, message, err := conn.ReadMessage()
		if err != nil {
			if websocket.IsCloseError(err, websocket.CloseNormalClosure) {
				fmt.Println("WebSocket connection closed normally")
				break
			}
			log.Fatal("Failed to read message:", err)
		}

		var subData SubscriptionData
		if err := json.Unmarshal(message, &subData); err != nil {
			log.Fatal("Failed to unmarshal message:", err)
		}
		eventChan <- subData

	}
}
