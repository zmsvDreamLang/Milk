package lua

import (
	"io"
	"net/http"
	"strings"
)

func OpenHttp(L *LState) int {
	httpModule := L.RegisterModule(HttpLibName, httpFuncs)
	L.Push(httpModule)
	return 1
}

var httpFuncs = map[string]LGFunction{
	"get":  httpGet,
	"post": httpPost,
}

func httpGet(L *LState) int {
	url := L.CheckString(1)
	resp, err := httpGetRaw(url)
	if err != nil {
		L.Push(LNil)
		L.Push(LString(err.Error()))
		return 2
	}
	L.Push(LString(resp))
	return 1
}

func httpPost(L *LState) int {
	url := L.CheckString(1)
	data := L.CheckString(2)
	resp, err := httpPostRaw(url, data)
	if err != nil {
		L.Push(LNil)
		L.Push(LString(err.Error()))
		return 2
	}
	L.Push(LString(resp))
	return 1
}

func httpGetRaw(url string) (string, error) {
	httpResp, err := http.Get(url)
	if err != nil {
		return "", err
	}
	defer httpResp.Body.Close()
	bodyBytes, err := io.ReadAll(httpResp.Body)
	if err != nil {
		return "", err
	}
	return string(bodyBytes), nil
}

func httpPostRaw(url string, data string) (string, error) {
	httpResp, err := http.Post(url, "application/json", strings.NewReader(data))
	if err != nil {
		return "", err
	}
	defer httpResp.Body.Close()
	bodyBytes, err := io.ReadAll(httpResp.Body)
	if err != nil {
		return "", err
	}
	return string(bodyBytes), nil
}
