package notifications

import (
	"log"

	"github.com/valentineejk/trust-go/models"
)

type NotificationService interface {
	SendNotification(address string, txn models.Transaction)
}

type SimpleNotificationService struct{}

func (s *SimpleNotificationService) SendNotification(address string, txn models.Transaction) {
	log.Printf("📩 Notification: Address %s involved in transaction %s (Value: %s)", address, txn.Hash, txn.Value)
}
