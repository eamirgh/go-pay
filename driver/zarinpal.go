package driver

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"

	"github.com/eamirgh/go-pay/payment"
)

const ZARINPAL_SANDBOX = "sandbox"
const ZARINPAL_NORMAL = "normal"
const ZARINPAL_GATEWAY = "gateway"

var normalAPI = map[string]string{
	"apiPurchaseUrl":     "https://api.zarinpal.com/pg/v4/payment/request.json",
	"apiPaymentUrl":      "https://www.zarinpal.com/pg/StartPay/",
	"apiVerificationUrl": "https://api.zarinpal.com/pg/v4/payment/verify.json",
}
var sandboxAPI = map[string]string{
	"apiPurchaseUrl":     "https://sandbox.zarinpal.com/pg/v4/payment/request.json",
	"apiPaymentUrl":      "https://sandbox.zarinpal.com/pg/StartPay/",
	"apiVerificationUrl": "https://sandbox.zarinpal.com/pg/v4/payment/verify.json",
}
var zarinGateAPI = map[string]string{
	"apiPurchaseUrl":     "https://ir.zarinpal.com/pg/services/WebGate/wsdl",
	"apiPaymentUrl":      "https://www.zarinpal.com/pg/StartPay/:authority/ZarinGate",
	"apiVerificationUrl": "https://ir.zarinpal.com/pg/services/WebGate/wsdl",
}

type Zarinpal struct {
	cfg       *ZarinpalConfig
	endpoints map[string]string
}

type ZarinpalConfig struct {
	Mode        string
	MerchantID  string
	Callback    string
	Description string
}

func NewZarinpalConfig(mode, merchantID, callback, description string) *ZarinpalConfig {
	return &ZarinpalConfig{
		Mode:        mode,
		MerchantID:  merchantID,
		Callback:    callback,
		Description: description,
	}
}

// Gateway creates new Zarinpal gateway from the credentials in config
func (z *ZarinpalConfig) Gateway() (*Zarinpal, error) {
	if z.Mode != ZARINPAL_GATEWAY && z.Mode != ZARINPAL_NORMAL && z.Mode != ZARINPAL_SANDBOX {
		return nil, errors.New("invalid mode for Zarinpal driver")
	}
	var endpoints map[string]string
	switch z.Mode {
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
	MerchantID  string            `json:"merchant_id"`
	Amount      uint64            `json:"amount"`
	CallbackURL string            `json:"callback_url"`
	Description string            `json:"description"`
	Metadata    map[string]string `json:"metadata"`
}

func (r *purchaseReq) toJSON() ([]byte, error) {
	return json.Marshal(r)
}

func (z *Zarinpal) Purchase(ctx context.Context, i *payment.Invoice) (*payment.Invoice, error) {
	bs, err := (&purchaseReq{
		MerchantID:  z.cfg.MerchantID,
		Amount:      i.Amount,
		CallbackURL: z.cfg.Callback + "/" + i.TransactionID,
		Description: z.cfg.Description,
		Metadata:    i.Details,
	}).toJSON()
	if err != nil {
		return nil, err
	}
	body := bytes.NewReader(bs)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, z.endpoints["apiPurchaseUrl"], body)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	b, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("invalid status code: %d from %s returned %s", resp.StatusCode, z.endpoints["apiPurchaseUrl"], string(b))
	}
	var res struct {
		Data struct {
			Status    int    `json:"code"`
			Authority string `json:"authority"`
			Errors    []struct {
				Code    int    `json:"code"`
				Message string `json:"message"`
			} `json:"errors,omitempty"`
		} `json:"data"`
	}
	if err := json.Unmarshal(b, &res); err != nil {
		return nil, err
	}
	if res.Data.Status != 100 {
		return nil, fmt.Errorf("could not complete: %s status was %d", string(b), res.Data.Status)
	}
	i.TransactionID = res.Data.Authority
	return i, nil
}
func (z *Zarinpal) Pay(i *payment.Invoice) *payment.PayResponse {
	return &payment.PayResponse{
		URL:         z.endpoints["apiPaymentUrl"] + i.TransactionID,
		HasRedirect: true,
	}
}

type verifyReq struct {
	MerchantID string `json:"merchant_id"`
	Authority  string `json:"authority"`
	Amount     uint64 `json:"amount"`
}

func (r *verifyReq) toJSON() ([]byte, error) {
	return json.Marshal(r)
}

func (z *Zarinpal) Verify(ctx context.Context, amount uint64, args map[string]string) (*payment.Receipt, error) {
	bs, err := (&verifyReq{
		MerchantID: z.cfg.MerchantID,
		Authority:  args["transactionID"],
		Amount:     amount,
	}).toJSON()
	if err != nil {
		return nil, err
	}
	body := bytes.NewReader(bs)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, z.endpoints["apiVerificationUrl"], body)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	b, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	var res struct {
		Data struct {
			Status  int               `json:"code"`
			RefID   string            `json:"ref_id"`
			Details map[string]string `json:"details"`
			Errors  []struct {
				Code    int    `json:"code"`
				Message string `json:"message"`
			} `json:"errors,omitempty"`
		} `json:"data"`
	}
	if err := json.Unmarshal(b, &res); err != nil {
		return nil, err
	}
	msg := "خطای ناشناخته رخ داده است. در صورت کسر مبلغ از حساب حداکثر پس از 72 ساعت به حسابتان برمیگردد"
	var isSuccess bool
	switch res.Data.Status {
	case 100:
		isSuccess = true
		msg = "تراکنش با موفقیت انجام گردید"
	case 101:
		isSuccess = true
		msg = "عمليات پرداخت موفق بوده و قبلا عملیات وریفای تراكنش انجام شده است"
	case -9:
		msg = "خطای اعتبار سنجی"
	case -10:
		msg = "ای پی و يا مرچنت كد پذيرنده صحيح نمی باشد"
	case -11:
		msg = "مرچنت کد فعال نیست لطفا با تیم پشتیبانی ما تماس بگیرید"
	case -12:
		msg = "تلاش بیش از حد در یک بازه زمانی کوتاه"
	case -15:
		msg = "ترمینال شما به حالت تعلیق در آمده با تیم پشتیبانی تماس بگیرید"
	case -16:
		msg = "سطح تاييد پذيرنده پايين تر از سطح نقره ای می باشد"
	case -30:
		msg = "اجازه دسترسی به تسویه اشتراکی شناور ندارید"
	case -31:
		msg = "حساب بانکی تسویه را به پنل اضافه کنید مقادیر وارد شده برای تسهیم صحيح نمی باشد"
	case -32:
		msg = "مقادیر وارد شده برای تسهیم صحيح نمی باشد"
	case -33:
		msg = "درصد های وارد شده صحيح نمی باشد"
	case -34:
		msg = "مبلغ از کل تراکنش بیشتر است"
	case -35:
		msg = "تعداد افراد دریافت کننده تسهیم بیش از حد مجاز است"
	case -40:
		msg = "پارامترهای اضافی نامعتبر، expire_in معتبر نیست"
	case -50:
		msg = "مبلغ پرداخت شده با مقدار مبلغ در وریفای متفاوت است"
	case -51:
		msg = "پرداخت ناموفق"
	case -52:
		msg = "خطای غیر منتظره با پشتیبانی تماس بگیرید"
	case -53:
		msg = "اتوریتی برای این مرچنت کد نیست"
	case -54:
		msg = "اتوریتی نامعتبر است"
	}
	if isSuccess {
		return &payment.Receipt{
			RefID:   res.Data.RefID,
			Details: res.Data.Details,
		}, nil
	}
	return nil, errors.New(msg)
}
