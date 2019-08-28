package util

import "testing"

func TestGetPrimaryDomain(t *testing.T) {
	str1 := "www.baidu.com"

	r1, err := GetPrimaryDomain(str1)
	if err != nil {
		t.Error("Error:", err)
	}
	if r1 != "baidu.com" {
		t.Error("Wrong:", r1)
	}
}