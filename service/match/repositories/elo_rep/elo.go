//// repositories
//// @author LanguageY++2013 2022/10/31 16:33
//// @company soulgame
package elo_rep
//
//import (
//	"github.com/Languege/flexmatch/service/match/proto/open"
//	"math"
//	"log"
//)
//
////https://zhuanlan.zhihu.com/p/28190267
//
//const(
//	InitialRating = 1200
//
//	WinScore = 1.0
//
//	LossScore = 0.5
//
//	RedCamp = "red"
//	BlueCamp = "blue"
//)
//
////已知分差求胜率
//func Rating(d float64) float64 {
//	return 1.0 / (1 + math.Pow(10, d/400))
//}
//
////已知胜率求分差
//func Diff(rating float64) float64 {
//	return math.Log10(1.0/rating - 1) * 400
//}
//
////ELO匹配机制实现
//
//type Elo struct {
//	store 	*PlayerBattleRecordManager
//	playerRating map[int64]float64 //保存玩家当前战力
//}
//
//func NewElo() *Elo {
//	return &Elo{playerRating: map[int64]float64{},store: NewPlayerBattleRecordManager()}
//}
//
////计算当前得分
////rn 是玩家比赛结束后的新的排位分值
////ro 是比赛前玩家的排位分
////K是一个加成系数，由玩家当前分值水平决定（分值越高K越小）
////G是玩家实际对局得分（赢得1分，输得0.5分）
////Ge是原排位分基础上玩家的预期得分（根据胜率来算，多名对手情况就是和多名对手对战的胜率求和）
//func (e *Elo) RateNow(ro float64, g, ge float64) (rn float64) {
//	rn = ro + e.KFactor(ro) * (g - ge)
//	log.Printf("战力更新: %0.2f -> %0.2f\n", ro, rn)
//	return
//}
//
////K是一个加成系数，由玩家当前分值水平决定（分值越高K越小）
//// 等级分 < 2000, K=30
//// 等级分 2000-2400， K=130-R/2
//// 等级分 > 2400, K=10
//func(e *Elo) KFactor(score float64) float64 {
//	if score < 2000 {
//		return 30
//	}
//
//	if score <= 2400 {
//		return 130 - score/2
//	}
//
//	return 10
//}
//
//
//func(e *Elo) HandleBattleRecord(battleResult *open.BattleResult) {
//	//获胜阵营
//	winCampPlayerList := map[int64]*open.PlayerBattleInfo{}
//	lossCampPlayerList := map[int64]*open.PlayerBattleInfo{}
//	playerBattleResult := map[int64]*open.PlayerBattleResult{}
//	var winCampTotalRating, lossCampTotalRating float64
//
//	for _, player := range battleResult.PlayerList {
//		if player.Rating == 0 {
//			player.Rating = InitialRating
//		}
//
//		if player.Camp == battleResult.WinCamp {
//			winCampPlayerList[player.UserId] = player
//			winCampTotalRating += player.Rating
//		}else{
//			lossCampPlayerList[player.UserId] = player
//			lossCampTotalRating += player.Rating
//		}
//		playerBattleResult[player.UserId] = &open.PlayerBattleResult{
//			UserId: player.UserId,
//		}
//	}
//
//	//win阵营玩家 该场对局预期胜率
//	for _, player := range  winCampPlayerList {
//		result := playerBattleResult[player.UserId]
//		//计算分差
//		result.Diff = player.Rating / winCampTotalRating * lossCampTotalRating - player.Rating
//		//实际得分
//		result.GainScore = WinScore
//		//预期得分（胜率）
//		result.GainExpectScore = Rating(result.Diff)
//		log.Printf("win玩家 %d  result:%#v\n", player.UserId, result)
//	}
//
//	//loss阵营玩家 该场对局预期胜率
//	for _, player := range  lossCampPlayerList {
//		result := playerBattleResult[player.UserId]
//		//计算分差
//		result.Diff = player.Rating / lossCampTotalRating * winCampTotalRating - player.Rating
//		//实际得分
//		result.GainScore = LossScore
//		//预期得分（胜率）
//		result.GainExpectScore = Rating(result.Diff)
//		log.Printf("loss玩家 %d  result:%#v\n",  player.UserId, result)
//	}
//
//	//保存玩家对局结果
//	for _, result := range playerBattleResult {
//		e.store.AddPlayerRecord(result)
//	}
//
//	//更新玩家战力
//	for _, player := range battleResult.PlayerList {
//		//获取最近对战记录，计算G和Ge
//		queue, ok := e.store.Queue(player.UserId)
//		if !ok {
//			continue
//		}
//
//		player.Rating = e.RateNow(player.Rating, queue.G(), queue.Ge())
//
//		//保存玩家当前排位分
//		e.playerRating[player.UserId] = player.Rating
//	}
//}