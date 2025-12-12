package main

import (
	"bytes"
	"context"
	"fmt"
	"net"
	"os"
	"time"

	"github.com/cwww3/bifrost"
	"github.com/cwww3/bifrost/providers/utils/clientx"
	"github.com/cwww3/bifrost/schemas"
)

type MyAccount struct {
	cm clientx.ConnManager
}

// Account interface needs to implement these 3 methods
func (a *MyAccount) GetConfiguredProviders() ([]schemas.ModelProvider, error) {
	return []schemas.ModelProvider{schemas.OpenAI}, nil
}

func (a *MyAccount) GetKeysForProvider(ctx *context.Context, provider schemas.ModelProvider) ([]schemas.Key, error) {
	if provider == schemas.OpenAI {
		return []schemas.Key{{
			Value:  os.Getenv("OPENAI_API_KEY"),
			Models: []string{}, // Keep Models empty to use any model
			Weight: 1.0,
		}}, nil
	}
	return nil, fmt.Errorf("provider %s not supported", provider)
}

func (a *MyAccount) GetConfigForProvider(provider schemas.ModelProvider) (*schemas.ProviderConfig, error) {
	c := schemas.NetworkConfig{
		DefaultRequestTimeoutInSeconds: schemas.DefaultRequestTimeoutInSeconds,
		MaxRetries:                     schemas.DefaultMaxRetries,
		RetryBackoffInitial:            schemas.DefaultRetryBackoffInitial,
		RetryBackoffMax:                schemas.DefaultRetryBackoffMax,
	}
	c.BaseURL = "https://www.baidu.com"

	return &schemas.ProviderConfig{
		NetworkConfig:            c,
		ConcurrencyAndBufferSize: schemas.DefaultConcurrencyAndBufferSize,
		ConnManager:              a.cm,
	}, nil
}

func main() {
	var cm clientx.ConnManager = &connManager{}
	client, initErr := bifrost.Init(context.Background(), schemas.BifrostConfig{
		Account: &MyAccount{cm: cm},
	})
	if initErr != nil {
		panic(initErr)
	}
	defer client.Shutdown()

	messages := []schemas.ChatMessage{
		{
			Role: schemas.ChatMessageRoleUser,
			Content: &schemas.ChatMessageContent{
				ContentStr: schemas.Ptr("Hello, Bifrost!"),
			},
		},
	}

	ctx := context.Background()

	response, err := client.ChatCompletionRequest(ctx, &schemas.BifrostChatRequest{
		Provider: schemas.OpenAI,
		Model:    "gpt-4o-mini",
		Input:    messages,
	})

	a, b := cm.GetDump()
	fmt.Println(string(a))
	fmt.Println(string(b))

	if err != nil {
		panic(err)
	}

	fmt.Println("Response:", *response.Choices[0].Message.Content.ContentStr)
}

type connManager struct {
	conn *Conn
}

func (c *connManager) Wrap(conn net.Conn) net.Conn {
	wc := &Conn{
		conn: conn,
	}
	c.conn = wc
	return wc
}

func (c *connManager) GetDump() ([]byte, []byte) {
	if c.conn == nil {
		return nil, nil
	}

	request, response := c.conn.request.Bytes(), c.conn.response.Bytes()
	c.conn.request.Reset()
	c.conn.response.Reset()
	return request, response
}

type Conn struct {
	conn     net.Conn
	request  bytes.Buffer
	response bytes.Buffer
}

func (c *Conn) Read(b []byte) (n int, err error) {
	c.response.Write(b)
	return c.conn.Read(b)
}

func (c *Conn) Write(b []byte) (n int, err error) {
	c.request.Write(b)
	return c.conn.Write(b)
}

func (c *Conn) Close() error {
	return c.conn.Close()
}

func (c *Conn) LocalAddr() net.Addr {
	return c.conn.LocalAddr()
}

func (c *Conn) RemoteAddr() net.Addr {
	return c.conn.RemoteAddr()
}

func (c *Conn) SetDeadline(t time.Time) error {
	return c.conn.SetDeadline(t)
}

func (c *Conn) SetReadDeadline(t time.Time) error {
	return c.conn.SetDeadline(t)
}

func (c *Conn) SetWriteDeadline(t time.Time) error {
	return c.conn.SetWriteDeadline(t)
}

func (c *Conn) Handshake() error {
	return nil
}
