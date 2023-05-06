// parser
// @author LanguageY++2013 2023/4/3 22:19
// @company soulgame
package parser

import (
	"github.com/Languege/flexmatch/service/match/parser/chain"
	"context"
	"log"
	"github.com/Languege/flexmatch/service/match/proto/open"
	"strconv"
)

func AvgHandlerWrapper(ctx context.Context, handler chain.UserHandler) {
	//从上下文获取 参数
	v := ctx.Value(chain.CtxReturnKey)
	if v == nil {
		panic("avg call need args, no args in ctx")
	}
	switch args := v.(type) {
	case []float64:
		ret := AvgFloat64(args)
		ctx = context.WithValue(ctx, chain.CtxReturnKey, ret)
	case []*[]float64:
		ret := make([]float64, 0, len(args))
		for _, arr := range args {
			ret = append(ret, AvgFloat64(*arr))
		}

		ctx = context.WithValue(ctx, chain.CtxReturnKey, ret)
	default:
		log.Panicf("args %v not support", v)
	}

	handler(ctx)
}

//MinHandlerWrapper 最小值处理
func MinHandlerWrapper(ctx context.Context, handler chain.UserHandler) {
	//从上下文获取 参数
	v := ctx.Value(chain.CtxReturnKey)
	if v == nil {
		panic("max call need args, no args in ctx")
	}
	switch args := v.(type) {
	case []float64:
		ret := MinFloat64(args)
		ctx = context.WithValue(ctx, chain.CtxReturnKey, ret)
	case []*[]float64:
		ret := make([]float64, 0, len(args))
		for _, arr := range args {
			ret = append(ret, MinFloat64(*arr))
		}

		ctx = context.WithValue(ctx, chain.CtxReturnKey, ret)
	default:
		log.Panicf("min %v not support", v)
	}

	handler(ctx)
}

//MaxHandlerWrapper 最大值
func MaxHandlerWrapper(ctx context.Context, handler chain.UserHandler) {
	//从上下文获取 参数
	v := ctx.Value(chain.CtxReturnKey)
	if v == nil {
		panic("max call need args, no args in ctx")
	}
	switch args := v.(type) {
	case []float64:
		ret := MaxFloat64(args)
		ctx = context.WithValue(ctx, chain.CtxReturnKey, ret)
	case []*[]float64:
		ret := make([]float64, 0, len(args))
		for _, arr := range args {
			ret = append(ret, MaxFloat64(*arr))
		}

		ctx = context.WithValue(ctx, chain.CtxReturnKey, ret)
	default:
		log.Panicf("max %v not support", v)
	}

	handler(ctx)
}


//CountHandlerWrapper 计数
func CountHandlerWrapper(ctx context.Context, handler chain.UserHandler) {
	//从上下文获取 参数
	v := ctx.Value(chain.CtxReturnKey)
	if v == nil {
		panic("count call need args, no args in ctx")
	}
	switch args := v.(type) {
	case []float64:
		ret := len(args)
		ctx = context.WithValue(ctx, chain.CtxReturnKey, ret)
	case []*[]float64:
		ret := make([]float64, 0, len(args))
		for _, arr := range args {
			ret = append(ret, float64(len(*arr)))
		}

		ctx = context.WithValue(ctx, chain.CtxReturnKey, ret)
	default:
		log.Panicf("max %v not support", v)
	}

	handler(ctx)
}

//FlattenHandlerWrapper 多维数组扁平化（降维）
func FlattenHandlerWrapper(ctx context.Context, handler chain.UserHandler) {
	//从上下文获取 参数
	v := ctx.Value(chain.CtxReturnKey)
	if v == nil {
		panic("flatten call need args, no args in ctx")
	}
	switch args := v.(type) {
	case []*[]float64:
		ret := make([]float64, 0, len(args))
		for _, arr := range args {
			ret = append(ret, *arr...)
		}

		ctx = context.WithValue(ctx, chain.CtxReturnKey, ret)
	default:
		log.Panicf("flatten %v not support", v)
	}

	handler(ctx)
}


//FilterRuleClosure 规则过滤闭包
func FilterRuleClosure(rules []*open.MatchmakingRule, ruleNames []string)(out []*open.MatchmakingRule) {
	ruleNameMap := make(map[string]bool, len(ruleNames))
	for _, v := range ruleNames {
		ruleNameMap[v] = true
	}

	out = make([]*open.MatchmakingRule, 0, len(ruleNames))
	for _, rule := range rules {
		if _, ok := ruleNameMap[rule.Name];ok {
			out = append(out, rule)
		}
	}

	return
}

//FilterTeamClosure 团队过滤闭包
func FilterTeamClosure(teams []*open.MatchTeam, names []string)(out []*open.MatchTeam) {
	nameMap := make(map[string]bool, len(names))
	for _, v := range names {
		nameMap[v] = true
	}

	out = make([]*open.MatchTeam, 0, len(names))
	for _, team := range teams {
		if _, ok := nameMap[team.Conf.Name];ok {
			out = append(out, team)
		}
	}

	return
}

//FilterPlayerAttribute 过滤玩家属性
func FilterPlayerAttribute(p *open.MatchPlayer, attrName string) float64 {
	for _, v := range p.Attributes {
		if v.Name == attrName {
			data, _ := strconv.ParseFloat(v.Value, 10)
			return data
		}
	}

	return 0
}

//TeamConfValue 读取team配置值
func TeamConfValue(team *open.MatchTeam, fieldName string) (ret float64) {
	switch fieldName {
	case "PlayerNumber":
		return float64(team.Conf.PlayerNumber)
	default:
		return 0
	}
	//fv := reflect.ValueOf(team.Conf).Elem().FieldByName(fieldName)
	//switch fv.Kind() {
	//case reflect.Int32,reflect.Int64:
	//	ret = float64(fv.Int())
	//case reflect.Float32,reflect.Float64:
	//	ret = fv.Float()
	//default:
	//	log.Panicf("Team.Conf.%s type %s not supported", fieldName, fv.Kind())
	//}
	//return
}

//RuleConfValue 读取rule配置值
func RuleConfValue(rule *open.MatchmakingRule, fieldName string) (ret float64) {
	switch fieldName {
	case "MaxDistance":
		return rule.MaxDistance
	default:
		return 0
	}
	//fv := reflect.ValueOf(rule).Elem().FieldByName(fieldName)
	//switch fv.Kind() {
	//case reflect.Int32,reflect.Int64:
	//	ret = float64(fv.Int())
	//case reflect.Float32,reflect.Float64:
	//	ret = fv.Float()
	//default:
	//	log.Panicf("Rule.%s type %s not supported", fieldName, fv.Kind())
	//}
	//return
}