package driver

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"mime/multipart"
	"net/http"

	"github.com/eamirgh/go-pay/payment"
)

type Paytr struct {
	cfg *PaytrConfig
}

type PaytrConfig struct {
	IsTest       bool
	MerchantID   string
	MerchantSalt string
	MerchantKey  string
	CallbackURL  string
}

func NewPaytrConfig(isTest bool, callbackURL, merchantID, merchantSalt, merchantKey string) *PaytrConfig {
	return &PaytrConfig{
		IsTest:       isTest,
		MerchantID:   merchantID,
		MerchantSalt: merchantSalt,
		MerchantKey:  merchantKey,
		CallbackURL:  callbackURL,
	}
}

func (c *PaytrConfig) Prepare() *Paytr {
	return &Paytr{
		cfg: c,
	}
}

func checkMetadata(i *payment.Invoice) bool {
	return i.Has("user_phone") && i.Has("user_ip") && i.Has("user_basket") &&
		i.Has("user_name") && i.Has("user_address") && i.Has("email") &&
		i.Has("currency") && i.Has("no_installment") && i.Has("max_installment") &&
		i.Has("lang") && i.Has("merchant_oid")
}

type iframeTokenRes struct {
	Status string `json:"status"`
	Token  string `json:"token"`
}

func (p *Paytr) Purchase(ctx context.Context, i *payment.Invoice) (*payment.Invoice, error) {

	if !checkMetadata(i) {
		return nil, fmt.Errorf("missing metadata")
	}
	url := "https://www.paytr.com/odeme/api/get-token"

	payload := &bytes.Buffer{}
	writer := multipart.NewWriter(payload)
	_ = writer.WriteField("merchant_id", p.cfg.MerchantID)
	_ = writer.WriteField("timeout_limit", "30")
	_ = writer.WriteField("payment_amount", fmt.Sprint(i.Amount))
	for k, v := range i.Details {
		_ = writer.WriteField(k, v)
	}
	_ = writer.WriteField("merchant_ok_url", p.cfg.CallbackURL+"?status=success")
	_ = writer.WriteField("merchant_fail_url", p.cfg.CallbackURL+"?status=fail")
	if p.cfg.IsTest {
		_ = writer.WriteField("debug_on", "1")
		_ = writer.WriteField("test_mode", "1")
	} else {
		_ = writer.WriteField("debug_on", "0")
		_ = writer.WriteField("test_mode", "0")
	}
	itest := 0
	if p.cfg.IsTest {
		itest = 1
	}
	hashStr := fmt.Sprintf("%s%s%s%s%s%s%s%s%s%d%s",
		p.cfg.MerchantID, i.Get("user_ip"), i.Get("merchant_oid"), i.Get("email"), fmt.Sprint(i.Amount),
		i.Get("user_basket"), i.Get("no_installment"), i.Get("max_installment"), i.Get("currency"),
		itest, p.cfg.MerchantSalt)
	hash := hmac.New(sha256.New, []byte(p.cfg.MerchantKey))
	_, err := hash.Write([]byte(hashStr))
	if err != nil {
		return nil, err
	}
	token := base64.StdEncoding.EncodeToString(hash.Sum(nil))
	_ = writer.WriteField("paytr_token", token)
	err = writer.Close()
	if err != nil {
		return nil, err
	}
	client := &http.Client{}
	req, err := http.NewRequest(http.MethodPost, url, payload)

	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", writer.FormDataContentType())
	res, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}
	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("paytr request failed: %s", body)
	}
	var tRes iframeTokenRes
	err = json.Unmarshal(body, &tRes)
	if err != nil {
		return nil, err
	}
	if tRes.Status != "success" {
		return nil, fmt.Errorf("paytr request failed: %s", body)
	}
	i.TransactionID = tRes.Token
	return i, nil
}

func (p *Paytr) Pay(i *payment.Invoice) *payment.PayResponse {
	return &payment.PayResponse{
		HasRedirect: false,
		URL:         fmt.Sprintf("https://www.paytr.com/odeme/guvenli/%s", i.TransactionID),
	}
}

func (p *Paytr) Verify(ctx context.Context, amount uint64, transactionID, hash, status string) (*payment.Receipt, error) {
	hashStr := fmt.Sprintf("%s%s%s%s", transactionID, p.cfg.MerchantSalt, status, fmt.Sprint(amount))
	hasher := hmac.New(sha256.New, []byte(p.cfg.MerchantKey))
	_, err := hasher.Write([]byte(hashStr))
	if err != nil {
		return nil, err
	}
	if hash != base64.StdEncoding.EncodeToString(hasher.Sum(nil)) {
		return nil, fmt.Errorf("hash not match")
	}
	if status != "success" {
		return nil, fmt.Errorf("payment failed")
	}
	return &payment.Receipt{}, nil
}
