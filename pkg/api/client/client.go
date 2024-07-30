package client

import (
	"bytes"
	"drone/pkg/api/server"
	"encoding/json"
	"fmt"
	"github.com/sirupsen/logrus"
	"net/http"
	"time"
)

type Client struct {
	url    string
	client *http.Client
}

func New(domain string) *Client {
	httpClient := &http.Client{Timeout: 3 * time.Second}
	return &Client{url: fmt.Sprintf("http://%s/put", domain), client: httpClient}
}

func (c *Client) SendWithTimeout(msg server.Message) {
	data, _ := json.Marshal(msg)
	_, err := c.client.Post(c.url, "application/json", bytes.NewBuffer(data))
	if err != nil {
		logrus.Debugf("Post: %s", err.Error())
	}
}
