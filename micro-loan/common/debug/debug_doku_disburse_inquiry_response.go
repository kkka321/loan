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
			Address string
			Country struct {
				Code string
			}
			FirstName   string
			LastName    string
			PhoneNumber string
		}
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
			IdToken string
		}
		Sender struct {
			Address   string
			BirthDate string
			Country   struct {
				Code string
			}
			FirstName         string
			Gender            string
			LastName          string
			PersonalId        string
			PersonalIdCountry struct {
				Code string
			}
			PersonalIdExpireDate string
			PersonalIdIssueDate  string
			PersonalIdType       string
			PhoneNumber          string
		}

		SenderAmount  string
		SenderCountry struct {
			Code string
		}
		SenderCurrency struct {
			Code string
		}
		SenderNote string
	}

	logs.Debug(dokuRemitReq)

	var dokuDisburseInquiryResponse struct {
		Status  string
		Message string
		Inquiry struct {
			IdToken string
			Fund    struct {
				Origin struct {
					Amount   int64
					Currency string
				}
				Destination struct {
					Amount   int64
					Currency string
				}
				Fees struct {
					Total    int64
					Currency string
				}
				Components []struct {
					Description string
					Amount      int64
				}
			}
			Sendercountry struct {
				Code int
				Name string
			}
			SenderCurrency struct {
				Code int
			}
			BeneficiaryCountry struct {
				Code int
				Name string
			}
			BeneficiaryCurrency struct {
				Code int
			}
			Channel struct {
				Code int
				Name string
			}
			ForexReference struct {
				Id    int
				Forex struct {
					Origin struct {
						Code int
					}
					Destination struct {
						Code int
					}
				}
				Rate        int
				CreatedTime int64
			}
		}
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

	json.Unmarshal(body, &dokuDisburseInquiryResponse)

	logs.Debug(dokuDisburseInquiryResponse)
	//logs.Debug(dokuDisburseResponse.Status, dokuDisburseResponse.Message, dokuDisburseResponse.Inquiry.IdToken, dokuDisburseResponse.Inquiry.Fund.Origin.Amount, dokuDisburseResponse.Inquiry.Fund.Origin.Currency)

}
