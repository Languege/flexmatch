// grpc_middleware
// @author LanguageY++2013 2023/5/9 18:17
// @company soulgame
package grpc_middleware

import (
	"go.uber.org/zap"
	"github.com/Languege/flexmatch/common/logger"
	"github.com/spf13/viper"
)

var(
	rpcLogger *zap.Logger
)

func init() {
	cfg := logger.LoggerConfig{}
	err := viper.UnmarshalKey("log.rpc", &cfg)
	logger.FatalfIf(err != nil, "log.rpc unmarshal err %v", err)

	rpcLogger, err = logger.NewZapLogger(cfg)
	logger.FatalfIf(err != nil, "new zap logger err %v", err)
}
