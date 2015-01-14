qcloud_sign_golang
==================
Qcloud API 调试工具 Golang 版

# 下载

```sh
go get github.com/QcloudApi/qcloud_sign_golang
```
- 安装

```sh
go install github.com/QcloudApi/qcloud_sign_golang
```

# 使用

```C
package main

import(
    "fmt"
    "os"
    "encoding/json"
    "github.com/QcloudApi/qcloud_sign_golang"
)

func main() {
    // 替换实际的 SecretId 和 SecretKey
    secretId := "YOUR_SECRET_ID"
    secretKey := "YOUR_SECRET_KEY"

    // 配置
    config := map[string]interface{} {"secretId" : secretId, "secretKey" : secretKey, "debug" : false}

    // 请求参数
    params := map[string]interface{} {"Region" : "hk", "Action" : "DescribeInstances"}

    // 发送请求
    retData, err := QcloudApi.SendRequest("cvm", params, config)
    if err != nil{
        fmt.Print("Error.", err)
        return
    }

    // 解析 Json 字符串
    var jsonObj interface{}
    err = json.Unmarshal([]byte(retData), &jsonObj)
    if err != nil {
        fmt.Println(err);
        return
    }
    // 打印 Json
    jsonOut, _ := json.MarshalIndent(jsonObj, "", "  ");
    b2 := append(jsonOut, '\n')
    os.Stdout.Write(b2)

    return
}
```
