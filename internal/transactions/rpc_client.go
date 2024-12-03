package transactions

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"

	"github.com/valentineejk/trust-go/models"
	"github.com/valentineejk/trust-go/utils"
)

type EthereumRPCClient struct {
	URL string
}

func NewEthereumRPCClient(url string) *EthereumRPCClient {
	return &EthereumRPCClient{URL: url}
}

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

func (client *EthereumRPCClient) getBlockByNumber(number string) (models.Block, error) {
	params := []interface{}{number, true}
	response, err := client.makeRequest("eth_getBlockByNumber", params)
	if err != nil {
		return models.Block{}, err
	}

	blockData := response["result"].(map[string]interface{})
	block := models.Block{
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

func (client *EthereumRPCClient) getTransactionByHash(hash string) (models.Transaction, error) {
	params := []interface{}{hash}
	response, err := client.makeRequest("eth_getTransactionByHash", params)
	if err != nil {
		return models.Transaction{}, fmt.Errorf("failed to fetch transaction by hash: %v", err)
	}

	result, ok := response["result"].(map[string]interface{})
	if !ok || result == nil {
		log.Printf("❌ Transaction not found for hash: %s. Response: %+v", hash, response)
		return models.Transaction{}, fmt.Errorf("transaction not found or invalid response: %v", response)
	}

	tx := models.Transaction{
		Hash:  utils.SafeGetString(result, "hash"),
		From:  utils.SafeGetString(result, "from"),
		To:    utils.SafeGetString(result, "to"),
		Value: utils.SafeGetString(result, "value"),
	}

	// dbg log.Printf("✅ Fetched Transaction: %+v", tx)
	return tx, nil
}
