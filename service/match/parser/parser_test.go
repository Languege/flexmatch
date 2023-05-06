// parser
// @author LanguageY++2013 2023/3/31 10:55
// @company soulgame
package parser

import (
	"testing"
	"github.com/davyxu/golexer"
	"context"
	"github.com/Languege/flexmatch/service/match/proto/open"
	"github.com/google/uuid"
	"github.com/Languege/flexmatch/service/match/parser/chain"
	"reflect"
)

func TestParser(t *testing.T) {
	//TODO: 最终返回一个处理过程的切片，调用过程类似m1(m2(m3()))
	p := newPropertyExprParser()


	p.Lexer().Start(`avg(flatten(teams[*].players.attributes[score]))`)
	defer golexer.ErrorCatcher(func(err error) {
		t.Log(err)
	})

	p.NextToken()
	p.Parse()

	p.BuildChain()
	teams := []*open.MatchTeam{
		{
			Conf:&open.MatchmakingTeamConfiguration{
				Name: "test blue team",
				PlayerNumber: 1,
			},
			Tickets: []*open.MatchmakingTicket{
				{
					TicketId: uuid.NewString(),
					Players: []*open.MatchPlayer{
						{
							UserId: 1,
							Attributes: []*open.PlayerAttribute{
								{
									Name: "score",
									Type: "float64",
									Value: "5",
								},
							},
						},
						{
							UserId: 2,
							Attributes: []*open.PlayerAttribute{
								{
									Name: "score",
									Type: "float64",
									Value: "5.5",
								},
							},
						},
					},
				},
			},
		},
	}
	ctx := context.WithValue(context.TODO(), chain.CtxTeamsKey, teams)
	p.Do(ctx, func(ctx context.Context) {
		v := ctx.Value(chain.CtxReturnKey)
		switch args := v.(type) {
		case []float64:
			if args[0] != 5.25 {
				t.Fatalf("args should be %0.2f", 5.25)
			}
		case float64:
			if args != 5.25 {
				t.Fatalf("args should be %0.2f", 5.25)
			}
		default:
			t.Fatalf("shouldn't type %v", reflect.TypeOf(args))
		}
	})
}


func TestTeamConf(t *testing.T) {
	//最终返回一个处理过程的切片，调用过程类似m1(m2(m3()))
	p := newPropertyExprParser()


	p.Lexer().Start(`teams[Blue].PlayerNumber`)
	defer golexer.ErrorCatcher(func(err error) {
		t.Log(err)
	})

	p.NextToken()
	p.Parse()

	p.BuildChain()
	teams := []*open.MatchTeam{
		{
			Conf:&open.MatchmakingTeamConfiguration{
				Name: "Blue",
				PlayerNumber: 1,
			},
			Tickets: []*open.MatchmakingTicket{
				{
					TicketId: uuid.NewString(),
					Players: []*open.MatchPlayer{
						{
							UserId: 1,
							Attributes: []*open.PlayerAttribute{
								{
									Name: "score",
									Type: "float64",
									Value: "5",
								},
							},
						},
						{
							UserId: 2,
							Attributes: []*open.PlayerAttribute{
								{
									Name: "score",
									Type: "float64",
									Value: "5.5",
								},
							},
						},
					},
				},
			},
		},
	}
	ctx := context.WithValue(context.TODO(), chain.CtxTeamsKey, teams)
	p.Do(ctx, func(ctx context.Context) {
		v := ctx.Value(chain.CtxReturnKey)
		switch args := v.(type) {
		case float64:
			if args != 1 {
				t.Fatalf("args should be %d", 1)
			}
		default:
			t.Fatalf("shouldn't type %v", reflect.TypeOf(args))
		}
	})
}

func TestRuleConf(t *testing.T) {
	//最终返回一个处理过程的切片，调用过程类似m1(m2(m3()))
	p := newPropertyExprParser()


	p.Lexer().Start(`rules[FairTeamSkill].MaxDistance`)
	defer golexer.ErrorCatcher(func(err error) {
		t.Log(err)
		t.Log(p.TokenPos().String())
	})

	p.NextToken()
	p.Parse()

	p.BuildChain()

	maxDistance := 100.0
	rules := []*open.MatchmakingRule{
		{
			Name: "FairTeamSkill",
			MaxDistance: maxDistance,
		},
	}

	ctx := context.WithValue(context.TODO(), chain.CtxRulesKey, rules)
	p.Do(ctx, func(ctx context.Context) {
		v := ctx.Value(chain.CtxReturnKey)
		switch args := v.(type) {
		case float64:
			if args != maxDistance {
				t.Fatalf("args should be %v", maxDistance)
			}
		default:
			t.Fatalf("shouldn't type %v", reflect.TypeOf(args))
		}
	})
}