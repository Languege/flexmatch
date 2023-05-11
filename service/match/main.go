package main

import (
	_ "github.com/Languege/flexmatch/service/match/conf"
	_ "github.com/Languege/flexmatch/common/bootstraps"

	"fmt"
	"github.com/spf13/viper"
	"google.golang.org/grpc"
	"github.com/Languege/flexmatch/service/match/proto/open"
	"github.com/Languege/flexmatch/service/match/service"
	"google.golang.org/grpc/reflection"
	"log"
	"net"
	"github.com/Languege/flexmatch/service/match/entities"
	"github.com/Languege/flexmatch/common/logger"

	"github.com/Languege/flexmatch/common/bootstraps"
	"github.com/Languege/flexmatch/common/grpc_middleware"
	logger_pubsub "github.com/Languege/flexmatch/service/match/pubsub/logging"
	kafka_pubsub "github.com/Languege/flexmatch/service/match/pubsub/kafka"
	"github.com/Languege/flexmatch/service/match/pubsub"
)

var(
	BuildVersion string
	BuildDate string
)

func init() {
	//服务发布
	bootstraps.PublishService(viper.GetString("rpc.service"), viper.GetInt("rpc.port"))

	//匹配事件发布器设置
	loggerPublisher := logger_pubsub.NewLoggerPublisher()
	kafkaPublisher := kafka_pubsub.NewKafkaPublisher(viper.GetStringSlice("kafka.bootstrapServers"))
	multiPublisher := pubsub.NewMultiPublisher(loggerPublisher, kafkaPublisher)
	entities.SetPublisher(multiPublisher)

	//TODO: 临时使用mock
	entities.SetGameClient(entities.NewMockGameClient())
}


func main() {
	logger.Debugw("", "BuildVersion", BuildVersion, "BuildDate", BuildDate)

	address := fmt.Sprintf("%s:%d", viper.GetString("rpc.host"), viper.GetInt("rpc.port"))
	lis, err := net.Listen("tcp", address)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	s := grpc.NewServer(grpc_middleware.ServerOptions()...)

	open.RegisterFlexMatchServer(s, service.NewMatchServerImpl())

	//注册反射服务, 便于通过GM调试面板调试
	reflection.Register(s)

	if err = s.Serve(lis);err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}