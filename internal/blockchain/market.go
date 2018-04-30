package blockchain

import (
	"fmt"
	"github.com/ethereum/go-ethereum/common"
	"github.com/nfeld9807/rest-api/internal/blockchain/generated"
	"log"
)

// ConnectMarket - Connect and return configured market
func ConnectMarket() *generated.Market {

	conn := ConnectClient()

	market, err := generated.NewMarket(common.HexToAddress("0xc4dfb5c9e861eeae844795cfb8d30b77b78bbc38"), conn)

	if err != nil {
		log.Fatalf("Failed to instantiate a Market contract: %v", err)
	}

	return market
}

// MarketPools - List all available market pools
func MarketPools() ([]common.Address, error) {
	market := ConnectMarket()

	pools, err := market.GetAllPools(nil)
	if err != nil {
		return nil, err
	}

	return pools, nil
}

//MarketCreatePool - Create new pool
func MarketCreatePool(publicKey string) (common.Hash, error) {
	market := ConnectMarket()
	auth := GetAuth("password")

	transaction, err := market.CreatePool(auth, publicKey)
	if err != nil {
		log.Fatalf("Failed to request token transfer: %v", err)
	}

	fmt.Printf("Transfer pending: 0x%x\n", transaction.Hash())

	return transaction.Hash(), nil
}
