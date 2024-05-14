package interactives

import (
	"github.com/charmbracelet/bubbles/textinput"
)

func PostgresqlInitialModel() model {
	m := model{
		inputs: make([]textinput.Model, 4),
	}

	var t textinput.Model
	for i := range m.inputs {
		t = textinput.New()
		t.Cursor.Style = cursorStyle

		switch i {
		case 0:
			t.Placeholder = "Host"
			t.Focus()
			t.PromptStyle = focusedStyle
			t.TextStyle = focusedStyle
		case 1:
			t.Placeholder = "Database user"
		case 2:
			t.Placeholder = "Database name"
		case 3:
			t.Placeholder = "Password"
			t.EchoMode = textinput.EchoPassword
			t.EchoCharacter = '•'
		}

		m.inputs[i] = t
	}

	return m
}
