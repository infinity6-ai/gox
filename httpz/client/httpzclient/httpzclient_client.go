package httpzclient

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"time"

	"github.com/infinity6-ai/gox/commonz/deferz"
	"github.com/infinity6-ai/gox/commonz/validation/checker"
)

var defaultHttpClient = &http.Client{
	Timeout: 10 * time.Second,
	Transport: &http.Transport{
		MaxIdleConns:        100,
		MaxIdleConnsPerHost: 100,
		IdleConnTimeout:     90 * time.Second,
	},
}

type Options struct {
	BaseUrl    string
	OpenClient func(ctx context.Context) *http.Client
}

func (o *Options) fix() {
	checker.StrNotEmpty(o.BaseUrl, "baseurl")
}

type Client struct {
	Context context.Context
	Options Options
	filters []Filter
	dfz     *deferz.Deferz
	client  *http.Client
}

func New(ctx context.Context, opts Options) *Client {
	opts.fix()
	ret := &Client{
		Context: ctx,
		Options: opts,
		client:  defaultHttpClient,
		dfz:     deferz.New(ctx),
	}
	if opts.OpenClient != nil {
		ret.client = opts.OpenClient(ctx)
		ret.dfz.Add(ret.client.CloseIdleConnections)
	}
	return ret
}

func (c *Client) Close() {
	if c.dfz != nil {
		c.dfz.Close()
	}
}

func (c *Client) AddFilter(filter Filter) {
	c.filters = append(c.filters, filter)
}

func (c *Client) Do(req *Req) (*Resp, error) {
	var h Handler = c.send
	for i := len(c.filters) - 1; i >= 0; i-- {
		filter := c.filters[i]
		next := h
		h = func(req *Req) (*Resp, error) {
			return filter(req, next)
		}
	}
	return h(req)
}

func (c *Client) send(req *Req) (*Resp, error) {
	dfz := deferz.New(c.Context)
	defer dfz.Close()

	urlString, err := c.buildURL(req)
	if err != nil {
		return nil, fmt.Errorf("failed to build URL: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(c.Context, req.Method, urlString, req.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	for k, v := range req.Headers {
		httpReq.Header[k] = v
	}

	httpResp, err := c.client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed to execute request: %w", err)
	}
	dfz.AddCloserS(httpResp.Body)
	resp := &Resp{}
	resp.fromHttpResponse(httpResp)

	dfz.Detach()
	return resp, nil
}

func (c *Client) buildURL(req *Req) (string, error) {
	u, err := url.Parse(c.Options.BaseUrl)
	if err != nil {
		return "", fmt.Errorf("invalid base URL: %w", err)
	}

	if u.Scheme == "" || u.Host == "" {
		return "", fmt.Errorf("invalid base URL: must be an absolute URL: %s", c.Options.BaseUrl)
	}

	u.Path = req.Path.String()
	u.RawQuery = req.Query.Encode()

	return u.String(), nil
}
