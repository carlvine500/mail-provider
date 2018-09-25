package http

import (
	"net/http"
	"strings"

	"github.com/open-falcon/mail-provider/config"
	"github.com/toolkits/smtp"
	"github.com/toolkits/web/param"

	"gopkg.in/gomail.v2"
	"strconv"
	"crypto/tls"
)

func configProcRoutes() {

	http.HandleFunc("/sender/mail", func(w http.ResponseWriter, r *http.Request) {
		cfg := config.Config()
		token := param.String(r, "token", "")
		if cfg.Http.Token != token {
			http.Error(w, "no privilege", http.StatusForbidden)
			return
		}

		tos := param.MustString(r, "tos")
		subject := param.MustString(r, "subject")
		content := param.MustString(r, "content")
		tos = strings.Replace(tos, ",", ";", -1)

		//替换content中的 \r\n 为 <br/>
		content = strings.Replace(content, "\r\n", "<br/>", -1)

		tosArr := []string{}
		for _, tosTmp := range strings.Split(tos, ";") {
			tosArr = append(tosArr, strings.TrimSpace(tosTmp))
		}

		if cfg.Smtp.Type == "smtp_ssl" {
			m := gomail.NewMessage()
			m.SetHeader("From", cfg.Smtp.From)
			m.SetHeader("To", tosArr...)
			//m.SetAddressHeader("Cc", "dan@example.com", "Dan")
			m.SetHeader("Subject", subject)
			m.SetBody("text/html", content)
			//m.Attach("/home/Alex/lolcat.jpg")

			//d := gomail.NewDialer(cfg.Smtp.Addr, cfg.Smtp.Port, cfg.Smtp.Username, cfg.Smtp.Password)
			d := &gomail.Dialer{
				Host:     cfg.Smtp.Addr,
				Port:     cfg.Smtp.Port,
				Username: cfg.Smtp.Username,
				Password: cfg.Smtp.Password,
				SSL:      cfg.Smtp.Port == 465,
				TLSConfig: &tls.Config{ServerName: cfg.Smtp.Addr, InsecureSkipVerify: true},
			}
			// Send the email to Bob, Cora and Dan.
			if err := d.DialAndSend(m); err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
			}else {
				http.Error(w, "success", http.StatusOK)
			}
		}else {
			s := smtp.New(cfg.Smtp.Addr+":" + strconv.Itoa(cfg.Smtp.Port), cfg.Smtp.Username, cfg.Smtp.Password)
			err := s.SendMail(cfg.Smtp.From, tos, subject, content)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
			} else {
				http.Error(w, "success", http.StatusOK)
			}
		}
	})

}
