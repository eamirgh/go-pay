package payment

import (
	"context"

	"github.com/google/uuid"
)

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

type MetaData interface {
	Details() map[string]string
	Set(key, value string)
	Get(key string) string
	Has(key string) bool
}

type Invoice struct {
	ID            uuid.UUID
	Amount        uint64
	Currency      string
	TransactionID string
	Details       map[string]string
	Driver        *Driver
}

func (i *Invoice) Set(key, value string) {
	i.Details[key] = value
}
func (i *Invoice) Get(key string) string {
	return i.Details[key]
}
func (i *Invoice) Has(key string) bool {
	_, ok := i.Details[key]
	return ok
}

type Receipt struct {
	RefID   string
	Details map[string]string
}

type Driver interface {
	Purchase(ctx context.Context, i *Invoice) (*Invoice, error)
	Pay(i *Invoice) *PayResponse
	Verify(ctx context.Context, amount uint64, args map[string]string) (*Receipt, error)
}

type PayResponse struct {
	URL         string
	HasRedirect bool
}
