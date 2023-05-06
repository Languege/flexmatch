// entities
// @author LanguageY++2013 2022/11/9 15:11
// @company soulgame
package entities

import (
	"github.com/Languege/flexmatch/service/match/proto/open"
	"math"
	"strconv"
	"sync"
)

// Ticket 并发保护ticket
type Ticket struct {
	*open.MatchmakingTicket
	guard sync.Mutex
}

//TODO:属性值是否需要抽象成一个类

// getPlayerFloatAttributeValue 获取玩家float属性值
func getPlayerFloatAttributeValue(player *open.MatchPlayer, attrName string) (attrValue float64) {
	for _, attr := range player.Attributes {
		if attr.Name == attrName {
			attrValue, _ = strconv.ParseFloat(attr.Value, 64)
			return
		}
	}

	return
}

// getTicketFloatAttributeValue 获取票据float属性值
func getTicketFloatAttributeValue(ticket *open.MatchmakingTicket, attrName string, partyAggregation string) (attrValue float64) {
	if len(ticket.Players) == 0 {
		return getPlayerFloatAttributeValue(ticket.Players[0], attrName)
	}

	totalValue := 0.0
	minValue := math.MaxFloat64
	maxValue := -math.MaxFloat64
	for _, player := range ticket.Players {
		val := getPlayerFloatAttributeValue(player, attrName)
		if val > maxValue {
			maxValue = val
		}
		if val < minValue {
			minValue = val
		}
		totalValue += val
	}

	//多人 根据partyAggregation聚合函数计算
	switch partyAggregation {
	case "min":
		return minValue
	case "max":
		return maxValue
	case "sum":
		return totalValue
	default:
		//默认平均
		attrValue = totalValue/float64(len(ticket.Players)) + (maxValue - minValue)
		return
	}
}

// 票据组合
type CombinationTickets struct {
	com           *combination
	selectTickets []*open.MatchmakingTicket
}
