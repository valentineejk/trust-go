package transactions

import (
	"log"
	"strconv"
	"strings"

	"github.com/valentineejk/trust-go/internal/notifications"
	"github.com/valentineejk/trust-go/models"
	"github.com/valentineejk/trust-go/utils"
)

type EthereumParser struct {
	client            *EthereumRPCClient
	storage           *models.MemoryStorage
	notificationHooks []notifications.NotificationService
}

func (parser *EthereumParser) HookNotificationService(notificationService notifications.NotificationService) {
	panic("unimplemented")
}

func NewEthereumParser(client *EthereumRPCClient, storage *models.MemoryStorage) *EthereumParser {
	return &EthereumParser{
		client:            client,
		storage:           storage,
		notificationHooks: []notifications.NotificationService{},
	}
}

func (parser *EthereumParser) MonitorBlocks(startBlock int) {

	currentBlock := startBlock

	for {
		log.Printf("🔄 Checking latest block. Current block being processed: %d", currentBlock)

		block, err := parser.client.getBlockByNumber("latest")
		if err != nil {
			log.Println("❌ Error fetching latest block:", err)
			continue
		}

		blockNumber, err := strconv.ParseInt(block.Number[2:], 16, 64) // Convert hex to int
		if err != nil {
			log.Printf("❌ Error parsing block number %s: %v", block.Number, err)
			continue
		}

		if int(blockNumber) > currentBlock {
			log.Printf("📦 Processing new block: %d", blockNumber)
			parser.ProcessTransactions(block.Transactions)

			parser.storage.Mu.Lock()
			parser.storage.CurrentBlock = int(blockNumber)
			parser.storage.Mu.Unlock()
			currentBlock = int(blockNumber)
		} else {
			log.Printf("⏳ No new blocks to process. Current block: %d, Latest block: %d", currentBlock, blockNumber)
		}
	}
}

func (parser *EthereumParser) Subscribe(address string) bool {

	//SKIPED ADRRESS VALIDATION *old
	//ADDED ADDR VALIDATION
	if !utils.IsValidEthereumAddress(address) {
		log.Printf("❌ Invalid Ethereum address: %s", address)
		return false
	}

	parser.storage.Mu.Lock()
	defer parser.storage.Mu.Unlock()

	normalizedAddress := strings.ToLower(address)
	if _, exists := parser.storage.Subscribed[normalizedAddress]; exists {
		return false
	}
	parser.storage.Subscribed[normalizedAddress] = struct{}{}
	log.Printf("✅ Subscribed to address: %s", normalizedAddress)
	log.Printf("📜 Current Subscribed Addresses: %+v", parser.storage.Subscribed)
	return true
}

func (parser *EthereumParser) GetCurrentBlock() int {
	parser.storage.Mu.Lock()
	defer parser.storage.Mu.Unlock()
	return parser.storage.CurrentBlock
}

func (parser *EthereumParser) GetTransactions(address string) []models.Transaction {
	parser.storage.Mu.Lock()
	defer parser.storage.Mu.Unlock()

	// Log the entire transaction store for debugging
	log.Printf("📦 Full Transaction Store: %+v", parser.storage.Transactions)

	// Log specific transactions for the requested address
	if txns, exists := parser.storage.Transactions[address]; exists {
		log.Printf("🔍 Transactions for address %s: %+v", address, txns)
		return txns
	}

	log.Printf("❌ No transactions found for address %s", address)
	return nil
}

func (parser *EthereumParser) ProcessTransactions(txHashes []string) {

	log.Printf("🔍 Processing %d transactions", len(txHashes)) // Logging

	for _, hash := range txHashes {
		tx, err := parser.client.getTransactionByHash(hash)
		if err != nil {
			log.Printf("❌ Error fetching transaction %s: %v", hash, err)
			continue
		}

		parser.storage.Mu.Lock()
		normalizedTo := strings.ToLower(tx.To)
		normalizedFrom := strings.ToLower(tx.From)

		// Check if the `to` address is subscribed and handle
		if _, subscribed := parser.storage.Subscribed[normalizedTo]; subscribed {
			log.Printf("📝 Transaction matched for address %s: %+v", tx.To, tx)
			parser.storage.Transactions[normalizedTo] = append(parser.storage.Transactions[normalizedTo], tx)
			log.Printf("DBG Transactions for %s: %+v", normalizedTo, parser.storage.Transactions[normalizedTo])

			for _, hook := range parser.notificationHooks {
				hook.SendNotification(normalizedTo, tx)
			}
		}

		// Check if the `from` address is subscribed and handle
		if _, subscribed := parser.storage.Subscribed[normalizedFrom]; subscribed {
			parser.storage.Transactions[normalizedFrom] = append(parser.storage.Transactions[normalizedFrom], tx)
			log.Printf("DBG Transactions for %s: %+v", normalizedFrom, parser.storage.Transactions[normalizedFrom])
			// we can now plus the notifications here, i just did a simple one that logs on console
			for _, hook := range parser.notificationHooks {
				hook.SendNotification(normalizedFrom, tx)
			}
		}

		parser.storage.Mu.Unlock()
	}
}

func (parser *EthereumParser) NotificationService(notificationService notifications.NotificationService) {
	parser.notificationHooks = append(parser.notificationHooks, notificationService)
}
