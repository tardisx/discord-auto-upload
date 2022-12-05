//go:build darwin || linux
// +build darwin linux

package main

import "github.com/tardisx/discord-auto-upload/config"

func mainloop(c *config.ConfigService) {

	ch := make(chan bool)
	<-ch
}
