// entities
// @author LanguageY++2013 2022/11/8 18:28
// @company soulgame
package entities

import (
	"github.com/Languege/flexmatch/service/match/proto/open"
	"math"
	"sort"
	"time"
	"sync/atomic"
)

const (
	//最小潜在匹配耗时
	minPotentialMatchSeconds = 5
	//最近耗时评估所需票据的最大数量
	recentEstimateTicketMaxNum = 50
)

//票据批次处理器

type BatchTicketProcessor struct {
	//媒介
	Matchmaking *Matchmaking

	//票据通道
	batchCh chan []*open.MatchmakingTicket

	//请求用户接收匹配通道
	requiresAcceptanceMatchChan chan *Match

	//创建游戏会话匹配通道
	placingMatchChan chan *Match

	//保留最近10次潜在配对的耗时单位秒
	recentPotentialMatchTickets []*open.MatchmakingTicket
}

func MewBatchTicketProcessor(Matchmaking *Matchmaking) *BatchTicketProcessor {
	proc := &BatchTicketProcessor{
		Matchmaking:                 Matchmaking,
		batchCh:                     make(chan []*open.MatchmakingTicket, 10),
		requiresAcceptanceMatchChan: make(chan *Match, 10),
		placingMatchChan:            make(chan *Match, 10),
	}

	go proc.run()

	return proc
}

func (proc *BatchTicketProcessor) Input(batch []*open.MatchmakingTicket) {
	proc.batchCh <- batch
}

// sort 对票据进行排序
func (proc *BatchTicketProcessor) sort(tickets []*open.MatchmakingTicket) {
	sort.Slice(tickets, func(i, j int) bool {
		for _, sortRule := range proc.Matchmaking.rsw.sortRule {
			switch sortRule.Type {
			case open.MatchmakingRuleType_MatchmakingRuleType_AbsoluteSort:
				//绝对距离排序
				attrValuei := getTicketFloatAttributeValue(tickets[i], sortRule.SortAttribute, sortRule.PartyAggregation)
				attrValuej := getTicketFloatAttributeValue(tickets[j], sortRule.SortAttribute, sortRule.PartyAggregation)

				if attrValuei == attrValuej {
					//进行一个规则排序
					break
				}

				if sortRule.SortDirection == open.SortDirectionType_Ascending {
					return attrValuei < attrValuej
				} else {
					return attrValuei > attrValuej
				}
			case open.MatchmakingRuleType_MatchmakingRuleType_DistanceSort:
				//相对距离排序
				firstAttrValue := getTicketFloatAttributeValue(tickets[0], sortRule.SortAttribute, sortRule.PartyAggregation)
				attrValuei := getTicketFloatAttributeValue(tickets[i], sortRule.SortAttribute, sortRule.PartyAggregation)
				attrValuej := getTicketFloatAttributeValue(tickets[j], sortRule.SortAttribute, sortRule.PartyAggregation)

				distancei := math.Abs(attrValuei - firstAttrValue)
				distancej := math.Abs(attrValuej - firstAttrValue)
				if distancei == distancej {
					//进行一个规则排序
					break
				}

				if sortRule.SortDirection == open.SortDirectionType_Ascending {
					return distancei < distancej
				} else {
					return distancei > distancej
				}
			}
		}

		return false
	})

}

// checkRuleForReferenceAndMeasurements 检测参考值
//
//	返回值： 是否满足规则 true-满足规则 false-不满足规则
func (proc *BatchTicketProcessor) checkRuleForReferenceAndMeasurements(rule *MatchmakingRule,
	reference float64, measurements []float64, tryTimes int, firstTicketCreationTime int64) bool {
	switch rule.Type {
	case open.MatchmakingRuleType_MatchmakingRuleType_Distance:
		//距离规则
		maxDistance := rule.MaxDistance
		//TODO:最大距离计算
		epMaxDistance, ok := proc.Matchmaking.rsw.MaxDistance(rule.Name, rule.MaxDistance, tryTimes, firstTicketCreationTime)
		if ok && epMaxDistance > maxDistance {
			maxDistance = epMaxDistance
		}

		for _, measurement := range measurements {
			if math.Abs(measurement-reference) > maxDistance {
				return false
			}
		}

		return true
	case open.MatchmakingRuleType_MatchmakingRuleType_Comparison:
		//比对
		switch rule.Operation {
		case "=":
			for _, measurement := range measurements {
				if measurement != reference {
					return false
				}
			}
			return true
		case "!=":
			for _, measurement := range measurements {
				if measurement == reference {
					return false
				}
			}

			return true
		}
	}

	return false
}

// comparisonMeasurements 当规则为比对规则且无参考值时，测量值之间进行比对
//
//	返回值： 是否满足规则 true-满足规则 false-不满足规则
func (proc *BatchTicketProcessor) comparisonMeasurements(rule *MatchmakingRule, measurements []float64) bool {
	switch rule.Type {
	case open.MatchmakingRuleType_MatchmakingRuleType_Comparison:
		various := map[float64]bool{}
		for _, measurement := range measurements {
			various[measurement] = true
		}
		//比对
		switch rule.Operation {
		case "=":
			return len(various) == 1
		case "!=":
			return len(various) == len(measurements)
		}
	}

	return false
}

// groupTicketsByPlayerNum 将票据按人数划分
func (proc *BatchTicketProcessor) groupTicketsByPlayerNum(tickets []*open.MatchmakingTicket) (ticketGroupByPlayerNum map[int]*[]*open.MatchmakingTicket) {
	ticketGroupByPlayerNum = map[int]*[]*open.MatchmakingTicket{}
	for _, v := range tickets {
		playerNum := len(v.Players)
		if group, ok := ticketGroupByPlayerNum[playerNum]; ok {
			*group = append(*group, v)
		} else {
			ticketGroupByPlayerNum[playerNum] = &[]*open.MatchmakingTicket{v}
		}
	}

	return
}

// selectTeamTickets 选择团队所需票据
//
//	completedTeams: 已经完成配对的团队
//	team: 当前就行配对的团队
func (proc *BatchTicketProcessor) selectTeamTickets(completedTeams []*Team, team *Team,
	needPlayerNum int, tickets []*open.MatchmakingTicket, tryTimes int, creationTime int64) (selectTickets []*open.MatchmakingTicket) {
	//将票据按人数划分
	ticketGroupByPlayerNum := proc.groupTicketsByPlayerNum(tickets)

	combinations := arrangementCombination[needPlayerNum]
	availableCombinations := []*CombinationTickets{}
	for _, com := range combinations {
		//假定组合可用
		available := true
		availableTickets := []*open.MatchmakingTicket{}
		for _, v := range com {
			ticketsByPlayer, ok := ticketGroupByPlayerNum[v.Value]
			if !ok {
				available = false
				break
			}

			if len(*ticketsByPlayer) < v.Quality {
				available = false
				break
			}

			//复制team, 将票据添加到team
			teamCopy := team.Copy()

			availableTicketsByGroup := []*open.MatchmakingTicket{}
			for _, ticket := range *ticketsByPlayer {
				teamCopy.Tickets = append(teamCopy.Tickets, ticket)

				//对teamCopy进行规则集验证
				matched := true
				for _, rule := range proc.Matchmaking.rsw.filterRule {
					//计算测量值
					measurements := proc.Matchmaking.rsw.GetPropertyExpressionMeasurements(rule.MeasurementsParser, append(completedTeams, teamCopy))

					//判断参考值referenceValue是否配置，配置则测量值与参考值进行对比，未配置则测量值之间进行对比
					if rule.ReferenceValueParser != nil {
						//参考值
						reference := proc.Matchmaking.rsw.GetPropertyExpressionReferenceValue(rule.ReferenceValueParser, append(completedTeams, teamCopy))

						//检测测量值是否匹配规则
						matched = proc.checkRuleForReferenceAndMeasurements(rule, reference, measurements, tryTimes, creationTime)
					} else {
						matched = proc.comparisonMeasurements(rule, measurements)
					}

					if matched == false {
						//ticket不满足匹配
						break
					}
				}

				if matched {
					availableTicketsByGroup = append(availableTicketsByGroup, ticket)
					if len(availableTicketsByGroup) == v.Quality {
						break
					}
				}
			}

			if len(availableTicketsByGroup) < v.Quality {
				//可用票据不足，改组合不可用
				available = false
				break
			}

			availableTickets = append(availableTickets, availableTicketsByGroup...)
		}

		if !available {
			continue
		}

		availableCombinations = append(availableCombinations, &CombinationTickets{
			com:           &com,
			selectTickets: availableTickets,
		})
	}

	if len(availableCombinations) == 0 {
		return
	}

	if len(availableCombinations) > 1 {
		//优先票据玩家数多的组合
		sort.Slice(availableCombinations, func(i, j int) bool {
			gi := (*(availableCombinations[i].com))[0]
			gj := (*(availableCombinations[j].com))[0]

			return gi.Value > gj.Value
		})
	}

	return availableCombinations[0].selectTickets
}

// removeTickets 返回移除后的票据
func (proc *BatchTicketProcessor) removeTickets(tickets []*open.MatchmakingTicket, removeTickets []*open.MatchmakingTicket) (newTickets []*open.MatchmakingTicket) {
	removeTicketIds := map[string]bool{}
	for _, v := range removeTickets {
		removeTicketIds[v.TicketId] = true
	}

	newTickets = make([]*open.MatchmakingTicket, 0, len(tickets))
	for _, v := range tickets {
		_, ok := removeTicketIds[v.TicketId]
		if ok {
			continue
		}

		newTickets = append(newTickets, v)
	}

	return
}

// tryBuildOneMatch 尝试构建一个对局
func (proc *BatchTicketProcessor) tryBuildOneMatch(tickets []*open.MatchmakingTicket, tryTimes int,
	creationTime int64) (match *Match, leftTickets []*open.MatchmakingTicket) {
	matchTeams := []*Team{}
	//保留原票据
	leftTickets = make([]*open.MatchmakingTicket, len(tickets), len(tickets))
	copy(leftTickets, tickets)
	for _, teamConf := range proc.Matchmaking.Conf.RuleSet.Teams {
		team := newTeam(teamConf)

		team.Tickets = proc.selectTeamTickets(matchTeams, team, int(teamConf.PlayerNumber), leftTickets, tryTimes, creationTime)
		if len(team.Tickets) == 0 {
			//没有可用票据，进行下一轮匹配
			goto next
		}

		//移除已成团票据
		leftTickets = proc.removeTickets(leftTickets, team.Tickets)

		matchTeams = append(matchTeams, team)
	}

	//形成匹配
	match = NewMatch(proc.Matchmaking, matchTeams)

	//添加到匹配列表
	proc.Matchmaking.MatchInsert(match)

	//更新票据状态为REQUIRES_ACCEPTANCE
	for _, matchTeam := range matchTeams {
		for _, ticket := range matchTeam.Tickets {
			ticket.MatchId = match.MatchId
			ticket.Status = open.MatchmakingTicketStatus_REQUIRES_ACCEPTANCE.String()
		}
	}
next:
	//更新剩余票据
	if match == nil {
		leftTickets = tickets
	}

	return
}

//handleTicketCancelRequests 处理票据取消请求
func (proc *BatchTicketProcessor) handleTicketCancelRequests(tickets []*open.MatchmakingTicket) (newTickets []*open.MatchmakingTicket) {
	newTickets = make([]*open.MatchmakingTicket, 0, len(tickets))
	for _, ticket := range tickets {
		if ticket.CancelRequest {
			ticket.Status = open.MatchmakingTicketStatus_CANCELLED.String()
			cancelMatchEvent := &open.MatchEvent{
				MatchEventType: open.MatchEventType_MatchmakingCancelled,
				Tickets:        []*open.MatchmakingTicket{ticket},
				Message:        "Cancelled by request.",
			}

			proc.Matchmaking.eventSubs.MatchEventInput(cancelMatchEvent)
			continue
		}
		newTickets = append(newTickets, ticket)
	}

	return
}

// loopBuildMatch 循环构建票据
func (proc *BatchTicketProcessor) loopBuildMatch(tickets []*open.MatchmakingTicket) {
	go func() {
		//最大重试次数
		maxRetryTimes := 10
		for {

			//对票据进行排序
			proc.sort(tickets)

			var (
				match *Match
			)
			i := 0
			for ; i < maxRetryTimes; i++ {
				if len(tickets) == 0 {
					//没有剩余票据，直接返回
					return
				}
				match, tickets = proc.tryBuildOneMatch(tickets, i, tickets[0].StartTime)
				if match != nil {
					proc.matchPotential(match)
					break
				}

				//对取消请求进行处理
				tickets = proc.handleTicketCancelRequests(tickets)

				//主要距离规则的最大距离扩展
				time.Sleep(time.Second)
			}

			if i == maxRetryTimes {
				//超时检测
				tickets = proc.Matchmaking.checkTicketsTimeout(tickets)
				//无法形成对局，票据冲入队列
				for _, ticket := range tickets {
					proc.ticketMaybeQueue(ticket)
				}
				return
			}
		}
	}()
}

// matchSucceed 匹配成功
func (proc *BatchTicketProcessor) matchPotential(match *Match) {
	proc.requiresAcceptanceMatchChan <- match
	//添加统计
	proc.addMatchCost(match)
	atomic.AddInt64(&proc.Matchmaking.MatchSucceedCounter, 1)

	//匹配成功事件
	matchCreatedEvent := &open.MatchEvent{
		MatchEventType:     open.MatchEventType_PotentialMatchCreated,
		Tickets:            match.AllTickets(),
		MatchId:            match.MatchId,
		AcceptanceTimeout:  proc.Matchmaking.Conf.AcceptanceTimeoutSeconds,
		AcceptanceRequired: proc.Matchmaking.Conf.AcceptanceRequired,
	}

	proc.Matchmaking.eventSubs.MatchEventInput(matchCreatedEvent)
}

// ticketMaybeQueue 票据尝试重入
func (proc *BatchTicketProcessor) ticketMaybeQueue(ticket *open.MatchmakingTicket) {
	//状态判断 (用户可能取消匹配之类)
	if ticket.Status != open.MatchmakingTicketStatus_SEARCHING.String() {
		return
	}

	proc.Matchmaking.TicketQueued(ticket)
}

func (proc *BatchTicketProcessor) addMatchCost(match *Match) {
	now := time.Now().Unix()
	for _, ticket := range match.AllTickets() {
		if len(proc.recentPotentialMatchTickets) >= recentEstimateTicketMaxNum {
			copy(proc.recentPotentialMatchTickets, proc.recentPotentialMatchTickets[1:])
			proc.recentPotentialMatchTickets = proc.recentPotentialMatchTickets[:len(proc.recentPotentialMatchTickets)-1]
		}

		ticket.PotentialMatchCostSeconds = now - ticket.StartTime
		proc.recentPotentialMatchTickets = append(proc.recentPotentialMatchTickets, ticket)
	}
}

//EstimatedWaitTime 预计等待时间单位秒，不能低于保底值
func (proc *BatchTicketProcessor) EstimatedWaitTime() int64 {
	if len(proc.recentPotentialMatchTickets) == 0 {
		return minPotentialMatchSeconds
	}
	var totalCostSeconds int64 = 0
	for _, ticket := range proc.recentPotentialMatchTickets {
		totalCostSeconds += ticket.PotentialMatchCostSeconds
	}

	avg := totalCostSeconds / int64(len(proc.recentPotentialMatchTickets))

	if avg < minPotentialMatchSeconds {
		return minPotentialMatchSeconds
	}

	return avg
}

// 票据匹配主题
func (proc *BatchTicketProcessor) run() {
	for {
		select {
		case tickets := <-proc.batchCh:
			go proc.loopBuildMatch(tickets)
		case match := <-proc.requiresAcceptanceMatchChan:
			//保存票据匹配耗时

			//是否需要等待用户接受
			if proc.Matchmaking.Conf.AcceptanceRequired == false {
				proc.placingMatchChan <- match
				break
			}

			//开启独立协程定时检测对局用户是否接受
			match.StartAccept(proc.Matchmaking.Conf.AcceptanceTimeoutSeconds, proc.placingMatchChan)
		case match := <-proc.placingMatchChan:
			//请求游戏会话
			match.StartGameSession()
		}
	}
}
