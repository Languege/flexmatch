// etcd
// @author LanguageY++2013 2023/5/8 21:10
// @company soulgame
package etcd

import (
	"github.com/coreos/etcd/clientv3"
	"time"
)

const (
	defaultMaxCallSendMsgSize = 16 * 1024 * 1024
)

type Option func(c *clientv3.Config)

func WithAuth(username, password string)  Option {
	return func(c *clientv3.Config) {
		c.Username = username
		c.Password = password
	}
}

func WithMaxSendSize(maxCallSendMsgSize int) Option {
	return func(c *clientv3.Config) {
		c.MaxCallSendMsgSize = maxCallSendMsgSize
	}
}


func NewClient(addr []string, options... Option)(client *clientv3.Client, err error) {
	conf := clientv3.Config{
		Endpoints:   addr,
		DialTimeout: 5 * time.Second,
		MaxCallSendMsgSize: defaultMaxCallSendMsgSize,
	}

	if len(options) > 0 {
		for _, op := range options {
			op(&conf)
		}
	}

	client, err = clientv3.New(conf)
	return
}

