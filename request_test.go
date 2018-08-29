package request

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"
)

var handleFunc = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "GET":
		if r.URL.Path == "/name" {
			w.WriteHeader(404)
			w.Header().Set("Content-Type", "text/plain")
			w.Write([]byte("404 not found"))
			return
		}
		if r.UserAgent() == "TEST - BROWSER" && r.URL.RawQuery == "key=value&test=true" {
			fmt.Fprintf(w, "SUCCESS")
		}
	case "POST":
		if r.URL.Path == "/name" {
			w.WriteHeader(404)
			w.Header().Set("Content-Type", "text/plain")
			w.Write([]byte("404 not found"))
			return
		}
		r.ParseForm()
		if r.URL.RawQuery == "test=true" {
			fmt.Fprintf(w, "SUCCESS")
		}
	case "PUT":
		length, _ := strconv.Atoi(r.Header.Get("Content-Length"))
		body := make([]byte, length)
		r.Body.Read(body)
		var jsonBody map[string]interface{}
		fmt.Println(string(body))
		json.Unmarshal(body[:length], &jsonBody)
		fmt.Printf("%v\n", jsonBody)
		if r.URL.RawQuery == "test=true" {
			fmt.Fprintf(w, "SUCCESS")
		}
	case "PATCH":
		username, password, _ := r.BasicAuth()
		if username == "username" && password == "password" {
			fmt.Fprintf(w, "SUCCESS")
		}
	case "DELETE":
		fmt.Fprintf(w, "SUCCESS")
	}
})

func Test__GET_HTTPS_SUCCESS(t *testing.T) {
	// https サーバをテストで立てる
	ts := httptest.NewTLSServer(handleFunc)
	defer ts.Close()

	// リクエスト送信用の構造体をセット
	r := Request{
		URL:      ts.URL + "/?test=true",
		Insecure: true,
	}

	// ヘッダをセット
	r.Header().Add("User-Agent", "TEST - BROWSER")
	// 値をセット
	r.Values().Add("key", "value")
	// GETメソッドでリクエストを送信
	res, _, err := r.Get()
	// エラーが発生した場合は、Fatalとする
	if err != nil {
		t.Fatal(err)
	}
	if string(res) != "SUCCESS" {
		t.Fatal("ERROR - GET_HTTPS_SUCCESS Get()")
	}
	if r.Header().Get("User-Agent") != "TEST - BROWSER" {
		t.Fatal("ERROR - GET_HTTPS_SUCCESS Header().Get()")
	}
	r.Header().Del("User-Agent")
	r.Header().Clear()
}

func Test__GET_HTTPS_FAILED(t *testing.T) {
	// https サーバをテストで立てる
	ts := httptest.NewTLSServer(handleFunc)
	defer ts.Close()

	idx := strings.LastIndex(ts.URL, ":")
	var URL = ts.URL[:idx-1]
	// リクエスト送信用の構造体をセット
	r := Request{
		URL:      URL,
		Insecure: true,
	}

	if _, _, err := r.Get(); err == nil {
		t.Fatal("ERROR - GET_HTTPS_FAILED Get()")
	}

	r.URL = "://localhost///path//to/url"
	if _, _, err := r.Get(); err == nil {
		t.Fatal("ERROR - GET_HTTPS_FAILED url.Parse()")
	}
}

func Test__GET_HTTP_FAILED(t *testing.T) {
	// https サーバをテストで立てる
	ts := httptest.NewServer(handleFunc)
	defer ts.Close()

	idx := strings.LastIndex(ts.URL, ":")
	var URL = ts.URL[:idx-1]
	// リクエスト送信用の構造体をセット
	r := Request{
		URL:      URL,
		Insecure: true,
	}

	if _, _, err := r.Get(); err == nil {
		t.Fatal("ERROR - GET_HTTP_FAILED Get()")
	}
}

func Test__GET_HTTPS_NOTFOUND(t *testing.T) {
	// https サーバをテストで立てる
	ts := httptest.NewTLSServer(handleFunc)
	defer ts.Close()

	// リクエスト送信用の構造体をセット
	r := Request{
		URL:      ts.URL + "/name",
		Insecure: true,
	}

	// GETメソッドでリクエストを送信
	_, _, err := r.Get()
	// エラーが発生した場合は、Fatalとする
	if err == nil {
		t.Fatal("ERROR - GET_HTTPS_NOTFOUND Get()")
	}

	fmt.Println(err.Error())
}

func Test__POST_HTTPS_SUCCESS(t *testing.T) {
	// https サーバをテストで立てる
	ts := httptest.NewTLSServer(handleFunc)

	defer ts.Close()

	// リクエスト送信用の構造体をセット
	r := Request{
		URL:      ts.URL + "/?test=true",
		Insecure: true,
	}

	// ヘッダをセット
	r.Header().Add("User-Agent", "TEST - BROWSER")
	// 値をセット
	r.Values().Add("key", "value")

	// GETメソッドでリクエストを送信
	res, _, err := r.Post()
	// エラーが発生した場合は、Fatalとする
	if err != nil {
		t.Fatal(err)
	}
	if string(res) != "SUCCESS" {
		t.Fatal("ERROR - POST_HTTPS_SUCCESS Post()")
	}
	r.Header().Del("User-Agent")
	r.Header().Clear()
}

func Test__PUT_HTTPS_JSON_SUCCESS(t *testing.T) {
	// https サーバをテストで立てる
	ts := httptest.NewTLSServer(handleFunc)
	defer ts.Close()

	// リクエスト送信用の構造体をセット
	r := Request{
		URL:      ts.URL + "/?test=true",
		Insecure: true,
	}

	r.JSON([]byte(`{"key":"value"}`))

	// GETメソッドでリクエストを送信
	res, _, err := r.Put()
	// エラーが発生した場合は、Fatalとする
	if err != nil {
		t.Fatal(err)
	}
	if string(res) != "SUCCESS" {
		t.Fatal("ERROR - PUT_HTTPS_JSON_SUCCESS Put()")
	}

	// JSON形式ではないため、エラーとなる
	if err := r.JSON([]byte("hello world")); err == nil {
		t.Fatal("ERROR - PUT_HTTPS_JSON_SUCCESS JSON()")
	}
}

func Test__PATCH_HTTPS_BASIC_SUCCESS(t *testing.T) {
	// https サーバをテストで立てる
	ts := httptest.NewTLSServer(handleFunc)
	defer ts.Close()

	// リクエスト送信用の構造体をセット
	r := Request{
		URL:      ts.URL + "/?test=true",
		Insecure: true,
		Username: "username",
		Password: "password",
	}

	// GETメソッドでリクエストを送信
	res, _, err := r.Patch()
	// エラーが発生した場合は、Fatalとする
	if err != nil {
		t.Fatal(err)
	}
	if string(res) != "SUCCESS" {
		t.Fatal("ERROR - PUT_HTTPS_JSON_SUCCESS Put()")
	}
}

func Test__DELETE_HTTPS_SUCCESS(t *testing.T) {
	// https サーバをテストで立てる
	ts := httptest.NewTLSServer(handleFunc)
	defer ts.Close()

	// リクエスト送信用の構造体をセット
	r := Request{
		URL:      ts.URL + "/?test=true",
		Insecure: true,
	}

	// GETメソッドでリクエストを送信
	res, _, err := r.Delete()
	// エラーが発生した場合は、Fatalとする
	if err != nil {
		t.Fatal(err)
	}
	if string(res) != "SUCCESS" {
		t.Fatal("ERROR - PUT_HTTPS_JSON_SUCCESS Put()")
	}
}
