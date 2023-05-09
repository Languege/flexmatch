// etcd
// @author LanguageY++2013 2023/5/8 21:32
// @company soulgame
package etcd

import (
	"testing"
	"net/url"
)

func TestURL(t *testing.T) {
	u := &url.URL{
		Scheme: "http",
		User:url.UserPassword("root", "123456"),
		Host: "localhost",
		Path: "/index",
	}

	t.Log(u.String())
}

type ValueMethodDemo struct {
	A string
}

func(v *ValueMethodDemo) Set(x string) {
	v.A = x
}


func TestValueMethod(t *testing.T) {
	v := ValueMethodDemo{}
	//语法糖 v.Set -> (&v).Set()
	v.Set("test")

	t.Log(v.A)
}