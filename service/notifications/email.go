package notifications

import (
	"crypto/tls"
	"fmt"
	"go-ddns-client/service/config"
	"log"
	"net"
	"net/mail"
	"net/smtp"
	"os"
	"strings"
)

//EmailNotifier implements a simple email sender
type EmailNotifier struct {
	conf *config.Email
}

//Send sends the email notification
func (notifier EmailNotifier) Send(domainCount int, domainsStr, ipv4, ipv6 string) error {
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

	emailMsg := notifier.buildMessage(from, *recipients, domainCount, domainsStr, ipv4, ipv6)
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
		break

	case "TLS":
		client, err = smtp.Dial(notifier.conf.SmtpServer)
		if err != nil {
			break
		}

		err := client.StartTLS(tlsConfig)
		if err != nil {
			break
		}
		break
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
	domainCount int,
	domainsStr,
	ipv4,
	ipv6 string) string {

	plural := ""
	if domainCount > 1 {
		plural = "s"
	}
	subject := "go ddns client ip address update"

	hostname, err := os.Hostname()
	if err != nil {
		return ""
	}

	body := fmt.Sprintf("The IP addresses for domain%s '%s' were updated to:\n%s\n%s\nby: %s",
		plural, domainsStr, ipv4, ipv6, hostname)

	var builder strings.Builder
	for index, recipient := range recipients {
		if recipient.Name != "" {
			_, err := fmt.Fprintf(&builder, "\"%s\" ", recipient.Name)
			if err != nil {
				return ""
			}
		}

		_, err := fmt.Fprintf(&builder, "<%s>", recipient.Address)
		if err != nil {
			return ""
		}
		if index < (len(recipients) - 1) {
			_, err = fmt.Fprint(&builder, ", ")
			if err != nil {
				return ""
			}
		}
	}

	//setup headers
	headers := make(map[string]string)
	headers["From"] = from.String()
	headers["To"] = builder.String()
	headers["Subject"] = subject

	//setup message
	message := ""
	for k, v := range headers {
		message += fmt.Sprintf("%s: %s\r\n", k, v)
	}
	message += "\r\n" + body

	return message
}

func (notifier EmailNotifier) emailError(err error) error {
	return fmt.Errorf("email notifier error: %v", err)
}
