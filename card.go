package main

type ValueType int16

const (
	None ValueType = iota
	Ace
	Two
	Three
	Four
	Five
	Six
	Seven
	Eight
	Nine
	Ten
	Jack
	Queen
	King
)

type SuitType int16

const (
	Blank SuitType = iota
	Clubs
	Diamonds
	Hearts
	Aces
)

type Card struct {
	Value ValueType
	Suit  SuitType

	X, Y int16

	Selected bool
}

const (
	CardWidth    = 71
	CardHeight   = 96
	CardYPadding = 18
)

func (card *Card) Red() bool {
	return (card.Suit == Diamonds) || (card.Suit == Hearts)
}

func CanMove(src, dst *Card) bool {
	return (src != nil) && (src.Suit != Blank) && (dst != nil) && (dst.Suit != Blank) && (src.Red() != dst.Red()) && (dst.Value-src.Value == 1)
}

func CanMove2Goal(src, dst *Card) bool {
	return ((src.Suit == dst.Suit) && (src.Value-dst.Value == 1)) || ((src.Value == 1) && (dst.Suit == Blank))
}
