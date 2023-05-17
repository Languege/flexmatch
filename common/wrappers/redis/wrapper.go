// redis
// @author LanguageY++2013 2023/5/12 10:11
// @company soulgame
package redis

import (
	"errors"
	"fmt"
	"github.com/Languege/flexmatch/common/logger"
	"github.com/gomodule/redigo/redis"
	"github.com/mna/redisc"
	"log"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"
	"context"
)

const (
	//默认空闲超时时间
	defaultIdleTimeout = "1h"

	//最大空闲连接数
	defaultMaxIdle = 10

	//最大活跃连接数
	defaultMaxActive = 20
)

//redis 架构模型
type ConnectMode string

const (
	//直连模式
	ConnectModeDirect ConnectMode = "direct"

	// 哨兵模式
	ConnectModeSentinel ConnectMode = "sentinel"

	// cluster模式
	ConnectModeCluster ConnectMode = "cluster"
)

//格式：https://pkg.go.dev/github.com/mitchellh/mapstructure
type SentinelConfigure struct {
	Addrs      string `mapstructure:"addrs"`      //哨兵地址列表
	MasterName string `mapstructure:"masterName"` //master节点名 sentinel monitor mymaster 127.0.0.1 6379 2
}

type ClusterConfigure struct {
	Addrs []string `mapstructure:"addrs"` //哨兵地址列表
}

type Configure struct {
	Host         string            `mapstructure:"host"`
	Port         int               `mapstructure:"port"`
	Password     string            `mapstructure:"password"`
	MaxIdle      int               `mapstructure:"maxIdle"`
	MaxActive    int               `mapstructure:"maxActive"`
	IdleTimeout  string            `mapstructure:"idleTimeout"` //1s, 1m, 1h
	Prefix       string            `mapstructure:"prefix"`
	ConnectMode  ConnectMode        `mapstructure:"connectMode,string"` //集群模式:sentinel,cluster,proxy
	SentinelInfo SentinelConfigure `mapstructure:"sentinelInfo"`
	ClusterInfo  ClusterConfigure  `mapstructure:"clusterInfo"`

	idleTimeout time.Duration
	addr        string
}

type RedisWrapper struct {
	redis.Pool
	Configure Configure
	cluster   *redisc.Cluster
}

//type Option  func(w *RedisWrapper)
//
//func WithIdleTimeOut(timeout time.Duration) Option {
//	return func(w *RedisWrapper) {
//		w.IdleTimeout = timeout
//	}
//}

type NewWrapper func(configure Configure, options ...redis.DialOption) *RedisWrapper

var newFuncs = map[ConnectMode]NewWrapper{}

func RegisterRedisWrapperNewFunc(mode ConnectMode, newFunc NewWrapper) {
	if _, ok := newFuncs[mode]; ok {
		logger.Panicf("redis wrapper new func %s already has been registered previously", mode)
	}

	newFuncs[mode] = newFunc
}

//newSentinelWrapper 建立哨兵模式客户端
func newSentinelWrapper(configure Configure, options ...redis.DialOption) *RedisWrapper {
	logger.Infof("[redis] running in sentinel mode, %v", configure.SentinelInfo)
	sntnl := &Sentinel{
		Addrs:      strings.Split(configure.SentinelInfo.Addrs, ","),
		MasterName: configure.SentinelInfo.MasterName,
		Dial: func(addr string) (redis.Conn, error) {
			c, err := redis.Dial("tcp", addr, redis.DialConnectTimeout(time.Second*20),
				redis.DialReadTimeout(time.Second*20),
				redis.DialWriteTimeout(time.Second*20))
			if err != nil {
				return nil, err
			}
			return c, nil
		},
	}
	w := &RedisWrapper{
		Configure: configure,
		Pool: redis.Pool{
			MaxIdle:     configure.MaxIdle,
			IdleTimeout: configure.idleTimeout,
			MaxActive:   configure.MaxActive,
			Wait:        true,
			Dial: func() (redis.Conn, error) {
				masterAddr, err := sntnl.MasterAddr()
				if err != nil {
					return nil, err
				}
				logger.Infof("[redis] masterAddr:%s", masterAddr)
				c, err := redis.Dial("tcp", masterAddr, options...)
				if err != nil {
					return nil, err
				}
				return c, nil
			},
			TestOnBorrow: func(c redis.Conn, t time.Time) error {
				if !TestRole(c, "master") {
					return errors.New("Role check failed")
				}

				return nil
			},
		},
	}

	return w
}

//newDirectWrapper 建立直连客户端
func newDirectWrapper(configure Configure, options ...redis.DialOption) *RedisWrapper {
	w := &RedisWrapper{
		Configure: configure,
		Pool: redis.Pool{
			MaxIdle:     configure.MaxIdle,
			IdleTimeout: configure.idleTimeout,
			MaxActive:   configure.MaxActive,
			Wait:        true,

			Dial: func() (conn redis.Conn, err error) {
				conn, err = redis.Dial("tcp", configure.addr, options...)
				if err != nil {
					logger.DPanicf("redis dial err %s", err)
				}
				return
			},
			TestOnBorrow: func(c redis.Conn, t time.Time) error {
				_, err := c.Do("PING")
				if err != nil {
					logger.DPanicf("redis ping err %s", err)
				}
				return err
			},
		},
	}

	return w
}

//newClusterWrapper 新建cluster模式wrapper
func newClusterWrapper(configure Configure, options ...redis.DialOption) *RedisWrapper {
	logger.Info("[redis] running in cluster mode", configure.ClusterInfo)
	cluster := &redisc.Cluster{
		StartupNodes: configure.ClusterInfo.Addrs,
		DialOptions:  options,
		CreatePool: func(address string, options ...redis.DialOption) (*redis.Pool, error) {
			return &redis.Pool{
				MaxIdle:     configure.MaxIdle,
				IdleTimeout: configure.idleTimeout,
				MaxActive:   configure.MaxActive,
				Wait:        true,

				Dial: func() (conn redis.Conn, err error) {
					conn, err = redis.Dial("tcp", address, options...)
					if err != nil {
						logger.DPanicf("redis dail err %s", err)
					}
					return
				},

				TestOnBorrow: func(c redis.Conn, t time.Time) error {
					_, err := c.Do("PING")
					if err != nil {
						logger.DPanicf("redis ping err %s", err)
					}
					return err
				},
			}, nil
		},
	}
	// initialize its mapping
	if err := cluster.Refresh(); err != nil {
		log.Fatalf("[redis] cluster Refresh failed: %v", err)
	}
	w := &RedisWrapper{
		Configure: configure,
		cluster:   cluster,
	}

	return w
}

func init() {
	//cluster模式
	RegisterRedisWrapperNewFunc(ConnectModeCluster, newClusterWrapper)
	//哨兵模式
	RegisterRedisWrapperNewFunc(ConnectModeSentinel, newSentinelWrapper)
	//直连模式
	RegisterRedisWrapperNewFunc(ConnectModeDirect, newDirectWrapper)
}

func NewRedisWrapper(configure Configure) *RedisWrapper {

	configure.addr = fmt.Sprintf("%s:%d", configure.Host, configure.Port)

	options := []redis.DialOption{
		redis.DialConnectTimeout(time.Second * 20),
		redis.DialReadTimeout(time.Second * 20),
		redis.DialWriteTimeout(time.Second * 20),
	}

	if configure.Password != "" {
		options = append(options, redis.DialPassword(configure.Password))
	}
	//空闲超时解析
	if configure.IdleTimeout == "" {
		configure.IdleTimeout = defaultIdleTimeout
	}
	idleTimeout, err := time.ParseDuration(configure.IdleTimeout)
	if err != nil {
		logger.Panicf("RedisWrapper idleTimeout  err %s", err)
	}
	if idleTimeout == 0 {
		logger.Panicf("RedisWrapper idleTimeout cannot empty")
	}
	if configure.MaxIdle == 0 {
		configure.MaxIdle = defaultMaxIdle
	}

	if configure.MaxActive == 0 {
		configure.MaxActive = defaultMaxActive
	}

	newFunc, ok := newFuncs[ConnectMode(configure.ConnectMode)]
	if !ok {
		logger.Panicf("ConnectMode %s has not been registered", configure.ConnectMode)
	}

	w := newFunc(configure, options...)

	go w.closeConnection()

	return w
}

func (rw *RedisWrapper) closeConnection() {
	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-c
		rw.Close()
	}()
}


func(rw *RedisWrapper) GetConn() redis.Conn{
	if rw.Configure.ConnectMode == ConnectModeCluster  {
		//然而，一个连接可以通过调用 RetryConn 来包装，它返回一个 redis.Conn 接口，
		// 其中只有对 Do、Close 和 Err 的调用可以成功。这意味着不支持流水线，一次只能执行一条命令，
		// 但它会自动处理MOVED和ASK的回复，以及TRYAGAIN错误。
		conn := rw.cluster.Get()
		conn, _ = redisc.RetryConn(conn, 3, 100 * time.Millisecond)
		return conn
	}

	conn, err := rw.Pool.GetContext(context.Background())
	if err != nil {
		logger.Error("pool get conn err %s", err)
	}

	return conn
}


func(rw *RedisWrapper) buildKey(key string) string {
	if rw.Configure.Prefix != "" {
		key = rw.Configure.Prefix + string(os.PathListSeparator) + key
	}

	return key
}