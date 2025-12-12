package clientx

import (
	"bytes"
	"crypto/tls"
	"net"

	"github.com/valyala/fasthttp"
)

type FastHttpDoer interface {
	Do(req *fasthttp.Request, resp *fasthttp.Response) error
}

// Client 需要保证顺序请求 一个 client 同一时间 只有一个连接  只有一个请求
type Client struct {
	client   *fasthttp.Client
	request  bytes.Buffer
	response bytes.Buffer
}

func (c *Client) Do(req *fasthttp.Request, resp *fasthttp.Response) error {
	return c.client.Do(req, resp)
}

func WrapClient(client *fasthttp.Client, cm ConnManager) *Client {
	c := new(Client)

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

type ConnManager interface {
	Wrap(net.Conn) net.Conn
	GetDump() ([]byte, []byte)
}
