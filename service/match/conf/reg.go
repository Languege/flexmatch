package conf

import (
	"github.com/spf13/viper"
	"log"
	"runtime"
	"path/filepath"
	"flag"
	"os"
)

var(
	configFile string
	flagSet  = flag.NewFlagSet("config", flag.ContinueOnError)
)
/**
 *@author LanguageY++2013
 *2020/2/15 11:03 PM
 **/
func init()  {
	//修改flag的默认行为
	flag.Func("config", "yaml配置文件路径", func(s string) error {
		return nil
	})
	flagSet.StringVar(&configFile, "config", "", "yaml配置文件路径")
	flagSet.Parse(os.Args[1:])
	if configFile != "" {
		viper.SetConfigFile(configFile)
		goto readConfig
	}

	viper.AddConfigPath("./conf")
	viper.AddConfigPath(curPathDir())
	viper.SetConfigName("main")
	viper.SetConfigType("yaml")

readConfig:
	err := viper.ReadInConfig() // Find and read the config file
	if err != nil { // Handle errors reading the config file
		log.Fatalf("Fatal error config file: %s \n", err)
	}
}

func curPathDir() string {
	_, file, _, _ := runtime.Caller(1)
	return filepath.Dir(file)
}