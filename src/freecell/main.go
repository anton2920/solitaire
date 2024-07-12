package freecell

import (
	"github.com/anton2920/gofa/gui"
	"github.com/anton2920/gofa/gui/color"
	"github.com/anton2920/gofa/gui/gr"
	"github.com/anton2920/gofa/util"
)

type GameState int

const (
	GameNothing GameState = iota
	GameRunning
	GameEnd
)

type CardSuit int16

const (
	Blank CardSuit = iota
	Clubs
	Diamonds
	Hearts
	Aces
)

type CardState byte

const (
	Normal CardState = iota
	Pressed
)

type CursorState byte

const (
	DefaultCursor CursorState = iota
	UpArrow
	DownArrow
)

type Card struct {
	Value int16
	Suit  CardSuit

	X, Y int16

	State CardState
}

type Direction byte

const (
	Left Direction = iota
	Right
)

const (
	GameWidth  = 632
	GameHeight = 452

	MenuHeight = 20

	TableLeft    = 7
	TableTop     = 126
	TableColumns = 8

	CardWidth    = 71
	CardHeight   = 96
	CardYPadding = 18

	PlaceholderTop    = MenuHeight
	PlaceholderWidth  = 71
	PlaceholderHeight = 96
)

var (
	State GameState

	Table []Card

	Freecells [4]Card
	Goals     [4]Card

	FaceDirection Direction

	SelectedCard  *Card
	CurrentCursor CursorState

	Seed int32
)

func (card *Card) Red() bool {
	return (card.Suit == Diamonds) || (card.Suit == Hearts)
}

func Rand() int32 {
	Seed = (Seed*214013 + 2531011) & ((1 << 31) - 1)
	return Seed >> 16
}

func Deal(N int) {
	if Table == nil {
		Table = make([]Card, 0, 52)
	}
	Table = Table[:0]

	for j := 13; j > 0; j-- {
		for i := Aces; i > 0; i-- {
			Table = append(Table, Card{Value: int16(j), Suit: i})
		}
	}

	Seed = int32(N)
	for i := 0; i < 51; i++ {
		j := 51 - int(Rand())%(52-i)
		Table[i], Table[j] = Table[j], Table[i]
	}

	for k := 0; k < len(Table); k++ {
		card := &Table[k]

		i := k % TableColumns
		j := k / TableColumns
		card.X = int16(TableLeft + i*(TableLeft+CardWidth))
		card.Y = int16(TableTop + j*18)
	}
}

func NewGame(N int) {
	SelectedCard = nil

	if (len(Freecells) != 4) && (len(Freecells) != len(Goals)) {
		panic("number of freecells/goals must be 4")
	}
	for i := 0; i < len(Freecells); i++ {
		Freecells[i].Suit = Blank
		Freecells[i].X = int16(i * PlaceholderWidth)
		Freecells[i].Y = PlaceholderTop

		Goals[len(Goals)-i-1].Suit = Blank
		Goals[len(Goals)-i-1].X = int16(GameWidth - (i+1)*PlaceholderWidth)
		Goals[len(Goals)-i-1].Y = PlaceholderTop
	}

	State = GameRunning
	Deal(N)
}

func FindBottomCard(needle *Card) *Card {
	var bottomCard *Card
	var bottomY int16

	for i := 0; i < len(Table); i++ {
		card := &Table[i]
		if needle.X == card.X {
			if card.Y > bottomY {
				bottomCard = card
				bottomY = card.Y
			}
		}
	}

	return bottomCard
}

func AllowedToMove() int {
	var freecells, columns int
	for i := 0; i < len(Freecells); i++ {
		if Freecells[i].Suit == Blank {
			freecells++
		}
	}
	for i := 0; i < TableColumns; i++ {
		columnRect := TableColumnRect(i)
		if FindBottomCard(&Card{X: int16(columnRect.X0), Y: int16(columnRect.Y0)}) == nil {
			columns++
		}
	}
	return (freecells + 1) * (1 << columns)
}

func PowerMove(src *Card, dst *Card, pressed bool) bool {
	if (!CardOnTable(src)) || (!CardOnTable(dst)) {
		return false
	}

	var canPowerMove bool
	cards := make([]*Card, 0, 52)
	card := src
	for i := 0; (i < AllowedToMove()) && (card != nil); i++ {
		cards = append(cards, card)
		if CanMove(card, dst) {
			canPowerMove = true
			break
		}
		next := FindCardAbove(card)
		if !CanMove(card, next) {
			break
		}
		card = next
	}
	if canPowerMove {
		if pressed {
			dstY := dst.Y
			for i := len(cards) - 1; i >= 0; i-- {
				card := cards[i]
				card.X = dst.X
				card.Y = dstY + CardYPadding
				dstY += CardYPadding
			}
		}
	}
	return canPowerMove
}

func PowerMoveOnTable(window *gui.Window, src *Card, idx int, pressed bool) bool {
	if !CardOnTable(src) {
		return false
	}

	cards := make([]*Card, 0, 52)
	card := src
	for i := 0; (i < AllowedToMove()) && (card != nil); i++ {
		cards = append(cards, card)
		next := FindCardAbove(card)
		if !CanMove(card, next) {
			break
		}
		card = next
	}
	if pressed {
		/*
			dialog, err := window.NewSubwindow("Power move", 100, 100, 100, 100, 0)
			if err != nil {
				panic("failed to create window")
			}
			renderer := gui.NewSoftwareRenderer(dialog)
			ui := gui.NewUI(renderer)
			_ = ui

			renderer.Clear(color.Grey(200))
		*/

		dstX := int16(TableColumnRect(idx).X0)
		dstY := int16(TableTop)
		for i := len(cards) - 1; i >= 0; i-- {
			card := cards[i]
			card.X = dstX
			card.Y = dstY
			dstY += CardYPadding
		}
	}
	return len(cards) > 1

}

func CanMove(src, dst *Card) bool {
	return (src != nil) && (dst != nil) && (src.Red() != dst.Red()) && (dst.Value-src.Value == 1)
}

func CanMove2Goal(src, dst *Card) bool {
	return ((src.Suit == dst.Suit) && (src.Value-dst.Value == 1)) || ((src.Value == 1) && (dst.Suit == Blank))
}

func FindCardAbove(card *Card) *Card {
	for i := 0; i < len(Table); i++ {
		if (Table[i].X == card.X) && (Table[i].Y == card.Y-CardYPadding) {
			return &Table[i]
		}
	}
	return nil
}

func FindCardOnTable(card *Card) int {
	for i := 0; i < len(Table); i++ {
		if &Table[i] == card {
			return i
		}
	}
	return -1
}

func CardOnTable(card *Card) bool {
	return FindCardOnTable(card) != -1
}

func RemoveFromTable(card *Card) {
	Table = util.RemoveAtIndex(Table, FindCardOnTable(card))
}

func RemoveFromFreecell(card *Card) {
	for i := 0; i < len(Freecells); i++ {
		if &Freecells[i] == card {
			card.Suit = Blank
			card.X = int16(i * PlaceholderWidth)
			card.Y = PlaceholderTop
			break
		}
	}
}

func SetSelectedCard(card *Card) {
	if card != nil {
		SelectedCard = card
		SelectedCard.State = Pressed
	}
}

func RemoveSelection() {
	SelectedCard.State = Normal
	SelectedCard = nil
}

func MoveCard(src, dst *Card) {
	rect := CardRect(dst)

	*dst = *src
	dst.State = Normal
	dst.X = int16(rect.X0)
	dst.Y = int16(rect.Y0)
}

func DrawMenu(window *gui.Window, renderer gui.Renderer, ui *gui.UI) {
	renderer.RenderSolidRectWH(0, 0, window.Width, MenuHeight, color.RGB(0xD4, 0xD0, 0xC8))
}

func DrawRectWithShadow(renderer gui.Renderer, x0, y0, x1, y1 int, pclr, sclr color.Color) {
	renderer.RenderLine(x0, y0, x1-1, y0, pclr)
	renderer.RenderLine(x0, y0, x0, y1-1, pclr)
	renderer.RenderLine(x0+1, y1, x1, y1, sclr)
	renderer.RenderLine(x1, y0+1, x1, y1, sclr)
}

func DrawBackground(window *gui.Window, renderer gui.Renderer) {
	renderer.Clear(color.RGB(0, 127, 0))

	for i := 0; i < len(Freecells); i++ {
		const width = PlaceholderWidth - 1
		const height = PlaceholderHeight - 1

		x := i * (width + 1)
		y := PlaceholderTop

		DrawRectWithShadow(renderer, x, y, x+width, y+height, color.Black, color.Green)
		DrawRectWithShadow(renderer, window.Width-width-x-1, y, window.Width-x-1, y+height, color.Black, color.Green)
	}

	DrawRectWithShadow(renderer, 297, 38, 334, 75, color.Green, color.Black)
}

func DrawFace(window *gui.Window, renderer gui.Renderer, assets *gr.Pixmap) {
	const x = 298
	const y = 39
	const width = 36
	renderer.RenderPixmap(assets.Sub(320+int(width*FaceDirection), 453, 355+int(width*FaceDirection), 488), x, y)
}

func DrawGiantFace(window *gui.Window, renderer gui.Renderer, assets *gr.Pixmap) {
	renderer.RenderPixmap(assets.Sub(0, 453, 320, 773), 10, 126)
}

func GetCardPixmap(assets *gr.Pixmap, card *Card) gr.Pixmap {
	const x = 632
	const y = 0

	i := int(card.Value - 1)
	j := int(card.Suit-1) + int(card.State*4)

	return assets.Sub(x+i*CardWidth, y+j*CardHeight, x+(i+1)*CardWidth, y+(j+1)*CardHeight)
}

func DrawCard(renderer gui.Renderer, assets *gr.Pixmap, card *Card) {
	if card.Suit != Blank {
		renderer.RenderPixmap(GetCardPixmap(assets, card), int(card.X), int(card.Y))
	}
}

func DrawCards(window *gui.Window, renderer gui.Renderer, ui *gui.UI, assets *gr.Pixmap) {
	const marginLeft = 7
	const marginTop = 126

	const paddingLeft = marginLeft
	const paddingTop = CardYPadding

	for i := 0; i < len(Table); i++ {
		DrawCard(renderer, assets, &Table[i])
	}

	for i := 0; i < len(Freecells); i++ {
		DrawCard(renderer, assets, &Freecells[i])
	}

	for i := 0; i < len(Goals); i++ {
		DrawCard(renderer, assets, &Goals[i])
	}
}

func DrawCursor(window *gui.Window, renderer gui.Renderer, ui *gui.UI, assets *gr.Pixmap) {
	switch CurrentCursor {
	case DefaultCursor:
		window.EnableCursor()
	case UpArrow:
		window.DisableCursor()
		old := assets.Alpha
		assets.Alpha = gr.Alpha8bit
		renderer.RenderPixmap(assets.Sub(406, 453, 415, 472), ui.MouseX, ui.MouseY)
		assets.Alpha = old
	case DownArrow:
		window.DisableCursor()
		old := assets.Alpha
		assets.Alpha = gr.Alpha8bit
		renderer.RenderPixmap(assets.Sub(392, 453, 406, 480), ui.MouseX, ui.MouseY-27)
		assets.Alpha = old
	}
}

func CardRect(card *Card) gui.Rect {
	return gui.Rect{int(card.X), int(card.Y), int(card.X) + CardWidth - 1, int(card.Y) + CardHeight - 1}
}

func TableColumnRect(idx int) gui.Rect {
	return gui.Rect{TableLeft + idx*(TableLeft+CardWidth), TableTop, TableLeft + idx*(TableLeft+CardWidth) + CardWidth - 1, GameHeight - 1}
}

func HandleMouseInput(window *gui.Window, ui *gui.UI) {
	mouse := gui.Rect{ui.MouseX, ui.MouseY, ui.MouseX, ui.MouseY}
	CurrentCursor = DefaultCursor

	/* Handle face turn. */
	faceLeftRect := gui.Rect{0, 20, 284, 116}
	faceRightRect := gui.Rect{348, 20, 632, 116}
	if faceLeftRect.Contains(mouse) {
		FaceDirection = Left
	} else if faceRightRect.Contains(mouse) {
		FaceDirection = Right
	}

	for i := 0; i < len(Freecells); i++ {
		freecell := &Freecells[i]
		over := CardRect(freecell).Contains(mouse)
		pressed := ui.ButtonLogicDown(gui.ID(freecell), over)

		if over {
			if (pressed) && (SelectedCard == nil) && (freecell.Suit != Blank) {
				SetSelectedCard(freecell)
			} else if (pressed) && (SelectedCard == freecell) {
				RemoveSelection()
			} else if (SelectedCard != nil) && (freecell.Suit == Blank) {
				CurrentCursor = UpArrow
				if pressed {
					MoveCard(SelectedCard, freecell)
					RemoveFromFreecell(SelectedCard)
					RemoveFromTable(SelectedCard)
					RemoveSelection()
				}
			}
		}
	}

	for i := 0; i < len(Goals); i++ {
		goal := &Goals[i]
		over := CardRect(goal).Contains(mouse)
		pressed := ui.ButtonLogicDown(gui.ID(goal), over)

		if (over) && (SelectedCard != nil) && (CanMove2Goal(SelectedCard, goal)) {
			CurrentCursor = UpArrow
			if pressed {
				MoveCard(SelectedCard, goal)
				RemoveFromFreecell(SelectedCard)
				RemoveFromTable(SelectedCard)
				RemoveSelection()
			}
		}
	}

	for i := 0; i < TableColumns; i++ {
		columnRect := TableColumnRect(i)
		bottomCard := FindBottomCard(&Card{X: int16(columnRect.X0), Y: int16(columnRect.Y0)})

		overRect := columnRect
		if bottomCard != nil {
			overRect.Y1 = CardRect(bottomCard).Y1
		}
		over := overRect.Contains(mouse)
		pressed := ui.ButtonLogicDown(gui.ID(uintptr(TableLeft+i)), over)

		if (pressed) && (SelectedCard == nil) {
			SetSelectedCard(bottomCard)
		} else if (pressed) && (SelectedCard == bottomCard) {
			RemoveSelection()
		} else if (over) && (SelectedCard != nil) {
			if bottomCard != nil {
				if PowerMove(SelectedCard, bottomCard, pressed) {
					CurrentCursor = DownArrow
					if pressed {
						RemoveSelection()
					}
				} else if CanMove(SelectedCard, bottomCard) {
					CurrentCursor = DownArrow
					if pressed {
						SelectedCard.State = Normal
						SelectedCard.X = bottomCard.X
						SelectedCard.Y = bottomCard.Y + CardYPadding
						if !CardOnTable(SelectedCard) {
							Table = append(Table, *SelectedCard)
							RemoveFromFreecell(SelectedCard)
						}
						RemoveSelection()
					}
				}
			} else {
				if PowerMoveOnTable(window, SelectedCard, i, pressed) {
					CurrentCursor = DownArrow
					if pressed {
						RemoveSelection()
					}
				} else {
					CurrentCursor = DownArrow
					if pressed {
						SelectedCard.State = Normal
						SelectedCard.X = int16(columnRect.X0)
						SelectedCard.Y = int16(columnRect.Y0)
						if !CardOnTable(SelectedCard) {
							Table = append(Table, *SelectedCard)
							RemoveFromFreecell(SelectedCard)
						}
						RemoveSelection()
					}
				}
			}
		}
	}
}

func RemoveCardIfUseless(card *Card) bool {
	if (card != nil) && (card.Suit != Blank) {
		useless := true

		for i := 0; i < len(Table); i++ {
			if CanMove(&Table[i], card) {
				useless = false
				break
			}
		}
		if useless {
			for i := 0; i < len(Goals); i++ {
				goal := &Goals[i]

				if CanMove2Goal(card, goal) {
					MoveCard(card, goal)
					RemoveFromFreecell(card)
					RemoveFromTable(card)
					return true
				}
			}
		}
	}

	return false
}

func Autoplay() {
	removed := true
	for removed {
		removed = false
		for i := 0; i < TableColumns; i++ {
			columnRect := TableColumnRect(i)
			card := FindBottomCard(&Card{X: int16(columnRect.X0), Y: int16(columnRect.Y0)})
			removed = RemoveCardIfUseless(card) || removed
		}

		for i := 0; i < len(Freecells); i++ {
			removed = RemoveCardIfUseless(&Freecells[i]) || removed
		}
	}
}

func SortCards() {
	for i := 1; i < len(Table); i++ {
		for j := 0; j < i; j++ {
			if Table[j].Y > Table[i].Y {
				Table[i], Table[j] = Table[j], Table[i]
			}
		}
	}
}

func GameWon() bool {
	var kings int
	for i := 0; i < len(Goals); i++ {
		if Goals[i].Value == 13 {
			kings++
		}
	}
	return kings == 4
}

func Main(window *gui.Window, renderer gui.Renderer, ui *gui.UI, assets *gr.Pixmap) {
	DrawBackground(window, renderer)
	DrawCards(window, renderer, ui, assets)
	DrawMenu(window, renderer, ui)

	switch State {
	case GameNothing:
		DrawFace(window, renderer, assets)
	case GameRunning:
		DrawFace(window, renderer, assets)

		HandleMouseInput(window, ui)
		DrawCursor(window, renderer, ui, assets)

		Autoplay()
		SortCards()

		if GameWon() {
			CurrentCursor = DefaultCursor
			State = GameEnd
		}
	case GameEnd:
		DrawGiantFace(window, renderer, assets)
		DrawCursor(window, renderer, ui, assets)
	}
}
