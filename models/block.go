package models

type Block struct {
	Number       string   `json:"number"`
	Transactions []string `json:"transactions"`
}
