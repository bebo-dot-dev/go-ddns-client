package notifications

import (
	"bytes"
	"fmt"
	"go-ddns-client/service/config"
	"log"
	"net/http"
)

//SipGateSmsNotifier implements the sipgate IO sms sender
/*
sipgate docs:
	https://api.sipgate.com/v2/doc#/sessions/sendWebSms
	https://github.com/sipgate-io/sipgateio-sendsms-python

request url:
	https://api.sipgate.com/v2/sessions/sms

json payload:
	{
		"smsId": "smsId",
		"recipient": "0123456789",
		"message": "The IP address for domain 'example.com' has been updated to '127.0.0.1'"
	}
*/
type SipGateSmsNotifier struct {
	conf *config.SipgateSMS
}

//Send sends the sipgate IO sms notification
func (notifier SipGateSmsNotifier) Send(domainCount int, domainsStr string, ipaddress string) error {
	plural := ""
	if domainCount > 1 {
		plural = "s"
	}
	msg := fmt.Sprintf("The IP address for domain%s '%s' has been updated to '%s'", plural, domainsStr, ipaddress)

	jsonBody := fmt.Sprintf(`{
		"smsId": "%s",
		"recipient": "%s",
		"message": "%s"
	}`, notifier.conf.SmsId, notifier.conf.Recipient, msg)

	headers := make(map[string]string)
	headers["Content-Type"] = "application/json"

	_, _, err := PerformHttpRequest(
		http.MethodPost,
		"https://api.sipgate.com/v2/sessions/sms",
		notifier.conf.TokenId,
		notifier.conf.Token,
		bytes.NewBuffer([]byte(jsonBody)),
		headers)

	if err != nil {
		return notifier.sipgateError(err)
	}

	log.Println("Sipgate IO SMS notification sent")
	return nil
}

func (notifier SipGateSmsNotifier) sipgateError(err error) error {
	return fmt.Errorf("sipgate IO SMS error: %v", err)
}
