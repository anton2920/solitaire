package main

import (
	"math/rand/v2"
	"unsafe"

	"github.com/anton2920/gofa/gui"
	"github.com/anton2920/gofa/gui/color"
	"github.com/anton2920/gofa/gui/gr"
	"github.com/anton2920/gofa/slices"
	"github.com/anton2920/gofa/trace"
	"github.com/anton2920/gofa/util"
)

type GameState int

const (
	GameNothing GameState = iota
	GameRunning
	GameEnd
)

type CursorType int

const (
	CursorDefault CursorType = iota
	CursorUp
	CursorDown
)

type FreeCell struct {
	/* Window-related stuff. */
	Window   *gui.Window
	Renderer gui.Renderer
	UI       *gui.UI
	Assets   *gr.Pixmap

	/* Game-related stuff. */
	State GameState

	Table     []Card
	FreeCells [4]Card
	Goals     [4]Card

	SelectedCard    *Card
	AutoplayAllowed bool

	FaceDirection int
	Cursor        CursorType

	RandSeed int

	/* Measurements. */
	Width int

	/* TODO(anton2920): replace with UI Layout menu height. */
	MenuHeight int

	PlaceholderTop    int
	PlaceholderWidth  int
	PlaceholderHeight int

	TableLeft    int
	TableTop     int
	TableColumns int
}

func NewFreeCell(window *gui.Window, renderer gui.Renderer, ui *gui.UI, assets *gr.Pixmap) FreeCell {
	var game FreeCell

	game.Window = window
	game.Renderer = renderer
	game.UI = ui
	game.Assets = assets

	game.Table = make([]Card, 0, 52)

	game.Width = 632
	game.MenuHeight = 20

	game.PlaceholderTop = game.MenuHeight
	game.PlaceholderWidth = CardWidth
	game.PlaceholderHeight = CardHeight

	game.TableLeft = 7
	game.TableTop = 126
	game.TableColumns = 8

	return game
}

func (game *FreeCell) Rand() int {
	game.RandSeed = (game.RandSeed*214013 + 2531011) & ((1 << 31) - 1)
	return game.RandSeed >> 16
}

func (game *FreeCell) Deal(N int) {
	game.Table = game.Table[:0]
	game.AutoplayAllowed = false
	game.SelectedCard = nil

	for j := King; j >= Ace; j-- {
		for i := Aces; i >= Clubs; i-- {
			game.Table = append(game.Table, Card{Value: j, Suit: i})
		}
	}

	game.RandSeed = N
	for i := 0; i < len(game.Table)-1; i++ {
		j := (len(game.Table) - 1) - game.Rand()%(len(game.Table)-i)
		game.Table[i], game.Table[j] = game.Table[j], game.Table[i]
	}

	for k := 0; k < len(game.Table); k++ {
		card := &game.Table[k]

		i := k % game.TableColumns
		j := k / game.TableColumns
		card.X = int16(game.TableLeft + i*(game.TableLeft+CardWidth))
		card.Y = int16(game.TableTop + j*CardYPadding)
	}

	for i := 0; i < len(game.FreeCells); i++ {
		game.FreeCells[i].Suit = Blank
		game.FreeCells[i].Value = 0
		game.FreeCells[i].X = int16(i * game.PlaceholderWidth)
		game.FreeCells[i].Y = int16(game.PlaceholderTop)

		game.Goals[len(game.Goals)-i-1].Suit = Blank
		game.Goals[len(game.Goals)-i-1].Value = 0
		game.Goals[len(game.Goals)-i-1].X = int16(game.Width - (i+1)*game.PlaceholderWidth)
		game.Goals[len(game.Goals)-i-1].Y = int16(game.PlaceholderTop)
	}

	var n int
	buffer := make([]byte, 128)
	n += copy(buffer[n:], Title)
	n += copy(buffer[n:], ": FreeCell Game #")
	n += slices.PutInt(buffer[n:], N)
	title := unsafe.String(&buffer[0], n)
	game.Window.SetTitle(title)

	game.State = GameRunning
}

func (game *FreeCell) NewRandomGame() {
	game.Deal((rand.Int() % 30000) + 1)
}

func (game *FreeCell) NewSelectedGame(N int) {
	game.Deal(N)
}

func (game *FreeCell) FindBottomCard(needle *Card) *Card {
	defer trace.End(trace.Begin(""))

	var bottomCard *Card
	var bottomY int16

	for i := 0; i < len(game.Table); i++ {
		card := &game.Table[i]
		if needle.X == card.X {
			if card.Y > bottomY {
				bottomCard = card
				bottomY = card.Y
			}
		}
	}

	return bottomCard
}

func (game *FreeCell) AllowedToMove(onTable bool) int {
	defer trace.End(trace.Begin(""))

	var freecells, columns int
	for i := 0; i < len(game.FreeCells); i++ {
		if game.FreeCells[i].Suit == Blank {
			freecells++
		}
	}
	for i := 0; i < game.TableColumns; i++ {
		columnRect := game.TableColumnRect(i)
		if game.FindBottomCard(&Card{X: int16(columnRect.X0), Y: int16(columnRect.Y0)}) == nil {
			columns++
		}
	}
	if onTable {
		columns--
	}
	return (freecells + 1) * (1 << columns)
}

func (game *FreeCell) PowerMove(src *Card, dst *Card, pressed bool) bool {
	defer trace.End(trace.Begin(""))

	if (!game.CardOnTable(src)) || (!game.CardOnTable(dst)) {
		return false
	}

	var canPowerMove bool
	cards := make([]*Card, 0, 52)
	card := src
	for i := 0; (i < game.AllowedToMove(false)) && (card != nil); i++ {
		cards = append(cards, card)
		if CanMove(card, dst) {
			canPowerMove = true
			break
		}
		next := game.FindCardAbove(card)
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

func (game *FreeCell) PowerMoveOnTable(src *Card, idx int, pressed bool) bool {
	defer trace.End(trace.Begin(""))

	if !game.CardOnTable(src) {
		return false
	}

	cards := make([]*Card, 0, 52)
	card := src
	for i := 0; (i < game.AllowedToMove(true)) && (card != nil); i++ {
		cards = append(cards, card)
		next := game.FindCardAbove(card)
		if !CanMove(card, next) {
			break
		}
		card = next
	}
	if pressed {
		dstX := int16(game.TableColumnRect(idx).X0)
		dstY := int16(game.TableTop)
		for i := len(cards) - 1; i >= 0; i-- {
			card := cards[i]
			card.X = dstX
			card.Y = dstY
			dstY += CardYPadding
		}
	}
	return len(cards) > 1

}

func (game *FreeCell) FindCardAbove(card *Card) *Card {
	defer trace.End(trace.Begin(""))

	for i := 0; i < len(game.Table); i++ {
		if (game.Table[i].X == card.X) && (game.Table[i].Y == card.Y-CardYPadding) {
			return &game.Table[i]
		}
	}
	return nil
}

func (game *FreeCell) FindCardOnTable(card *Card) int {
	for i := 0; i < len(game.Table); i++ {
		if &game.Table[i] == card {
			return i
		}
	}
	return -1
}

func (game *FreeCell) CardOnTable(card *Card) bool {
	return game.FindCardOnTable(card) != -1
}

func (game *FreeCell) RemoveFromTable(card *Card) {
	game.Table = util.RemoveAtIndex(game.Table, game.FindCardOnTable(card))
}

func (game *FreeCell) RemoveFromFreecell(card *Card) {
	for i := 0; i < len(game.FreeCells); i++ {
		if &game.FreeCells[i] == card {
			card.Suit = Blank
			card.X = int16(i * game.PlaceholderWidth)
			card.Y = int16(game.PlaceholderTop)
			break
		}
	}
}

func (game *FreeCell) SetSelectedCard(card *Card) {
	if card != nil {
		game.SelectedCard = card
		game.SelectedCard.Selected = true
	}
}

func (game *FreeCell) RemoveSelection() {
	if game.SelectedCard != nil {
		game.SelectedCard.Selected = false
		game.SelectedCard = nil
		game.AutoplayAllowed = true
	}
}

func (game *FreeCell) MoveCard(src, dst *Card) {
	rect := game.CardRect(dst)

	*dst = *src
	dst.Selected = false
	dst.X = int16(rect.X0)
	dst.Y = int16(rect.Y0)
}

func (game *FreeCell) DrawMenu() {
	defer trace.End(trace.Begin(""))

	game.Renderer.RenderSolidRectWH(0, 0, game.Width, game.MenuHeight, color.RGB(0xD4, 0xD0, 0xC8))
}

func (game *FreeCell) DrawBackground() {
	defer trace.End(trace.Begin(""))

	game.Renderer.Clear(color.RGB(0, 127, 0))

	for i := 0; i < len(game.FreeCells); i++ {
		x := i * game.PlaceholderWidth
		y := game.PlaceholderTop

		DrawRectWithShadow(game.Renderer, x, y, x+game.PlaceholderWidth-1, y+game.PlaceholderHeight-1, color.Black, color.Green)
		DrawRectWithShadow(game.Renderer, game.Width-game.PlaceholderWidth-x, y, game.Width-x-1, y+game.PlaceholderHeight-1, color.Black, color.Green)
	}

	DrawRectWithShadow(game.Renderer, 297, 38, 334, 75, color.Green, color.Black)
}

func (game *FreeCell) DrawFace() {
	defer trace.End(trace.Begin(""))

	const x = 298
	const y = 39
	const width = 36
	game.Renderer.RenderPixmap(game.Assets.Sub(320+int(width*game.FaceDirection), 453, 355+int(width*game.FaceDirection), 488), x, y)
}

func (game *FreeCell) DrawGiantFace() {
	defer trace.End(trace.Begin(""))

	game.Renderer.RenderPixmap(game.Assets.Sub(0, 453, 320, 773), 10, 126)
}

/* TODO(anton2929): store it with card? */
func (game *FreeCell) GetCardPixmap(card *Card) gr.Pixmap {
	defer trace.End(trace.Begin(""))

	const x = 632
	const y = 0

	i := int(card.Value - 1)
	j := int(card.Suit-1) + int(util.Bool2Int(card.Selected)*4)

	return game.Assets.Sub(x+i*CardWidth, y+j*CardHeight, x+(i+1)*CardWidth, y+(j+1)*CardHeight)
}

func (game *FreeCell) DrawCard(card *Card) {
	defer trace.End(trace.Begin(""))

	if card.Suit != Blank {
		game.Renderer.RenderPixmap(game.GetCardPixmap(card), int(card.X), int(card.Y))
	}
}

func (game *FreeCell) DrawCards() {
	defer trace.End(trace.Begin(""))

	const marginLeft = 7
	const marginTop = 126

	const paddingLeft = marginLeft
	const paddingTop = CardYPadding

	for i := 0; i < len(game.Table); i++ {
		game.DrawCard(&game.Table[i])
	}

	for i := 0; i < len(game.FreeCells); i++ {
		game.DrawCard(&game.FreeCells[i])
	}

	for i := 0; i < len(game.Goals); i++ {
		game.DrawCard(&game.Goals[i])
	}
}

func (game *FreeCell) DrawCursor() {
	defer trace.End(trace.Begin(""))

	switch game.Cursor {
	case CursorDefault:
		game.Window.ShowCursor()
	case CursorUp:
		game.Window.HideCursor()
		old := game.Assets.Alpha
		game.Assets.Alpha = gr.Alpha8bit
		game.Renderer.RenderPixmap(game.Assets.Sub(406, 453, 415, 472), game.UI.MouseX, game.UI.MouseY)
		game.Assets.Alpha = old
	case CursorDown:
		game.Window.HideCursor()
		old := game.Assets.Alpha
		game.Assets.Alpha = gr.Alpha8bit
		game.Renderer.RenderPixmap(game.Assets.Sub(392, 453, 406, 480), game.UI.MouseX, game.UI.MouseY-27)
		game.Assets.Alpha = old
	}
}

func (game *FreeCell) CardRect(card *Card) gui.Rect {
	defer trace.End(trace.Begin(""))

	return gui.Rect{int(card.X), int(card.Y), int(card.X) + CardWidth - 1, int(card.Y) + CardHeight - 1}
}

func (game *FreeCell) TableColumnRect(idx int) gui.Rect {
	defer trace.End(trace.Begin(""))

	return gui.Rect{game.TableLeft + idx*(game.TableLeft+CardWidth), game.TableTop, game.TableLeft + idx*(game.TableLeft+CardWidth) + CardWidth - 1, game.Window.Height - 1}
}

func (game *FreeCell) HandleFaceInput() {
	defer trace.End(trace.Begin(""))

	mouse := gui.Rect{game.UI.MouseX, game.UI.MouseY, game.UI.MouseX, game.UI.MouseY}

	/* Handle face turn. */
	faceLeftRect := gui.Rect{0, 20, 284, 116}
	faceRightRect := gui.Rect{348, 20, 632, 116}
	if faceLeftRect.Contains(mouse) {
		game.FaceDirection = 0
	} else if faceRightRect.Contains(mouse) {
		game.FaceDirection = 1
	}
}

func (game *FreeCell) HandleCardsInput() {
	defer trace.End(trace.Begin(""))

	mouse := gui.Rect{game.UI.MouseX, game.UI.MouseY, game.UI.MouseX, game.UI.MouseY}
	game.Cursor = CursorDefault

	for i := 0; i < len(game.FreeCells); i++ {
		freecell := &game.FreeCells[i]
		over := game.CardRect(freecell).Contains(mouse)
		pressed := game.UI.ButtonLogicDown(gui.ID(freecell), over)

		if over {
			if (pressed) && (game.SelectedCard == nil) && (freecell.Suit != Blank) {
				game.SetSelectedCard(freecell)
			} else if (pressed) && (game.SelectedCard == freecell) {
				game.RemoveSelection()
			} else if (game.SelectedCard != nil) && (freecell.Suit == Blank) {
				game.Cursor = CursorUp
				if pressed {
					game.MoveCard(game.SelectedCard, freecell)
					game.RemoveFromFreecell(game.SelectedCard)
					game.RemoveFromTable(game.SelectedCard)
					game.RemoveSelection()
				}
			}
		}
	}

	for i := 0; i < len(game.Goals); i++ {
		goal := &game.Goals[i]
		over := game.CardRect(goal).Contains(mouse)
		pressed := game.UI.ButtonLogicDown(gui.ID(goal), over)

		if (over) && (game.SelectedCard != nil) && (CanMove2Goal(game.SelectedCard, goal)) {
			game.Cursor = CursorUp
			if pressed {
				game.MoveCard(game.SelectedCard, goal)
				game.RemoveFromFreecell(game.SelectedCard)
				game.RemoveFromTable(game.SelectedCard)
				game.RemoveSelection()
			}
		}
	}

	for i := 0; i < game.TableColumns; i++ {
		columnRect := game.TableColumnRect(i)
		bottomCard := game.FindBottomCard(&Card{X: int16(columnRect.X0), Y: int16(columnRect.Y0)})

		overRect := columnRect
		if bottomCard != nil {
			overRect.Y1 = game.CardRect(bottomCard).Y1
		}
		over := overRect.Contains(mouse)
		pressed := game.UI.ButtonLogicDown(gui.ID(uintptr(game.TableLeft+i)), over)

		if (pressed) && (game.SelectedCard == nil) {
			game.SetSelectedCard(bottomCard)
		} else if (pressed) && (game.SelectedCard == bottomCard) {
			game.RemoveSelection()
		} else if (over) && (game.SelectedCard != nil) {
			if bottomCard != nil {
				if game.PowerMove(game.SelectedCard, bottomCard, pressed) {
					game.Cursor = CursorDown
					if pressed {
						game.RemoveSelection()
					}
				} else if CanMove(game.SelectedCard, bottomCard) {
					game.Cursor = CursorDown
					if pressed {
						game.SelectedCard.Selected = false
						game.SelectedCard.X = bottomCard.X
						game.SelectedCard.Y = bottomCard.Y + CardYPadding
						if !game.CardOnTable(game.SelectedCard) {
							game.Table = append(game.Table, *game.SelectedCard)
							game.RemoveFromFreecell(game.SelectedCard)
						}
						game.RemoveSelection()
					}
				}
			} else {
				if game.PowerMoveOnTable(game.SelectedCard, i, pressed) {
					game.Cursor = CursorDown
					if pressed {
						game.RemoveSelection()
					}
				} else {
					game.Cursor = CursorDown
					if pressed {
						game.SelectedCard.Selected = false
						game.SelectedCard.X = int16(columnRect.X0)
						game.SelectedCard.Y = int16(columnRect.Y0)
						if !game.CardOnTable(game.SelectedCard) {
							game.Table = append(game.Table, *game.SelectedCard)
							game.RemoveFromFreecell(game.SelectedCard)
						}
						game.RemoveSelection()
					}
				}
			}
		}
	}
}

func (game *FreeCell) RemoveCardIfUseless(card *Card) bool {
	defer trace.End(trace.Begin(""))

	if (card != nil) && (card.Suit != Blank) {
		useless := true

		for i := 0; i < len(game.Table); i++ {
			if (game.Table[i].Value >= Two) && (CanMove(&game.Table[i], card)) {
				useless = false
				break
			}
		}
		for i := 0; i < len(game.FreeCells); i++ {
			if (game.FreeCells[i].Value >= Two) && (CanMove(&game.FreeCells[i], card)) {
				useless = false
				break
			}
		}
		if useless {
			for i := 0; i < len(game.Goals); i++ {
				goal := &game.Goals[i]

				if CanMove2Goal(card, goal) {
					game.MoveCard(card, goal)
					game.RemoveFromFreecell(card)
					game.RemoveFromTable(card)
					return true
				}
			}
		}
	}

	return false
}

func (game *FreeCell) Autoplay() {
	defer trace.End(trace.Begin(""))

	removed := true && game.AutoplayAllowed
	for removed {
		removed = false
		for i := 0; i < game.TableColumns; i++ {
			columnRect := game.TableColumnRect(i)
			card := game.FindBottomCard(&Card{X: int16(columnRect.X0), Y: int16(columnRect.Y0)})
			removed = game.RemoveCardIfUseless(card) || removed
		}

		for i := 0; i < len(game.FreeCells); i++ {
			removed = game.RemoveCardIfUseless(&game.FreeCells[i]) || removed
		}
	}
}

func (game *FreeCell) SortCards() {
	defer trace.End(trace.Begin(""))

	for i := 1; i < len(game.Table); i++ {
		for j := 0; j < i; j++ {
			if game.Table[j].Y > game.Table[i].Y {
				game.Table[i], game.Table[j] = game.Table[j], game.Table[i]
			}
		}
	}
}

func (game *FreeCell) GameWon() bool {
	defer trace.End(trace.Begin(""))

	var kings int
	for i := 0; i < len(game.Goals); i++ {
		if game.Goals[i].Value == King {
			kings++
		}
	}
	return kings == 4
}

func (game *FreeCell) UpdateAndRender() {
	defer trace.End(trace.Begin(""))

	game.HandleFaceInput()

	if game.UI.MiddleDown {
		game.Table = game.Table[:0]

		for i := 0; i < len(game.FreeCells); i++ {
			freeCell := &game.FreeCells[i]
			freeCell.Suit = Blank
		}

		for i := 0; i < len(game.Goals); i++ {
			goal := &game.Goals[i]
			goal.Suit = SuitType(i) + 1
			goal.Value = King
		}
	}

	if game.State == GameRunning {
		game.HandleCardsInput()
		game.Autoplay()
		game.SortCards()

		if game.GameWon() {
			game.State = GameEnd
			game.Cursor = CursorDefault
			game.UI.ClearActive()
		}
	}

	game.DrawBackground()
	game.DrawCards()

	game.DrawFace()
	if game.State == GameEnd {
		game.DrawGiantFace()
	}
	game.DrawCursor()
	game.DrawMenu()
}
