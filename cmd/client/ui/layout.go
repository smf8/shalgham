package ui

import "github.com/jroimartin/gocui"

// Layout creates chat UI
func Layout(g *gocui.Gui) error {
	maxX, maxY := g.Size()
	g.Cursor = true

	if messages, err := g.SetView("messages", 0, 0, maxX-20, maxY-5); err != nil {
		if err != gocui.ErrUnknownView {
			return err
		}
		messages.Title = " messages: "
		messages.Autoscroll = true
		messages.Wrap = true
	}

	if input, err := g.SetView("input", 0, maxY-5, maxX-20, maxY-1); err != nil {
		if err != gocui.ErrUnknownView {
			return err
		}
		input.Title = " send: "
		input.Autoscroll = false
		input.Wrap = true
		input.Editable = true
	}

	if users, err := g.SetView("users", maxX-20, 0, maxX-1, maxY/2-1); err != nil {
		if err != gocui.ErrUnknownView {
			return err
		}
		users.Title = " users: "
		users.Autoscroll = false
		users.Wrap = true
	}

	if conversations, err := g.SetView("conversations", maxX-20, maxY/2-1, maxX-1, maxY-1); err != nil {
		if err != gocui.ErrUnknownView {
			return err
		}

		conversations.Title = " conversations: "
		conversations.Autoscroll = false
		conversations.Wrap = true
	}

	if login, err := g.SetView("login", maxX/2-20, maxY/2-1, maxX/2+20, maxY/2+1); err != nil {
		if err != gocui.ErrUnknownView {
			return err
		}
		g.SetCurrentView("login")
		login.Title = "username:password (login) "
		login.Autoscroll = false
		login.Wrap = true
		login.Editable = true
	}

	if signup, err := g.SetView("signup", maxX/2-20, maxY/2+1, maxX/2+20, maxY/2+3); err != nil {
		if err != gocui.ErrUnknownView {
			return err
		}

		signup.Title = "username:password (signup) "
		signup.Autoscroll = false
		signup.Wrap = true
		signup.Editable = true
	}
	return nil
}
