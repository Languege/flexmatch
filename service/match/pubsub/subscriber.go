// match_api
// @author LanguageY++2013 2023/5/10 23:20
// @company soulgame
package pubsub

import (
	"github.com/Languege/flexmatch/service/match/proto/open"
)

type Subscriber interface {
	Name() string
	Receive(ev *open.MatchEvent) error
}
