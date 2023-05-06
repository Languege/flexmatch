// parser
// @author LanguageY++2013 2023/3/31 09:39
// @company soulgame
package parser

import (
	"context"
	"github.com/Languege/flexmatch/service/match/proto/open"
	"github.com/Languege/flexmatch/service/match/parser/chain"
	"github.com/davyxu/golexer"
	"log"
	"reflect"
	"strings"
	"strconv"
)

//属性表达式解析器

// 自定义token id
const (
	Token_EOF = iota
	Token_Unknown
	Token_Numeral //数字
	Token_Identifier
	Token_ParenL   // (
	Token_ParenR   // )
	Token_BracketL // [
	Token_BracketR // ]
	Token_Wildcard // * 通配符
	Token_Dot      // .
	Token_Comma    // ,
	Token_XMax     // max 函数
	Token_XMin     // min 函数
	Token_XFlatten // flatten 函数
	Token_XAvg     //  avg 平均
	//Token_XAnd		  // and 函数
	//Token_XOr		  // or 函数
	Token_XCount      // count 函数
	Token_ERules      // rules 实体
	Token_ETeams      // teams 实体
	Token_EPlayers    // players 实体
	Token_EAttributes // attributes 实体
)

type PropertyExprParser struct {
	*golexer.Parser

	//函数调用链
	funcChain *chain.Chain

	//实体调用链
	entityChain *chain.Chain

	doWrapper chain.UserHandlerWrapper
}

func (p *PropertyExprParser) Do(ctx context.Context, userHandler chain.UserHandler) {
	p.doWrapper(ctx, userHandler)
}

//构建调用链
func (p *PropertyExprParser) BuildChain() {
	if p.entityChain != nil {
		p.funcChain.Add(p.entityChain.BuildQueue())
	}

	p.doWrapper = p.funcChain.BuildStack()
}

//IsEntity Kind是否为实体
func (p *PropertyExprParser) IsEntity() bool {
	id := p.TokenID()
	for _, v := range []int{Token_ERules, Token_ETeams, Token_EPlayers, Token_EAttributes} {
		if v == id {
			return true
		}
	}

	return false
}

//IsFunc Kind是否为函数
func (p *PropertyExprParser) IsFunc() bool {
	id := p.TokenID()
	for _, v := range []int{Token_XMax, Token_XMin, Token_XFlatten, Token_XCount, Token_XAvg} {
		if v == id {
			return true
		}
	}

	return false
}

//parseFunc 解析函数
func (p *PropertyExprParser) Parse() {

	switch {
	case p.IsEntity():
		p.parseEntity()
	case p.IsFunc():
		switch p.TokenID() {
		case Token_XAvg:
			p.funcChain.Add(AvgHandlerWrapper)
		case Token_XMin:
			p.funcChain.Add(MinHandlerWrapper)
		case Token_XMax:
			p.funcChain.Add(MaxHandlerWrapper)
		case Token_XCount:
			p.funcChain.Add(CountHandlerWrapper)
		case Token_XFlatten:
			p.funcChain.Add(FlattenHandlerWrapper)
		default:
			log.Panicf("func %s not supported", p.TokenValue())
		}
		p.NextToken()
		p.Expect(Token_ParenL)
		p.Parse()
		p.Expect(Token_ParenR)
	default:
		switch p.TokenID() {
		case Token_Numeral:
			numberValue , _:= strconv.ParseFloat(p.TokenValue(), 10)
			p.funcChain.Add(func(ctx context.Context, handler chain.UserHandler) {
				ctx = context.WithValue(ctx, chain.CtxReturnKey, numberValue)
				handler(ctx)
			})
		default:
			log.Panicf("first token %s not supported", p.TokenValue())
		}
	}
}

//parseEntity 解析实体
func (p *PropertyExprParser) parseEntity() {
	switch p.TokenID() {
	case Token_ERules:
		p.parseRule()
	case Token_ETeams:
		p.parseTeam()
	case Token_EPlayers:
		p.parsePlayer()
	case Token_EAttributes:
		p.parseAttribute()
	}
}

//parseRule 解析规则 实体 先解析的先执行（queue）  而函数先解析后执行 (stack)
func (p *PropertyExprParser) parseRule() {
	p.Expect(Token_ERules)
	p.Expect(Token_BracketL)
	switch p.TokenID() {
	case Token_Wildcard:
		//所有规则
		p.entityChain.Add(func(ctx context.Context, handler chain.UserHandler) {
			v := ctx.Value(chain.CtxRulesKey)
			if v == nil {
				log.Panicln("rules expected in ctx")
			}
			rules, ok := v.([]*open.MatchmakingRule)
			if !ok {
				log.Panicf("rules should be *open.MatchmakingRule type not %s\n", reflect.TypeOf(v))
			}

			ctx = context.WithValue(ctx, chain.CtxReturnKey, rules)

			handler(ctx)
		})
	case Token_Identifier:
		//过滤规则
		ruleNames := strings.Split(p.TokenValue(), ",")

		p.entityChain.Add(func(ctx context.Context, handler chain.UserHandler) {
			v := ctx.Value(chain.CtxRulesKey)
			if v == nil {
				log.Panicln("rules expected in ctx")
			}
			rules, ok := v.([]*open.MatchmakingRule)
			if !ok {
				log.Panicf("rules should be *open.MatchmakingRule type not %s\n", reflect.TypeOf(v))
			}
			out := FilterRuleClosure(rules, ruleNames)
			if len(out) == 1 {
				ctx = context.WithValue(ctx, chain.CtxReturnKey, out[0])
			} else {
				ctx = context.WithValue(ctx, chain.CtxReturnKey, out)
			}

			handler(ctx)
		})
	default:
		log.Panicln("expect wildcard or identifier")
	}
	p.NextToken()
	p.Expect(Token_BracketR)

	if p.TokenID() == Token_Dot {
		//子属性
		p.NextToken()
		fieldName := p.TokenValue()
		rt := reflect.TypeOf((*open.MatchmakingRule)(nil)).Elem()
		if _, ok := rt.FieldByName(fieldName);ok {
			p.entityChain.Add(func(ctx context.Context, handler chain.UserHandler) {
				v := ctx.Value(chain.CtxReturnKey)
				switch args := v.(type) {
				case *open.MatchmakingRule:
					ret :=  RuleConfValue(args, fieldName)
					ctx = context.WithValue(ctx, chain.CtxReturnKey, ret)
				default:
					log.Panicf("expect *open.MatchmakingRule")
				}

				handler(ctx)
			})
		}else{
			p.parseEntity()
		}
	}
}

//parseTeam 解析团队 实体 先解析的先执行（queue）  而函数先解析后执行 (stack)
func (p *PropertyExprParser) parseTeam() {
	p.Expect(Token_ETeams)
	p.Expect(Token_BracketL)
	switch p.TokenID() {
	case Token_Wildcard:
		//所有规则
		p.entityChain.Add(func(ctx context.Context, handler chain.UserHandler) {
			v := ctx.Value(chain.CtxTeamsKey)
			if v == nil {
				log.Panicln("teams expected in ctx")
			}
			teams, ok := v.([]*open.MatchTeam)
			if !ok {
				log.Panicf("teams should be []*open.MatchTeam type not %s\n", reflect.TypeOf(v))
			}

			ctx = context.WithValue(ctx, chain.CtxReturnKey, teams)

			handler(ctx)
		})
	case Token_Identifier:
		//过滤规则
		ruleNames := strings.Split(p.TokenValue(), ",")

		p.entityChain.Add(func(ctx context.Context, handler chain.UserHandler) {
			v := ctx.Value(chain.CtxTeamsKey)
			if v == nil {
				log.Panicln("teams expected in ctx")
			}
			teams, ok := v.([]*open.MatchTeam)
			if !ok {
				log.Panicf("teams should be []*open.MatchTeam type not %s\n", reflect.TypeOf(v))
			}
			out := FilterTeamClosure(teams, ruleNames)
			if len(out) == 1 {
				ctx = context.WithValue(ctx, chain.CtxReturnKey, out[0])
			} else {
				ctx = context.WithValue(ctx, chain.CtxReturnKey, out)
			}

			handler(ctx)
		})
	default:
		log.Panicf("expect token wildcard or identifier")
	}
	p.NextToken()
	p.Expect(Token_BracketR)

	if p.TokenID() == Token_Dot {
		//子属性
		p.NextToken()
		//是否为配置属性
		fieldName := p.TokenValue()
		rt := reflect.TypeOf((*open.MatchmakingTeamConfiguration)(nil)).Elem()
		if _, ok := rt.FieldByName(fieldName);ok {
			p.entityChain.Add(func(ctx context.Context, handler chain.UserHandler) {
				v := ctx.Value(chain.CtxReturnKey)
				switch args := v.(type) {
				case []*open.MatchTeam:
					rets := make([]float64, 0, len(args))
					for _, arg := range args {
						rets = append(rets, TeamConfValue(arg, fieldName))
					}

					ctx = context.WithValue(ctx, chain.CtxReturnKey, rets)
				case *open.MatchTeam:
					ret :=  TeamConfValue(args, fieldName)
					ctx = context.WithValue(ctx, chain.CtxReturnKey, ret)
				default:
					log.Panicf("expect *open.MatchTeam or []*open.MatchTeam")
				}

				handler(ctx)
			})
		}else{
			p.parseEntity()
		}
	}
}

//parsePlayer 解析玩家 实体 先解析的先执行（queue）  而函数先解析后执行 (stack)
func (p *PropertyExprParser) parsePlayer() {
	p.Expect(Token_EPlayers)

	p.entityChain.Add(func(ctx context.Context, handler chain.UserHandler) {
		v := ctx.Value(chain.CtxReturnKey)
		switch args := v.(type) {
		case *open.MatchTeam:
			ret := make([]*open.MatchPlayer, 0, len(args.Tickets))
			for _, ticket := range args.Tickets {
				ret = append(ret, ticket.Players...)
			}
			ctx = context.WithValue(ctx, chain.CtxReturnKey, ret)
		case []*open.MatchTeam:
			rets := make([]*[]*open.MatchPlayer, 0, len(args))
			for _, team := range args {
				ret := make([]*open.MatchPlayer, 0, len((*team).Tickets))
				for _, ticket := range (*team).Tickets {
					ret = append(ret, ticket.Players...)
				}
				rets = append(rets, &ret)
			}

			ctx = context.WithValue(ctx, chain.CtxReturnKey, rets)
		default:
			log.Panicf("type %s not supported in parsing player\n", reflect.TypeOf(v))
		}

		handler(ctx)
	})

	if p.TokenID() == Token_Dot {
		//子属性
		p.NextToken()
		p.parseEntity()
	}
}

//parseAttribute 解析属性 实体 先解析的先执行（queue）  而函数先解析后执行 (stack)
func (p *PropertyExprParser) parseAttribute() {
	p.Expect(Token_EAttributes)

	p.Expect(Token_BracketL)
	//必须是具体某个属性
	attrName := p.Expect(Token_Identifier).Value()

	p.entityChain.Add(func(ctx context.Context, handler chain.UserHandler) {
		v := ctx.Value(chain.CtxReturnKey)
		if v == nil {
			log.Panicln("the values returned by previous call are expected in ctx")
		}
		switch args := v.(type) {
		case []*open.MatchPlayer:
			//一个玩家数组 （一个团队）
			ret := make([]float64, 0, len(args))
			for _, player := range args {
				attrValue := FilterPlayerAttribute(player, attrName)
				ret = append(ret, attrValue)
			}
			ctx = context.WithValue(ctx, chain.CtxReturnKey, ret)
		case []*[]*open.MatchPlayer:
			//多个玩家数组（多个团队）
			rets := make([]*[]float64, 0, len(args))
			for _, arr := range args {
				ret := make([]float64, 0, len(*arr))
				for _, player := range *arr {
					attrValue := FilterPlayerAttribute(player, attrName)
					ret = append(ret, attrValue)
				}
				rets = append(rets, &ret)
			}
			ctx = context.WithValue(ctx, chain.CtxReturnKey, rets)
		}

		handler(ctx)
	})

	p.Expect(Token_BracketR)
}

func newPropertyExprParser() *PropertyExprParser {

	l := golexer.NewLexer()

	// 匹配顺序从高到低
	l.AddMatcher(golexer.NewKeywordMatcher(Token_XFlatten, "flatten"))
	l.AddMatcher(golexer.NewKeywordMatcher(Token_XAvg, "avg"))
	l.AddMatcher(golexer.NewKeywordMatcher(Token_XMax, "max"))
	l.AddMatcher(golexer.NewKeywordMatcher(Token_XMin, "min"))
	l.AddMatcher(golexer.NewKeywordMatcher(Token_XCount, "count"))
	l.AddMatcher(golexer.NewSignMatcher(Token_ParenL, "("))
	l.AddMatcher(golexer.NewSignMatcher(Token_ParenR, ")"))
	l.AddMatcher(golexer.NewSignMatcher(Token_Comma, ","))

	l.AddMatcher(golexer.NewKeywordMatcher(Token_ERules, "rules"))
	l.AddMatcher(golexer.NewKeywordMatcher(Token_ETeams, "teams"))
	l.AddMatcher(golexer.NewKeywordMatcher(Token_EPlayers, "players"))
	l.AddMatcher(golexer.NewKeywordMatcher(Token_EAttributes, "attributes"))
	l.AddMatcher(golexer.NewSignMatcher(Token_BracketL, "["))
	l.AddMatcher(golexer.NewSignMatcher(Token_BracketR, "]"))
	l.AddMatcher(golexer.NewSignMatcher(Token_Wildcard, "*"))
	l.AddMatcher(golexer.NewSignMatcher(Token_Dot, "."))

	l.AddMatcher(golexer.NewPositiveNumeralMatcher(Token_Numeral))
	l.AddMatcher(golexer.NewIdentifierMatcher(Token_Identifier))

	l.AddMatcher(golexer.NewUnknownMatcher(Token_Unknown))

	return &PropertyExprParser{
		Parser:      golexer.NewParser(l, "expr"),
		funcChain:   &chain.Chain{},
		entityChain: &chain.Chain{},
	}
}

//NewPropertyExprParser 建立属性表达式
func NewPropertyExprParser(expr string) *PropertyExprParser {
	p := newPropertyExprParser()


	p.Lexer().Start(expr)
	defer golexer.ErrorCatcher(func(err error) {
		log.Println(err)
	})

	p.NextToken()
	p.Parse()

	p.BuildChain()

	return p
}
