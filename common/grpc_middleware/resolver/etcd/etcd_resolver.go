// etcd
// @author LanguageY++2013 2023/5/8 18:31
// @company soulgame
package etcd

import (
	"strings"
	"google.golang.org/grpc/resolver"
	"github.com/coreos/etcd/clientv3"
	"sync"
	etcd_wrapper "github.com/Languege/flexmatch/common/wrappers/etcd"
	"github.com/Languege/flexmatch/common/logger"
	"net/url"
	"fmt"
	"context"
	"github.com/coreos/etcd/mvcc/mvccpb"
	"encoding/json"
	"google.golang.org/grpc/attributes"
)

const (
	EtcdScheme = "etcd2"
)

func init() {
	resolver.Register(newEtcdBuilder())
}

type etcdBuilder struct {
	client     *clientv3.Client
	once 		*sync.Once
}

func newEtcdBuilder() *etcdBuilder {
	return &etcdBuilder{
		once: &sync.Once{},
	}
}

func(d *etcdBuilder) checkInit(addrs []string, username, password string) {
	d.once.Do(func() {
		options := []etcd_wrapper.Option{}
		if username != "" && password != "" {
			options = append(options, etcd_wrapper.WithAuth(username, password))
		}
		var err error
		d.client, err = etcd_wrapper.NewClient(addrs, options...)
		if err != nil {
			logger.Errorf("new etcd client err %v", err)
		}
	})
}


// Build
func (d *etcdBuilder) Build(target resolver.Target, cc resolver.ClientConn, opts resolver.BuildOptions) (
	resolver.Resolver, error) {

	tURL, err := url.Parse(fmt.Sprintf("%s://%s/%s", target.Scheme, target.Authority, target.Endpoint))
	if err != nil {
		logger.Errorf("url parse target err %s", err)
		return nil, err
	}

	password, _ := tURL.User.Password()
	//构建器初始化
	d.checkInit(strings.Split(tURL.Host, ","), tURL.User.Username(), password)

	//对 target.Endpoint 进行监听
	r := d.newResolver(cc, target.Endpoint)
	r.start()

	return r, nil
}

func(d *etcdBuilder) newResolver(cc resolver.ClientConn, endpoint string) *etcdResolver {
	r := &etcdResolver{
		cc:       cc,
		endpoint: endpoint,
		client:   d.client,
		services: map[string]*ServiceDescriptor{},
	}

	r.key = fmt.Sprintf("%s/%s", GrpcServicePrefix, r.endpoint)

	return r
}

func (d *etcdBuilder) Scheme() string {
	return EtcdScheme
}


type etcdResolver struct {
	cc resolver.ClientConn
	endpoint string
	key string
	client *clientv3.Client

	//服务集合
	services map[string]*ServiceDescriptor
	serviceGuard sync.Mutex
}


func (r *etcdResolver) Close() {
}

func (r *etcdResolver) ResolveNow(options resolver.ResolveNowOptions) {
}

func(r *etcdResolver) listener() {
	var addrs []resolver.Address

	for _, sd := range r.services {
		kvs := make([]interface{}, 0, len(sd.Tags))
		for _, v := range sd.Tags {
			kvs = append(kvs, v)
		}


		addrs = append(addrs, resolver.Address{
			Addr:       sd.ListenAddr,
			ServerName: sd.Name,
			Attributes: attributes.New(kvs...),
		})
	}

	r.cc.UpdateState(resolver.State{
		Addresses: addrs,
	})

	logger.Infof("UpdateState:%#v", addrs)
}

func (r *etcdResolver) start() {
	resp, err := r.client.Get(context.Background(), r.key, clientv3.WithPrefix())
	if err != nil {
		logger.Errorf("get %s err %v", r.key, err)
		return
	}

	for _, kv := range resp.Kvs {
		r.addService(kv)
	}

	watchChan := r.client.Watch(context.Background(), r.key, clientv3.WithPrefix(), clientv3.WithPrevKV())

	go func() {
		for resp := range watchChan {
			for _, ev := range resp.Events {
				switch ev.Type {
				case clientv3.EventTypePut:
					r.addService(ev.Kv)
				case clientv3.EventTypeDelete:
					r.removeService(ev.PrevKv)
				}
			}
		}
	}()
}


func(r *etcdResolver) addService(kv *mvccpb.KeyValue) {
	sd := &ServiceDescriptor{}
	err := json.Unmarshal(kv.Value, sd)
	if err != nil {
		logger.Errorf("add service unmarshal err %v, key %s value %s", err, string(kv.Key), string(kv.Value))
		return
	}
	r.serviceGuard.Lock()
	r.services[sd.ListenAddr] = sd
	r.serviceGuard.Unlock()

	r.notifyChange()
}

func(r *etcdResolver) removeService(kv *mvccpb.KeyValue) {
	sd := &ServiceDescriptor{}
	err := json.Unmarshal(kv.Value, sd)
	if err != nil {
		logger.Errorf("remove service unmarshal err %v, key %s value %s", err, string(kv.Key), string(kv.Value))
		return
	}

	r.serviceGuard.Lock()
	delete(r.services, sd.ListenAddr)
	r.serviceGuard.Unlock()
}

func(r *etcdResolver) notifyChange() {
	r.listener()
}

// 返回一个代表所给endpoints的字符串，keyPrefix代码服务前缀
func BuildTarget(endpoints []string, service string) string {
	return fmt.Sprintf("%s://%s/%s", EtcdScheme,
		strings.Join(endpoints, ","), service)
}

// 返回一个代表所给endpoints的字符串，keyPrefix代码服务前缀
func BuildTargetWithUserPassword(endpoints []string, service string, username, password string) string {
	return fmt.Sprintf("%s://%s:%s@%s/%s", EtcdScheme,
		strings.Join(endpoints, ","), username, password, service)
}