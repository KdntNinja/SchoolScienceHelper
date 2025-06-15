package achievements

type Achievement struct {
	ID       string
	UserID   string
	Name     string
	Desc     string
	EarnedAt int64
}
