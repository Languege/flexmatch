// entities
// @author LanguageY++2013 2022/11/8 15:49
// @company soulgame
package entities

import (
	"fmt"
	"github.com/Languege/flexmatch/service/match/proto/open"
	"reflect"
	"strings"
	"time"
	"github.com/juju/errors"
	"github.com/Languege/flexmatch/service/match/parser"
	"context"
	"github.com/Languege/flexmatch/service/match/parser/chain"
	"log"
)

type MatchmakingRule struct {
	open.MatchmakingRule
	MeasurementsParser   *parser.PropertyExprParser
	ReferenceValueParser 	 *parser.PropertyExprParser
}

func newMatchmakingRule(rule *open.MatchmakingRule) *MatchmakingRule {
	ruleWrapper := &MatchmakingRule{
		MatchmakingRule:*rule,
	}

	if ruleWrapper.Measurements != "" {
		ruleWrapper.MeasurementsParser = parser.NewPropertyExprParser(ruleWrapper.Measurements)
	}

	if ruleWrapper.ReferenceValue != "" {
		ruleWrapper.ReferenceValueParser = parser.NewPropertyExprParser(ruleWrapper.ReferenceValue)
	}


	return ruleWrapper
}

type MatchmakingRuleSetWrapper struct {
	*open.MatchmakingRuleSet
	ruleMap            map[string]*MatchmakingRule
	playerAttributeMap map[string]*open.PlayerAttribute
	teamMap            map[string]*open.MatchmakingTeamConfiguration

	//对局所需成员数
	MatchPlayerNum int32

	//批前排序算法
	preBatchSortAlgorithm *open.MatchmakingRuleAlgorithm

	//batchDistance 批次距离规则，任意两个票据之间距离不超过阈值，需求对票据按相同属性进行预先排序
	batchDistanceRule *open.MatchmakingRule

	//规则扩展
	ruleExpansionMap map[string]*open.MatchmakingExpansionRule

	//团队扩展
	teamExpansionMap map[string]*open.MatchmakingExpansionRule

	//批次内排序规则
	sortRule []*MatchmakingRule

	//批次内匹配规则
	filterRule []*MatchmakingRule
}

func NewMatchmakingRuleSetWrapper(rs *open.MatchmakingRuleSet) *MatchmakingRuleSetWrapper {
	rsw := &MatchmakingRuleSetWrapper{
		MatchmakingRuleSet: rs,
		ruleMap:            map[string]*MatchmakingRule{},
		playerAttributeMap: map[string]*open.PlayerAttribute{},
		teamMap:            map[string]*open.MatchmakingTeamConfiguration{},
		ruleExpansionMap:   map[string]*open.MatchmakingExpansionRule{},
		teamExpansionMap:   map[string]*open.MatchmakingExpansionRule{},
	}

	rsw.Init()

	return rsw
}

//Init 规则集初始化
func (rsw *MatchmakingRuleSetWrapper) Init() {
	//ruleMap
	for _, rule := range rsw.Rules {
		rsw.ruleMap[rule.Name] = newMatchmakingRule(rule)


		switch rule.Type {
		case open.MatchmakingRuleType_MatchmakingRuleType_AbsoluteSort, open.MatchmakingRuleType_MatchmakingRuleType_DistanceSort:
			rsw.sortRule = append(rsw.sortRule, rsw.ruleMap[rule.Name])
		case open.MatchmakingRuleType_MatchmakingRuleType_Comparison, open.MatchmakingRuleType_MatchmakingRuleType_Distance:
			rsw.filterRule = append(rsw.filterRule, rsw.ruleMap[rule.Name])
		}
	}

	//playerAttributeMap
	for _, attr := range rsw.PlayerAttributes {
		rsw.playerAttributeMap[attr.Name] = attr
	}

	//teamMap
	for _, team := range rsw.Teams {
		rsw.teamMap[team.Name] = team
		rsw.MatchPlayerNum += team.PlayerNumber
	}

	if rsw.Algorithm != nil && rsw.Algorithm.BatchingPreference == "sorted" {
		rsw.preBatchSortAlgorithm = rsw.Algorithm
	}

	//扩展规则
	for _, ep := range rsw.Expansions {
		switch ep.Target.ComponentType {
		case open.ComponentType_ComponentType_Rules:
			rsw.ruleExpansionMap[ep.Target.ComponentName] = ep
		case open.ComponentType_ComponentType_Teams:
			rsw.teamExpansionMap[ep.Target.ComponentName] = ep
		}
	}
}

func (rs *MatchmakingRuleSetWrapper) CheckTeams() (err error) {
	if len(rs.Teams) == 0 {
		err = errors.Trace(fmt.Errorf("RuleSet.Teams cannot be empty"))
		return
	}

	for _, team := range rs.Teams {
		err = rs.CheckTeam(team)
		if err != nil {
			err = errors.Trace(err)
			return
		}
	}

	return
}

//CheckRuleTeam 验证Team定义
func (rs *MatchmakingRuleSetWrapper) CheckTeam(team *open.MatchmakingTeamConfiguration) (err error) {
	if team.Name == "" {
		err = errors.Trace(fmt.Errorf("RuleSet.Teams[%s].Name cannot be empty", team.Name))
		return
	}

	if team.PlayerNumber <= 0 {
		err = errors.Trace(fmt.Errorf("RuleSet.Teams[%s].PlayerNumber must be greater than 1", team.Name))
		return
	}

	return
}

//CheckRules 验证规则列表
func (rs *MatchmakingRuleSetWrapper) CheckRules() (err error) {
	for _, rule := range rs.Rules {
		err = rs.CheckRule(rule)
		if err != nil {
			err = errors.Trace(err)
			return
		}
	}

	return
}

//CheckRule 验证规则
func (rs *MatchmakingRuleSetWrapper) CheckRule(rule *open.MatchmakingRule) (err error) {
	if rule.Name == "" {
		err = errors.Trace(fmt.Errorf("RuleSet.Rules[%s].Name cannot be empty", rule.Name))
		return
	}

	//规则类型
	switch rule.Type {
	case open.MatchmakingRuleType_MatchmakingRuleType_Comparison:
		//比较规则验证
		//参数非空校验
		if rule.ReferenceValue == "" {
			//没有参考值时，operation仅支持=,!=操作
			if (rule.Operation == "=" || rule.Operation == "!=") == false {
				err = errors.Trace(fmt.Errorf("RuleSet.Rules[%s].Operation %s is not allowed when ReferenceValue is nil", rule.Name, rule.Operation))
				return
			}
		} else {
			if (rule.Operation == "=" || rule.Operation == "!=" || rule.Operation == "<" || rule.Operation == "<=" ||
				rule.Operation == ">" || rule.Operation == ">=") == false {
				err = errors.Trace(fmt.Errorf("RuleSet.Rules[%s].Operation %s is not allowed", rule.Name, rule.Operation))
				return
			}
		}

		//玩家属性
		if rule.Measurements == "" {
			err = errors.Trace(fmt.Errorf("RuleSet.Rules[%s].Measurements cannot be nil", rule.Name))
			return
		}

	case open.MatchmakingRuleType_MatchmakingRuleType_Distance:
		//最大距离
		if rule.MaxDistance <= 0 {
			err = fmt.Errorf("RuleSet.Rules[%s].MaxDistance is less than 0", rule.Name)
			return
		}
	case open.MatchmakingRuleType_MatchmakingRuleType_Collection:
		//集合规则（未实现）
		err = errors.Trace(fmt.Errorf("RuleSet.Rules[%s].Type %s is not implemented", rule.Name, rule.Type.String()))
		return
	case open.MatchmakingRuleType_MatchmakingRuleType_BatchDistance:
		//批次距离规则 (玩家属性之间的距离)
		//batchAttribute属性是否存在玩家属性中
		_, ok := rs.playerAttributeMap[rule.BatchAttribute]
		if !ok {
			err = errors.Trace(fmt.Errorf("RuleSet.Rules[%s].BatchAttribute %s is not defined previously in RuleSet.PlayerAttributes", rule.Name, rule.Type.String()))
			return
		}

		//最大距离
		if rule.MaxDistance <= 0 {
			err = errors.Trace(fmt.Errorf("RuleSet.Rules[%s].MaxDistance is less than 0", rule.Name))
			return
		}
	//以下两个均为排序规则 属性值排序,相对于票据批次中第一个票据的属性值差
	case open.MatchmakingRuleType_MatchmakingRuleType_AbsoluteSort, open.MatchmakingRuleType_MatchmakingRuleType_DistanceSort:
		//升序降序排列
		if (rule.SortDirection == open.SortDirectionType_Ascending || rule.SortDirection == open.SortDirectionType_Descending) == false {
			err = errors.Trace(fmt.Errorf("RuleSet.Rules[%s].SortDirection %s is invalid", rule.Name, rule.SortDirection.String()))
			return
		}

		//排序属性
		_, ok := rs.playerAttributeMap[rule.SortAttribute]
		if !ok {
			err = errors.Trace(fmt.Errorf("RuleSet.Rules[%s].SortAttribute %s is not defined previously in RuleSet.PlayerAttributes", rule.Name, rule.Type.String()))
			return
		}
	default:
		err = errors.Trace(fmt.Errorf("RuleSet.Rules[%s].Type %s is not supported", rule.Name, rule.Type.String()))
		return
	}

	return
}

//CheckExpansionRules 验证扩展规则列表
func (rs *MatchmakingRuleSetWrapper) CheckExpansionRules() (err error) {
	//扩展规则
	for _, expansion := range rs.Expansions {
		err = rs.CheckExpansionRule(expansion)
		if err != nil {
			return
		}
	}

	return
}

//CheckExpansionRule 验证扩展规则
func (rs *MatchmakingRuleSetWrapper) CheckExpansionRule(ep *open.MatchmakingExpansionRule) (err error) {
	//验证target
	switch ep.Target.ComponentType {
	case open.ComponentType_ComponentType_Rules:
		//验证规则名是否存在
		rule, ok := rs.ruleMap[ep.Target.ComponentName]
		if !ok {
			err = errors.Trace(fmt.Errorf("ExpansionRule.Target.ComponentName %s not defined previously in RuleSet.Rules", ep.Target.ComponentName))
			return
		}

		//rule 的属性ep.Target.Attribute非空
		ruleRv := reflect.ValueOf(rule).Elem()
		_, fieldOk := ruleRv.Type().FieldByName(ep.Target.Attribute)
		if !fieldOk {
			err =  errors.Trace(fmt.Errorf("ExpansionRule.Target.Attribute %s not defined previously in RuleSet.Rules[%s]",
				ep.Target.Attribute, ep.Target.ComponentName))
			return
		}
	case open.ComponentType_ComponentType_Teams:
		//验证规则名是否存在
		team, ok := rs.teamMap[ep.Target.ComponentName]
		if !ok {
			err = errors.Trace(fmt.Errorf("ExpansionRule.Target.ComponentName %s not defined previously in RuleSet.Rules", ep.Target.ComponentName))
			return
		}

		//rule 的属性ep.Target.Attribute非空
		ruleRv := reflect.ValueOf(team).Elem()
		_, fieldOk := ruleRv.Type().FieldByName(ep.Target.Attribute)
		if !fieldOk {
			err = errors.Trace(fmt.Errorf("ExpansionRule.Target.Attribute %s not defined previously in RuleSet.Teams[%s]",
				ep.Target.Attribute, ep.Target.ComponentName))
			return
		}
	default:
		err = errors.Trace(fmt.Errorf("ExpansionRule.Target.ComponentType %s not supported", ep.Target.ComponentType))
		return
	}

	return
}

func (rsw *MatchmakingRuleSetWrapper) CheckParams() (err error) {
	//验证团队
	err = rsw.CheckTeams()
	if err != nil {
		err = errors.Trace(err)
		return
	}

	//规则验证
	err = rsw.CheckRules()
	if err != nil {
		err = errors.Trace(err)
		return
	}

	err = rsw.CheckExpansionRules()
	if err != nil {
		err = errors.Trace(err)
		return
	}

	return
}

//GetPlayerAttributeType 获取玩家属性类型
func (rsw *MatchmakingRuleSetWrapper) GetPlayerAttributeType(attrName string) string {
	return rsw.playerAttributeMap[attrName].Type
}

//DivideBatch划分批次
func (rsw *MatchmakingRuleSetWrapper) DivideBatch(tickets []*open.MatchmakingTicket) (batches []*[]*open.MatchmakingTicket) {
	rule := rsw.batchDistanceRule
	if rule == nil {
		//未设置批前排序，应该怎么划分批次？ 根据批次处理器的数量，随机划分数量
		return
	}
	//锚点
	anchor := tickets[0]
	anchorAttrValue := getTicketFloatAttributeValue(anchor, rule.BatchAttribute, rule.PartyAggregation)
	batch := []*open.MatchmakingTicket{}
	for i := 1; i < len(tickets); i++ {
		ticketAttrValue := getTicketFloatAttributeValue(tickets[i], rule.BatchAttribute, rule.PartyAggregation)
		if ticketAttrValue-anchorAttrValue < rule.MaxDistance {
			batch = append(batch, tickets[i])
			continue
		}

		batches = append(batches, &batch)

		//重置锚点
		anchor = tickets[i]
		anchorAttrValue = getTicketFloatAttributeValue(anchor, rule.BatchAttribute, rule.PartyAggregation)
		batch = []*open.MatchmakingTicket{}
	}

	if len(batch) > 0 {
		batches = append(batches, &batch)
	}

	return
}

//GetPropertyExpressionReferenceValue 获取规则的属性表达式值的参考值
func (rs *MatchmakingRuleSetWrapper) GetPropertyExpressionReferenceValue(p *parser.PropertyExprParser, teams []*Team) (referenceValue float64) {
	ctxTeams := make([]*open.MatchTeam, 0, len(teams))
	for _, v := range teams {
		ctxTeams = append(ctxTeams, (*open.MatchTeam)(v))
	}

	ctx := context.WithValue(context.TODO(), chain.CtxTeamsKey, ctxTeams)
	p.Do(ctx, func(ctx context.Context) {
		v := ctx.Value(chain.CtxReturnKey)
		switch args := v.(type) {
		case float64:
			referenceValue = args
		default:
			log.Panicf("args should be float64, but get %v\n", v)
		}
	})

	return
}

//测量值
type MeasurementResult struct {
	Aggregation string //聚合函数 min/max/avg时等于团队数量, 空时仅执行flatten, 等于所有选择的团队的玩家的数量
	Values      []*MeasurementValue
}

type MeasurementValue struct {
	team   *Team
	ticket *open.MatchmakingTicket
	player *open.MatchPlayer
	value  float64
}

//GetPropertyExpressionMeasurements 获取规则的属性表达式值的测量值
func (rs *MatchmakingRuleSetWrapper) GetPropertyExpressionMeasurements(p *parser.PropertyExprParser, teams []*Team) (measurements []float64) {
	ctxTeams := make([]*open.MatchTeam, 0, len(teams))
	for _, v := range teams {
		ctxTeams = append(ctxTeams, (*open.MatchTeam)(v))
	}

	ctx := context.WithValue(context.TODO(), chain.CtxTeamsKey, ctxTeams)
	p.Do(ctx, func(ctx context.Context) {
		v := ctx.Value(chain.CtxReturnKey)
		switch args := v.(type) {
		case []float64:
			measurements = args
		default:
			log.Panicf("args should be []float64, but get %v\n", v)
		}
	})

	return
}

//MaxDistance 根据扩展规则计算最大距离
func (rsw *MatchmakingRuleSetWrapper) MaxDistance(ruleName string, maxDistance float64, tryTimes int, creationTime int64) (newMaxDistance float64, ok bool) {
	//查看是否配置了扩展规则
	var ep *open.MatchmakingExpansionRule
	ep, ok = rsw.ruleExpansionMap[ruleName]
	if !ok {
		return
	}

	//是否对最大距离进行设置
	if strings.ToLower(ep.Target.Attribute) != "maxdistance" {
		ok = false
		return
	}

	if ep.FixedExpansionDistance > 0 {
		//根据次数计算最大距离
		newMaxDistance = maxDistance + ep.FixedExpansionDistance*float64(tryTimes)
	} else {
		//根据时间计算
		intervalTime := time.Now().Unix() - creationTime
		var selectStep *open.MatchmakingExpansionRuleStep
		for _, step := range ep.Steps {
			if intervalTime >= step.WaitTimeSeconds {
				selectStep = step
				continue
			}

			break
		}

		if selectStep != nil {
			newMaxDistance = selectStep.Value
			return
		}
	}

	return
}