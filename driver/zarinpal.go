package driver

import (
	"errors"

	"github.com/eamirgh/go-pay/payment"
)

type Zarinpal struct {
	MerchantID string
	Mode       string
}

const ZARINPAL_SANDBOX = "sandbox"
const ZARINPAL_NORMAL = "normal"
const ZARINPAL_GATEWAY = "gateway"

// NewZarinpalGateway creates new Zarinpal gateway with given credentials
func NewZarinpalGateway(merchantID, mode string) (*Zarinpal, error) {
	if mode != ZARINPAL_GATEWAY && mode != ZARINPAL_NORMAL && mode != ZARINPAL_SANDBOX {
		return nil, errors.New("invalid mode for Zarinpal driver")
	}
	return &Zarinpal{
		MerchantID: merchantID,
		Mode:       mode,
	}, nil
}

func (z *Zarinpal) Pay(i *payment.Invoice) *payment.PayResponse { return &payment.PayResponse{} }
func (z *Zarinpal) Verify() *payment.Receipt                    { return &payment.Receipt{} }
