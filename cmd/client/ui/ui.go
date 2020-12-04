package ui

import (
	"fmt"

	"github.com/jroimartin/gocui"
	"github.com/sirupsen/logrus"
	"github.com/smf8/shalgham/client"
	"github.com/smf8/shalgham/model"
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
	//g.SetKeybinding("input", gocui.KeyEnter, gocui.ModNone, client.Send)
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

func WriteMessage(msg model.Message, v *gocui.View) {
	fmt.Fprintf(v, "\u001B[3%d;%dm[%s]\u001B[0m  \u001B[3%d;%dm%s\u001B[0m: %s\n", 3, 1, msg.CreatedAT, 2, 7, msg.Author)
}
