package utils

import "github.com/gdamore/tcell/v2"

type ExecuteEvent struct {
	tcell.EventTime
	Command string
}
