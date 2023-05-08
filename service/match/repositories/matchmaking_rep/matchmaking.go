// repositories
// @author LanguageY++2013 2022/11/7 18:59
// @company soulgame
package matchmaking_rep

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/Languege/flexmatch/service/match/proto/open"
	"github.com/Languege/flexmatch/service/match/sgerrors"
	"github.com/Languege/flexmatch/service/match/entities"
	"github.com/coreos/etcd/clientv3"
	"github.com/juju/errors"
	"log"
	"strings"
	"sync"
	"time"
	"github.com/Languege/flexmatch/service/match/logger"
	etcd_wrapper "github.com/Languege/flexmatch/service/match/wrappers/etcd"
)

const (
	//默认etcd路径配置
	default_ConfPath = "open.gorm.matchmaking."
	//最大请求超时时间
	maxRequestTimeoutSeconds = 43200
	//最大接受超时时间
	maxAcceptanceTimeoutSeconds = 600
	//票据最大保留时长
	maxTicketTTLSeconds = 3600
)

// 可选项
type Option func(rep *MatchmakingRepository)

// WithConfPath 指定配置路径
func WithConfPath(path string) Option {
	return func(rep *MatchmakingRepository) {
		rep.confPath = path
	}
}

// 媒人仓储
type MatchmakingRepository struct {
	m     map[string]*entities.Matchmaking
	guard sync.RWMutex

	etcdClient *clientv3.Client

	//etcd 配置路径
	confPath string

	//票据内存存储 （TODO:后面使用redis后mysql替换）
	memoryStore    map[string]*open.MatchmakingTicket
	memoryStoreIds []string
	storeGuard     *sync.RWMutex
	storeCond      *sync.Cond
}

func NewMatchmakingRepository(opts ...Option) *MatchmakingRepository {
	rep := &MatchmakingRepository{
		m:           map[string]*entities.Matchmaking{},
		etcdClient:  etcd_wrapper.GlobalClient,
		confPath:    default_ConfPath,
		memoryStore: map[string]*open.MatchmakingTicket{},
		storeGuard:  &sync.RWMutex{},
	}

	rep.storeCond = sync.NewCond(rep.storeGuard)

	//可选项
	for _, opt := range opts {
		opt(rep)
	}
	//配置监听
	go rep.watch()

	//票据过期检查
	go rep.loopCheckExpiredTickets()

	return rep
}

func (rep *MatchmakingRepository) watch() {
	resp, err := rep.etcdClient.Get(context.TODO(), rep.confPath, clientv3.WithPrefix())
	if err != nil {
		log.Panicln(err.Error())
		return
	}

	for _, kv := range resp.Kvs {
		conf := &open.MatchmakingConfiguration{}
		err = json.Unmarshal(kv.Value, conf)
		if err != nil {
			logger.Warnf("Matchmaking configuration unmarshal err %s, k:%s v:%s", err.Error(),
				string(kv.Key), string(kv.Value))
			continue
		}
		//添加媒人实体
		rep.save(conf)
	}

	//监听信号
	watchChan := rep.etcdClient.Watch(context.TODO(), rep.confPath, clientv3.WithPrefix(), clientv3.WithPrevKV())
	for watchResp := range watchChan {
		for _, ev := range watchResp.Events {
			switch ev.Type {
			case clientv3.EventTypePut:
				//添加
				conf := &open.MatchmakingConfiguration{}
				err = json.Unmarshal(ev.Kv.Value, conf)
				if err != nil {
					logger.Warnf("Matchmaking configuration unmarshal err %s, k:%s v:%s", err.Error(),
						string(ev.Kv.Key), string(ev.Kv.Value))
					break
				}
				//添加/更新媒人实体
				rep.save(conf)
			case clientv3.EventTypeDelete:
				//删除
				conf := &open.MatchmakingConfiguration{}
				err = json.Unmarshal(ev.PrevKv.Value, conf)
				if err != nil {
					logger.Warnf("Matchmaking configuration unmarshal err %s, k:%s v:%s", err.Error(),
						string(ev.PrevKv.Key), string(ev.PrevKv.Value))
					break
				}

				rep.remove(conf.Name)
			}
		}
	}
}

func (rep *MatchmakingRepository) save(conf *open.MatchmakingConfiguration) {
	rep.guard.Lock()
	defer rep.guard.Unlock()

	e, ok := rep.m[conf.Name]
	if !ok {
		//新建实体
		e = entities.NewMatchmaking(conf)

		rep.m[e.Conf.Name] = e
		return
	}

	//已存在，更新配置
	e.Reload(conf)

	return
}

func (rep *MatchmakingRepository) remove(name string) {
	rep.guard.Lock()
	defer rep.guard.Unlock()

	e, ok := rep.m[name]
	if !ok {
		return
	}

	//做清理工作
	e.CleanUp()

	delete(rep.m, name)
}

func (rep *MatchmakingRepository) get(name string) (m *entities.Matchmaking, ok bool) {
	rep.guard.RLock()
	defer rep.guard.RUnlock()

	m, ok = rep.m[name]

	return
}

func (rep *MatchmakingRepository) CheckConfiguration(conf *open.MatchmakingConfiguration) (sgErr error) {
	conf.Name = strings.TrimSpace(conf.Name)
	if conf.Name == "" {
		sgErr = sgerrors.NewSGError(open.ResultCode_ParamInvalid, "Name cannot be empty")
		return
	}

	//匹配超时设置
	if conf.RequestTimeoutSeconds <= 0 || conf.RequestTimeoutSeconds > maxRequestTimeoutSeconds {
		sgErr = sgerrors.NewSGError(open.ResultCode_ParamInvalid,
			fmt.Sprintf("RequestTimeoutSeconds must be in 1~%d, current value is %d", maxRequestTimeoutSeconds, conf.RequestTimeoutSeconds))
		return
	}

	//允许用户主动接受匹配
	if conf.AcceptanceRequired {
		if conf.AcceptanceTimeoutSeconds <= 0 || conf.AcceptanceTimeoutSeconds > maxAcceptanceTimeoutSeconds {
			sgErr = sgerrors.NewSGError(open.ResultCode_ParamInvalid,
				fmt.Sprintf("AcceptanceTimeoutSeconds must be in 1~%d, current value is %d", maxAcceptanceTimeoutSeconds, conf.AcceptanceTimeoutSeconds))
			return
		}
	}

	//规则集非空
	if conf.RuleSet == nil {
		sgErr = sgerrors.NewSGError(open.ResultCode_ParamInvalid, "RuleSet cannot be empty")
		return
	}

	//验证规则集
	err := entities.NewMatchmakingRuleSetWrapper(conf.RuleSet).CheckParams()
	if err != nil {
		sgErr = sgerrors.NewSGError(open.ResultCode_ParamInvalid, errors.ErrorStack(err))
		return
	}

	return
}

// SaveConfiguration 保存配置
func (rep *MatchmakingRepository) SaveConfiguration(ctx context.Context, conf *open.MatchmakingConfiguration) (e *entities.Matchmaking, sgErr sgerrors.SGError) {
	rep.guard.Lock()
	//配置是否已存在
	_, ok := rep.m[conf.Name]
	if ok {
		sgErr = sgerrors.NewSGError(open.ResultCode_MatchmakingConfigurationHasExist, fmt.Sprintf("configuration name %s has exist", conf.Name))
		rep.guard.Unlock()
		return
	}
	rep.guard.Unlock()

	//持久化配置
	confKey := rep.confPath + conf.Name
	//数据序列化
	data, _ := json.Marshal(conf)
	_, err := rep.etcdClient.Put(ctx, confKey, string(data))
	if err != nil {
		sgErr = sgerrors.NewSGError(open.ResultCode_MatchmakingConfigurationSaveFailure, err.Error())
		return
	}

	return
}

//MatchmakingConf 媒介配置
func (rep *MatchmakingRepository) MatchmakingConf(ctx context.Context, name string) (conf *open.MatchmakingConfiguration, sgErr error) {
	//查询Matchmaking是否建立
	matchmaking, ok := rep.get(name)
	if !ok {
		sgErr = sgerrors.NewSGError(open.ResultCode_MatchmakingNotSetting, fmt.Sprintf("Matchmaking %s is not configured", name))
		return
	}

	return matchmaking.Conf, nil
}

// StartMatchmaking 开始匹配
func (rep *MatchmakingRepository) StartMatchmaking(ctx context.Context, name string, ticketId string, players []*open.MatchPlayer) (sgErr error) {
	ticket := &open.MatchmakingTicket{
		TicketId:          ticketId,
		Players:           players,
		ConfigurationName: name,
		StartTime:         time.Now().Unix(),
	}

	//查询Matchmaking是否建立
	matchmaking, ok := rep.get(name)
	if !ok {
		sgErr = sgerrors.NewSGError(open.ResultCode_MatchmakingNotSetting, fmt.Sprintf("Matchmaking %s is not configured", name))
		return
	}

	err := matchmaking.TicketInput(ticket)
	if err != nil {
		sgErr = sgerrors.NewSGError(open.ResultCode_MatchmakingTicketCannotQueued,
			errors.ErrorStack(err))
		return
	}

	//票据存储
	rep.ticketStore(ticket)

	return
}

// StopMatchmaking 取消匹配 （是否成功依据取消事件，而不是当前返回）
func (rep *MatchmakingRepository) StopMatchmaking(ctx context.Context, ticketId string) (sgErr error) {
	var (
		ticket      *open.MatchmakingTicket
		matchmaking *entities.Matchmaking
		ok          bool
	)
	//获取票据对应的对局配置名
	ticket, ok = rep.ticketGet(ticketId)
	if !ok {
		sgErr = sgerrors.NewSGError(open.ResultCode_MatchmakingTicketCannotQueued, fmt.Sprintf("ticket %s is not exist", ticketId))
		return
	}

	//查询Matchmaking是否建立
	matchmaking, ok = rep.get(ticket.ConfigurationName)
	if !ok {
		sgErr = sgerrors.NewSGError(open.ResultCode_MatchmakingNotSetting, fmt.Sprintf("Matchmaking %s is not configured", ticket.ConfigurationName))
		return
	}

	err := matchmaking.StopMatch(ticketId)
	if err != nil {
		sgErr = sgerrors.NewSGError(open.ResultCode_StopMatchmakingFailed,
			errors.ErrorStack(err))
		return
	}

	return
}

func (rep *MatchmakingRepository) AcceptMatch(ctx context.Context, name string, ticketId string, acceptType open.AcceptanceType, playerIds []int64) (sgErr sgerrors.SGError) {
	//查询Matchmaking是否建立
	matchmaking, ok := rep.get(name)
	if !ok {
		sgErr = sgerrors.NewSGError(open.ResultCode_MatchmakingNotSetting, fmt.Sprintf("Matchmaking %s is not configured", name))
		return
	}
	for _, userId := range playerIds {
		err := matchmaking.AcceptMatch(ticketId, userId, acceptType)
		if err != nil {
			sgErr = sgerrors.NewSGError(open.ResultCode_MatchmakingTicketCannotQueued,
				errors.ErrorStack(err))
			return
		}
	}

	return
}

// ticketStore 票据存储
func (m *MatchmakingRepository) ticketStore(ticket *open.MatchmakingTicket) {
	m.storeGuard.Lock()

	m.memoryStore[ticket.TicketId] = ticket
	m.memoryStoreIds = append(m.memoryStoreIds, ticket.TicketId)
	m.storeGuard.Unlock()

	m.storeCond.Signal()
}

// ticketGet 票据获取
func (m *MatchmakingRepository) ticketGet(ticketId string) (ticket *open.MatchmakingTicket, ok bool) {
	m.storeGuard.RLock()
	defer m.storeGuard.RUnlock()

	ticket, ok = m.memoryStore[ticketId]

	return
}

//loopCheckExpiredTickets 票据过期检查
func (m *MatchmakingRepository) loopCheckExpiredTickets() {
	for {
		m.storeGuard.Lock()
		if len(m.memoryStoreIds) == 0 {
			m.storeCond.Wait()
		}
		m.storeGuard.Unlock()

		m.storeGuard.RLock()
		ttl := m.memoryStore[m.memoryStoreIds[0]].StartTime + maxTicketTTLSeconds - time.Now().Unix()
		if ttl < 0 {
			ttl = 1
		}
		ticker := time.NewTicker(time.Second * time.Duration(ttl))
		m.storeGuard.RUnlock()

		select {
		case now := <-ticker.C:
			ticker.Stop()
			//删除过期的票据
			m.storeGuard.Lock()
			i := 0
			for ; i < len(m.memoryStoreIds); i++ {
				if m.memoryStore[m.memoryStoreIds[i]].StartTime+maxTicketTTLSeconds <= now.Unix() {
					continue
				}

				break
			}

			removeTicketIds := m.memoryStoreIds[0:i]
			log.Printf("remove ticket len %d\n", len(removeTicketIds))
			for _, ticketId := range removeTicketIds {
				//删除媒介中的票据
				matchmaking, ok := m.m[m.memoryStore[m.memoryStoreIds[i]].ConfigurationName]
				if ok {
					matchmaking.TicketRemove(ticketId)
				}
				delete(m.memoryStore, ticketId)
			}
			copy(m.memoryStoreIds, m.memoryStoreIds[i:])
			m.memoryStoreIds = m.memoryStoreIds[:len(m.memoryStoreIds)-len(removeTicketIds)]
			m.storeGuard.Unlock()
		}
	}
}

//DescribeMatchmaking 票证信息查询
func (m *MatchmakingRepository) DescribeMatchmaking(ctx context.Context, ticketIds []string) (tickets []*open.MatchmakingTicket) {
	m.storeGuard.RLock()
	defer m.storeGuard.RUnlock()

	for _, ticketId := range ticketIds {
		ticket, ok := m.memoryStore[ticketId]
		if ok {
			tickets = append(tickets, ticket)
		}
	}

	return
}
