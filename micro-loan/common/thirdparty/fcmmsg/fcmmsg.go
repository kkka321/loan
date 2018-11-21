package fcmmsg

import (
	"context"
	"encoding/json"
	"io/ioutil"

	"github.com/astaxie/beego"
	"github.com/astaxie/beego/logs"
	"golang.org/x/oauth2/google"

	"micro-loan/common/tools"
)

type FcmMessage struct {
	Message struct {
		Token string `json:"token"`
		Data  struct {
			Skipto string `json:"skipto"`
		} `json:"data"`
		Notification struct {
			Body  string `json:"body"`
			Title string `json:"title"`
		} `json:"notification"`
	} `json:"message"`
}

var (
	token []byte
	url   string
)

func init() {
	tokenFile := beego.AppConfig.String("google_token")
	url = beego.AppConfig.String("google_url")

	token, _ = ioutil.ReadFile(tokenFile)
}

func SendMessage(accountToken []string, title, body, skipTo string) (int, error) {
	ctx := context.Background()

    num := 0
	creds, err := google.CredentialsFromJSON(ctx, token, "https://www.googleapis.com/auth/cloud-platform", "https://www.googleapis.com/auth/firebase.messaging")
	if err != nil {
		logs.Error("[SendMessage] CredentialsFromJSON error err:%v", err)
		return num, err
	}

	t, _ := creds.TokenSource.Token()

	for i := 0; i < len(accountToken); i++ {
		if accountToken[i] == "" {
			logs.Warn("[SendMessage] fcm token empty")
			continue
		}

		msg := FcmMessage{}
		msg.Message.Token = accountToken[i]
		msg.Message.Notification.Body = body
		msg.Message.Notification.Title = title
		msg.Message.Data.Skipto = skipTo

		reqHeader := map[string]string{
			"Content-Type":  "application/json;charset=UTF-8",
			"Authorization": "Bearer " + t.AccessToken,
		}

		bytesData, err := json.Marshal(msg)
		reqBody := string(bytesData)

		_, httpCode, err := tools.SimpleHttpClient("POST", url, reqHeader, reqBody, tools.DefaultHttpTimeout())
		if err != nil {
			logs.Error("[SendMessage] SimpleHttpClient error err:%s", err.Error())
			continue
		}

		if httpCode != 200 && httpCode != 404 {
			logs.Error("[SendMessage] return error token:%s, code:%d", accountToken[i], httpCode)
			continue
		}

		num++
	}

	logs.Info("[SendMessage] send message done size:%d", num)

	return num, nil
}
