// pubsub
// @author LanguageY++2013 2023/5/11 09:26
// @company soulgame
package pubsub

import (
	"github.com/Languege/flexmatch/service/match/proto/open"
)

type Publisher interface {
	Name() string
	Send(topic string, ev *open.MatchEvent) error
}


type MultiPublisher []Publisher

func NewMultiPublisher(multi... Publisher) MultiPublisher {
	return multi
}

func(p MultiPublisher) Name() string {
	return "multi"
}

func(p MultiPublisher) Send(topic string, ev *open.MatchEvent) (errSend error){
	for _, sub := range p {
		err := sub.Send(topic, ev)
		if err != nil  && errSend == nil {
			//保留首次错误
			errSend = err
		}
	}

	return
}