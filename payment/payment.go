package payment

import "github.com/google/uuid"

type Paymenter interface {
	Via(driver *Driver)
	Driver() string
	Callback(url string) string
	Amount(amount uint64)
	Detail(details map[string]string)
	TransactionID(id string)
	Invoice() *Invoice
	Purchase(i *Invoice)
	Verify() *Receipt
}

type Invoice struct {
	ID            uuid.UUID
	Amount        uint64
	Currency      string
	TransactionID string
	Details       map[string]string
	Driver        *Driver
}

type Receipt struct {
	Details map[string]string
}

type Driver interface {
	Pay(i *Invoice) *PayResponse
	Verify(r *Receipt) *Receipt
}

type PayResponse struct {
	URL         string
	HasRedirect bool
}
