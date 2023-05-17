// pubsub
// @author LanguageY++2013 2023/5/11 09:26
// @company soulgame
package pubsub

import (
	"github.com/Languege/flexmatch/service/match/proto/open"
	"github.com/Languege/flexmatch/common/logger"
)

type Publisher interface {
	Name() string
	Send(topic string, ev *open.MatchEvent) error
}


type MultiPublisher []Publisher

func NewMultiPublisher(multi... Publisher) MultiPublisher {
	return multi
}

func(mp *MultiPublisher) Add(p Publisher) {
	for _, v := range *mp {
		if p.Name() == v.Name() {
			logger.Panicf("publisher %s has been registered previously", p.Name())
		}
	}

	*mp = append(*mp, p)
}

func(mp MultiPublisher) Name() string {
	return "multi"
}

func(mp MultiPublisher) Send(topic string, ev *open.MatchEvent) (errSend error){
	for _, sub := range mp {
		err := sub.Send(topic, ev)
		if err != nil  && errSend == nil {
			//保留首次错误
			errSend = err
		}
	}

	return
}