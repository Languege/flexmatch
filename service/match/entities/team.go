// entities
// @author LanguageY++2013 2022/11/9 18:30
// @company soulgame
package entities

import (
	"github.com/Languege/flexmatch/service/match/proto/open"
)
//
//type Team struct {
//	//团队名
//	Name string
//
//	//期望玩家数
//	ExpectPlayerNum int
//
//	//票据
//	Tickets []*open.MatchmakingTicket
//}
type Team open.MatchTeam

func newTeam(conf *open.MatchmakingTeamConfiguration) *Team {
	return &Team{
		Conf: conf,
	}
}
func (t *Team) Copy() *Team {
	cp := &Team{
		Conf:            t.Conf,
		Tickets:         make([]*open.MatchmakingTicket, 0, len(t.Tickets)),
	}

	for _, v := range t.Tickets {
		cp.Tickets = append(cp.Tickets, v)
	}

	return cp
}
