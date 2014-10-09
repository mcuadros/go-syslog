package syslog

import (
	"github.com/jeromer/syslogparser"
)

//The handler receive every syslog entry at Handle method
type Handler interface {
	Handle(syslogparser.LogParts)
}

type LogPartsChannel chan syslogparser.LogParts

//The ChannelHandler will send all the syslog entries into the given channel
type ChannelHandler struct {
	channel LogPartsChannel
}

//NewChannelHandler returns a new ChannelHandler
func NewChannelHandler(channel LogPartsChannel) *ChannelHandler {
	handler := new(ChannelHandler)
	handler.SetChannel(channel)

	return handler
}

//The channel to be used
func (self *ChannelHandler) SetChannel(channel LogPartsChannel) {
	self.channel = channel
}

//Syslog entry receiver
func (self *ChannelHandler) Handle(logParts syslogparser.LogParts) {
	self.channel <- logParts
}
