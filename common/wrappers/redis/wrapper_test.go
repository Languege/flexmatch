// redis
// @author LanguageY++2013 2023/5/16 16:27
// @company soulgame
package redis

import (
	"testing"
	"github.com/Languege/flexmatch/service/match/proto/open"
	"github.com/google/uuid"
)

type TestStruct struct {
	StreamBaseMsg
	Test string `redis:"test"`
}

type StreamObject struct {
	StreamBaseMsg
	open.MatchEvent
	MatchEventType int `redis:"MatchEventType"`
}

func TestRedisWrapper_XAdd(t *testing.T) {
	wrapper := NewRedisWrapper(Configure{
		ConnectMode: ConnectModeDirect,
		Host:        "10.10.10.16",
		Port:        6389,
		MaxIdle:     10,
		MaxActive:   10,
	})

	xqueue := "xqueue"
	in := &TestStruct{
		Test: "hello",
	}
	err := wrapper.XAdd(xqueue, in)
	if err != nil {
		t.Fatal(err)
	}

	t.Run("args", func(t *testing.T) {
		entry := &StreamObject{
			MatchEvent:open.MatchEvent{
				MatchEventType: open.MatchEventType_AcceptMatchCompleted,
				Tickets: []*open.MatchmakingTicket{
					{TicketId: uuid.NewString()},
				},
			},
			MatchEventType: int(open.MatchEventType_AcceptMatchCompleted),
		}
		err := wrapper.XAdd(xqueue, entry)
		if err != nil {
			t.Fatal(err)
		}
	})
}