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

type Currency struct {
	MintAddress string `json:"MintAddress"`
}

type Trade struct {
	Currency Currency `json:"Currency"`
	Close    float64  `json:"close"`
	High     float64  `json:"high"`
	Low      float64  `json:"low"`
	Open     float64  `json:"open"`
}

type Block struct {
	Timefield string `json:"Timefield"`
}

type Event struct {
	Block  Block  `json:"Block"`
	Trade  Trade  `json:"Trade"`
	Count  string `json:"count"`
	Volume string `json:"volume"`
}

type DEXTradeByTokens struct {
	Block  Block  `json:"Block"`
	Trade  Trade  `json:"Trade"`
	Count  string `json:"count"`
	Volume string `json:"volume"`
}

type Data struct {
	Payload struct {
		Data struct {
			Solana struct {
				DEXTradeByTokens []DEXTradeByTokens `json:"DEXTradeByTokens"`
			} `json:"Solana"`
		} `json:"data"`
	} `json:"payload"`
	ID   string `json:"id"`
	Type string `json:"type"`
}

var eventChan = make(chan string)

const (
	serverURL = "wss://streaming.bitquery.io/eap?token=ory_at_e_wAazl2aa4UXhgSxe_1Nsk1-nAB2HUnQMDBT43n4RI.LvBbAqVy4_vWSxf6OGQrXqe4PeHcGjx32l37k7JqX3I"
	ohlcQuery = `query {
		Solana {
		  DEXTradeByTokens(
			limit: {count:1}
			orderBy: { descendingByField: "Block_Timefield" }
			where: {
			  Trade: {
				Currency: {MintAddress: {in: $Addresses}} 
				Dex: {
				  ProgramAddress: {
					is: "6EF8rrecthR5Dkzon8Nwu78hRvfCKubJ14M5uBEwF6P"
				  }
				}
			  }
			}
		  ) {
			Block {
			  Timefield: Time(interval: { in: minutes, count: 1 })
			}
			volume: sum(of: Trade_Amount)
			Trade {
			  high: PriceInUSD(maximum: Trade_Price)
			  low: PriceInUSD(minimum: Trade_Price)
			  open: PriceInUSD(minimum: Block_Slot)
			  close: PriceInUSD(maximum: Block_Slot)
			  Currency {
				MintAddress
			  }
		
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

	var addresses = []string{}

	addresses = append(addresses, address)

	quotedAddresses := make([]string, len(addresses))
	for i, addr := range addresses {
		quotedAddresses[i] = "\"" + addr + "\""
	}

	joinedAddresses := "[" + strings.Join(quotedAddresses, ", ") + "]"

	return strings.ReplaceAll(query, "$Addresses", joinedAddresses)
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
	fmt.Printf("Query is ----> %v\n", query)
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
