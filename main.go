// I WIILL PUT ALL THE CODE IN THE MAIN DIRECTORY THIS IS NOT THE BEST PRACTICE,
// PART OF THE INTRUSCTIONS FOR THE TAKE HOME, IS TO KEEP IT SIMPLE

//FOR THE TX I USED POLLING, ON PROD WEBSOCKETS OR BATCH PROCESSING WILL BE IMPLEMENTED, TRYING TO KEEP EVERYTHING BASIC

package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strconv"
	"strings"
	"sync"
)

const (
	mainnnet = "https://ethereum-rpc.publicnode.com"
	testnet  = "https://ethereum-sepolia-rpc.publicnode.com"
)

// models:
type Transaction struct {
	Hash  string `json:"hash"`
	From  string `json:"from"`
	To    string `json:"to"`
	Value string `json:"value"`
}

type Block struct {
	Number       string   `json:"number"`
	Transactions []string `json:"transactions"`
}

type MemoryStorage struct {
	mu           sync.Mutex
	currentBlock int
	subscribed   map[string]struct{}
	transactions map[string][]Transaction
}

// Memory:
func NewMemoryStorage() *MemoryStorage {
	return &MemoryStorage{
		currentBlock: 0,
		subscribed:   make(map[string]struct{}),
		transactions: make(map[string][]Transaction),
	}
}

// eth client:
type EthereumRPCClient struct {
	URL string
}

func NewEthereumRPCClient(url string) *EthereumRPCClient {
	return &EthereumRPCClient{URL: url}
}

// /////////////////// api calll ////////////////////
func (client *EthereumRPCClient) makeRequest(method string, params []interface{}) (map[string]interface{}, error) {

	payload := map[string]interface{}{
		"jsonrpc": "2.0",
		"id":      1,
		"method":  method,
		"params":  params,
	}

	data, _ := json.Marshal(payload)
	resp, err := http.Post(client.URL, "application/json", bytes.NewBuffer(data))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	var result map[string]interface{}
	json.Unmarshal(body, &result)

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("HTTP error: %s", resp.Status)
	}

	// Check for JSON-RPC errors in the response
	if errMsg, ok := result["error"].(map[string]interface{}); ok {
		return nil, fmt.Errorf("JSON-RPC error: code %v, message: %s", errMsg["code"], errMsg["message"])
	}

	return result, nil
}

// //////////////////////////////////////////////////
func (client *EthereumRPCClient) getBlockByNumber(number string) (Block, error) {
	params := []interface{}{number, true}
	response, err := client.makeRequest("eth_getBlockByNumber", params)
	if err != nil {
		return Block{}, err
	}

	blockData := response["result"].(map[string]interface{})
	block := Block{
		Number:       blockData["number"].(string),
		Transactions: []string{},
	}

	if txns, ok := blockData["transactions"].([]interface{}); ok {
		for _, txn := range txns {
			txHash := txn.(map[string]interface{})["hash"].(string)
			block.Transactions = append(block.Transactions, txHash)
		}
	}

	return block, nil
}

// Helper function to safely extract string values
func safeGetString(data map[string]interface{}, key string) string {
	if value, ok := data[key]; ok && value != nil {
		return value.(string)
	}
	return ""
}

// ///////////////fetches transactions with hash///////////////////////////
func (client *EthereumRPCClient) getTransactionByHash(hash string) (Transaction, error) {
	params := []interface{}{hash}
	response, err := client.makeRequest("eth_getTransactionByHash", params)
	if err != nil {
		return Transaction{}, fmt.Errorf("failed to fetch transaction by hash: %v", err)
	}

	result, ok := response["result"].(map[string]interface{})
	if !ok || result == nil {
		log.Printf("❌ Transaction not found for hash: %s. Response: %+v", hash, response)
		return Transaction{}, fmt.Errorf("transaction not found or invalid response: %v", response)
	}

	tx := Transaction{
		Hash:  safeGetString(result, "hash"),
		From:  safeGetString(result, "from"),
		To:    safeGetString(result, "to"),
		Value: safeGetString(result, "value"),
	}

	// log.Printf("✅ Fetched Transaction: %+v", tx)
	return tx, nil
}

type Parser interface {
	GetCurrentBlock() int
	Subscribe(address string) bool
	GetTransactions(address string) []Transaction
	HookNotificationService(notificationService NotificationService)
}

// /////////////// notifications///////////////////////
type NotificationService interface {
	SendNotification(address string, txn Transaction)
}

type SimpleNotificationService struct{}

func (s *SimpleNotificationService) SendNotification(address string, txn Transaction) {
	log.Printf("📩 Notification: Address %s involved in transaction %s (Value: %s)", address, txn.Hash, txn.Value)
}

/////////////////////////////////////////////

type EthereumParser struct {
	client            *EthereumRPCClient
	storage           *MemoryStorage
	notificationHooks []NotificationService
}

func NewEthereumParser(client *EthereumRPCClient, storage *MemoryStorage) *EthereumParser {
	return &EthereumParser{
		client:            client,
		storage:           storage,
		notificationHooks: []NotificationService{},
	}
}

// ///////////////// Get the current block ////////////////////////
func (parser *EthereumParser) GetCurrentBlock() int {
	parser.storage.mu.Lock()
	defer parser.storage.mu.Unlock()
	return parser.storage.currentBlock
}

func (parser *EthereumParser) Subscribe(address string) bool {

	//SKIPED ADRRESS VALIDATION

	parser.storage.mu.Lock()
	defer parser.storage.mu.Unlock()

	normalizedAddress := strings.ToLower(address)
	if _, exists := parser.storage.subscribed[normalizedAddress]; exists {
		return false
	}
	parser.storage.subscribed[normalizedAddress] = struct{}{}
	log.Printf("✅ Subscribed to address: %s", normalizedAddress)
	log.Printf("📜 Current Subscribed Addresses: %+v", parser.storage.subscribed)
	return true
}

func (parser *EthereumParser) GetTransactions(address string) []Transaction {
	parser.storage.mu.Lock()
	defer parser.storage.mu.Unlock()

	// Log the entire transaction store for debugging
	log.Printf("📦 Full Transaction Store: %+v", parser.storage.transactions)

	// Log specific transactions for the requested address
	if txns, exists := parser.storage.transactions[address]; exists {
		log.Printf("🔍 Transactions for address %s: %+v", address, txns)
		return txns
	}

	log.Printf("❌ No transactions found for address %s", address)
	return nil
}

func (parser *EthereumParser) NotificationService(notificationService NotificationService) {
	parser.notificationHooks = append(parser.notificationHooks, notificationService)
}

func (parser *EthereumParser) processTransactions(txHashes []string) {
	log.Printf("🔍 Processing %d transactions", len(txHashes)) // Logging

	for _, hash := range txHashes {
		tx, err := parser.client.getTransactionByHash(hash)
		if err != nil {
			log.Printf("❌ Error fetching transaction %s: %v", hash, err)
			continue
		}

		parser.storage.mu.Lock()
		normalizedTo := strings.ToLower(tx.To)
		if _, subscribed := parser.storage.subscribed[normalizedTo]; subscribed {
			log.Printf("📝 Transaction matched for address %s: %+v", tx.To, tx)
			parser.storage.transactions[normalizedTo] = append(parser.storage.transactions[normalizedTo], tx)

			for _, hook := range parser.notificationHooks {
				hook.SendNotification(normalizedTo, tx)
			}
		}

		// Check if the `from` address is subscribed
		if _, subscribed := parser.storage.subscribed[tx.From]; subscribed {
			parser.storage.transactions[tx.From] = append(parser.storage.transactions[tx.From], tx)

			// we can now plus the notifications here, i just did a simple one that logs on console
			for _, hook := range parser.notificationHooks {
				hook.SendNotification(tx.From, tx)
			}
		}

		parser.storage.mu.Unlock()
	}
}

func (parser *EthereumParser) monitorBlocks(startBlock int) {

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
			parser.processTransactions(block.Transactions)

			parser.storage.mu.Lock()
			parser.storage.currentBlock = int(blockNumber)
			parser.storage.mu.Unlock()
			currentBlock = int(blockNumber)
		} else {
			log.Printf("⏳ No new blocks to process. Current block: %d, Latest block: %d", currentBlock, blockNumber)
		}
	}
}

// ///////////endpoints////////////////////////
var globalParser *EthereumParser

func getCurrentBlockHandler(w http.ResponseWriter, r *http.Request) {
	block := globalParser.GetCurrentBlock()
	response := map[string]int{"current_block": block}
	json.NewEncoder(w).Encode(response)
}

func subscribeHandler(w http.ResponseWriter, r *http.Request) {
	var payload struct {
		Address string `json:"address"`
	}
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	if payload.Address == "" {
		http.Error(w, "Address is required", http.StatusBadRequest)
		return
	}

	subscribed := globalParser.Subscribe(payload.Address)
	response := map[string]bool{"subscribed": subscribed}
	json.NewEncoder(w).Encode(response)
}

func getTransactionsHandler(w http.ResponseWriter, r *http.Request) {
	address := r.URL.Query().Get("address")
	if address == "" {
		http.Error(w, "Address is required", http.StatusBadRequest)
		return
	}
	address = strings.ToLower(address)

	transactions := globalParser.GetTransactions(address)
	json.NewEncoder(w).Encode(transactions)
}

// ///////////////////////////////////////////////

func main() {

	client := NewEthereumRPCClient(testnet)
	storage := NewMemoryStorage()
	globalParser = NewEthereumParser(client, storage)

	notificationService := &SimpleNotificationService{}
	globalParser.NotificationService(notificationService)

	// Start monitoring
	go globalParser.monitorBlocks(0)

	// HTTP server setup
	http.HandleFunc("/current-block", getCurrentBlockHandler)
	http.HandleFunc("/subscribe", subscribeHandler)
	http.HandleFunc("/transactions", getTransactionsHandler)

	log.Println("Trust 🛡️ Server started on port 8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
