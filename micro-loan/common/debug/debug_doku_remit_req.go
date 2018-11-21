package main

import (
	"io/ioutil"
	"net/http"

	"encoding/json"

	"github.com/astaxie/beego/logs"
)

func main() {

	var dokuRemitReq struct {
		SendType    string
		Auth1       string
		Beneficiary struct {
			/*
				Address string
				Country struct {
					Code string
				}
				FirstName   string
				LastName    string
				PhoneNumber string
			*/
			IdToken string
		}
		/*
			BeneficiaryAccount struct {
				Address string
				Bank    struct {
					Code        string
					CountryCode string
					Id          string
					Name        string
				}
				City   string
				Name   string
				Number string
			}
		*/
		BeneficiaryCity    string
		BeneficiaryCountry struct {
			Code string
		}
		BeneficiaryCurrency struct {
			Code string
		}
		Channel struct {
			Code string
		}
		Inquiry struct {
			Idtoken string
		}
		Sender struct {
			Address   string
			Birthdate string
			Country   struct {
				Code string
			}
			Firstname struct {
			}
		}
		SenderNote string
	}

	client := &http.Client{}
	req, err := http.NewRequest("GET", "http://localhost/json.php", nil)
	if err != nil {
		logs.Error("localhost http.NewRequest err: ", err)
	}

	//req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	//req.Header.Set("Cookie", "name=anny")
	//req.SetBasicAuth(secretKey, "")
	resp, err := client.Do(req)

	if err != nil {
		logs.Error("Xendit get virtual banks client.Do err: ", err)
	}

	defer resp.Body.Close()

	logs.Debug(resp.Body)

	body, err := ioutil.ReadAll(resp.Body)

	json.Unmarshal(body, &dokuRemitReq)

	logs.Debug(dokuRemitReq)
	//logs.Debug(dokuDisburseResponse.Status, dokuDisburseResponse.Message, dokuDisburseResponse.Inquiry.IdToken, dokuDisburseResponse.Inquiry.Fund.Origin.Amount, dokuDisburseResponse.Inquiry.Fund.Origin.Currency)

}
