// entities
// @author LanguageY++2013 2023/5/18 23:16
// @company soulgame
package prometheus

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"net/http"
)

type MatchMetrics struct {
	//票据入队列计数
	TicketQueuedCounter *prometheus.CounterVec
	//票据回填计数
	TicketBackfillCounter *prometheus.CounterVec
	//票据取消计数
	TicketCancelledCounter *prometheus.CounterVec
	//票据超时计数
	TicketTimeOutCounter *prometheus.CounterVec
	//票据匹配失败计数
	TicketFailedCounter *prometheus.CounterVec

	//处于队列中的票据数
	TicketQueuedGather *prometheus.GaugeVec
	//处于搜索中的票据数
	TicketSearchingGather *prometheus.GaugeVec
	//处于请求接收状态中的票据数
	TicketRequiresAcceptanceGather *prometheus.GaugeVec
	//处于安排游戏状态的票据数
	TicketPlacingGather *prometheus.GaugeVec

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

	m.TicketQueuedCounter = prometheus.NewCounterVec(prometheus.CounterOpts{
		Namespace: "flex_match",
		Subsystem: "ticket",
		Name:      "queued",
		Help:      "票据入队列计数",
	}, []string{"topic"})

	prometheus.DefaultRegisterer.Register(m.TicketQueuedCounter)

	m.TicketBackfillCounter = prometheus.NewCounterVec(prometheus.CounterOpts{
		Namespace: "flex_match",
		Subsystem: "ticket",
		Name:      "backfill",
		Help:      "票据回填计数",
	}, []string{"topic"})

	prometheus.DefaultRegisterer.Register(m.TicketBackfillCounter)

	m.TicketCancelledCounter = prometheus.NewCounterVec(prometheus.CounterOpts{
		Namespace: "flex_match",
		Subsystem: "ticket",
		Name:      "canceled",
		Help:      "票据取消计数",
	}, []string{"topic"})

	prometheus.DefaultRegisterer.Register(m.TicketCancelledCounter)

	m.TicketTimeOutCounter = prometheus.NewCounterVec(prometheus.CounterOpts{
		Namespace: "flex_match",
		Subsystem: "ticket",
		Name:      "timeout",
		Help:      "票据超时计数",
	}, []string{"topic"})

	prometheus.DefaultRegisterer.Register(m.TicketCancelledCounter)

	m.TicketQueuedGather = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: "flex_match",
		Subsystem: "ticket",
		Name:      "in_queued",
		Help:      "处于队列中的票据数",
	}, []string{"topic"})

	prometheus.DefaultRegisterer.Register(m.TicketQueuedGather)

	m.TicketSearchingGather = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: "flex_match",
		Subsystem: "ticket",
		Name:      "in_searching",
		Help:      "处于搜索中的票据数",
	}, []string{"topic"})

	prometheus.DefaultRegisterer.Register(m.TicketSearchingGather)

	m.TicketRequiresAcceptanceGather = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: "flex_match",
		Subsystem: "ticket",
		Name:      "in_requires_acceptance",
		Help:      "处于请求接收状态中的票据数",
	}, []string{"topic"})

	prometheus.DefaultRegisterer.Register(m.TicketRequiresAcceptanceGather)

	m.TicketPlacingGather = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: "flex_match",
		Subsystem: "ticket",
		Name:      "in_placing",
		Help:      "处于安排游戏状态的票据数",
	}, []string{"topic"})

	prometheus.DefaultRegisterer.Register(m.TicketPlacingGather)

	m.PotentialMatchCounter = prometheus.NewCounterVec(prometheus.CounterOpts{
		Namespace: "flex_match",
		Subsystem: "match",
		Name:      "potential",
		Help:      "潜在匹配计数",
	}, []string{"topic"})

	prometheus.DefaultRegisterer.Register(m.PotentialMatchCounter)

	m.CompletedMatchCounter = prometheus.NewCounterVec(prometheus.CounterOpts{
		Namespace: "flex_match",
		Subsystem: "match",
		Name:      "completed",
		Help:      "接收完成对局计数 (超时、部分接收、任意拒绝)",
	}, []string{"topic"})

	prometheus.DefaultRegisterer.Register(m.CompletedMatchCounter)

	m.TimeOutMatchCounter = prometheus.NewCounterVec(prometheus.CounterOpts{
		Namespace: "flex_match",
		Subsystem: "match",
		Name:      "accept_timeout",
		Help:      "接收超时对局计数",
	}, []string{"topic"})

	prometheus.DefaultRegisterer.Register(m.TimeOutMatchCounter)

	m.SucceedMatchCounter = prometheus.NewCounterVec(prometheus.CounterOpts{
		Namespace: "flex_match",
		Subsystem: "match",
		Name:      "succeed",
		Help:      "成功对局计数",
	}, []string{"topic"})

	prometheus.DefaultRegisterer.Register(m.SucceedMatchCounter)

	m.RejectedMatchCounter = prometheus.NewCounterVec(prometheus.CounterOpts{
		Namespace: "flex_match",
		Subsystem: "match",
		Name:      "rejected",
		Help:      "拒绝对局计数",
	}, []string{"topic"})

	prometheus.DefaultRegisterer.Register(m.RejectedMatchCounter)

	return m
}

func init() {
	http.Handle("/metrics", promhttp.Handler())
}
