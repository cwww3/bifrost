package clientx

import (
	"bytes"
	"context"
	"crypto/tls"
	"net"
	"net/http"

	"github.com/valyala/fasthttp"
)

type ConnManager interface {
	Wrap(net.Conn) net.Conn
	GetDump() ([]byte, []byte)
}

type FastHttpDoer interface {
	Do(req *fasthttp.Request, resp *fasthttp.Response) error
}

// FastHttpClientWrap 需要保证顺序请求 一个 client 同一时间 只有一个连接  只有一个请求
type FastHttpClientWrap struct {
	client   *fasthttp.Client
	request  bytes.Buffer
	response bytes.Buffer
}

func (c *FastHttpClientWrap) Do(req *fasthttp.Request, resp *fasthttp.Response) error {
	return c.client.Do(req, resp)
}

type HttpDoer interface {
	Do(req *http.Request) (*http.Response, error)
}

func WrapFastHttpClient(client *fasthttp.Client, cm ConnManager) *FastHttpClientWrap {
	c := new(FastHttpClientWrap)

	config := &tls.Config{InsecureSkipVerify: true}

	client.Dial = func(addr string) (net.Conn, error) {
		conn, err := fasthttp.Dial(addr)
		if err != nil {
			return nil, err
		}

		// 尝试 TLS 握手
		tlsConn := tls.Client(conn, config)
		err = tlsConn.Handshake()
		if err != nil {
			// tls 握手失败 返回原始连接
			conn, err = fasthttp.Dial(addr)
			if err != nil {
				return nil, err
			}
			return cm.Wrap(conn), nil
		}

		return cm.Wrap(tlsConn), nil
	}

	c.client = client
	return c
}

// HttpClientWrap 需要保证顺序请求 一个 client 同一时间 只有一个连接  只有一个请求
type HttpClientWrap struct {
	client   *http.Client
	request  bytes.Buffer
	response bytes.Buffer
}

func WrapHttpClient(cm ConnManager) *HttpClientWrap {
	c := new(HttpClientWrap)

	config := &tls.Config{InsecureSkipVerify: true}
	netTransport := &http.Transport{
		DialContext: func(ctx context.Context, network, addr string) (net.Conn, error) {
			var d net.Dialer
			conn, err := d.DialContext(ctx, network, addr)
			if err != nil {
				return nil, err
			}

			return cm.Wrap(conn), nil
		},
		DialTLSContext: func(ctx context.Context, network, addr string) (net.Conn, error) {
			conn, err := tls.Dial(network, addr, config)
			if err != nil {
				return nil, err
			}

			return cm.Wrap(conn), nil
		},
		TLSClientConfig: config, // 跳过tls认证
	}
	c.client = &http.Client{Transport: netTransport}
	return c
}

func (c *HttpClientWrap) Do(req *http.Request) (resp *http.Response, err error) {
	return c.client.Do(req)
}
