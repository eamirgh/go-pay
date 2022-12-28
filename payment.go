package main

import "github.com/google/uuid"

type Payment interface {
	Driver() string
	CallbackURL(url string) string
	Amount(amount uint64)
	Invoice() *Invoice
	Detail(details map[string]string)
	TransactionID(id string)
	Via(driver string)
	Purchase(i Invoice)
	Verify() *Receipt
}

type Invoice struct {
	ID            uuid.UUID
	Amount        uint64
	Currency      string
	TransactionID string
	Driver        string
}

type Receipt struct {
	Details map[string]string
}

type Driver interface {
	Amount(amount uint64)     //sets Amount
	Detail(map[string]string) //sets Details
	Purchase() string
	Pay() *PayResponse
	Verify() *Receipt
}

type PayResponse struct {
	URL         string
	HasRedirect bool
}
