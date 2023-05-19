// prometheus
// @author LanguageY++2013 2023/5/18 23:41
// @company soulgame
package prometheus

import (
	"github.com/Languege/flexmatch/service/match/entities"
	"github.com/Languege/flexmatch/service/match/proto/open"
)

type MetricsPublisher struct {
	MatchMetrics
}

func NewMetricsPublisher() *MetricsPublisher {
	return &MetricsPublisher{
		MatchMetrics: *NewMatchMetrics(),
	}
}

func (p MetricsPublisher) Name() string {
	return "metrics"
}

//票据的状态 searching -> potential_match;    TODO: 需要解决票据重新入队列的指标问题, 缺少票据入队列事件
func (p MetricsPublisher) Send(topic string, ev *open.MatchEvent) error {
	ticketCount := len(ev.Tickets)
	switch ev.MatchEventType {
	case open.MatchEventType_MatchmakingQueued:
		//searching 阶段的 queue -> searching -> queue -> searching 不计入
		if ev.Reason != entities.StatusReasonSearchingFailedBackfill {
			//处于队列中的票据数增加
			p.TicketQueuedGather.WithLabelValues(topic).Add(float64(ticketCount))
			//票据入队列
			p.TicketQueuedCounter.WithLabelValues(topic).Add(float64(ticketCount))
		}

		if ev.Reason == entities.StatusReasonSearchingFailedBackfill || ev.Reason == entities.StatusReasonMatchRejectBackfill {
			p.TicketBackfillCounter.WithLabelValues(topic).Add(float64(ticketCount))
		}
	case open.MatchEventType_MatchmakingSearching:
		if ev.Reason != entities.StatusReasonSearchingFailedBackfill {
			//处于队列中的票据数减少
			p.TicketQueuedGather.WithLabelValues(topic).Sub(float64(ticketCount))
			//处于搜索中的票据增加
			p.TicketSearchingGather.WithLabelValues(topic).Add(float64(ticketCount))
		}
	case open.MatchEventType_PotentialMatchCreated:
		//处于搜索中的票据减少
		p.TicketSearchingGather.WithLabelValues(topic).Sub(float64(ticketCount))
		//潜在对局+1
		p.PotentialMatchCounter.WithLabelValues(topic).Add(1)
		if ev.AcceptanceRequired {
			p.TicketRequiresAcceptanceGather.WithLabelValues(topic).Add(float64(ticketCount))
		}
	//case open.MatchEventType_AcceptMatch:
	//	//处于请求接收的票据增加 接收事件会重复多次，不能作为指标
	//	p.TicketRequiresAcceptanceGather.WithLabelValues(topic).Add(float64(ticketCount))
	case open.MatchEventType_AcceptMatchCompleted:
		//匹配完成+1
		p.CompletedMatchCounter.WithLabelValues(topic).Add(1)

		switch ev.AcceptanceCompletedReason {
		case entities.AcceptanceCompletedReasonTimeOut:
			p.TimeOutMatchCounter.WithLabelValues(topic).Add(float64(ticketCount))
		case entities.AcceptanceCompletedReasonRejection:
			p.RejectedMatchCounter.WithLabelValues(topic).Add(float64(ticketCount))
		default:
			p.TicketPlacingGather.WithLabelValues(topic).Add(float64(ticketCount))
		}
		if ev.AcceptanceRequired {
			p.TicketRequiresAcceptanceGather.WithLabelValues(topic).Sub(float64(ticketCount))
		}
	case open.MatchEventType_MatchmakingSucceeded:
		p.SucceedMatchCounter.WithLabelValues(topic).Add(1)

		p.TicketPlacingGather.WithLabelValues(topic).Sub(float64(ticketCount))
	case open.MatchEventType_MatchmakingTimedOut:
		//匹配超时
		p.TicketTimeOutCounter.WithLabelValues(topic).Add(1)
		//处于搜索中的票据减少
		p.TicketSearchingGather.WithLabelValues(topic).Sub(float64(ticketCount))
	case open.MatchEventType_MatchmakingCancelled:
		//票据取消次数增加
		p.TicketCancelledCounter.WithLabelValues(topic).Add(float64(ticketCount))
	case open.MatchEventType_MatchmakingFailed:
		//票据匹配失败计数增加
		p.TicketFailedCounter.WithLabelValues(topic).Add(float64(ticketCount))
		if ev.MatchId == "" {
			//玩法形成对局，处于搜索中的票据数量减少
			p.TicketSearchingGather.WithLabelValues(topic).Sub(float64(ticketCount))
		}
	}

	return nil
}
