package leaderboard

type LeaderboardEntry struct {
	UserID   string
	Username string
	Score    int
	Streak   int
	Rank     int
}
