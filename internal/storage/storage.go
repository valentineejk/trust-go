package storage

import "github.com/valentineejk/trust-go/models"

// Memory:
func NewMemoryStorage() *models.MemoryStorage {
	return &models.MemoryStorage{
		CurrentBlock: 0,
		Subscribed:   make(map[string]struct{}),
		Transactions: make(map[string][]models.Transaction),
	}
}
