package ui

import (
	"github.com/jroimartin/gocui"
	"github.com/sirupsen/logrus"
	"github.com/smf8/shalgham/client"
)

func ShowUI() {
	g, err := gocui.NewGui(gocui.OutputNormal)
	if err != nil {
		logrus.Fatal(err)
	}

	defer g.Close()

	g.SetManagerFunc(Layout)
	g.SetKeybinding("login", gocui.KeyEnter, gocui.ModNone, client.Login)
	//g.SetKeybinding("input", gocui.KeyEnter, gocui.ModNone, client.Send)
	//g.SetKeybinding("", gocui.KeyCtrlC, gocui.ModNone, client.Disconnect)
	g.MainLoop()
}
