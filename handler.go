package syslog

import (
	"fmt"
)

import "github.com/jeromer/syslogparser"

type Handler interface {
	Handle(syslogparser.LogParts)
}

type LogPartsChannel chan syslogparser.LogParts

func NewChannelHandler(channel LogPartsChannel) *ChannelHandler {
	handler := new(ChannelHandler)
	handler.SetChannel(channel)

	return handler
}

type ChannelHandler struct {
	channel LogPartsChannel
}

func (self *ChannelHandler) SetChannel(channel LogPartsChannel) {
	self.channel = channel
}

func (self *ChannelHandler) Handle(logParts syslogparser.LogParts) {
	fmt.Println(logParts)
	self.channel <- logParts
}
