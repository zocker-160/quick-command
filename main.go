package main

import (
	"fmt"

	"github.com/gdamore/tcell/v2"
	"github.com/lithammer/fuzzysearch/fuzzy"
	"github.com/mattn/go-runewidth"

	"github.com/zocker-160/quick-command/utils"
	//"github.com/rivo/tview"
	//"github.com/jesseduffield/gocui"
)

const VERSION = "0.1"

var screen tcell.Screen

var styleActive = tcell.StyleDefault.Foreground(tcell.ColorGreen)
var stylePassive = tcell.StyleDefault

var searchText string
var cursorPosition int

type OMode int

var overlayMode OMode = overlayNone
var overlayTextNameMode bool = true
var overlayTextName string
var overlayTextCommand string

const (
	overlayNone OMode = iota
	overlayAdd
	overlayEdit
	overlayDelete
)

var config *utils.Config
var filteredList []*utils.ListEntry
var listIndex int

func initScreen() error {
	s, err := tcell.NewScreen()
	if err != nil {
		return err
	}

	err = s.Init()
	if err != nil {
		return err
	}

	s.Clear()

	screen = s

	return nil
}

func draw() {
	screen.HideCursor()

	if w, h := screen.Size(); w < 50 || h < 7 {
		drawNotEnoughSpace()
		return
	}

	drawSearch()
	drawList()
	drawFooter()

	if overlayMode == overlayAdd || overlayMode == overlayEdit {
		drawInputDialog()
	}

	if overlayMode == overlayDelete {
		drawDeleteDialog()
	}
}

func handleKey(ev *tcell.EventKey) {
	if overlayMode != overlayNone {
		switch ev.Key() {
		case tcell.KeyEsc:
			overlayMode = overlayNone

		case tcell.KeyEnter:
			if overlayMode == overlayAdd {
				entry := utils.ListEntry{
					Name:    overlayTextName,
					Command: overlayTextCommand,
				}
				config.AddEntry(entry)
				resetInputMode()
			}
			if overlayMode == overlayDelete {
				config.RemoveEntry(filteredList[listIndex])
				if l := len(filteredList); l > 0 && listIndex >= l-1 {
					listIndex--
				}

				resetInputMode()
			}
			if overlayMode == overlayEdit {
				e := filteredList[listIndex]
				e.Name = overlayTextName
				e.Command = overlayTextCommand
				config.Save()
				resetInputMode()
			}

		case tcell.KeyRune:
			if overlayTextNameMode {
				overlayTextName += string(ev.Rune())
			} else {
				overlayTextCommand += string(ev.Rune())
			}

		case tcell.KeyDEL:
			if overlayTextNameMode {
				overlayTextName = utils.RemoveLastRune(overlayTextName)
			} else {
				overlayTextCommand = utils.RemoveLastRune(overlayTextCommand)
			}

		case tcell.KeyTab:
			overlayTextNameMode = !overlayTextNameMode

		}
		return
	}

	switch ev.Key() {
	case tcell.KeyRune:
		searchText += string(ev.Rune())
		cursorPosition++
		listIndex = 0

	case tcell.KeyDEL:
		if ev.Modifiers()&tcell.ModAlt != 0 {
			searchText = ""
			cursorPosition = 0
		} else {
			searchText = utils.RemoveLastRune(searchText)
			if cursorPosition > 0 {
				cursorPosition--
			}
		}

	case tcell.KeyDown:
		if listIndex < len(filteredList)-1 {
			listIndex++
		}

	case tcell.KeyUp:
		if listIndex > 0 {
			listIndex--
		}

	case tcell.KeyCtrlN:
		overlayMode = overlayAdd

	case tcell.KeyCtrlD:
		overlayMode = overlayDelete
		overlayTextName = ""
		overlayTextCommand = ""

	case tcell.KeyCtrlE:
		overlayMode = overlayEdit
		e := filteredList[listIndex]
		overlayTextName = e.Name
		overlayTextCommand = e.Command

	case tcell.KeyEnter:
		if len(filteredList) == 0 {
			break
		}

		event := new(utils.ExecuteEvent)
		event.SetEventNow()
		event.Command = filteredList[listIndex].Command

		screen.PostEvent(event)
	}
}

func drawNotEnoughSpace() {
	screen.Fill(' ', styleActive)
	drawText("not enough space", 0, 0, false, styleActive)
}

func drawSearch() {
	w, _ := screen.Size()

	s := styleActive
	if overlayMode != overlayNone {
		s = stylePassive
	}

	drawBorders(0, 0, w, 3, s)
	drawText("Filter", 2, 0, false, s)
	drawText("<Alt-Back to clear>", w-2, 2, true, s)

	clearLine(1, 1, w-1, stylePassive)
	drawText(searchText, 1, 1, false, stylePassive)
	screen.ShowCursor(cursorPosition+1, 1)
}

func drawList() {
	w, h := screen.Size()
	hl := w / 3

	drawBorders(0, 3, w, h-1, stylePassive)
	drawText("Name", 2, 3, false, stylePassive)
	drawText("Command", hl+2, 3, false, stylePassive)

	for i := 4; i < h-2; i++ {
		clearLine(1, i, w-1, stylePassive)
		screen.SetContent(hl, i, tcell.RuneVLine, nil, stylePassive)
	}

	applyListFilter()

	lCap := h - 6

	sI := listIndex - (lCap - 1)
	if sI < 0 {
		sI = 0
	}

	for i := 0; i < min(lCap, len(filteredList)); i++ {
		s := stylePassive
		y := 4 + i
		vI := i + sI

		if vI == listIndex {
			s = tcell.StyleDefault.Background(tcell.NewHexColor(0x424242))
			clearLine(1, y, w-1, s)
			screen.SetContent(hl, y, tcell.RuneVLine, nil, s)
		}
		//drawText(fmt.Sprintf("%d", vI), 1, y, false, s)
		drawText(filteredList[vI].Name, 1, y, false, s)
		drawText(filteredList[vI].Command, hl+1, y, false, s)
	}

	//t := fmt.Sprintf(" %d %d %d ", listIndex, lCap, sI)
	t := fmt.Sprintf(" %d of %d ", listIndex+1, len(filteredList))
	drawText(t, w-2, 3, true, stylePassive)
}

func drawFooter() {
	w, h := screen.Size()

	s := tcell.StyleDefault.Foreground(tcell.ColorLightBlue)
	//s := tcell.StyleDefault.Foreground(tcell.ColorBlack).Background(tcell.ColorWhiteSmoke)
	footer := "<Ctrl-q>: quit, <up/down>: selection, <Ctrl-n>: new, <Ctrl-d>: delete, <Ctrl-e>: edit, <enter>: execute"

	if overlayMode != overlayNone {
		footer = "<ESC>: cancel, <enter>: confirm"
	}

	clearLine(0, h-1, w, stylePassive)
	drawText(footer, 1, h-1, false, s)

	v := fmt.Sprintf("%s v%s", "quick-command", VERSION)
	drawText(v, w-1, h-1, true, s.Foreground(tcell.ColorGreen))
}

func drawInputDialog() {
	w, h := screen.Size()

	// TODO fixed width?

	x1 := w / 4
	x2 := w - x1
	y := h/2 - 1

	sN := stylePassive
	sD := styleActive
	if overlayTextNameMode {
		sN = styleActive
		sD = stylePassive
	}

	clearLine(x1, y, x2, stylePassive)
	drawBorders(x1-1, y-1, x2+1, y+2, sN)
	drawText("Name", x1+1, y-1, false, sN)

	clearLine(x1, y+3, x2, stylePassive)
	drawBorders(x1-1, y+2, x2+1, y+5, sD)
	drawText("Command", x1+1, y+2, false, sD)

	drawText("press <tab> to switch", x2, y+5, true, styleActive)

	drawText(overlayTextName, x1, y, false, stylePassive)
	drawText(overlayTextCommand, x1, y+3, false, stylePassive)

	if overlayTextNameMode {
		screen.ShowCursor(x1+runewidth.StringWidth(overlayTextName), y)
	} else {
		screen.ShowCursor(x1+runewidth.StringWidth(overlayTextCommand), y+3)
	}
}

func drawDeleteDialog() {
	w, h := screen.Size()

	// TODO fixed width?

	x1 := w / 4
	x2 := w - x1
	y := h/2 - 1
	s := styleActive.Foreground(tcell.ColorRed)

	clearLine(x1, y, x2, stylePassive)
	drawBorders(x1-1, y-1, x2+1, y+2, s)
	drawText("Confirm delete", x1+1, y-1, false, s)
	drawText(filteredList[listIndex].Name, x1, y, false, stylePassive)
	screen.HideCursor()
}

func resetInputMode() {
	overlayTextName = ""
	overlayTextCommand = ""
	overlayTextNameMode = true
	overlayMode = overlayNone
}

func run() error {
	defer screen.Fini()

	for {
		draw()
		screen.Show()

		event := screen.PollEvent()

		switch ev := event.(type) {
		case *tcell.EventKey:
			if ev.Key() == tcell.KeyCtrlQ {
				return nil
			} else {
				handleKey(ev)
			}

		case *tcell.EventResize:
			draw()
			screen.Sync()

		case *utils.ExecuteEvent:
			screen.Fini()

			fmt.Println(">", ev.Command)

			cmd, err := utils.GenerateCommand(ev.Command)
			if err != nil {
				return err
			}
			cmd.Run()
			return nil
		}
	}
}

func drawText(text string, x, y int, alignRight bool, style tcell.Style) {
	w := 0
	for _, ch := range text {
		p := x + w
		if alignRight {
			p = x - runewidth.StringWidth(text) + w
		}
		screen.SetContent(p, y, ch, nil, style)
		w += runewidth.RuneWidth(ch)
	}
}

func drawBorders(x0, y0, x1, y1 int, style tcell.Style) {
	x1--
	y1--

	for x := x0 + 1; x < x1; x++ {
		screen.SetContent(x, y0, tcell.RuneHLine, nil, style)
		screen.SetContent(x, y1, tcell.RuneHLine, nil, style)
	}

	for y := y0 + 1; y < y1; y++ {
		screen.SetContent(x0, y, tcell.RuneVLine, nil, style)
		screen.SetContent(x1, y, tcell.RuneVLine, nil, style)
	}

	// corners
	if x1 > x0 && y1 > y0 {
		screen.SetContent(x0, y0, tcell.RuneULCorner, nil, style)
		screen.SetContent(x1, y0, tcell.RuneURCorner, nil, style)
		screen.SetContent(x0, y1, tcell.RuneLLCorner, nil, style)
		screen.SetContent(x1, y1, tcell.RuneLRCorner, nil, style)
	}
}

func clearLine(x0, y, x1 int, style tcell.Style) {
	for i := x0; i < x1; i++ {
		screen.SetContent(i, y, ' ', nil, style)
	}
}

func applyListFilter() {
	list := config.GetList()
	filteredList = make([]*utils.ListEntry, 0, len(list))

	for i := range list {
		e := &list[i]

		if len(searchText) == 0 || fuzzy.MatchNormalizedFold(searchText, e.Name) {
			filteredList = append(filteredList, e)
		}
	}
}

func main() {
	fmt.Println("Starting quick-command")

	fmt.Println("Loading config")
	c, err := utils.NewConfig()
	if err != nil {
		panic(err)
	}
	config = c

	err = initScreen()
	if err != nil {
		panic(err)
	}

	/*
	config.AddEntry(utils.ListEntry{Name: "Midnight Commander", Command: "mc"})
	config.AddEntry(utils.ListEntry{Name: "Terminal", Command: "konsole"})
	config.AddEntry(utils.ListEntry{Name: "Neofetch", Command: "neofetch"})
	config.AddEntry(utils.ListEntry{Name: "test", Command: "notify-send test"})
	for i := range 50 {
		config.AddEntry(utils.ListEntry{
			Name: strconv.Itoa(i),
			Command: fmt.Sprintf("commmand from %d", i),
		})
	}
	*/

	err = run()
	if err != nil {
		panic(err)
	}
}
