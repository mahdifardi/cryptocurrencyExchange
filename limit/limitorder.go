package limit

import (
	"fmt"
	"math/rand"
	"time"
)

type LimitOrder struct {
	ID        int64
	UserId    int64
	Size      float64
	Bid       bool
	Limit     *Limit
	Timestamp int64
}

type LimitOrders []*LimitOrder

func (o LimitOrders) Len() int           { return len(o) }
func (o LimitOrders) Swap(i, j int)      { o[i], o[j] = o[j], o[i] }
func (o LimitOrders) Less(i, j int) bool { return o[i].Timestamp < o[j].Timestamp }

func (o *LimitOrder) String() string {
	return fmt.Sprintf("[size: %.2f]", o.Size)
}

func (o *LimitOrder) IsFilled() bool {
	return o.Size == 0.0
}

func NewLimitOrder(bid bool, size float64, userID int64) *LimitOrder {
	return &LimitOrder{
		ID:        int64(rand.Intn(1000000)),
		UserId:    userID,
		Size:      size,
		Bid:       bid,
		Timestamp: time.Now().UnixNano(),
	}
}
