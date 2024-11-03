package client

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"go.elastic.co/apm/module/apmhttp"
	"golang.org/x/net/context/ctxhttp"
	"io/ioutil"
	"net"
	"net/http"
	"time"
)

type HTTPClient struct {
	name       string
	httpClient *http.Client
}

func (h *HTTPClient) SendHTTPRequest(ctx context.Context, method, uri string, requestData []byte, res interface{}) error {
	req, err := http.NewRequestWithContext(ctx, method, uri, bytes.NewReader(requestData))
	if err != nil {
		return nil
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := ctxhttp.Do(ctx, h.httpClient, req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return errors.New(resp.Status)
	}
	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	err = json.Unmarshal(data, res)
	return err
}

func NewHttpClient(name string) *HTTPClient {
	trans := &http.Transport{
		DialContext: (&net.Dialer{
			Timeout:   5 * time.Second,
			KeepAlive: 5 * time.Second,
		}).DialContext,
		TLSHandshakeTimeout:   4 * time.Second,
		ResponseHeaderTimeout: 6 * time.Second,
		ExpectContinueTimeout: 4 * time.Second,
		DisableKeepAlives:     false,
		MaxIdleConnsPerHost:   1024,
		MaxConnsPerHost:       2048,
	}
	client := http.Client{
		Transport: trans,
		Timeout:   5 * time.Second,
	}
	httpClient := apmhttp.WrapClient(&client)
	return &HTTPClient{
		name:       name,
		httpClient: httpClient,
	}
}
