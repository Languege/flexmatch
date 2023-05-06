// entities
// @author LanguageY++2013 2022/11/8 18:31
// @company soulgame
package entities

import (
	"errors"
	"github.com/Languege/flexmatch/service/match/proto/open"
	"time"
)

const (
	//票据队列缓冲区
	maxTicketQueueBufferSize = 2000

	//票据默认超时时间
	defaultTicketTimeoutSeconds = 30

	//票据输入通道大小
	inputTicketChanBufferSize = 2000
)

var (
	ErrOutOfTicketQueueSize = errors.New("out of ticket queue size")
)

type TicketQueue interface {
	//票据输入
	Input(ticket *open.MatchmakingTicket) error

	//票据监听
	Watch()
}

type realTicketQueue struct {
	//票据输入通道
	input chan *open.MatchmakingTicket
	//票据缓冲队列
	tickets []*open.MatchmakingTicket
	//首个票据的创建时间
	firstTicketCreationTime int64

	//批次票据发送通道
	batchTicketChan chan<- []*open.MatchmakingTicket

	//批次最小所需票据
	batchMinTicketNum int32

	//票据超时时间(s)
	ticketTimeoutSeconds int64
}

type TicketQueueOption func(q *realTicketQueue)

func WithTicketQueueTimeout(timeoutSeconds int64) TicketQueueOption {
	return func(q *realTicketQueue) {
		q.ticketTimeoutSeconds = timeoutSeconds
	}
}

func NewRealTicketQueue(batchTicketChan chan<- []*open.MatchmakingTicket, batchMinTicketNum int32, opts ...TicketQueueOption) TicketQueue {
	q := &realTicketQueue{
		input:                make(chan *open.MatchmakingTicket, inputTicketChanBufferSize),
		tickets:              make([]*open.MatchmakingTicket, 0, maxTicketQueueBufferSize),
		batchTicketChan:      batchTicketChan,
		batchMinTicketNum:    batchMinTicketNum,
		ticketTimeoutSeconds: defaultTicketTimeoutSeconds,
	}

	for _, opt := range opts {
		opt(q)
	}

	go q.Watch()

	return q
}

// Input 票据输入，票据状态变更为队列中
func (q *realTicketQueue) Input(ticket *open.MatchmakingTicket) error {
	//判断通道大小
	if len(q.input) >= inputTicketChanBufferSize {
		return ErrOutOfTicketQueueSize
	}

	q.input <- ticket

	return nil
}

// 判断是否可以复制元素出队列
func (q *realTicketQueue) canMove() bool {
	return time.Now().Unix()-q.firstTicketCreationTime >= q.ticketTimeoutSeconds/2 || len(q.tickets) >= int(q.batchMinTicketNum)
}

// Reset 队列重置
func (q *realTicketQueue) Reset() {
	q.tickets = q.tickets[0:0]
}

func (q *realTicketQueue) Watch() {
	tick := time.Tick(time.Second)
	for {
		select {
		case <-tick:
			if q.canMove() == false {
				continue
			}
			//复制元素出队列
			batch := make([]*open.MatchmakingTicket, 0, len(q.tickets))
			for _, ticket := range q.tickets {
				batch = append(batch, ticket)
			}
			q.Reset()

			q.batchTicketChan <- batch
		case ticket := <-q.input:
			if len(q.tickets) == 0 {
				q.firstTicketCreationTime = ticket.StartTime
			}
			q.tickets = append(q.tickets, ticket)
		}
	}
}
