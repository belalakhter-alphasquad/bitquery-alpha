package main

import (
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:   1024,
	WriteBufferSize:  1024,
	HandshakeTimeout: time.Hour,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

var (
	eventChan = make(chan interface{}, 100)
)

const (
	serverURL = "wss://streaming.bitquery.io/eap?token=ory_at_t9d0ZJVrgBD0TYAr91Meb8kfiAnPpAHX87iE4YWXw6I.ayTNpiPmRIvGX6G0Pfsu_XT3jX2F4GYlAOhfNh1Q_F4"
	ohlcQuery = `subscription {
		Solana {
		  DEXTradeByTokens(
			orderBy: {descendingByField: "Block_Timefield"}
			where: {Trade: {Currency: {MintAddress: {is: $Address}}, Side: {Currency: {MintAddress: {is: "So11111111111111111111111111111111111111112"}}}, Dex: {ProgramAddress: {is: "675kPX9MHTjS2zt1qfr1NYHuzeLXfQM9H24wFSUt1Mp8"}}}}
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
	balanceTransferQuery = `subscription {
		Solana {
		  Transfers(
			where: {Transfer: {Sender: {Address: {is: $Address}}}}
		  ) {
			Transaction {
			  Signature
			}
			Transfer {
			  Amount
			  AmountInUSD
			  Sender {
				Address
			  }
			  Receiver {
				Address
			  }
			  Currency {
				Name
				Symbol
				MintAddress
			  }
			}
		  }
		}
	  }
	  `
	splTransferQuery = `subscription {
		Solana {
		  Transfers(
			where: {Transfer: {Currency: {MintAddress: {is: $Address}}}}
		  ) {
			Transfer {
			  Currency {
				MintAddress
				Symbol
				Name
				Fungible
				Native
			  }
			  Receiver {
				Address
			  }
			  Sender {
				Address
			  }
			  Amount
			  AmountInUSD
			}
			Transaction{
			  Signature
			}
		  }
		}
	  }
	  `
)

func substituteAddress(query string, address string) string {
	quotedAddress := "\"" + address + "\""
	return strings.ReplaceAll(query, "$Address", quotedAddress)
}

func handleRequest(w http.ResponseWriter, r *http.Request) {
	queryParam := r.URL.Query().Get("query")
	Address := r.URL.Query().Get("address")

	var query string
	switch queryParam {
	case "ohlc":
		query = ohlcQuery
	case "wallettransfer":
		query = balanceTransferQuery
	case "spltransfer":
		query = splTransferQuery
	default:
		http.Error(w, "Invalid query parameter", http.StatusBadRequest)
		return
	}

	if Address != "" {
		query = substituteAddress(query, Address)
	}

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("Failed to upgrade to WebSocket: %v", err)
		return
	}
	conn.SetReadDeadline(time.Now().Add(60000 * time.Second))
	conn.SetWriteDeadline(time.Now().Add(60000 * time.Second))

	go func() {
		Connect(serverURL, query)

	}()

	fmt.Fprintf(w, "Subscribed to %s\n", queryParam)

	go func() {
		for eventMsg := range eventChan {
			err := conn.WriteJSON(eventMsg)
			if err != nil {
				log.Printf("Error sending event message: %v", err)
				break
			}
			fmt.Printf("Received message: %s\n", eventMsg)

		}
		conn.Close()
	}()
}

func main() {

	http.HandleFunc("/ws", handleRequest)
	port := "8080"
	fmt.Printf("Server listening on port %s...\n", port)
	log.Fatal(http.ListenAndServe(":"+port, nil))

}
