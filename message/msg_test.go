package message

import (
	"fmt"
	"testing"
)

func Test_Image(t *testing.T){

	a:=Image("http://www.baidu.com/img/bd_logo1.png", ImageOptions().SetCache(false).SetProxy(true))
	fmt.Printf("%v\n",a)
}
