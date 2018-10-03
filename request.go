package request

import (
	"bytes"
	"crypto/tls"
	"encoding/base64"
	"encoding/json"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// ResponseStatus は、Send関数コール時のHTTPステータス管理用構造体
type ResponseStatus struct {
	Code    int
	Message string
}

// Error はエラーメッセージを返却する関数
func (status *ResponseStatus) Error() string {
	return status.Message
}

// HeaderType 型は、HTTPヘッダを構築する型
type HeaderType map[string]string

// Add は、データを追加する
func (header HeaderType) Add(key string, value string) {
	header[key] = value
}

// Del は、指定されたキーを削除する
func (header HeaderType) Del(key string) {
	if _, ok := header[key]; ok {
		delete(header, key)
	}
}

// Get は、指定されたキーが所持する情報を返却する
func (header HeaderType) Get(key string) string {
	if v, ok := header[key]; ok {
		return v
	}
	return ""
}

// Clear 関数は、ヘッダを削除する
func (header HeaderType) Clear() {
	header = make(HeaderType)
}

// Request 構造体は、リクエストを送信する構造体
type Request struct {
	URL      string      // リクエスト送信先のURL
	Username string      // 送信先URLのベーシック認証ユーザID
	Password string      // 送信先URLのベーシック認証パスワード
	Timeout  int         // 送信先URLのタイムアウト時間
	Insecure bool        // 送信先URLが自己証明書の場合でも、送信できるようにするフラグ
	Proxy    Proxy       // プロキシサーバ設定情報
	values   *url.Values // 送信するデータ
	header   HeaderType  // HTTPヘッダ
}

// Proxy 構造体は、プロキシサーバの情報を取り扱う構造体
type Proxy struct {
	URL      string // プロキシサーバのURL
	Username string // プロキシサーバの認証ユーザID
	Password string // プロキシサーバの認証パスワード
}

// Header は、HTTPヘッダのヘッダ情報を返却する
func (r *Request) Header() HeaderType {
	if r.header == nil {
		r.header = make(HeaderType)
	}
	return r.header
}

// Values は、サーバへ送信するデータ情報を返却する
func (r *Request) Values() url.Values {
	if r.values == nil {
		r.values = &url.Values{}
	}
	return *r.values
}

// Request 関数は、リクエスト情報を作成する
func (r *Request) Request(method, username, password string, data io.Reader) (*http.Request, error) {
	// 指定されたURLへアクセスする *http.Request を生成する
	req, err := http.NewRequest(method, r.URL, data)
	if err != nil {
		return nil, err
	}

	// HTTPヘッダが設定されている場合は、設定を実施する
	for k, v := range r.header {
		req.Header.Set(k, v)
	}

	// ベーシック認証の場合は、ユーザID、パスワードを設定
	if username != "" && password != "" {
		req.SetBasicAuth(username, password)
	}

	// 生成したリクエスト情報構造体を返却する
	return req, nil
}

// Transport 構造体は、プロキシやHTTPSの設定情報をTransport構造体に設定する
func (r *Request) Transport(req *http.Request) (*http.Transport, error) {
	var transport = &http.Transport{}
	// プロキシサーバを介する場合の設定
	if r.Proxy.URL != "" {
		// Transport にプロキシサーバ情報を登録
		proxy, err := url.Parse(r.Proxy.URL)
		if err != nil {
			return nil, err
		}
		transport.Proxy = http.ProxyURL(proxy)
		// プロキシサーバの認証設定
		if r.Proxy.Username != "" && r.Proxy.Password != "" {
			auth := r.Proxy.Username + ":" + r.Proxy.Password
			basic := "Basic " + base64.StdEncoding.EncodeToString([]byte(auth))
			req.Header.Add("Proxy-Ahthorization", basic)
		}
	}

	// httpsでかつ、自己証明書を許可する場合
	if strings.Index(r.URL, "https://") == 0 && r.Insecure {
		transport.TLSClientConfig = &tls.Config{InsecureSkipVerify: r.Insecure}
	}

	// 設定したTransport構造体を返却する
	return transport, nil
}

// Send は、リクエスト情報をサーバへ送信する
func (r *Request) Send(req *http.Request, transport *http.Transport) ([]byte, *http.Response, error) {
	// タイムアウト時間を設定
	var timeout = r.Timeout
	if timeout <= 0 {
		timeout = 10
	}

	// リクエスト送信用インスタンスを生成
	client := http.Client{
		Timeout:   time.Duration(timeout) * time.Millisecond,
		Transport: transport,
	}

	// リクエストを送信し、結果を受け取る
	res, err := client.Do(req)
	if err != nil {
		return nil, nil, err
	}
	// 結果取得後、必ずBodyをクローズ
	defer res.Body.Close()

	// 取得したデータを読み取る
	buf, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, nil, err
	}

	// 接続したが、正常に値を取得できなかった場合
	if !(res.StatusCode >= 200 && res.StatusCode <= 299) {
		return buf, res, &ResponseStatus{Code: res.StatusCode, Message: res.Status}
	}

	return buf, res, nil
}

// Get 関数は、GET メソッドでリクエスト情報を送信する
func (r *Request) Get() ([]byte, *http.Response, error) {
	// クエリパラメータが設定されている場合、パースする
	if idx := strings.Index(r.URL, "?"); idx != -1 {
		if u, err := url.Parse(r.URL); err == nil {
			for k, v1 := range u.Query() {
				for _, v2 := range v1 {
					r.Values().Add(k, v2)
				}
			}
		}
		r.URL = r.URL[:idx]
	}

	// 送信するリクエスト情報の構築
	req, err := r.Request("GET", r.Username, r.Password, nil)
	if err != nil {
		return nil, nil, err
	}

	// 送信データが存在する場合は、設定する
	if r.values != nil {
		req.URL.RawQuery = r.values.Encode()
	}

	// Transportの設定
	transport, err := r.Transport(req)
	if err != nil {
		return nil, nil, err
	}

	// リクエストを送信する
	return r.Send(req, transport)
}

// Post は、POSTメソッドでサーバに情報を送信する
func (r *Request) Post() ([]byte, *http.Response, error) {
	return r.Submit("POST")
}

// Put 関数は、PUTメソッドでリクエストを投げる
func (r *Request) Put() ([]byte, *http.Response, error) {
	return r.Submit("PUT")
}

// Delete 関数は、DELETE メソッドでリクエスト投げる
func (r *Request) Delete() ([]byte, *http.Response, error) {
	return r.Submit("DELETE")
}

// Patch 関数は、PATCHメソッドでリクエストを投げる
func (r *Request) Patch() ([]byte, *http.Response, error) {
	return r.Submit("PATCH")
}

// Submit は、指定されたメソッドでサーバへリクエストを送信する
func (r *Request) Submit(method string) ([]byte, *http.Response, error) {
	// クエリパラメータが設定されている場合、パースする
	var rawQuery string
	if idx := strings.Index(r.URL, "?"); idx != -1 {
		values := url.Values{}
		if u, err := url.Parse(r.URL); err == nil {
			for k, v1 := range u.Query() {
				for _, v2 := range v1 {
					values.Add(k, v2)
				}
			}
			rawQuery = values.Encode()
		}
		r.URL = r.URL[:idx]
	}

	// ヘッダ情報が空の場合は設定する
	if r.header == nil || r.Header().Get("Content-Type") == "" {
		r.Header().Add("Content-Type", "application/x-www-form-urlencoded")
	}

	// 送信するデータを整形する
	var data io.Reader
	if r.values != nil {
		switch r.header.Get("Content-Type") {
		// JSON タイプを指定している場合
		case "application/json", "text/json", "text/x-json":
			if buf, err := json.Marshal(r.values); err == nil {
				data = bytes.NewBuffer(buf)
			}
		// 上記以外の場合
		default:
			data = strings.NewReader(r.values.Encode())
		}
	}

	// 送信するリクエスト情報を構築
	req, err := r.Request(method, r.Username, r.Password, data)
	if err != nil {
		return nil, nil, err
	}

	if rawQuery != "" {
		req.URL.RawQuery = rawQuery
	}

	// Transportの設定
	transport, err := r.Transport(req)
	if err != nil {
		return nil, nil, err
	}

	// リクエストを送信する
	return r.Send(req, transport)
}

// JSON は、JSON文字列を整形し、サーバへリクエストを飛ばす準備をする
func (r *Request) JSON(j []byte) error {
	values := make(map[string]string)
	if err := json.Unmarshal(j, &values); err != nil {
		return err
	}

	for k, v := range values {
		r.Values().Add(k, v)
	}

	r.Header().Add("Content-Type", "application/json")
	return nil
}
