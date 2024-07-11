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
	Blank    CardSuit = iota
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
	MenuHeight = 20

	TableLeft = 7
	TableTop  = 126
	TableCols = 8

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

	SelectedCard *Card

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

		i := k % TableCols
		j := k / TableCols
		card.X = int16(TableLeft + i*(TableLeft+CardWidth))
		card.Y = int16(TableTop + j*18)
	}
}

func NewGame(N int) {
	SelectedCard = nil
	clear(Freecells[:])
	clear(Goals[:])

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

func CanMove(src, dst *Card) bool {
	return (src.Red() != dst.Red()) && (dst.Value-src.Value == 1)
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
			break
		}
	}
}

func SetSelectedCard(card *Card) {
	SelectedCard = card
	SelectedCard.State = Pressed
}

func RemoveSelection() {
	SelectedCard.State = Normal
	SelectedCard = nil
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

	for i := 0; i < 4; i++ {
		const width = PlaceholderWidth - 1
		const height = PlaceholderHeight - 1

		x := i * (width + 1)
		y := PlaceholderTop

		DrawRectWithShadow(renderer, x, y, x+width, y+height, color.Black, color.Green)
		DrawRectWithShadow(renderer, window.Width-width-x-1, y, window.Width-x-1, y+height, color.Black, color.Green)
	}

	DrawRectWithShadow(renderer, 297, 38, 334, 75, color.Green, color.Black)
}

func DrawFace(window *gui.Window, renderer gui.Renderer, ui *gui.UI, assets *gr.Pixmap) {
	const x = 298
	const y = 39
	const width = 36
	renderer.RenderPixmap(assets.Sub(320+int(width*FaceDirection), 453, 355+int(width*FaceDirection), 488), x, y)
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

func HandleMouseInput(window *gui.Window, renderer gui.Renderer, ui *gui.UI) {
	const debug = true
	const alphaMask = 0x7FFFFFFF

	mouseRect := gui.Rect{ui.MouseX, ui.MouseY, ui.MouseX, ui.MouseY}

	if SelectedCard != nil {
		if debug {
			renderer.RenderSolidRectWH(int(SelectedCard.X), int(SelectedCard.Y), CardWidth, CardHeight, color.Red&alphaMask)
		}
	}

	/* Handle face turn. */
	faceLeftRect := gui.Rect{0, 20, 297, 116}
	faceRightRect := gui.Rect{335, 20, 632, 116}
	if faceLeftRect.Contains(mouseRect) {
		FaceDirection = Left
	} else if faceRightRect.Contains(mouseRect) {
		FaceDirection = Right
	}

	/* Handle freecell moves. */
	for i := 0; i < len(Freecells); i++ {
		freecell := &Freecells[i]
		freecellRect := gui.Rect{i * PlaceholderWidth, PlaceholderTop, (i+1)*PlaceholderWidth - 1, PlaceholderTop + PlaceholderHeight - 1}

		if ui.ButtonLogic(gui.ID(freecell), freecellRect.Contains(mouseRect)) {
			if SelectedCard == nil {
				SetSelectedCard(freecell)
			} else if SelectedCard == freecell {
				RemoveSelection()
			} else if freecell.Suit == Blank {
				*freecell = *SelectedCard
				freecell.State = Normal
				freecell.X = int16(freecellRect.X0)
				freecell.Y = int16(freecellRect.Y0)

				RemoveFromTable(SelectedCard)
				RemoveSelection()
			}
			return
		} else if ui.IsHot {
			if debug {
				renderer.RenderSolidRect(freecellRect.X0, freecellRect.Y0, freecellRect.X1, freecellRect.Y1, color.Lite(color.Green)&alphaMask)
			}
			return
		}
	}

	/* Handle goal moves. */
	for i := 0; i < len(Goals); i++ {
		goal := &Goals[i]
		goalRect := gui.Rect{window.Width - (i+1)*PlaceholderWidth, PlaceholderTop, window.Width - i*PlaceholderWidth - 1, PlaceholderTop + PlaceholderHeight - 1}

		if ui.ButtonLogic(gui.ID(goal), goalRect.Contains(mouseRect)) {
			if SelectedCard != nil {
				if ((SelectedCard.Suit == goal.Suit) && (SelectedCard.Value-goal.Value == 1)) || ((SelectedCard.Value == 1) && (goal.Suit == Blank)) {
					*goal = *SelectedCard
					goal.State = Normal
					goal.X = int16(goalRect.X0)
					goal.Y = int16(goalRect.Y0)

					RemoveFromFreecell(SelectedCard)
					RemoveFromTable(SelectedCard)
					RemoveSelection()
				}
			}
			return
		} else if ui.IsHot {
			if debug {
				renderer.RenderSolidRect(goalRect.X0, goalRect.Y0, goalRect.X1, goalRect.Y1, color.Lite(color.Red)&alphaMask)
			}
			return
		}
	}

	/* Handle table cards select and move. */
	for i := 0; i < len(Table); i++ {
		card := &Table[i]
		cardRect := gui.Rect{int(card.X), int(card.Y), int(card.X) + CardWidth - 1, int(card.Y) + CardHeight - 1}
		if ui.ButtonLogic(gui.ID(card), cardRect.Contains(mouseRect)) {
			bottomCard := FindBottomCard(card)

			if SelectedCard == nil {
				SetSelectedCard(bottomCard)
			} else if bottomCard == SelectedCard {
				RemoveSelection()
			} else {
				if bottomCard != nil {
					if CanMove(SelectedCard, bottomCard) {
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
			}
			return
		} else if ui.IsHot {
			if debug {
				renderer.RenderSolidRect(cardRect.X0, cardRect.Y0, cardRect.X1, cardRect.Y1, color.Lite(color.Blue)&alphaMask)
			}
			return
		}
	}

	/* Handle moves on empty table. */
	for i := 0; i < TableCols; i++ {
		columnRect := gui.Rect{TableLeft + i*(TableLeft+CardWidth), TableTop, TableLeft + i*(TableLeft+CardWidth) + CardWidth - 1, window.Height - 1}
		if FindBottomCard(&Card{X: int16(columnRect.X0), Y: int16(columnRect.Y0)}) == nil {
			if ui.ButtonLogic(gui.ID(uintptr(100+i)), columnRect.Contains(mouseRect)) {
				if SelectedCard != nil {
					SelectedCard.State = Normal
					SelectedCard.X = int16(columnRect.X0)
					SelectedCard.Y = int16(columnRect.Y0)
					if !CardOnTable(SelectedCard) {
						Table = append(Table, *SelectedCard)
						RemoveFromFreecell(SelectedCard)
					}
					RemoveSelection()
				}
				return
			} else if ui.IsHot {
				if debug {
					renderer.RenderSolidRect(columnRect.X0, columnRect.Y0, columnRect.X1, columnRect.Y1, color.Blue&alphaMask)
				}
				return
			}
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

func Main(window *gui.Window, renderer gui.Renderer, ui *gui.UI, assets *gr.Pixmap) {
	DrawBackground(window, renderer)
	DrawFace(window, renderer, ui, assets)
	DrawCards(window, renderer, ui, assets)
	DrawMenu(window, renderer, ui)

	HandleMouseInput(window, renderer, ui)

	SortCards()
}
