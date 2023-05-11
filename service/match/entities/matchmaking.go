// entities
// @author LanguageY++2013 2022/11/8 09:49
// @company soulgame
package entities

import (
	"fmt"
	"github.com/Languege/flexmatch/service/match/proto/open"
	"github.com/juju/errors"
	"log"
	"sort"
	"sync"
	"time"
	"github.com/Languege/flexmatch/service/match/pubsub"
)

type Matchmaking struct {
	//对局配置
	Conf *open.MatchmakingConfiguration
	//规则集封装
	rsw *MatchmakingRuleSetWrapper

	//事件处理通知
	eventPubs pubsub.MultiPublisher

	//票据队列
	TicketQueue TicketQueue

	//票据批次通道
	batchTicketChan chan []*open.MatchmakingTicket

	//票据批次处理器 (TODO:如何根据负载动态调整处理器的数量？)
	BatchTicketProcessors []*BatchTicketProcessor

	//票据字典 key-票据唯一ID (TODO:所有媒介共用一个，还是每个媒介独立)
	ticketMap map[string]*open.MatchmakingTicket
	//票据并发保护
	ticketGuard *sync.Mutex

	//对局字典 key-对局唯一ID
	matchMap map[string]*Match
	//对局数据保护
	matchGuard *sync.Mutex
	//匹配成功计数
	MatchSucceedCounter int64
}

func NewMatchmaking(conf *open.MatchmakingConfiguration) *Matchmaking {
	e := &Matchmaking{
		Conf:            conf,
		//eventPubs:       newMatcheventPubscribeManager(conf.MatchEventQueueTopic),
		batchTicketChan: make(chan []*open.MatchmakingTicket, 10),
		ticketMap:       map[string]*open.MatchmakingTicket{},
		ticketGuard:     &sync.Mutex{},
		matchMap:        map[string]*Match{},
		matchGuard:      &sync.Mutex{},
	}

	//规则集解析封装
	e.rsw = NewMatchmakingRuleSetWrapper(e.Conf.RuleSet)

	e.TicketQueue = NewRealTicketQueue(e.batchTicketChan, e.rsw.MatchPlayerNum,
		WithTicketQueueTimeout(e.Conf.RequestTimeoutSeconds))

	//启动批前排序监听，对批次进行排序化划分
	go e.TicketWatch()

	//根据配置启动处理器相关
	e.BatchTicketProcessors = []*BatchTicketProcessor{
		MewBatchTicketProcessor(e),
	}

	return e
}

// TicketInput 票据输入
func (e *Matchmaking) TicketInput(ticket *open.MatchmakingTicket) (err error) {
	//保存票据
	err = e.TicketSave(ticket)
	if err != nil {
		err = errors.Trace(err)
		return
	}

	return e.TicketQueued(ticket)
}

// TicketQueued  票据入队列
func (e *Matchmaking) TicketQueued(ticket *open.MatchmakingTicket) (err error) {
	//票据入队列
	err = e.TicketQueue.Input(ticket)
	if err != nil {
		err = errors.Trace(err)
		return
	}

	//更新票据状态-队列中 (默认值)
	ticket.Status = open.MatchmakingTicketStatus_QUEUED.String()
	//设置预计时间
	ticket.EstimatedWaitTime = e.EstimatedWaitTime()
	return
}

// checkTicketsTimeout 检测票据是否超时
func (e *Matchmaking) checkTicketsTimeout(tickets []*open.MatchmakingTicket) (ret []*open.MatchmakingTicket) {
	ret = make([]*open.MatchmakingTicket, 0, len(tickets))
	now := time.Now().Unix()
	timeoutTickets := make([]*open.MatchmakingTicket, 0, len(tickets)/2)
	for _, t := range tickets {
		if now-t.StartTime >= e.Conf.RequestTimeoutSeconds {
			t.Status = open.MatchmakingTicketStatus_TIMED_OUT.String()
			timeoutTickets = append(timeoutTickets, t)
			continue
		}

		ret = append(ret, t)
	}

	if len(timeoutTickets) > 0 {
		//构建票据超时事件，并推送到事件队列
		timeOutEvent := &open.MatchEvent{
			MatchEventType: open.MatchEventType_MatchmakingTimedOut,
			Tickets:        timeoutTickets,
			Reason:         "TimedOut",
			Message:        "tickets are timeout when they were popped from batch tickets channel",
		}

		publisher.Send(e.Conf.MatchEventQueueTopic, timeOutEvent)
	}

	return
}

// preBatchSort 批前排序
func (e *Matchmaking) preBatchSort(tickets []*open.MatchmakingTicket) {
	if e.rsw.preBatchSortAlgorithm == nil || len(e.rsw.Algorithm.SortByAttributes) == 0 {
		return
	}

	//排序属性
	sort.Slice(tickets, func(i, j int) bool {
		//获取票据属性
		for _, attrName := range e.rsw.Algorithm.SortByAttributes {
			vi := getTicketFloatAttributeValue(tickets[i], attrName, "avg")
			vj := getTicketFloatAttributeValue(tickets[j], attrName, "avg")
			if vi != vj {
				return vi < vj
			}
		}

		return false
	})
}

func (e *Matchmaking) batchProcessorNum() int {
	return len(e.BatchTicketProcessors)
}

// randomDivideBatch 随机划分批次
func (e *Matchmaking) randomDivideBatch(tickets []*open.MatchmakingTicket) (batches []*[]*open.MatchmakingTicket) {
	batchProcessorNum := e.batchProcessorNum()
	batchSize := len(tickets) / batchProcessorNum
	if batchSize < int(e.rsw.MatchPlayerNum) {
		batchSize = int(e.rsw.MatchPlayerNum)
	}

	if batchSize >= len(tickets) {
		return []*[]*open.MatchmakingTicket{&tickets}
	}

	batchNum := len(tickets) / batchSize
	for i := 0; i < batchNum; i++ {
		var batch []*open.MatchmakingTicket
		if i < batchNum-1 {
			batch = tickets[i*batchSize : (i+1)*batchSize]
		} else {
			batch = tickets[i*batchSize : (i+1)*batchSize]
		}

		batches = append(batches, &batch)
	}

	return
}

// divideBatch 划分批次
func (e *Matchmaking) divideBatch(tickets []*open.MatchmakingTicket) (batches []*[]*open.MatchmakingTicket) {
	if e.rsw.batchDistanceRule != nil {
		//设置了批次距离规则
		batches = e.rsw.DivideBatch(tickets)
		return
	}

	//默认规则，随机划分批次
	return e.randomDivideBatch(tickets)
}

// TicketWatch 信号监听：
//  1. 票据输入
//  2. 票据超时检测
//  3. 关闭检测
func (e *Matchmaking) TicketWatch() {
	for {
		select {
		case tickets, ok := <-e.batchTicketChan:
			if !ok && len(tickets) == 0 {
				log.Println("票据已关闭并且队列内无缓冲数据,协程退出")
				return
			}

			//票据超时检测
			tickets = e.checkTicketsTimeout(tickets)
			if len(tickets) == 0 {
				break
			}

			//修改票据状态为-SEARCHING
			e.TicketSearching(tickets)
			//批前排序算法
			e.preBatchSort(tickets)

			//划分批次
			batches := e.divideBatch(tickets)

			//随机分配到批次处理器
			e.DispatchBatches(batches)
		}
	}
}

func (e *Matchmaking) TicketSearching(tickets []*open.MatchmakingTicket) {
	//更新状态
	for _, ticket := range tickets {
		ticket.Status = open.MatchmakingTicketStatus_SEARCHING.String()
	}

	//发送事件
	ev := &open.MatchEvent{
		MatchEventType:      open.MatchEventType_MatchmakingSearching,
		Tickets:             tickets,
		EstimatedWaitMillis: 5000,
	}

	publisher.Send(e.Conf.MatchEventQueueTopic, ev)
}

// DispatchBatches 票据批次分派
func (e *Matchmaking) DispatchBatches(batches []*[]*open.MatchmakingTicket) {
	for _, batch := range batches {
		e.BatchTicketProcessors[0].Input(*batch)
	}
}

// 数据清理
func (e *Matchmaking) CleanUp() {

}

// 配置重载
func (e *Matchmaking) Reload(conf *open.MatchmakingConfiguration) {

}

// TicketFind	票据查找
func (t *Matchmaking) TicketFind(ticketId string) (ticket *open.MatchmakingTicket, ok bool) {
	t.ticketGuard.Lock()
	defer t.ticketGuard.Unlock()

	ticket, ok = t.ticketMap[ticketId]
	return
}

// TicketInsert 票据插入
func (t *Matchmaking) TicketInsert(ticket *open.MatchmakingTicket) {
	t.ticketGuard.Lock()
	defer t.ticketGuard.Unlock()

	t.ticketMap[ticket.TicketId] = ticket
	return
}

// TicketSave 票据保存
func (t *Matchmaking) TicketSave(ticket *open.MatchmakingTicket) error {
	t.ticketGuard.Lock()
	defer t.ticketGuard.Unlock()

	_, ok := t.ticketMap[ticket.TicketId]
	if ok {
		return fmt.Errorf("ticket %s has exist", ticket.TicketId)
	}

	if (ticket.Status == "" || ticket.Status == open.MatchmakingTicketStatus_QUEUED.String()) &&
		ticket.StartTime == 0 {
		ticket.StartTime = time.Now().Unix()
	}

	t.ticketMap[ticket.TicketId] = ticket

	return nil
}

//TicketRemove 票据删除
func (t *Matchmaking) TicketRemove(ticketId string) {
	t.ticketGuard.Lock()
	defer t.ticketGuard.Unlock()

	delete(t.ticketMap, ticketId)
}

// MatchInsert 对局数据插入
func (e *Matchmaking) MatchInsert(match *Match) {
	e.matchGuard.Lock()
	defer e.matchGuard.Unlock()

	e.matchMap[match.MatchId] = match
}

// MatchFind	匹配查找
func (e *Matchmaking) MatchFind(matchId string) (match *Match, ok bool) {
	e.matchGuard.Lock()
	defer e.matchGuard.Unlock()

	match, ok = e.matchMap[matchId]
	return
}

//MatchRemove 匹配移除,接收匹配完成时调用（接受超时、全部接收、任意拒绝）
func (e *Matchmaking) MatchRemove(matchId string) {
	e.matchGuard.Lock()
	defer e.matchGuard.Unlock()

	delete(e.matchMap, matchId)
	return
}

// AcceptMatch 玩家接受对局
func (e *Matchmaking) AcceptMatch(ticketId string, userId int64, acceptType open.AcceptanceType) (err error) {
	var (
		ticket *open.MatchmakingTicket
		match  *Match
		ok     bool
	)
	//查找票据
	ticket, ok = e.TicketFind(ticketId)
	if !ok {
		err = errors.Trace(fmt.Errorf("ticket %s not exist", ticketId))
		return
	}

	//查找对局
	match, ok = e.MatchFind(ticket.MatchId)
	if !ok {
		err = errors.Trace(fmt.Errorf("match %s not exist", ticket.MatchId))
		return
	}

	//更新玩家接受对局状态
	for _, player := range ticket.Players {
		if player.UserId == userId {
			player.Accepted = acceptType == open.AcceptanceType_ACCEPT
			if player.Accepted {
				//对局接受事件上报
				match.AcceptMatch(ticketId)
			} else {
				//直接取消对局
				match.RejectMatch(ticketId)
			}
			break
		}
	}

	return
}

// StopMatch 停止/取消匹配
func (m *Matchmaking) StopMatch(ticketId string) (err error) {
	//查看票据是否存在
	ticket, ok := m.TicketFind(ticketId)
	if !ok {
		err = fmt.Errorf("ticket %s not exit", ticketId)
		return
	}

	//验证用户是否可以取消匹配，不可取消匹配的阶段：PLACING、COMPLETED
	if ticket.Status == open.MatchmakingTicketStatus_PLACING.String() ||
		ticket.Status == open.MatchmakingTicketStatus_COMPLETED.String() {
		err = fmt.Errorf("ticket %s cannot be canceled when it's in %s", ticketId, ticket.Status)
		return
	}

	//标识取消请求
	ticket.CancelRequest = true

	return
}

//EstimatedWaitTime 预计耗时
func (m *Matchmaking) EstimatedWaitTime() int64 {
	var total int64
	for _, p := range m.BatchTicketProcessors {
		total += p.EstimatedWaitTime()
	}

	return total / int64(len(m.BatchTicketProcessors))
}
