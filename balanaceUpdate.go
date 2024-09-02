package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/gorilla/websocket"
)

const (
	serverURL         = "wss://streaming.bitquery.io/eap?token=ory_at_aPzjXLYjGhpsW16NmIiBsspdDTo7XxWZ0UdHeuFTeCE.aN7TkdMNL87ZzqG0YG6CsBFvJXyZJlWb8Qm_mBWxtUQ"
	subscriptionQuery = `subscription {
		Solana {
		  DEXTradeByTokens(
			orderBy: {descendingByField: "Block_Timefield"}
			where: {Trade: {Currency: {MintAddress: {is: "6D7NaB2xsLd7cauWu1wKk6KBsJohJmP2qZH9GEfVi5Ui"}}, Side: {Currency: {MintAddress: {is: "So11111111111111111111111111111111111111112"}}}, Dex: {ProgramAddress: {is: "675kPX9MHTjS2zt1qfr1NYHuzeLXfQM9H24wFSUt1Mp8"}}}}
		  ) {
			Block {
			  Timefield: Time(interval: {in: minutes, count: 1})
			}
			volume: sum(of: Trade_Amount)
			Trade {
			  high: Price(maximum: Trade_Price)
			  low: Price(minimum: Trade_Price)
			  open: Price(minimum: Block_Slot)
			  close: Price(maximum: Block_Slot)
			}
			count
		  }
		}
	  }`
)

func main() {
	headers := http.Header{}
	headers.Set("Sec-WebSocket-Protocol", "graphql-ws")
	headers.Set("Content-Type", "application/json")

	conn, _, err := websocket.DefaultDialer.Dial(serverURL, headers)
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
		"payload": map[string]string{"query": subscriptionQuery},
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
		fmt.Printf("Received message: %s\n", message)
	}
}
