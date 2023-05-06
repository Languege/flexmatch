package main

import (
	_ "github.com/Languege/flexmatch/service/match/conf"
	"fmt"
	"github.com/spf13/viper"
	"google.golang.org/grpc"
	"github.com/Languege/flexmatch/service/match/proto/open"
	"github.com/Languege/flexmatch/service/match/service"
	"google.golang.org/grpc/reflection"
	"log"
	"net"
	"github.com/Languege/flexmatch/service/match/entities"
	"github.com/Languege/flexmatch/service/match/event_subscribers"
)

var(
	BuildVersion string
	BuildDate string
)

func init() {
	entities.RegisterSubscribe(event_subscribers.KafkaMatchEventSubscribe)
}


func main() {
	log.Println("BuildVersion", BuildVersion, "BuildDate", BuildDate)

	address := fmt.Sprintf("%s:%d", viper.GetString("rpc.host"), viper.GetInt("rpc.port"))
	lis, err := net.Listen("tcp", address)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	s := grpc.NewServer()

	open.RegisterFlexMatchServer(s, service.NewMatchServerImpl())

	//注册反射服务, 便于通过GM调试面板调试
	reflection.Register(s)

	if err = s.Serve(lis);err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}