package ui

import (
	"github.com/jroimartin/gocui"
	"github.com/sirupsen/logrus"
	"github.com/smf8/shalgham/client"
)

func ShowUI(c *client.Client) {
	g, err := gocui.NewGui(gocui.OutputNormal)
	if err != nil {
		logrus.Fatal(err)
	}

	defer g.Close()

	c.SetUI(g)
	g.SetManagerFunc(Layout)
	g.SetKeybinding("login", gocui.KeyEnter, gocui.ModNone, c.Login)
	g.SetKeybinding("login", gocui.KeyTab, gocui.ModNone, switchLogin)
	g.SetKeybinding("signup", gocui.KeyEnter, gocui.ModNone, c.Signup)
	g.SetKeybinding("signup", gocui.KeyTab, gocui.ModNone, switchLogin)
	g.SetKeybinding("input", gocui.KeyEnter, gocui.ModNone, c.SubmitInput)
	g.SetKeybinding("", gocui.KeyCtrlC, gocui.ModNone, c.Disconnect)
	g.MainLoop()
}

func switchLogin(g *gocui.Gui, v *gocui.View) error {
	nextView := "login"

	if v.Name() == "login" {
		nextView = "signup"
	}

	if _, err := setCurrentViewOnTop(g, nextView); err != nil {
		return err
	}

	g.Cursor = true

	return nil
}

func setCurrentViewOnTop(g *gocui.Gui, name string) (*gocui.View, error) {
	if _, err := g.SetCurrentView(name); err != nil {
		return nil, err
	}
	return g.SetViewOnTop(name)
}
