package notifications

import (
	"go-ddns-client/service/config"
	"io"
	"log"
	"net/http"
	"time"
)

//INotification describes the interface of a type able to send a notification
type INotification interface {
	Send(domain, currentIP string) error
}

//Manager wraps types that have the ability to send a notification
type Manager struct {
	Notifiers []INotification
}

//GetManager returns the notification manager
func GetManager(conf *config.Notifications) *Manager {
	var notifiers []INotification

	if conf.SipgateSMS.Enabled {
		notifiers = append(notifiers, &SipGateSmsNotifier{conf: &conf.SipgateSMS})
	}

	return &Manager{
		Notifiers: notifiers,
	}
}

//Send sends one or more notifications
func (manager *Manager) Send(domain, currentIP string) {
	for _, notifier := range manager.Notifiers {
		if err := notifier.Send(domain, currentIP); err != nil {
			log.Println("Send notification failed with error:", err)
		}
	}
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
