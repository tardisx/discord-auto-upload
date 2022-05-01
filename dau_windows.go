package main

import (
	"fmt"
	"github.com/getlantern/systray"
	"github.com/skratchdot/open-golang/open"
	"github.com/tardisx/discord-auto-upload/config"
	daulog "github.com/tardisx/discord-auto-upload/log"
	"github.com/tardisx/discord-auto-upload/version"
)

func mainloop(c *config.ConfigService) {
	systray.Run(func() { onReady(c) }, onExit)
}

func onReady(c *config.ConfigService) {

	systray.SetIcon(appIcon)
	//systray.SetTitle("DAU")
	systray.SetTooltip(fmt.Sprintf("discord-auto-upload %s", version.CurrentVersion))
	openApp := systray.AddMenuItem("Open", "Open in web browser")
	gh := systray.AddMenuItem("Github", "Open project page")
	ghr := systray.AddMenuItem("Release Notes", "Open project release notes")
	quit := systray.AddMenuItem("Quit", "Quit")

	go func() {
		<-quit.ClickedCh
		systray.Quit()
	}()

	go func() {
		for {
			select {
			case <-openApp.ClickedCh:
				openWebBrowser(c.Config.Port)
			case <-gh.ClickedCh:
				open.Start("https://github.com/tardisx/discord-auto-upload")
			case <-ghr.ClickedCh:
				open.Start(fmt.Sprintf("https://github.com/tardisx/discord-auto-upload/releases/tag/%s", version.CurrentVersion))
			}
		}
	}()

	// Sets the icon of a menu item. Only available on Mac and Windows.
	// mQuit.SetIcon(icon.Data)
}

func onExit() {
	// clean up here
	daulog.Info("quitting on user request")
}
