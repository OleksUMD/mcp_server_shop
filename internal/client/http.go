package client

import (
	"fmt"

	"github.com/sirupsen/logrus"
	"resty.dev/v3"
)

type RestClient struct {
	client       *resty.Client
	baseURL      string
	defaultToken string
	useToken     bool
	logger       *logrus.Logger
}

func NewRestClient(baseURL, defaultToken string, logger *logrus.Logger) *RestClient {
	client := resty.New()
	client.SetBaseURL(baseURL)
	client.SetHeader("Content-Type", "application/json")
	client.AddResponseMiddleware(func(c *resty.Client, resp *resty.Response) error {
		logger.WithFields(logrus.Fields{
			"method": resp.Request.Method,
			"url":    resp.Request.URL,
			"status": resp.StatusCode(),
		}).Debug("REST API call")
		return nil
	})
	return &RestClient{
		client:       client,
		baseURL:      baseURL,
		defaultToken: defaultToken,
		logger:       logger,
	}
}

func (c *RestClient) prepareRequest() *resty.Request {
	req := c.client.R()
	if c.useToken && c.defaultToken != "" {
		req.SetAuthToken(c.defaultToken)
	}
	return req
}

func (c *RestClient) WithToken() *RestClient {
	clone := *c
	clone.useToken = true
	return &clone
}

func (c *RestClient) Get(path string, params map[string]string) ([]byte, error) {
	req := c.prepareRequest()

	if len(params) > 0 {
		req.SetQueryParams(params)
	}
	resp, err := req.Get(path)
	if err != nil {
		return nil, err
	}
	if resp.ResultError() != nil {
		return nil, fmt.Errorf("API error from GET(): %s", resp.Status())
	}
	return resp.Bytes(), nil
}

func (c *RestClient) Post(path string, body any) ([]byte, error) {
	req := c.prepareRequest().
		SetBody(body).
		SetHeader("Content-Type", "application/json")

	resp, err := req.Post(path)
	if err != nil {
		return nil, err
	}
	if resp.ResultError() != nil {
		return nil, fmt.Errorf("API error from GET(): %s", resp.Status())
	}
	return resp.Bytes(), nil
}
