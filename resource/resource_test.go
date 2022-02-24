package resource

import (
	"path"
	"testing"
)

func Test_LoadNetworkRes(t *testing.T) {
	rm := NewResourceManager("res")
	res, err := rm.LoadNetworkResource("https://www.baidu.com/img/bd_logo1.png", false, "")
	if err != nil {
		t.Error(err)
	}
	t.Log(res)
}

func Test(t *testing.T) {
	a:=path.Join("a", "/b")
	t.Log(a)
}