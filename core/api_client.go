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
	"strings"
	"sync/atomic"
	"time"

	"github.com/bytedance/sonic"
	"golang.org/x/sync/singleflight"
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
	HttpResponse *http.Response `json:"http_response,omitempty"`
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

const (
	defaultTokenPath = "/oapi/auth/v1/app/token"
)

type defaultApiClient struct {
	config         *Config
	token          atomic.Value
	tokenExpiresAt atomic.Value
	tokenSingle    singleflight.Group
}

func NewDefaultApiClient(config *Config) ApiClient {
	return &defaultApiClient{
		config: config,
	}
}

func (c *defaultApiClient) Request(ctx context.Context, req *ApiRequest) (*ApiResponse, *ApiError) {
	var url string
	var body io.Reader

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

	// build body
	{
		if req.Stream != nil {
			body = req.Stream
		} else if b, ok := req.Body.([]byte); ok {
			body = bytes.NewBuffer(b)
		} else {
			b, err := sonic.Marshal(req.Body)
			if err != nil {
				return nil, &ApiError{Code: -1, Msg: err.Error()}
			}
			body = bytes.NewBuffer(b)
		}
	}

	httpReq, err := http.NewRequestWithContext(ctx, req.Method, url, body)
	if err != nil {
		return nil, &ApiError{Code: -1, Msg: err.Error()}
	}

	// build header
	httpReq.Header.Add("Content-Type", "application/json")
	httpReq.Header.Add("User-Agent", UserAgent)

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

	httpResp, err := c.config.HttpClient.Do(httpReq)
	if err != nil {
		return nil, &ApiError{Code: -1, Msg: err.Error()}
	}

	return &ApiResponse{
		HttpResponse: httpResp,
	}, nil
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
		_, err, _ := c.tokenSingle.Do("getToken", func() (interface{}, error) {
			f := func() error {
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
				body, _ := sonic.Marshal(map[string]interface{}{
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
				err = sonic.Unmarshal(b, &resp)
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
				return nil, err
			}
			return "", nil
		})
		if err != nil {
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
	defer e.HttpResponse.Body.Close()
	b, err := io.ReadAll(e.HttpResponse.Body)
	if err != nil {
		return &ApiError{Code: -1, Msg: err.Error()}
	}
	var data struct {
		Code  int         `json:"code"`
		Msg   string      `json:"msg"`
		LogId string      `json:"log_id"`
		Data  interface{} `json:"data,omitempty"`
	}
	data.Data = value
	err = sonic.Unmarshal(b, &data)
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
