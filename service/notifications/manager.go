package notifications

import (
	"go-ddns-client/service/config"
	"io"
	"log"
	"net/http"
	"time"
)

//INotificationManager describes the interface the notifications.Manager
type INotificationManager interface {
	GetNotifierCount() int
	Send(domainCount int, domainsStr, ipv4, ipv6 string) error
}

//INotification describes the interface of a type able to send a notification
type INotification interface {
	Send(domainCount int, domainsStr, ipv4, ipv6 string) error
}

//Manager wraps types that have the ability to send a notification
type Manager struct {
	Notifiers []INotification
}

//GetManager returns the notification manager
func GetManager(conf *config.Notifications) INotificationManager {
	var notifiers []INotification

	if conf.SipgateSMS.Enabled {
		notifiers = append(notifiers, &SipGateSmsNotifier{conf: &conf.SipgateSMS})
	}
	if conf.Email.IsEnabled {
		notifiers = append(notifiers, &EmailNotifier{conf: &conf.Email})
	}

	return &Manager{
		Notifiers: notifiers,
	}
}

func (manager *Manager) GetNotifierCount() int {
	return len(manager.Notifiers)
}

//Send sends one or more notifications
func (manager *Manager) Send(domainCount int, domainsStr string, ipv4, ipv6 string) error {
	for _, notifier := range manager.Notifiers {
		if err := notifier.Send(domainCount, domainsStr, ipv4, ipv6); err != nil {
			return err
		}
	}
	return nil
}

// PerformHttpRequest performs a HTTP request and returns the status code and the response
func PerformHttpRequest(
	method string,
	url string,
	username string,
	password string,
	body io.Reader,
	headers map[string]string) (int, []byte, error) {

	request, err := http.NewRequest(method, url, body)
	if err != nil {
		return 0, nil, err
	}

	if username != "" && password != "" {
		request.SetBasicAuth(username, password)
	}

	if headers != nil {
		for key, value := range headers {
			request.Header.Set(key, value)
		}
	}

	client := &http.Client{Timeout: 5 * time.Second}
	response, err := client.Do(request)
	if err != nil {
		return 0, nil, err
	}

	defer func() {
		err := response.Body.Close()
		if err != nil {
			log.Println(err)
		}
	}()

	responseBytes, err := io.ReadAll(response.Body)
	if err != nil {
		return response.StatusCode, nil, err
	}

	return response.StatusCode, responseBytes, nil
}
