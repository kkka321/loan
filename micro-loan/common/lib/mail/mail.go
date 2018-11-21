package mail

import (
	"fmt"
	"net/smtp"

	"github.com/astaxie/beego/logs"
)

func SendMail(title, body, addr, sender, rcpter string) {
	c, err := smtp.Dial(addr)
	if err != nil {
		logs.Error("[doSendMail] Dial error err:%v", err)
		return
	}

	content_type := "Content-Type: text/plain" + "; charset=UTF-8"

	// Set the sender and recipient.
	c.Mail(sender)
	c.Rcpt(rcpter)

	logs.Warn("[doSendMail] addr:%s, sender:%s, rcpter:%s, body:%s", addr, sender, rcpter, body)

	// Send the email body.
	wc, err := c.Data()
	if err != nil {
		logs.Error("[doSendMail] Data error err:%v", err)
		c.Quit()
		return
	}

	str := "To: " + rcpter + "\r\nFrom: " + sender + "<" + sender + ">\r\nSubject: " + title + "\r\n" + content_type + "\r\n\r\n" + body
	if _, err = fmt.Fprintf(wc, str); err != nil {
		logs.Error("[doSendMail] WriteTo error err:%v", err)
		c.Quit()
		return
	}

	if err = wc.Close(); err != nil {
		logs.Error("[doSendMail] Close error err:%v", err)
		c.Quit()
		return
	}

	c.Quit()
}
