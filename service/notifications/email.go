package notifications

import (
	"crypto/tls"
	"fmt"
	"github.com/bebo-dot-dev/go-ddns-client/service/config"
	"log"
	"net"
	"net/mail"
	"net/smtp"
	"strings"
)

//EmailNotifier implements a simple email sender
type EmailNotifier struct {
	conf *config.Email
}

//Send sends the email notification
func (notifier EmailNotifier) Send(hostname string, domainCount int, domainsStr, ipv4, ipv6 string) error {
	host, _, _ := net.SplitHostPort(notifier.conf.SmtpServer)
	auth := smtp.PlainAuth("", notifier.conf.Username, notifier.conf.Password, host)

	client, err := notifier.getSmtpClient(host)
	if err != nil {
		return notifier.emailError(err)
	}

	// Auth
	if err := client.Auth(auth); err != nil {
		return notifier.emailError(err)
	}

	from, recipients, err := notifier.getAddresses(client)
	if err != nil {
		return notifier.emailError(err)
	}

	// Data
	w, err := client.Data()
	if err != nil {
		return notifier.emailError(err)
	}

	emailMsg, err := notifier.buildMessage(from, *recipients, hostname, domainCount, domainsStr, ipv4, ipv6)
	if err != nil {
		return notifier.emailError(err)
	}
	_, err = w.Write([]byte(emailMsg))
	if err != nil {
		return notifier.emailError(err)
	}

	if err = w.Close(); err != nil {
		return notifier.emailError(err)
	}

	if err = client.Quit(); err != nil {
		return notifier.emailError(err)
	}

	log.Println("Email notification sent")
	return err
}

//getSmtpClient returns an smtp.Client setup according to the configured notifier.conf.SecurityType (SSL or TLS)
func (notifier EmailNotifier) getSmtpClient(host string) (*smtp.Client, error) {
	var client *smtp.Client
	var err error

	tlsConfig := &tls.Config{
		InsecureSkipVerify: true,
		ServerName:         host,
	}

	switch notifier.conf.SecurityType {
	case "SSL":
		conn, err := tls.Dial("tcp", notifier.conf.SmtpServer, tlsConfig)
		if err != nil {
			break
		}

		client, err = smtp.NewClient(conn, host)
		if err != nil {
			break
		}
	case "TLS":
		client, err = smtp.Dial(notifier.conf.SmtpServer)
		if err != nil {
			break
		}

		err := client.StartTLS(tlsConfig)
		if err != nil {
			break
		}
	default:
		log.Panic("unsupported email security type " + notifier.conf.SecurityType)
	}

	return client, err
}

//getAddresses constructs email addresses and validates them against the supplied smtp.Client
func (notifier EmailNotifier) getAddresses(client *smtp.Client) (*mail.Address, *[]mail.Address, error) {
	from := mail.Address{Name: notifier.conf.From.Name, Address: notifier.conf.From.Address}
	var recipients []mail.Address
	for _, recipient := range notifier.conf.Recipients {
		recipients = append(recipients, mail.Address{Name: recipient.Name, Address: recipient.Address})
	}

	if err := client.Mail(from.Address); err != nil {
		return nil, nil, notifier.emailError(err)
	}
	for _, recipient := range recipients {
		if err := client.Rcpt(recipient.Address); err != nil {
			return nil, nil, notifier.emailError(err)
		}
	}
	return &from, &recipients, nil
}

//buildMessage builds the email message to be sent
func (notifier EmailNotifier) buildMessage(
	from *mail.Address,
	recipients []mail.Address,
	hostname string,
	domainCount int,
	domainsStr,
	ipv4,
	ipv6 string) (string, error) {

	plural := ""
	if domainCount > 1 {
		plural = "s"
	}

	subject := "go ddns client ip address update"
	body := fmt.Sprintf("The IP addresses for domain%s '%s' were updated to:\n%s\n%s\nby: %s",
		plural, domainsStr, ipv4, ipv6, hostname)

	var rb strings.Builder
	for index, recipient := range recipients {
		if recipient.Name != "" {
			_, err := fmt.Fprintf(&rb, "\"%s\" ", recipient.Name)
			if err != nil {
				return "", nil
			}
		}

		_, err := fmt.Fprintf(&rb, "<%s>", recipient.Address)
		if err != nil {
			return "", nil
		}
		if index < (len(recipients) - 1) {
			_, err = fmt.Fprint(&rb, ", ")
			if err != nil {
				return "", nil
			}
		}
	}

	headers := make(map[string]string)
	headers["From"] = from.String()
	headers["To"] = rb.String()
	headers["Subject"] = subject

	//construct the message
	message := ""
	for k, v := range headers {
		message += fmt.Sprintf("%s: %s\r\n", k, v)
	}
	message += "\r\n" + body

	return message, nil
}

func (notifier EmailNotifier) emailError(err error) error {
	return fmt.Errorf("email notifier error: %v", err)
}
