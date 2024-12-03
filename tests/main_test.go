package tests

import (
	"testing"

	"github.com/valentineejk/trust-go/internal/storage"
	"github.com/valentineejk/trust-go/internal/transactions"
	"github.com/valentineejk/trust-go/utils"
)

func Test_SubscribeValidAndDuplicateAddress(t *testing.T) {

	storage := storage.NewMemoryStorage()
	client := transactions.NewEthereumRPCClient(utils.Testnet)
	parser := transactions.NewEthereumParser(client, storage)

	validAddress := "0x1234567890abcdef1234567890abcdef12345678"
	if !parser.Subscribe(validAddress) {
		t.Errorf("Expected valid address to be subscribed, got false")
	}

	if parser.Subscribe(validAddress) {
		t.Errorf("Expected duplicate address to be rejected, got true")
	}
}

func Test_InvalidAddress(t *testing.T) {

	storage := storage.NewMemoryStorage()
	client := transactions.NewEthereumRPCClient(utils.Testnet)
	parser := transactions.NewEthereumParser(client, storage)

	invalidAddress := "123xyz"
	if parser.Subscribe(invalidAddress) {
		t.Errorf("Expected invalid address to be rejected, got true")
	}
}
