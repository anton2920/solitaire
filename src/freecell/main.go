package freecell

import (
	"github.com/anton2920/gofa/gui"
	"github.com/anton2920/gofa/gui/color"
	"github.com/anton2920/gofa/gui/gr"
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

type Card struct {
	Value int16
	Suit  CardSuit

	X, Y int16

	State CardState
}

const (
	MenuHeight = 20

	TableLeft = 7
	TableTop  = 126
	TableCols = 8

	CardWidth  = 71
	CardHeight = 96
)

var (
	State GameState

	Table []Card

	Freecells [4]Card
	Goals     [4]Card

	Seed int32
)

func (card *Card) SetCoordinates(i, j int) {
	card.X = int16(TableLeft + i*(TableLeft+CardWidth))
	card.Y = int16(TableTop + j*18)
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
		const width = 70
		const height = 95

		x := i * (width + 1)
		y := MenuHeight

		DrawRectWithShadow(renderer, x, y, x+width, y+height, color.Black, color.Green)
		DrawRectWithShadow(renderer, window.Width-width-x-1, y, window.Width-x-1, y+height, color.Black, color.Green)
	}

	DrawRectWithShadow(renderer, 297, 38, 334, 75, color.Green, color.Black)
}

func DrawFace(window *gui.Window, renderer gui.Renderer, ui *gui.UI, assets *gr.Pixmap) {
	const x = 298
	const y = 39

	renderer.RenderPixmap(assets.Sub(320, 453, 356, 488), x, y)
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
	const paddingTop = 18

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

func Main(window *gui.Window, renderer gui.Renderer, ui *gui.UI, assets *gr.Pixmap) {
	DrawBackground(window, renderer)
	DrawFace(window, renderer, ui, assets)
	DrawCards(window, renderer, ui, assets)

	DrawMenu(window, renderer, ui)
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

	for i := 0; i < len(Table); i++ {
		card := &Table[i]
		card.SetCoordinates(i%TableCols, i/TableCols)
	}
}

func NewGame(N int) {
	State = GameRunning
	Deal(N)
}
