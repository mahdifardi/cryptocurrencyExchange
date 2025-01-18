package limit

type Match struct {
	Ask        *LimitOrder
	Bid        *LimitOrder
	SizeFilled float64
	Price      float64
}

func (l *Limit) Fill(o *LimitOrder) []Match {
	var (
		matches        = []Match{}
		ordersToDelete = []*LimitOrder{}
	)

	for _, order := range l.Orders {

		if o.IsFilled() {
			break
		}

		match := l.fillOrder(order, o)
		matches = append(matches, match)

		l.TotalVolume -= match.SizeFilled

		if order.IsFilled() {
			ordersToDelete = append(ordersToDelete, order)
		}

	}

	for _, order := range ordersToDelete {
		l.DeleteOrder(order)
	}

	return matches
}

func (l *Limit) fillOrder(a, b *LimitOrder) Match {
	var (
		bid        *LimitOrder
		ask        *LimitOrder
		sizeFilled float64
	)

	if a.Bid {
		bid = a
		ask = b
	} else {
		bid = b
		ask = a
	}

	if a.Size > b.Size {
		a.Size -= b.Size
		sizeFilled = b.Size
		b.Size = 0

	} else {
		b.Size -= a.Size
		sizeFilled = a.Size
		a.Size = 0
	}

	return Match{
		Ask:        ask,
		Bid:        bid,
		SizeFilled: sizeFilled,
		Price:      l.Price,
	}
}
