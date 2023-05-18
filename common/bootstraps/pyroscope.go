// bootstraps
// @author LanguageY++2013 2023/5/18 11:13
// @company soulgame
package bootstraps

import (
	"github.com/spf13/viper"
	pyroscope_wrapper "github.com/Languege/flexmatch/common/wrappers/pyroscope"
	"github.com/Languege/flexmatch/common/logger"
)

func InitPyroscope() {
	if viper.IsSet("pyroscope") {
		cfg := pyroscope_wrapper.Config{}
		err := viper.UnmarshalKey("pyroscope", &cfg)
		if err != nil {
			logger.Panic(err)
		}
		pw := pyroscope_wrapper.NewPyroscopeWrapper(cfg, pyroscope_wrapper.WithHostname())
		pw.Start()
	}
}
