package main

import (
	"fmt"
)

import "github.com/mcuadros/go-syslog"

func main() {
	var channel syslog.LogPartsChannel
	handler := syslog.NewChannelHandler(channel)

	server := syslog.NewServer()
	server.SetFormat(syslog.RFC3164_NO_STRICT)
	server.SetHandler(handler)
	server.ListenUDP("0.0.0.0:514")
	server.Boot()

	go func(channel syslog.LogPartsChannel) {
		for logParts := range channel {
			fmt.Println(logParts)
		}
	}(channel)

	server.Wait()
}
