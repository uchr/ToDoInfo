package todo

type TaskRottenness int

const (
	ZombieTaskRottenness = iota
	TiredTaskRottenness
	RipeTaskRottenness
	FreshTaskRottenness
)

type Daysums struct {
	Overall     int
	ListDaysums map[string]int
}

type TaskRottennessInfo struct {
	TaskName   string
	TaskList   string
	Age        int
	Rottenness TaskRottenness
}

func (r TaskRottenness) String() string {
	switch r {
	case ZombieTaskRottenness:
		return "🤢"
	case TiredTaskRottenness:
		return "🥱"
	case RipeTaskRottenness:
		return "😏"
	case FreshTaskRottenness:
		return "😊"
	}
	return "❓"
}
