package perfect_money

import (
	"crypto/tls"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"regexp"
)

const (
	CreateURL   = "https://perfectmoney.is/acct/ev_create.asp"
	ReturnURL   = "https://perfectmoney.is/acct/ev_remove.asp"
	ActivateURL = "https://perfectmoney.is/acct/ev_activate.asp"
	BalanceURL  = "https://perfectmoney.is/acct/balance.asp"
)

type PerfectMoney struct {
	AccountID    string
	PassPhrase   string
	PayeeAccount string
}

type Array map[string]string

func New(AccountID, PassPhrase, PayeeAccount string) *PerfectMoney {
	return &PerfectMoney{
		AccountID:    AccountID,
		PassPhrase:   PassPhrase,
		PayeeAccount: PayeeAccount,
	}
}

func (pm *PerfectMoney) send(requestUrl string, params map[string]interface{}) Array {
	values := url.Values{}
	values.Add("AccountID", pm.AccountID)
	values.Add("PassPhrase", pm.PassPhrase)

	for key, value := range params {
		values.Add(key, value.(string))
	}

	req, err := http.NewRequest("GET", fmt.Sprintf("%s?%s", requestUrl, values.Encode()), nil)
	if err != nil {
		fmt.Println("Error: Can not create the thing: ", err.Error())
		return nil
	}

	tr := &http.Transport{
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: true,
		},
	}
	c := http.Client{Transport: tr}
	resp, err := c.Do(req)

	if err != nil {
		fmt.Println("Request Failed:: ", err.Error())
		return nil
	}

	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			fmt.Println("Error:: ", err.Error())
		}
	}(resp.Body)
	b, err := io.ReadAll(resp.Body)

	extracted := make(Array)
	re := regexp.MustCompile("<input name='(.*)' type='hidden' value='(.*)'>")
	matches := re.FindAllStringSubmatch(string(b), -1)
	if matches == nil {
		fmt.Println("failed: invalid input")
		return nil
	}

	for _, match := range matches {
		key := match[1]
		value := match[2]
		extracted[key] = value
	}

	return extracted
}

func (pm *PerfectMoney) Redeem(evNumber, evCode string) Array {
	return pm.send(ActivateURL, map[string]interface{}{
		"Payee_Account": pm.PayeeAccount,
		"ev_number":     evNumber,
		"ev_code":       evCode,
	})
}

func (pm *PerfectMoney) Balance() Array {
	return pm.send(BalanceURL, map[string]interface{}{})
}
