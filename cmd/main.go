// I WIILL PUT ALL THE CODE IN THE MAIN DIRECTORY THIS IS NOT THE BEST PRACTICE,
// PART OF THE INTRUSCTIONS FOR THE TAKE HOME, IS TO KEEP IT SIMPLE

//FOR THE TX I USED POLLING, ON PROD WEBSOCKETS OR BATCH PROCESSING WILL BE IMPLEMENTED, TRYING TO KEEP EVERYTHING BASIC

package main

import (
	"log"
	"net/http"

	"github.com/valentineejk/trust-go/internal/api"
	"github.com/valentineejk/trust-go/internal/notifications"
	"github.com/valentineejk/trust-go/internal/storage"
	"github.com/valentineejk/trust-go/internal/transactions"
	"github.com/valentineejk/trust-go/utils"
)

func main() {

	client := transactions.NewEthereumRPCClient(utils.Testnet)
	storage := storage.NewMemoryStorage()
	parser := transactions.NewEthereumParser(client, storage)
	handler := api.NewHandler(parser)
	notificationService := &notifications.SimpleNotificationService{}
	parser.NotificationService(notificationService)

	// Start monitoring
	go parser.MonitorBlocks(0)

	// HTTP server setup
	http.HandleFunc("/current-block", handler.GetCurrentBlockHandler)
	http.HandleFunc("/subscribe", handler.SubscribeHandler)
	http.HandleFunc("/transactions", handler.GetTransactionsHandler)

	log.Println("************************************")
	log.Println("Trust 🛡️ Server started on port 8080")
	log.Println("************************************")

	log.Fatal(http.ListenAndServe(":8080", nil))

}
