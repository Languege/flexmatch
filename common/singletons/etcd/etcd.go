// singletons
// @author LanguageY++2013 2023/5/8 23:30
// @company soulgame
package etcd

import (
	"github.com/Languege/flexmatch/common/wrappers/etcd"
	"github.com/coreos/etcd/clientv3"
	"log"
)

var (
	EtcdClient *clientv3.Client
)

type Config struct {
	Addrs    []string `mapstructure:"addrs"`
	Username string   `mapstructrue:"username"`
	Password string   `mapstructure:"password"`
}

func LoadConfig(cfg Config) {
	var err error
	options := []etcd.Option{}
	if cfg.Username != "" {
		options = append(options, etcd.WithAuth(cfg.Username, cfg.Password))
	}
	EtcdClient, err = etcd.NewClient(cfg.Addrs, options...)
	if err != nil {
		log.Panicln(err)
	}
}
