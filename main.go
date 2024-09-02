package main

const (
	serverURL = "wss://streaming.bitquery.io/eap?token=ory_at_aPzjXLYjGhpsW16NmIiBsspdDTo7XxWZ0UdHeuFTeCE.aN7TkdMNL87ZzqG0YG6CsBFvJXyZJlWb8Qm_mBWxtUQ"
	ohlcQuery = `subscription {
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
	balanceTransferQuery = `subscription {
		Solana {
		  Transfers(
			where: {Transfer: {Sender: {Address: {is: "2g9NLWUM6bPm9xq2FBsb3MT3F3G5HDraGqZQEVzcCWTc"}}}}
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
			where: {Transfer: {Currency: {MintAddress: {is: "3LAjGfLUSEomZdfgsEAN1Chb4ZrtoDLQPkBoWQDq7WkK"}}}}
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

func main() {

	Connect(serverURL, splTransferQuery)

}
