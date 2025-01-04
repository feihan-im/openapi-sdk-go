package fhcore

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	urlLib "net/url"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/feihan-im/openapi-sdk-go/internal/model"
	"github.com/gogo/protobuf/proto"
)

type ApiClient interface {
	Preheat(ctx context.Context) error
	Request(ctx context.Context, req *ApiRequest) (*ApiResponse, error)
	OnEvent(eventType string, handler EventHandler)
	OffEvent(eventType string, handler EventHandler)
	Close() error
}

type ApiRequest struct {
	Method             string            `json:"method"`
	Path               string            `json:"path"`
	PathParams         map[string]string `json:"path_params,omitempty"`
	QueryParams        map[string]string `json:"query_params,omitempty"`
	HeaderParams       map[string]string `json:"header_params,omitempty"`
	Body               interface{}       `json:"body,omitempty"`
	Stream             io.Reader         `json:"stream,omitempty"`
	WithAppAccessToken bool              `json:"with_app_access_token,omitempty"`
	WithWebSocket      bool              `json:"with_websocket,omitempty"`
}

type ApiResponse struct {
	config  *Config
	GetBody func() ([]byte, error)
}

type ApiError struct {
	Code  int         `json:"code"` // code < 0 means local error
	Msg   string      `json:"msg"`
	LogId string      `json:"log_id"`
	Data  interface{} `json:"data,omitempty"`
}

func (e *ApiError) Error() string {
	return fmt.Sprintf("ApiError[code=%d, logId=%s, msg=%s]", e.Code, e.LogId, e.Msg)
}

type defaultApiClient struct {
	config         *Config
	secret         string
	token          string
	tokenRefreshAt int64
	tokenExpiresAt int64
	tokenFetching  uint64
	tokenMu        sync.RWMutex
	pingCalled     bool
	pingExpiresAt  time.Time
	pingFetching   uint64
	pingMu         sync.RWMutex
	cryptoManager  *defaultCryptoManager
	ws             *defaultWsClient
}

func NewDefaultApiClient(config *Config) ApiClient {
	secretBytes := sha256.Sum256([]byte(fmt.Sprintf(
		"%s:%s",
		config.AppId,
		config.AppSecret,
	)))
	secret := hex.EncodeToString(secretBytes[:])
	c := &defaultApiClient{
		config: config,
		secret: secret,
	}
	c.cryptoManager = newDefaultCryptoManager(config)
	c.ws = newDefaultWsClient(&defaultWsClientOptions{
		config:        config,
		secret:        secret,
		getToken:      c.getToken,
		ensurePing:    c.ensurePing,
		cryptoManager: c.cryptoManager,
	})
	return c
}

func (c *defaultApiClient) Preheat(ctx context.Context) error {
	err := c.ensurePing(ctx)
	if err != nil {
		return err
	}
	_, err = c.getToken(ctx)
	if err != nil {
		return err
	}
	return nil
}

func (c *defaultApiClient) Close() error {
	return c.ws.Close()
}

func (c *defaultApiClient) OnEvent(eventType string, handler EventHandler) {
	c.ws.OnEvent(eventType, handler)
}

func (c *defaultApiClient) OffEvent(eventType string, handler EventHandler) {
	c.ws.OffEvent(eventType, handler)
}

func (c *defaultApiClient) Request(ctx context.Context, req *ApiRequest) (*ApiResponse, error) {
	err := c.ensurePing(ctx)
	if err != nil {
		return nil, err
	}

	var url string

	// build url
	{
		urlBuilder := strings.Builder{}
		urlBuilder.Grow(len(c.config.BackendUrl) + len(req.Path))
		urlBuilder.WriteString(c.config.BackendUrl)
		for i, j := 0, 0; j <= len(req.Path); j++ {
			if j == len(req.Path) || req.Path[j] == '/' {
				if i < j {
					urlBuilder.WriteRune('/')
				}
				if req.Path[i] == ':' {
					if req.PathParams == nil {
						return nil, &ApiError{Code: -1, Msg: "PathParams is nil"}
					}
					v, ok := req.PathParams[req.Path[i+1:j]]
					if !ok || len(v) == 0 {
						return nil, &ApiError{Code: -1, Msg: fmt.Sprintf("PathParams[%s] is not found or empty", req.Path[i+1:j])}
					}
					urlBuilder.WriteString(v)
				} else {
					urlBuilder.WriteString(req.Path[i:j])
				}
				i = j + 1
			}
		}
		if len(req.QueryParams) > 0 {
			urlBuilder.WriteRune('?')
			and := false
			for k, v := range req.QueryParams {
				if and {
					urlBuilder.WriteRune('&')
				}
				urlBuilder.WriteString(urlLib.QueryEscape(k))
				urlBuilder.WriteRune('=')
				urlBuilder.WriteString(urlLib.QueryEscape(v))
				and = true
			}
		}
		url = urlBuilder.String()
	}

	if c.config.EnableEncryption && req.WithAppAccessToken {
		// encryption

		// build body
		var body []byte
		{
			if req.Stream != nil {
				b, err := ioutil.ReadAll(req.Stream)
				if err != nil {
					return nil, &ApiError{Code: -1, Msg: err.Error()}
				}
				body = b
			} else if b, ok := req.Body.([]byte); ok {
				body = b
			} else {
				b, err := c.config.JsonMarshal(req.Body)
				if err != nil {
					return nil, &ApiError{Code: -1, Msg: err.Error()}
				}
				body = b
			}
		}

		// build new body
		{
			httpReq := &model.HttpRequest{
				Method:  req.Method,
				Path:    url[len(c.config.BackendUrl):],
				Headers: req.HeaderParams,
				Body:    body,
			}

			// send by websocket
			if req.WithWebSocket {
				if httpReq.Headers == nil {
					httpReq.Headers = map[string]string{}
				}
				if len(httpReq.Headers["Authorization"]) == 0 {
					token, err := c.getToken(ctx)
					if err != nil {
						return nil, err
					}
					httpReq.Headers["Authorization"] = "Bearer " + token
				}
				httpResp, err := c.ws.HttpRequest(ctx, httpReq)
				if err != nil {
					return nil, err
				}
				return &ApiResponse{
					config: c.config,
					GetBody: func() ([]byte, error) {
						return httpResp.Body, nil
					},
				}, nil
			}

			httpReqBody, err := proto.Marshal(httpReq)
			if err != nil {
				return nil, &ApiError{Code: -1, Msg: err.Error()}
			}
			secureMessage, err := c.cryptoManager.encryptMessage(c.secret, httpReqBody)
			if err != nil {
				return nil, &ApiError{Code: -1, Msg: err.Error()}
			}
			body, _ = proto.Marshal(secureMessage)
		}

		// build request
		httpReq, err := http.NewRequest("POST", c.config.BackendUrl+defaultGatewayPath, bytes.NewBuffer(body))
		if err != nil {
			return nil, &ApiError{Code: -1, Msg: err.Error()}
		}

		// build header
		httpReq.Header.Add("User-Agent", UserAgent)
		token, err := c.getToken(ctx)
		if err != nil {
			return nil, err
		}
		httpReq.Header.Add("Authorization", "Bearer "+token)

		// request
		httpResp, err := c.config.HttpClient.Do(httpReq)
		if err != nil {
			return nil, &ApiError{Code: -1, Msg: err.Error()}
		}

		return &ApiResponse{
			config: c.config,
			GetBody: func() ([]byte, error) {
				defer httpResp.Body.Close()
				body, err := ioutil.ReadAll(httpResp.Body)
				if err != nil {
					return nil, &ApiError{Code: -1, Msg: err.Error()}
				}
				// decrypt
				var secureMessage model.SecureMessage
				err = proto.Unmarshal(body, &secureMessage)
				if err != nil {
					return nil, &ApiError{Code: -1, Msg: err.Error()}
				}
				data, err := c.cryptoManager.decryptMessage(c.secret, &secureMessage)
				if err != nil {
					return nil, &ApiError{Code: -1, Msg: err.Error()}
				}
				var httpResponse model.HttpResponse
				err = proto.Unmarshal(data, &httpResponse)
				if err != nil {
					return nil, &ApiError{Code: -1, Msg: err.Error()}
				}
				return httpResponse.Body, nil
			},
		}, nil
	} else {
		// no encryption

		// build body
		var body io.Reader
		{
			if req.Stream != nil {
				body = req.Stream
			} else if b, ok := req.Body.([]byte); ok {
				body = bytes.NewBuffer(b)
			} else {
				b, err := c.config.JsonMarshal(req.Body)
				if err != nil {
					return nil, &ApiError{Code: -1, Msg: err.Error()}
				}
				body = bytes.NewBuffer(b)
			}
		}

		// build request
		httpReq, err := http.NewRequest(req.Method, url, body)
		if err != nil {
			return nil, &ApiError{Code: -1, Msg: err.Error()}
		}

		// build header
		httpReq.Header.Add("Content-Type", "application/json")
		httpReq.Header.Add("User-Agent", UserAgent)
		httpReq.Header.Add("X-Feihan-Timestamp", strconv.FormatInt(c.config.TimeManager.GetServerTimestamp(), 10))
		httpReq.Header.Add("X-Feihan-Nonce", c.cryptoManager.getNonce())

		if req.WithAppAccessToken {
			token, err := c.getToken(ctx)
			if err != nil {
				return nil, err
			}
			httpReq.Header.Add("Authorization", "Bearer "+token)
		}

		for k, v := range req.HeaderParams {
			httpReq.Header.Add(k, v)
		}

		// request
		httpResp, err := c.config.HttpClient.Do(httpReq)
		if err != nil {
			return nil, &ApiError{Code: -1, Msg: err.Error()}
		}

		return &ApiResponse{
			config: c.config,
			GetBody: func() ([]byte, error) {
				defer httpResp.Body.Close()
				body, err := ioutil.ReadAll(httpResp.Body)
				if err != nil {
					return nil, &ApiError{Code: -1, Msg: err.Error()}
				}
				return body, nil
			},
		}, nil
	}
}

type apiResp struct {
	Code  int         `json:"code"`
	Msg   string      `json:"msg"`
	LogId string      `json:"log_id"`
	Data  interface{} `json:"data,omitempty"`
}

func (e *ApiResponse) JSON(value interface{}) error {
	body, err := e.GetBody()
	if err != nil {
		return err
	}
	var data apiResp
	data.Data = value
	err = e.config.JsonUnmarshal(body, &data)
	if err != nil {
		return &ApiError{Code: -1, Msg: err.Error()}
	}
	if data.Code != 0 {
		return &ApiError{
			Code:  data.Code,
			Msg:   data.Msg,
			LogId: data.LogId,
			Data:  data.Data,
		}
	}
	return nil
}

type tokenResp struct {
	Code  int    `json:"code"`
	Msg   string `json:"msg"`
	LogId string `json:"log_id"`
	Data  *struct {
		AppAccessToken          string `json:"app_access_token"`
		AppAccessTokenExpiresIn int    `json:"app_access_token_expires_in"`
	} `json:"data,omitempty"`
}

func (c *defaultApiClient) getToken(ctx context.Context) (string, error) {
	c.tokenMu.RLock()
	token := c.token
	tokenExpiresAt := c.tokenExpiresAt
	tokenRefreshAt := c.tokenRefreshAt
	c.tokenMu.RUnlock()

	timestamp := c.config.TimeManager.GetServerTimestamp()
	if tokenRefreshAt > timestamp {
		return token, nil
	}

	callFetchToken := func(lock bool) {
		resp, err := c.fetchToken()
		if err != nil {
			c.config.Logger.Errorf(ctx, "Get token error: %s", err.Error())
		} else {
			if lock {
				c.tokenMu.Lock()
			}
			c.token = resp.Data.AppAccessToken
			c.tokenExpiresAt =
				c.config.TimeManager.GetServerTimestamp() + int64(resp.Data.AppAccessTokenExpiresIn*1000) - 60*1000
			c.tokenRefreshAt = c.tokenExpiresAt - 5*60*1000
			if lock {
				c.tokenMu.Unlock()
			}
			c.config.Logger.Infof(ctx, "Get token successfully: token_expires_in=%d", resp.Data.AppAccessTokenExpiresIn)
		}
	}

	if len(token) == 0 || tokenExpiresAt <= timestamp {
		c.tokenMu.Lock()
		token = c.token
		if len(token) == 0 || c.tokenExpiresAt <= c.config.TimeManager.GetServerTimestamp() {
			callFetchToken(false)
		}
		token = c.token
		c.tokenMu.Unlock()
	} else {
		if atomic.CompareAndSwapUint64(&c.tokenFetching, 0, 1) {
			go func() {
				callFetchToken(true)
				atomic.StoreUint64(&c.tokenFetching, 0)
			}()
		}
	}

	if len(token) == 0 {
		return "", &ApiError{Code: -1, Msg: "Get token failed"}
	}

	return token, nil
}

func (c *defaultApiClient) fetchToken() (*tokenResp, error) {
	timestamp := c.config.TimeManager.GetServerTimestamp()
	nonce := randIntn(1e12)

	s := sha256.New()
	_, _ = s.Write([]byte(fmt.Sprintf(
		"%s:%d:%s:%d",
		c.config.AppId,
		timestamp,
		c.config.AppSecret,
		nonce,
	)))
	signature := hex.EncodeToString(s.Sum(nil))

	url := c.config.BackendUrl + defaultTokenPath
	body, _ := c.config.JsonMarshal(map[string]interface{}{
		"app_id":            c.config.AppId,
		"signature_version": "v1",
		"signature":         signature,
		"timestamp":         timestamp,
		"nonce":             nonce,
	})

	httpReq, err := http.NewRequest(http.MethodPost, url, bytes.NewBuffer(body))
	if err != nil {
		return nil, &ApiError{Code: -1, Msg: err.Error()}
	}

	httpReq.Header.Add("Content-Type", "application/json")
	httpReq.Header.Add("User-Agent", UserAgent)

	httpResp, err := c.config.HttpClient.Do(httpReq)
	if err != nil {
		return nil, &ApiError{Code: -1, Msg: err.Error()}
	}

	defer httpResp.Body.Close()
	b, err := ioutil.ReadAll(httpResp.Body)
	if err != nil {
		return nil, &ApiError{Code: -1, Msg: err.Error()}
	}

	var resp tokenResp
	err = c.config.JsonUnmarshal(b, &resp)
	if err != nil {
		return nil, &ApiError{Code: -1, Msg: err.Error()}
	}
	if resp.Code != 0 {
		return nil, &ApiError{
			Code:  resp.Code,
			Msg:   resp.Msg,
			LogId: resp.LogId,
		}
	}

	return &resp, nil
}

type pingResp struct {
	Code  int    `json:"code"`
	Msg   string `json:"msg"`
	LogId string `json:"log_id"`
	Data  *struct {
		Version   string `json:"version"`
		Timestamp int64  `json:"timestamp"`
		OrgCode   string `json:"org_code"`
	} `json:"data,omitempty"`
}

func (c *defaultApiClient) ensurePing(ctx context.Context) error {
	c.pingMu.RLock()
	pingCalled := c.pingCalled
	pingExpiresAt := c.pingExpiresAt
	c.pingMu.RUnlock()

	if pingExpiresAt.After(time.Now()) {
		return nil
	}

	callFetchPing := func(lock bool) {
		resp, err := c.fetchPing()
		if err != nil {
			c.config.Logger.Errorf(ctx, "Ping server error: %s", err.Error())
		} else {
			if lock {
				c.pingMu.Lock()
			}
			c.pingCalled = true
			c.pingExpiresAt = time.Now().Add(time.Hour)
			if lock {
				c.pingMu.Unlock()
			}
			c.config.TimeManager.SyncServerTimestamp(resp.Data.Timestamp)
			c.config.Logger.Infof(
				ctx,
				"Ping server successfully: org_code=%s, server_version=%s, server_time=%s",
				resp.Data.OrgCode,
				resp.Data.Version,
				time.Unix(resp.Data.Timestamp/1000, resp.Data.Timestamp%1000).Format(time.RFC3339),
			)
		}
	}

	if !pingCalled {
		c.pingMu.Lock()
		pingCalled = c.pingCalled
		if !pingCalled {
			callFetchPing(false)
		}
		pingCalled = c.pingCalled
		c.pingMu.Unlock()
	} else {
		if atomic.CompareAndSwapUint64(&c.pingFetching, 0, 1) {
			go func() {
				callFetchPing(true)
				atomic.StoreUint64(&c.tokenFetching, 0)
			}()
		}
	}

	if !pingCalled {
		return &ApiError{Code: -1, Msg: "Ping server failed"}
	}

	return nil
}

func (c *defaultApiClient) fetchPing() (*pingResp, error) {
	url := c.config.BackendUrl + defaultPingPath
	httpReq, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, &ApiError{Code: -1, Msg: err.Error()}
	}

	httpReq.Header.Add("Content-Type", "application/json")
	httpReq.Header.Add("User-Agent", UserAgent)

	httpResp, err := c.config.HttpClient.Do(httpReq)
	if err != nil {
		return nil, &ApiError{Code: -1, Msg: err.Error()}
	}

	defer httpResp.Body.Close()
	b, err := ioutil.ReadAll(httpResp.Body)
	if err != nil {
		return nil, &ApiError{Code: -1, Msg: err.Error()}
	}

	var resp pingResp
	err = c.config.JsonUnmarshal(b, &resp)
	if err != nil {
		return nil, &ApiError{Code: -1, Msg: err.Error()}
	}
	if resp.Code != 0 {
		return nil, &ApiError{
			Code:  resp.Code,
			Msg:   resp.Msg,
			LogId: resp.LogId,
		}
	}

	return &resp, nil
}
