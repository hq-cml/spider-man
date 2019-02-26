package main

import (
	"fmt"
	"strings"
	"testing"
	"net/http"

)

func TestStringTrim(t *testing.T) {
	fmt.Println(" Trim 函数的用法")
	fmt.Printf("[%q]", strings.Trim(" !!! Ac!htung ! ! !", "!"))
}

func TestStringSplit(t *testing.T) {
	t.Log(" Trim 函数的用法")
	tmp := strings.Split("https://www.360.cn/#1", "#")
	t.Log(tmp[0])

	tmp = strings.Split("https://www.360.cn/#1", "a")
	t.Log(tmp[0])
}

func TestHttpHeader(t *testing.T) {
	//url := "http://sd.360.cn/downloadoffline.html"
	url := "https://www.360.cn/"
	httpReq, err := http.NewRequest(http.MethodHead, url, nil)
	if err != nil {
		t.Fatal(err)
	}

	resp, err := http.DefaultClient.Do(httpReq)
	if err != nil {
		t.Fatal(err)
	}

	defer resp.Body.Close()
	t.Log(resp.Header.Get("Content-Type"))
}

