// pyroscope
// @author LanguageY++2013 2023/5/18 09:25
// @company soulgame
package pyroscope

import (
	"sync/atomic"
	"github.com/coreos/etcd/clientv3"
	"context"
	"github.com/Languege/flexmatch/common/logger"
	"github.com/pyroscope-io/client/pyroscope"
	"encoding/json"
	single_etcd "github.com/Languege/flexmatch/common/singletons/etcd"
	"github.com/spf13/viper"
	"os"
	"runtime"
)

const (
	confProviderFile = "file"
	confProviderEtcd = "etcd"
)

type Config struct {
	Provider  string //配置提供来源 viper-本地配置文件(默认)、etcd-远端etcd服务
	Key       string //viper配置Key 或者 etcd监听Key
	ApplicationName string  //应用名，透传给pyroscope.Config
	LocalTags map[string]string //本地标签，用于进行标签匹配，仅完全匹配的实例才进行指标上报
}

type PyroscopeWrapper struct {
	Config

	//动态Profile, 仅监听Key中的开关信息
	profile 	atomic.Value

	// Provider为etcd时生效
	client 		*clientv3.Client
}

type Option func(w *PyroscopeWrapper)

func WithEtcdClient(client *clientv3.Client) Option {
	return func(w *PyroscopeWrapper) {
		w.client = client
	}
}

func WithHostname() Option {
	return func(w *PyroscopeWrapper) {
		if w.LocalTags == nil {
			w.LocalTags = map[string]string{}
		}

		w.LocalTags["hostname"], _ = os.Hostname()
	}
}

func NewPyroscopeWrapper(cfg Config, opts... Option) *PyroscopeWrapper {
	w := &PyroscopeWrapper{
		Config: cfg,
	}

	for _, opt := range opts {
		opt(w)
	}

	if w.Provider == confProviderEtcd {
		if w.client == nil {
			w.client = single_etcd.EtcdClient
		}
	}

	return w
}

func(w *PyroscopeWrapper) Start() {
	switch w.Provider {
	case confProviderEtcd:
		resp, err := w.client.Get(context.Background(), w.Key)
		if err != nil {
			logger.Errorf("PyroscopeWrapper get %s err %s", w.Key, err)
			return
		}

		if len(resp.Kvs) > 0 {
			for _, kv := range resp.Kvs {
				cfg := pyroscope.Config{}
				if err := json.Unmarshal(kv.Value, &cfg);err == nil {
					w.reset()
					//标签匹配
					if w.matchTags(cfg.Tags) {
						w.start(cfg)
					}
				}

				break
			}
		}

		go w.watchConfig()
	default:
		//本地配置
		cfg := pyroscope.Config{}
		err := viper.UnmarshalKey(w.Key, &cfg)
		if err != nil {
			logger.Panic(err)
		}

		w.start(cfg)
	}

}

func(w *PyroscopeWrapper) handleWatchEvent(ev *clientv3.Event) {
	defer func() {
		if r := recover(); r != nil {
			logger.DPanicf("handleWatchEvent recover %s", r)
		}
	}()
	switch ev.Type {
	case clientv3.EventTypePut:
		cfg := pyroscope.Config{}
		if err := json.Unmarshal(ev.Kv.Value, &cfg);err == nil {
			w.reset()
			//标签匹配
			if w.matchTags(cfg.Tags) {
				w.start(cfg)
			}
		}
	case clientv3.EventTypeDelete:
		w.reset()
	}
}

func(w *PyroscopeWrapper) watchConfig() {
	wch := w.client.Watch(context.Background(), w.Key)
	go func() {
		defer func() {
			if r := recover(); r != nil {
				logger.DPanicf("watchConfig recover %s", r)
			}
		}()
		for wresp := range wch {
			for _, ev := range wresp.Events {
				w.handleWatchEvent(ev)
			}
		}
	}()
}

//matchTags 标签匹配
func(w *PyroscopeWrapper) matchTags(tags map[string]string) bool {
	if len(tags) == 0 {
		return true
	}

	for k, v := range tags {
		lv, ok := w.LocalTags[k]
		if !ok {
			return false
		}

		if lv != v {
			return false
		}
	}

	return true
}

func(w *PyroscopeWrapper) reset() {
	if v := w.profile.Load(); v != nil {
		if err := v.(*pyroscope.Profiler).Stop();err != nil {
			logger.Errorf("PyroscopeWrapper reset  err %s", err)
		}
		//关闭互斥、阻塞采样
		runtime.SetMutexProfileFraction(0)
		runtime.SetBlockProfileRate(0)
	}
}

func(w *PyroscopeWrapper) start(cfg pyroscope.Config) {
	//互斥、阻塞采样率设置
	for _, pt := range  cfg.ProfileTypes {
		if pt == pyroscope.ProfileMutexCount || pt == pyroscope.ProfileMutexDuration {
			runtime.SetMutexProfileFraction(5)
		}

		if pt == pyroscope.ProfileBlockCount || pt == pyroscope.ProfileBlockDuration {
			runtime.SetBlockProfileRate(5)
		}
	}


	cfg.Logger = logger.GlobalInstance()
	cfg.Tags = w.LocalTags
	cfg.ApplicationName = w.ApplicationName
	p, err := pyroscope.Start(cfg)
	if err != nil {
		logger.Error(err)
		return
	}

	w.profile.Store(p)
}