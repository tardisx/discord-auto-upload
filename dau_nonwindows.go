//go:build !windows

package main

import "github.com/tardisx/discord-auto-upload/config"

func mainloop(c *config.ConfigService) {

	ch := make(chan bool)
	<-ch
}
