package db

type User struct {
	Name   string `json:"name"`
	UID    string `json:"uid" binding:"required"`
	Avatar string `json:"avatar" binding:"required"`
}

// func (db *Database) SaveUser(user *User) error {
// 	member := &redis.Z{
// 		Score:  float64(user.Points),
// 		Member: user.UID,
// 	}
// 	pipe := db.RedisClient.TxPipeline()
// 	pipe.ZAdd(Ctx, "leaderboard", member)
// 	rank := pipe.ZRank(Ctx, leaderboardKey, user.UID)
// 	_, err := pipe.Exec(Ctx)
// 	if err != nil {
// 		return err
// 	}
// 	fmt.Println(rank.Val(), err)
// 	return nil
// }
