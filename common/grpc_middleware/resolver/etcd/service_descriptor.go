// etcd
// @author LanguageY++2013 2023/5/8 22:08
// @company soulgame
package etcd

import (
	"fmt"
	"encoding/json"
)

const GrpcServicePrefix = "services.grpc"

type ServiceDescriptor struct {
	Name 		string  // 服务名
	ListenAddr string   // 服务地址
	Tags       []string // 标签信息, 例如版本
}

func(sd *ServiceDescriptor) Key() string {
	return fmt.Sprintf("%s/%s/%s", GrpcServicePrefix, sd.Name, sd.ListenAddr)
}

func(sd *ServiceDescriptor) Value() string {
	data, _ := json.Marshal(sd)
	return string(data)
}