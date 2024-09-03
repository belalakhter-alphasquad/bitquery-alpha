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
	serverURL = "wss://streaming.bitquery.io/eap?token=ory_at_e_wAazl2aa4UXhgSxe_1Nsk1-nAB2HUnQMDBT43n4RI.LvBbAqVy4_vWSxf6OGQrXqe4PeHcGjx32l37k7JqX3I"
	ohlcQuery = `query {
		Solana {
		  DEXTradeByTokens(
			limit: {count:1000}
			orderBy: { descendingByField: "Block_Timefield" }
			where: {
			  Trade: {
				Currency: { mintAddress: { is: "EsY9oWzqj94ZiqEEWfwNzVVixKis4zZHfKddQrub1YpT" } }
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

	var addresses = []string{
		"7hfftzJ5QVUwsZxhUHyi8SfPUrXDinGkU7wujqN6pump", "BhHDowogMZ5YvGVvnkBYJVmwdnHtMjUMBAgaxk8fpump", "AEBnyoEZQRaat6RR3N2LVwdFqN9zWtgeSkMevgUfpump", "AakLXoLVQNCTKAWyczcxYDkHozpvhvNEmHsQLKFzpump", "DWx8pAjXt4T9LAJCJyPY1dRnPe7hSpGi8Qsg2Pfkpump",
		"993CrgJj5iV2vnE14B5qKzsKfCDpwXLdcNjpBafVpump", "EXCKNw3tUWpQ3Df8wN86CU3K4EeKh9WZgE1xBodApump", "7pYDzBtv7xr3QiF4cXEuF6any5dxzB5opMKBMiEgpump", "3giJ6SEP22ocpNxUt58Ry3cUb5MKUhntg6mtCULfpump", "7pYDzBtv7xr3QiF4cXEuF6any5dxzB5opMKBMiEgpump",
		"HzhGr945HZhyvcYCmzXFYHd8S5s4tJgv1Ask2su4pump", "BC76VfY4hNiNZiCYaPdoC9iRqfYCCZPqfTAEJHWLpump", "3H8ekRwUpgTijjNDU77qxtCrgSt24d1s1pByNYo2pump", "D8JgEYSNA8593S5zKdxky64HaUdj2rVyCaPsR1akpump", "4hbpRHxo4WoxxoVjayC71i7LjuLMmb9v3ckuFMDMpump",
		"4LQyfbSoawZLemacQNDktMpCaSY8i5oxLXKhvKvNpump", "97NRmqq1iLLnr3gKbu5p1qYdEygapPhhQG1PRJ4Tpump", "34injJbUYZthAnfeJ36KQokJgD6riCEPTQbNxKHNpump", "8bSTr3Rn185gUNxMbeEzyoccQQvpZqXFgVPQRuCFpump", "bjpCbRDQvPjTK8oukaGXBWTWNfYvYy9z5p1GzFapump",
		"AHEfpyu51jkt37LL1nb4Yiva9sgAqK4JA7veKcBGpump", "ERMagsFMBiB7XYJ2qY74kepSfpTd8vdiKrHQsGPgpump", "AHEfpyu51jkt37LL1nb4Yiva9sgAqK4JA7veKcBGpump", "H2Nj8ijdaRVjgkZQCj6XRofxsoCivXNK7e8jHR9ypump", "34kvJZAvad3b9VD6YCSxDetPnq6RbtLGkt48L8S4pump",
		"HDJRGG2Y8fzQHdaS71QhzeKD5UqZ8CCbTX7PYo9ppump", "DnmQRUBBg89z3QqGJqGi7J4jpb9n1nEDb6mZGJ8Gpump", "42zd8aPC79dFxSNkRDLoaXnKV68xdxxeokoCyVF9pump", "6bDYTWZbxCHZgRLyEELphJv55tHJXRHmyHUEbpXXpump", "FibAyTch6pqDUk5LQmW3PUvExqGHzLQBrRSb8rHdpump",
		"8HWFVn4DnqdZ14AG4AsDLwUwSPbbUCy9sqokSXmapump", "2bg7i2nueRUJ1KLSiFa2ewJ5By1Q39MH9vfn1Tvtpump", "EXLAATFGbUBPkAVPnLcF64vJhZc5ikkbKCRGXuVvpump", "ECs4ZLGaRte5iPfkaYznBEUviMC539aXzcF4SUmdpump", "sQB8ng2BgciryiHNvwUyFZVgLYdYM1XnQsnKg96pump",
		"51dM85YyP6LNo1cei7deBJY4uTq2h3gS2esz3woZpump", "X1CCaXbhNwb53qoShVoSoWxCJQWR3fxon96K2Rrpump", "CX3EZnnVnw7PE7mCyuVP5YaCLsgLFskw8PYyWfFzpump", "7ChcuVcY1yUkHqXNnj6vwYhKXFEmTK97vpyquS6Zpump", "9xEBdrwkYJhJfsRZH8yiTszWAk3Y3Jc3shbFoy1zpump",
		"8PN525m8UDdmYpVhh3D1A92VGCkx2xCM6VyFPK5npump", "F5K2P31i8ta3dJSvJe4wZF5tbdvYNTtp7ZENWTk2pump", "Desz5fga9AbBCs8weD2pdSFXiAgdCFNsuMoqrj8dpump", "B61w7iLAnatA5rZkK18niohTwTda3bhzY51wF5zkpump", "HP2kVY9Fz4zR41sHQT4XZ4udpWXLCQNbSnouAZiMvq5y",
		"sQB8ng2BgciryiHNvwUyFZVgLYdYM1XnQsnKg96pump", "DPxY4KD723wsLwsiAGgjznXCMxDfUbjruxXeFLsspump", "CDCAVw72kXzW9X57sn9Knmh1ttX7uZPahrdhSePqpump", "Gr4t1Aw1qUFrWLj2wgKrYvQ66dT7m2iZ6EWnTi4pump",
	}

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
