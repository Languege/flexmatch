// repositories
// @author LanguageY++2013 2022/11/1 10:19
// @company soulgame
package elo_rep

import (
	_ "github.com/Languege/flexmatch/service/match/proto/open"
	"github.com/Languege/flexmatch/service/match/proto/open"
)

const(
	//最多保留对局记录
	maxBattleRecordLength = 10
)

type BattleRecordQueue struct {
	open.PlayerBattleRecordList
}

func(q *BattleRecordQueue) Add(record *open.PlayerBattleResult) {
	if len(q.RecordList) >= maxBattleRecordLength {
		//移除队列头部元素
		copy(q.RecordList, q.RecordList[1:])
		q.RecordList[maxBattleRecordLength - 1] = record
	}else{
		q.RecordList = append(q.RecordList, record)
	}
}

func(q *BattleRecordQueue) G() (g float64) {
	for _, v := range q.RecordList {
		g += v.GainScore
	}

	return
}

func(q *BattleRecordQueue) Ge() (ge float64) {
	for _, v := range q.RecordList {
		ge += v.GainExpectScore
	}

	return
}

func(q *BattleRecordQueue) D() (diff float64) {
	return q.G() - q.Ge()
}

//玩家对局记录
type PlayerBattleRecordManager struct {
	store 		map[int64]*BattleRecordQueue
}

func NewPlayerBattleRecordManager() *PlayerBattleRecordManager {
	return &PlayerBattleRecordManager{store: map[int64]*BattleRecordQueue{}}
}
func(m *PlayerBattleRecordManager) AddPlayerRecord(record *open.PlayerBattleResult) {
	queue, ok := m.store[record.UserId]
	if !ok {
		queue = &BattleRecordQueue{}
		m.store[record.UserId] = queue
	}

	queue.Add(record)
}

func(m *PlayerBattleRecordManager) Queue(userId int64) (queue *BattleRecordQueue, ok bool) {
	queue, ok = m.store[userId]
	return
}