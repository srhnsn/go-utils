package email

import (
	"crypto/tls"
	"fmt"
	"net/mail"
	"net/smtp"
	"time"

	"github.com/srhnsn/go-utils/log"
)

type Email struct {
	From              string
	To                string
	Subject           string
	AdditionalHeaders map[string]string
	Text              string

	Error chan error
}

type EmailConfig struct {
	Server   string
	Port     uint16
	Username string
	Password string
	From     string
}

const PauseBetweenEmails = 2 * time.Second
const EmailWorkerRestartTime = 30 * time.Second

var emailConfig EmailConfig
var emailQueue chan Email
var smtpAuth smtp.Auth
var tlsConfig tls.Config

func InitEmails(config EmailConfig) {
	emailConfig = config
	emailQueue = make(chan Email, 1000)
	smtpAuth = smtp.PlainAuth("", config.Username, config.Password, config.Server)
	tlsConfig = tls.Config{ServerName: config.Server}

	go initWorker()
}

func SendEmail(email Email) chan error {
	if email.From == "" {
		email.From = emailConfig.From
	}

	if email.Error == nil {
		email.Error = make(chan error, 1)
	}

	log.Trace.Printf("New email to %s queued (subject: %s)", email.To, email.Subject)
	emailQueue <- email
	return email.Error
}

func initWorker() {
	var conn *smtp.Client
	var email Email
	var err error

	defer func() {
		go func() {
			log.Warning.Println("Email worker go routine returned, restarting in %d seconds", EmailWorkerRestartTime.Seconds())
			time.Sleep(EmailWorkerRestartTime)
			go initWorker()
		}()
	}()

	log.Trace.Println("Email worker initialized")

	for {
		log.Trace.Println("Waiting for new emails to send (without server connection)")
		email = <-emailQueue
		log.Trace.Println("New queued emails, trying to connect to SMTP server")
		conn, err = connect()

		if err != nil {
			log.Error.Panicf("Connection to SMTP server failed: %s", err)
		}

		log.Trace.Printf("Connected")
		sendEmailWithNotify(conn, email)
		log.Trace.Printf("Checking if there are more emails to send")

	EmailSendLoop:
		for {
			select {
			case email := <-emailQueue:
				time.Sleep(PauseBetweenEmails)
				sendEmailWithNotify(conn, email)
			default:
				break EmailSendLoop
			}
		}

		log.Trace.Println("No more emails to send, closing SMTP connection")

		conn.Close()
		conn = nil
	}
}

func connect() (*smtp.Client, error) {
	conn, err := smtp.Dial(fmt.Sprintf("%s:%d", emailConfig.Server, emailConfig.Port))

	if err != nil {
		return nil, err
	}

	if err := conn.StartTLS(&tlsConfig); err != nil {
		return nil, err
	}

	if err := conn.Auth(smtpAuth); err != nil {
		return nil, err
	}

	return conn, nil
}

func encodeRFC2047(String string) string {
	addr := mail.Address{String, ""}
	str := addr.String()
	str = str[1 : len(str)-4]
	return str
}

func sendEmail(conn *smtp.Client, email Email) error {
	log.Trace.Printf("Sending email to %s", email.To)

	fromParsed, err := mail.ParseAddress(email.From)

	if err != nil {
		return err
	}

	toParsed, err := mail.ParseAddress(email.To)

	if err != nil {
		return err
	}

	if err := conn.Mail(fromParsed.Address); err != nil {
		return err
	}

	if err := conn.Rcpt(toParsed.Address); err != nil {
		return err
	}

	wc, err := conn.Data()

	if err != nil {
		return err
	}

	headers := map[string]string{
		"Content-Type": "text/plain; charset=\"utf-8\"",
		"To":           toParsed.String(),
		"From":         fromParsed.String(),
		"Subject":      encodeRFC2047(email.Subject),
	}

	if email.AdditionalHeaders != nil {
		for key, value := range email.AdditionalHeaders {
			headers[key] = value
		}
	}

	for key, value := range headers {
		_, err = fmt.Fprintf(wc, "%s: %s\n", key, value)

		if err != nil {
			return err
		}
	}

	_, err = fmt.Fprintf(wc, "\n%s", email.Text)

	if err != nil {
		return err
	}

	err = wc.Close()

	if err != nil {
		return err
	}

	log.Trace.Println("Success!")

	return nil
}

func sendEmailWithNotify(conn *smtp.Client, email Email) {
	err := sendEmail(conn, email)

	if err == nil {
		email.Error <- nil
	} else {
		email.Error <- err
		log.Error.Printf("Sending email failed: %s", err)
	}
}
