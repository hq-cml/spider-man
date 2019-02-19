package plugin

import (
	"testing"
	"net/http"
)

func TestParseATag(t *testing.T) {
	resp, err := http.DefaultClient.Get("https://www.360.cn")
	if err != nil {
		t.Fatal(err)
	}

	items, reqs, _ := parseForATag(resp, 0)

	t.Log("分析出的URL列表:")
	for _, req := range reqs {
		t.Logf("Depth: %d, URL: %s", req.Depth(), req.HttpReq().URL.String())
	}

	t.Log("分析出的Item列表:")
	for _, item := range items {
		t.Log(*item)
	}

	//t.Log(errs)

}