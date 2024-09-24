package controls

import (
	input "github.com/quasilyte/ebitengine-input"
)

const (
	ActionNone input.Action = iota

	ActionMoveRight
	ActionMoveDown
	ActionMoveLeft
	ActionMoveUp

	ActionConfirm
	ActionRestart
)
