package email

import (
	"crypto/tls"
	"fmt"
	"net/smtp"
	"time"

	"github.com/srhnsn/go-utils/log"
)

type Email struct {
	From    string
	To      string
	Subject string
	Text    string

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

func sendEmail(conn *smtp.Client, email Email) error {
	log.Trace.Printf("Sending email to %s", email.To)

	if err := conn.Mail(emailConfig.From); err != nil {
		return err
	}

	if err := conn.Rcpt(email.To); err != nil {
		return err
	}

	wc, err := conn.Data()

	if err != nil {
		return err
	}

	_, err = fmt.Fprintf(wc, "Subject: %s\n\n%s", email.Subject, email.Text)

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
