package models

import "sync"

type MemoryStorage struct {
	Mu           sync.Mutex
	CurrentBlock int
	Subscribed   map[string]struct{}
	Transactions map[string][]Transaction
}
