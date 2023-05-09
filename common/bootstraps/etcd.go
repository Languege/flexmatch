// bootstraps
// @author LanguageY++2013 2023/5/8 23:35
// @company soulgame
package bootstraps

import (
	single_etcd "github.com/Languege/flexmatch/common/singletons/etcd"
	"github.com/spf13/viper"
	"github.com/Languege/flexmatch/common/logger"
)

func InitEtcd() {
	cfg := single_etcd.Config{}
	err := viper.UnmarshalKey("etcd", &cfg)
	if err != nil {
		logger.Panicf("viper unmarshal 'etcd' err %v", err)
	}
	if len(cfg.Addrs) == 0 {
		logger.Panicf("'etcd.addrs' cannot empty")
	}

	single_etcd.LoadConfig(cfg)
}
