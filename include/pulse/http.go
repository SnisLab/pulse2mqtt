package pulse

import (
	"io"
	"net/http"
	"sync"
	"time"

	"pulse2mqtt/include/settings"
)

var httpState struct {
	sync.Mutex
	client *http.Client
}

type HTTPResponse struct {
	StatusCode int
	Status     string
	Body       []byte
}

func Get(url string) (HTTPResponse, error) {
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return HTTPResponse{}, err
	}
	req.SetBasicAuth(settings.Load.Service.Pulse.User, settings.Load.Service.Pulse.Password)

	httpState.Lock()
	defer httpState.Unlock()
	if httpState.client == nil {
		httpState.client = &http.Client{Timeout: 10 * time.Second}
	}
	resp, err := httpState.client.Do(req)
	if err != nil {
		return HTTPResponse{}, err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return HTTPResponse{}, err
	}
	return HTTPResponse{StatusCode: resp.StatusCode, Status: resp.Status, Body: body}, nil
}
