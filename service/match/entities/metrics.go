// entities
// @author LanguageY++2013 2022/11/24 10:02
// @company soulgame
package entities

import (
	"github.com/prometheus/client_golang/prometheus"
)

type MatchMetrics struct {
	//票据输入计数
	TicketInputCounter *prometheus.CounterVec
	//票据回填计数
	TicketBackfillCounter *prometheus.CounterVec
	//票据取消计数
	TicketCancelledCounter *prometheus.CounterVec
	//票据超时计数
	TicketTimeOutCounter *prometheus.CounterVec
	//票据匹配失败计数
	TicketFailedCounter *prometheus.CounterVec

	//处于队列中的票据数
	TicketQueuedGather *prometheus.Gatherer
	//处于搜索中的票据数
	TicketSearchingGather *prometheus.Gatherer
	//处于请求接收状态中的票据数
	TicketRequiresAcceptanceGather *prometheus.Gatherer
	//处于安排游戏状态的票据数
	TicketPlacingGather *prometheus.Gatherer

	//潜在对局计数
	PotentialMatchCounter *prometheus.CounterVec
	//接收完成对局计数 (超时、部分接收、任意拒绝)
	CompletedMatchCounter *prometheus.CounterVec
	//接收超时对局计数
	TimeOutMatchCounter *prometheus.CounterVec
	//成功对局计数
	SucceedMatchCounter *prometheus.CounterVec
	//拒绝对局计数
	RejectedMatchCounter *prometheus.CounterVec
}

func NewMatchMetrics() *MatchMetrics {
	m := &MatchMetrics{}

	m.PotentialMatchCounter = prometheus.NewCounterVec(prometheus.CounterOpts{
		Namespace: "flex_match",
		Subsystem: "match",
		Name:      "potential",
	}, []string{"name"})

	prometheus.DefaultRegisterer.Register(m.PotentialMatchCounter)

	return m
}
