// bootstraps
// @author LanguageY++2013 2023/5/9 15:45
// @company soulgame
package bootstraps

import (
	resolver_etcd "github.com/Languege/flexmatch/common/grpc_middleware/resolver/etcd"
	single_etcd "github.com/Languege/flexmatch/common/singletons/etcd"
	"github.com/Languege/flexmatch/common/utils/network"
	"github.com/Languege/flexmatch/common/logger"
	"fmt"
	"context"
)

func PublishService(name string, port int, tags... string) {
	//获取内网ip
	ips, err := network.LocalIPv4s()
	if err != nil {
		logger.Panicf("network ipv4 err %v", err)
	}

	sd := &resolver_etcd.ServiceDescriptor{
		Name:       name,
		ListenAddr: fmt.Sprintf("%s:%d", ips[0], port),
		Tags:       tags,
	}

	publisher, err := resolver_etcd.NewServicePublisher(context.Background(), single_etcd.EtcdClient)
	if err != nil {
		logger.Panicf("NewServicePublisher err %v", err)
	}

	err = publisher.Publish(sd)
	if err != nil {
		logger.Panicf("publish sd  err %v, key '%s' value '%s'", err, sd.Key(), sd.Value())
	}

	logger.Infof("publish sd success, key '%s' value '%s'", sd.Key(), sd.Value())
}
