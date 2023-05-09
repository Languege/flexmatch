// bootstraps
// @author LanguageY++2013 2023/5/8 23:18
// @company soulgame
package bootstraps

import (
	"github.com/Languege/flexmatch/common/logger"
	"github.com/spf13/viper"
)

func InitLogger() {
	cfg := logger.LoggerConfig{}
	err := viper.UnmarshalKey("log", &cfg)
	logger.FatalfIf(err != nil, "viper unmarshal 'log' err %v", err)
	logger.LoadConfig(cfg)
}
