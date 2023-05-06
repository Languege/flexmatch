// etcd
// @author LanguageY++2013 2023/5/6 15:44
// @company soulgame
package etcd_wrapper

import (
	"github.com/coreos/etcd/clientv3"
	"time"
	"github.com/spf13/viper"
	"log"
)

var(
	GlobalClient *clientv3.Client
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
		MaxCallSendMsgSize: 16 * 1024 * 1024,
	}

	if len(options) > 0 {
		for _, op := range options {
			op(&conf)
		}
	}

	client, err = clientv3.New(conf)
	return
}


func init() {
	//初始化全局客户端
	addr := viper.GetStringSlice("etcd.addr")
	if len(addr) == 0 {
		log.Panicln("etcd.addr can not empty")
	}

	var err error
	options := []Option{}
	if viper.GetString("etcd.username") != "" {
		options = append(options, WithAuth(viper.GetString("etcd.username"), viper.GetString("etcd.password")))
	}
	GlobalClient, err = NewClient(addr,  options...)
	if err != nil {
		log.Panicln(err)
	}
}



