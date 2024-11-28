package fhcore

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"math/rand/v2"
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
	Request(ctx context.Context, req *ApiRequest) (*ApiResponse, *ApiError)
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
}

type ApiResponse struct {
	httpResponse    *http.Response
	encryptedSecret string
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
	token          atomic.Value
	tokenExpiresAt atomic.Value
	tokenMu        sync.Mutex
}

func NewDefaultApiClient(config *Config) ApiClient {
	secret := sha256.Sum256([]byte(fmt.Sprintf(
		"%s:%s",
		config.AppId,
		config.AppSecret,
	)))
	return &defaultApiClient{
		config: config,
		secret: hex.EncodeToString(secret[:]),
	}
}

func (c *defaultApiClient) Request(ctx context.Context, req *ApiRequest) (*ApiResponse, *ApiError) {
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
					v, ok := req.PathParams[req.Path[i:j]]
					if !ok {
						return nil, &ApiError{Code: -1, Msg: fmt.Sprintf("PathParams[%s] is not found", req.Path[i:j])}
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
				b, err := io.ReadAll(req.Stream)
				if err != nil {
					return nil, &ApiError{Code: -1, Msg: err.Error()}
				}
				body = b
			} else if b, ok := req.Body.([]byte); ok {
				body = b
			} else {
				b, err := JsonMarshal(req.Body)
				if err != nil {
					return nil, &ApiError{Code: -1, Msg: err.Error()}
				}
				body = b
			}
		}

		// build new body
		{
			req := &model.HttpRequest{
				Method:  req.Method,
				Path:    url[len(c.config.BackendUrl):],
				Headers: req.HeaderParams,
				Body:    body,
			}
			reqBody, err := proto.Marshal(req)
			if err != nil {
				return nil, &ApiError{Code: -1, Msg: err.Error()}
			}
			secureMessage, err := encryptMessage(c.secret, reqBody)
			if err != nil {
				return nil, &ApiError{Code: -1, Msg: err.Error()}
			}
			body, _ = proto.Marshal(secureMessage)
		}

		// build request
		httpReq, err := http.NewRequestWithContext(ctx, req.Method, c.config.BackendUrl+defaultGatewayPath, bytes.NewBuffer(body))
		if err != nil {
			return nil, &ApiError{Code: -1, Msg: err.Error()}
		}

		// build header
		httpReq.Header.Add("User-Agent", UserAgent)
		{
			token, err := c.getToken(ctx)
			if err != nil {
				return nil, err
			}
			httpReq.Header.Add("Authorization", "Bearer "+token)
		}

		// request
		httpResp, err := c.config.HttpClient.Do(httpReq)
		if err != nil {
			return nil, &ApiError{Code: -1, Msg: err.Error()}
		}

		return &ApiResponse{
			httpResponse:    httpResp,
			encryptedSecret: c.secret,
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
				b, err := JsonMarshal(req.Body)
				if err != nil {
					return nil, &ApiError{Code: -1, Msg: err.Error()}
				}
				body = bytes.NewBuffer(b)
			}
		}

		// build request
		httpReq, err := http.NewRequestWithContext(ctx, req.Method, url, body)
		if err != nil {
			return nil, &ApiError{Code: -1, Msg: err.Error()}
		}

		// build header
		httpReq.Header.Add("Content-Type", "application/json")
		httpReq.Header.Add("User-Agent", UserAgent)
		httpReq.Header.Add("X-Feihan-Timestamp", strconv.FormatInt(time.Now().UnixMilli(), 10))
		httpReq.Header.Add("X-Feihan-Nonce", randomAlphaNumString(16))

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
			httpResponse: httpResp,
		}, nil
	}
}

func (c *defaultApiClient) getToken(ctx context.Context) (string, *ApiError) {
	loadTokenExpiresAt := func() time.Time {
		v := c.tokenExpiresAt.Load()
		if v == nil {
			return time.Time{}
		}
		return v.(time.Time)
	}

	loadToken := func() string {
		v := c.token.Load()
		if v == nil {
			return ""
		}
		return v.(string)
	}

	if loadTokenExpiresAt().Before(time.Now()) {
		f := func() error {
			c.tokenMu.Lock()
			defer c.tokenMu.Unlock()

			if loadTokenExpiresAt().After(time.Now()) {
				return nil
			}

			ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
			defer cancel()

			timestamp := time.Now().UnixMilli()
			nonce := rand.Int64N(1e12)

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
			body, _ := JsonMarshal(map[string]interface{}{
				"app_id":            c.config.AppId,
				"signature_version": "v1",
				"signature":         signature,
				"timestamp":         timestamp,
				"nonce":             nonce,
			})

			httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewBuffer(body))
			if err != nil {
				return &ApiError{Code: -1, Msg: err.Error()}
			}

			httpReq.Header.Add("Content-Type", "application/json")
			httpReq.Header.Add("User-Agent", UserAgent)

			httpResp, err := c.config.HttpClient.Do(httpReq)
			if err != nil {
				return &ApiError{Code: -1, Msg: err.Error()}
			}

			defer httpResp.Body.Close()
			b, err := io.ReadAll(httpResp.Body)
			if err != nil {
				return &ApiError{Code: -1, Msg: err.Error()}
			}
			var resp struct {
				Code  int    `json:"code"`
				Msg   string `json:"msg"`
				LogId string `json:"log_id"`
				Data  *struct {
					AppAccessToken          string `json:"app_access_token"`
					AppAccessTokenExpiresIn int    `json:"app_access_token_expires_in"`
				} `json:"data,omitempty"`
			}
			err = JsonUnmarshal(b, &resp)
			if err != nil {
				return &ApiError{Code: -1, Msg: err.Error()}
			}
			if resp.Code != 0 {
				return &ApiError{
					Code:  resp.Code,
					Msg:   resp.Msg,
					LogId: resp.LogId,
				}
			}
			c.token.Store(resp.Data.AppAccessToken)
			c.tokenExpiresAt.Store(
				time.Now().
					Add(time.Second * time.Duration(resp.Data.AppAccessTokenExpiresIn)).
					Add(time.Minute * -5),
			)

			return nil
		}
		if err := f(); err != nil {
			c.tokenExpiresAt.Store(time.Now().Add(time.Minute * 5))
			if v, ok := err.(*ApiError); ok {
				return "", v
			}
			return "", &ApiError{Code: -1, Msg: err.Error()}
		}
	}

	token := loadToken()
	if len(token) == 0 {
		return "", &ApiError{Code: -1, Msg: "get token failed"}
	}

	return token, nil
}

func (e *ApiResponse) JSON(value interface{}) *ApiError {
	defer e.httpResponse.Body.Close()
	body, err := io.ReadAll(e.httpResponse.Body)
	if err != nil {
		return &ApiError{Code: -1, Msg: err.Error()}
	}
	// decrypt
	if len(e.encryptedSecret) > 0 {
		var secureMessage model.SecureMessage
		err = proto.Unmarshal(body, &secureMessage)
		if err != nil {
			return &ApiError{Code: -1, Msg: err.Error()}
		}
		data, err := decryptMessage(e.encryptedSecret, &secureMessage)
		if err != nil {
			return &ApiError{Code: -1, Msg: err.Error()}
		}
		var httpResponse model.HttpResponse
		err = proto.Unmarshal(data, &httpResponse)
		if err != nil {
			return &ApiError{Code: -1, Msg: err.Error()}
		}
		body = httpResponse.Body
	}
	var data struct {
		Code  int         `json:"code"`
		Msg   string      `json:"msg"`
		LogId string      `json:"log_id"`
		Data  interface{} `json:"data,omitempty"`
	}
	data.Data = value
	err = JsonUnmarshal(body, &data)
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
