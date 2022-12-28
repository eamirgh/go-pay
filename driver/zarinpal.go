package driver

import "github.com/eamirgh/go-pay/payment"

type Zarinpal struct {
}

func (z *Zarinpal) Amount(amount uint64)      {}
func (z *Zarinpal) Detail(map[string]string)  {}
func (z *Zarinpal) Purchase() string          { return "" }
func (z *Zarinpal) Pay() *payment.PayResponse { return &payment.PayResponse{} }
func (z *Zarinpal) Verify() *payment.Receipt  { return &payment.Receipt{} }
