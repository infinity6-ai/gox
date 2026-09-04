package httpzclient

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/infinity6-ai/gox/commonz/deferz"
	"github.com/infinity6-ai/gox/commonz/urlz"
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
	BaseUrl   *urlz.Url
	GetClient func() *http.Client
}

func (o *Options) fix() {
}

type Client struct {
	Options Options
	filters []Filter
	client  *http.Client
}

func New(opts Options) *Client {
	opts.fix()
	ret := &Client{
		Options: opts,
		client:  defaultHttpClient,
	}
	if opts.GetClient != nil {
		ret.client = opts.GetClient()
	}
	return ret
}

func (c *Client) AddFilter(filter Filter) {
	c.filters = append(c.filters, filter)
}

func (c *Client) Do(ctx context.Context, req *Req) (*Resp, error) {
	var h Handler = c.send
	for i := len(c.filters) - 1; i >= 0; i-- {
		filter := c.filters[i]
		next := h
		h = func(ctx context.Context, req *Req) (*Resp, error) {
			return filter(ctx, req, next)
		}
	}
	return h(ctx, req)
}

func (c *Client) send(ctx context.Context, req *Req) (*Resp, error) {
	dfz := deferz.New(ctx)
	defer dfz.Close()

	urlString, err := c.buildURL(req)
	if err != nil {
		return nil, fmt.Errorf("failed to build URL: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, req.Method, urlString, req.Body)
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
	u, err := req.ResolveUrl(c.Options.BaseUrl)
	// u, err := url.Parse(c.Options.BaseUrl)
	if err != nil {
		return "", fmt.Errorf("invalid base URL: %w", err)
	}

	if u.Scheme == "" || u.Host == "" {
		return "", fmt.Errorf("invalid base URL: must be an absolute URL: %s", c.Options.BaseUrl)
	}

	// u.Path = path.Join(u.Path, req.Path.String())
	u.Query = req.Query.Encode()

	return u.String(), nil
}
