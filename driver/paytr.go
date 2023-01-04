package driver

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"mime/multipart"
	"net/http"
)

type Paytr struct {
	cfg       *PaytrConfig
	endpoints map[string]string
}

type PaytrConfig struct {
	Mode         string
	MerchantID   string
	MerchantSalt string
}

func iframe() {

	url := "https://www.paytr.com/odeme/api/get-token"
	method := "POST"

	payload := &bytes.Buffer{}
	writer := multipart.NewWriter(payload)
	_ = writer.WriteField("merchant_id", "123456789")
	_ = writer.WriteField("merchant_oid", "STR-123456789")
	_ = writer.WriteField("paytr_token", "token")
	_ = writer.WriteField("payment_amount", "1000")
	_ = writer.WriteField("user_name", "User Name")
	_ = writer.WriteField("user_address", "User Address")
	_ = writer.WriteField("email", "example@mail.com")
	_ = writer.WriteField("user_phone", "55512345670")
	_ = writer.WriteField("user_ip", "127.0.0.1")
	_ = writer.WriteField("user_basket", "W1siw5xyw7xuIEFkxLEgMSIsIjE4LjAwIiwxXSxbIsOccsO8biBBZMSxIDEiLCIxOC4wMCIsMV1d")
	_ = writer.WriteField("currency", "TL")
	_ = writer.WriteField("no_installment", "0")
	_ = writer.WriteField("max_installment", "0")
	_ = writer.WriteField("lang", "tr")
	_ = writer.WriteField("merchant_ok_url", "http://www.siteniz.com/odeme_basarili.php")
	_ = writer.WriteField("merchant_fail_url", "http://www.siteniz.com/odeme_hata.php")
	_ = writer.WriteField("debug_on", "1")
	_ = writer.WriteField("test_mode", "1")
	_ = writer.WriteField("timeout_limit", "30")
	err := writer.Close()
	if err != nil {
		fmt.Println(err)
		return
	}

	client := &http.Client{}
	req, err := http.NewRequest(method, url, payload)

	if err != nil {
		fmt.Println(err)
		return
	}
	req.Header.Set("Content-Type", writer.FormDataContentType())
	res, err := client.Do(req)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer res.Body.Close()

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println(string(body))
}
