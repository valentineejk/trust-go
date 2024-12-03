package transactions

import (
	"github.com/valentineejk/trust-go/internal/notifications"
	"github.com/valentineejk/trust-go/models"
)

type Parser interface {
	GetCurrentBlock() int
	Subscribe(address string) bool
	GetTransactions(address string) []models.Transaction
	HookNotificationService(notificationService notifications.NotificationService)
}
