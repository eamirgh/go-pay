package driver

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"io/ioutil"
	"net/http"

	"github.com/eamirgh/go-pay/payment"
)

const ZARINPAL_SANDBOX = "sandbox"
const ZARINPAL_NORMAL = "normal"
const ZARINPAL_GATEWAY = "gateway"

var normalAPI = map[string]string{
	"apiPurchaseUrl":"https://api.zarinpal.com/pg/v4/payment/request.json",
	"apiPaymentUrl" : "https://www.zarinpal.com/pg/StartPay/",
	"apiVerificationUrl" : "https://api.zarinpal.com/pg/v4/payment/verify.json"
}
var sandboxAPI = map[string]string{
	"apiPurchaseUrl" : "https://sandbox.zarinpal.com/pg/services/WebGate/wsdl",
	"apiPaymentUrl" : "https://sandbox.zarinpal.com/pg/StartPay/",
	"apiVerificationUrl" : "https://sandbox.zarinpal.com/pg/services/WebGate/wsdl",
}
var zarinGateAPI = map[string]string{
	"apiPurchaseUrl" : "https://ir.zarinpal.com/pg/services/WebGate/wsdl",
	"apiPaymentUrl" : "https://www.zarinpal.com/pg/StartPay/:authority/ZarinGate",
	"apiVerificationUrl" : "https://ir.zarinpal.com/pg/services/WebGate/wsdl",
}

type Zarinpal struct {
	cfg *ZarinpalConfig
	endpoints map[string]string
}

type ZarinpalConfig struct {
	Mode       string
	MerchantID string
	Callback   string
	Description string
}

func NewZarinpalConfig(mode, merchantID, callback, description string) *ZarinpalConfig {
	return &ZarinpalConfig{
		Mode:       mode,
		MerchantID: merchantID,
		Callback:   callback,
		Description: description,
	}
}
// Gateway creates new Zarinpal gateway from the credentials in config
func (z *ZarinpalConfig) Gateway() (*Zarinpal, error) {
	if z.mode != ZARINPAL_GATEWAY && z.mode != ZARINPAL_NORMAL && z.mode != ZARINPAL_SANDBOX {
		return nil, errors.New("invalid mode for Zarinpal driver")
	}
	var endpoints map[string]string
	switch z.mode {
	case ZARINPAL_GATEWAY:
		endpoints = zarinGateAPI
	case ZARINPAL_NORMAL:
		endpoints = normalAPI
	case ZARINPAL_SANDBOX:
		endpoints = sandboxAPI
	}
	return &Zarinpal{
		z,
		endpoints,
	}, nil
}


type purchaseReq struct {
	MerchantID string `json:"merchant_id"`
	Amount     uint64 `json:"amount"`
	CallbackURL string `json:"callback_url"`
	Description string `json:"description"`
	Metadata   map[string]string `json:"metadata"`
}

func (r *purchaseReq) toJSON() ([]byte, error) {
	return json.Marshal(r)
}

func (z *Zarinpal) Purchase(ctx context.Context,i *payment.Invoice) (*payment.PayResponse, error) { 
	bs, err := (&purchaseReq{
		MerchantID: z.cfg.MerchantID,
		Amount:     i.Amount,
		CallbackURL: z.cfg.Callback,
		Description: z.cfg.Description,
		Metadata:   i.Details,
	}).toJSON()
	if err != nil {
		return nil, err
	}
	body := bytes.NewReader(bs)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, z.z.endpoints["apiPurchaseUrl"], body)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, errors.New("invalid status code")
	}
	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	var res struct {
		Status int `json:"status"`
		Authority string `json:"authority"`
		RefID string `json:"ref_id"`
		Errors struct {
			Code int `json:"code"`
			Message string `json:"message"`
		} `json:"errors"`
	}
	if err := json.Unmarshal(b, &res); err != nil {
		return nil, err
	}
	if res.Status != 100 {
		return nil, errors.New(res.Errors.Message)
	}
	return &payment.PayResponse{
		Authority: res.Authority,
		RefID: res.RefID,
	}, nil
	return &payment.PayResponse{}
}
func (z *Zarinpal) Pay(i *payment.Invoice) *payment.PayResponse { 
	return &payment.PayResponse{}
}
func (z *Zarinpal) Verify() *payment.Receipt { 
	return &payment.Receipt{}
}
