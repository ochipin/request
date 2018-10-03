サーバにリクエスト情報を送信するライブラリ
===


```go
package main

import (
    "fmt"
    "github.com/ochipin/request"
)

func main() {
    var r = &request.Request{
        // 送信先URL
        URL:      "https://localhost/index.html?nickname=ochipin",
        // 自己証明書の場合は true に設定
        Insecure: true,
        // Username: "ID",      // Basic認証が必要な場合は、値を設定
        // Password: "Password" // Basic認証が必要な場合は、値を設定
        // 送信先URLに対するタイムアウト時間(ms単位)
        Timeout:  1000,
        /* プロキシ設定情報が必要な場合は設定する
        Proxy: request.Proxy{
            URL: "http://proxy.server:8080",
            // Username: "Proxy User",
            // Password: "Proxy Password",
        },
        */
    }

    // 送信するリクエストヘッダ情報を設定
    r.Header().Add("User-Agent", "TEST - BROWSER")
    // JSON 形式でサーバヘリクエストを飛ばす場合は、下記のようにする
    // --> r.Header().Add("Content-Type", "application/json")
    // または、
    // --> r.JSON(`{"key": "value"}`) とする

    // 送信するリクエストデータを設定
    r.Values().Add("firstname", "suguru")
    r.Values().Add("lastname", "ochiai")

    // https://localhost/index.html?nickname=ochipin&firstname=suguru&lastname=ochiai を送信
    // r.Post/r.Delete/r.Put/r.Patch が使用可能
    // r.Submit("CustomMethod") で、リクエストメソッドを指定した以外のものを指定可能 
    res, status, err := r.Get()
    if err != nil {
        if status.StatusCode != 0 {
            // StatusCode Error.
        } else {
            // Error
        }
    }
    
    // <!DOCTYPE html> ... <h1>Hello World!</h1> を取得
    fmt.Println(string(res))
}
```