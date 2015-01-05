package QcloudApi

import (
	"crypto/hmac"
	"crypto/sha1"
	"encoding/base64"
	"fmt"
	"io/ioutil"
	"log"
	"math/rand"
	"net/http"
	"net/url"
	"sort"
	"strconv"
	"strings"
	"time"
)

var debug bool = false

func makePlainText(requestMethod string, requestHost string, requestPath string, params map[string]interface{}) (plainText string, err error) {

	plainText += strings.ToUpper(requestMethod)
	plainText += requestHost
	plainText += requestPath
	plainText += "?"

	// 排序
	keys := make([]string, 0, len(params))
	for k := range params {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	var plainParms string
	for i := range keys {
		k := keys[i]
		plainParms += "&" + fmt.Sprintf("%v", k) + "=" + fmt.Sprintf("%v", params[k])
	}
	plainText += plainParms[1:]

	if debug {
		log.Printf("plainText[%s]\n", plainText)
	}

	return plainText, nil
}

func sign(requestMethod string, requestHost string, requestPath string, params map[string]interface{}, secretKey string) (sign string, err error) {

	var source string

	source, err = makePlainText(requestMethod, requestHost, requestPath, params)
	if err != nil {
		panic(err)
		log.Fatalln("Make PlainText error.", err)
		return sign, err
	}

	hmacObj := hmac.New(sha1.New, []byte(secretKey))
	hmacObj.Write([]byte(source))

	sign = base64.StdEncoding.EncodeToString(hmacObj.Sum(nil))
	if debug {
		log.Printf("Sign[%s]\n", sign)
	}

	return sign, nil
}

func SendRequest(mod string, params map[string]interface{}, config map[string]interface{}) (retData string, err error) {

	if config["debug"] != nil {
		debug, _ = strconv.ParseBool(fmt.Sprintf("%t", config["debug"]))
	}

	secretId := fmt.Sprintf("%s", config["secretId"])
	secretKey := fmt.Sprintf("%s", config["secretKey"])

	requestMethod := "POST"
	requestHost := mod + ".api.qcloud.com"
	requestPath := "/v2/index.php"

	paramValues := url.Values{}
	if params["SecretId"] == nil {
		params["SecretId"] = secretId
	}
	if params["Timestamp"] == nil {
		params["Timestamp"] = fmt.Sprintf("%v", time.Now().Unix())
	}
	if params["Nonce"] == nil {
		rand.Seed(time.Now().UnixNano())
		params["Nonce"] = fmt.Sprintf("%v", rand.Int())
	}
	if params["Region"] == nil {
		params["Region"] = "gz"
	}
	if params["Action"] == nil {
		params["Action"] = "DescribeInstances"
	}

	sign, err := sign(requestMethod, requestHost, requestPath, params, secretKey)
	paramValues.Add("Signature", sign)

	for k, v := range params {
		paramValues.Add(fmt.Sprintf("%v", k), fmt.Sprintf("%v", v))
	}
	if debug {
		log.Printf("req[%v]\n", paramValues)
	}

	urlStr := "https://" + requestHost + requestPath

	rsp, err := http.PostForm(urlStr, paramValues)

	if err != nil {
		panic(err)
		log.Fatal("http post error.", err)
		return "", err
	}

	defer rsp.Body.Close()

	retData_, err := ioutil.ReadAll(rsp.Body)
	if err != err {
		panic(err)
		return "", err
	}

	if debug {
		log.Printf("rsp[%v]\n", string(retData_))
	}

	return string(retData_), nil
}
