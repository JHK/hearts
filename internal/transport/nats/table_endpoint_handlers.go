package natswire

import (
	"github.com/JHK/hearts/internal/protocol"
	"github.com/nats-io/nats.go"
)

func (e *TableEndpoint) handleDiscover(msg *nats.Msg) {
	if msg.Reply == "" || e.handlers.OnDiscover == nil {
		return
	}

	if len(msg.Data) > 0 {
		var request protocol.DiscoverRequest
		if err := unmarshalJSON(msg.Data, &request); err != nil {
			return
		}
	}

	e.reply(msg, e.handlers.OnDiscover())
}

func (e *TableEndpoint) handleJoin(msg *nats.Msg) {
	if msg.Reply == "" {
		return
	}

	var request protocol.JoinRequest
	if err := unmarshalJSON(msg.Data, &request); err != nil {
		e.reply(msg, protocol.JoinResponse{Accepted: false, Reason: "invalid join request"})
		return
	}

	if e.handlers.OnJoin == nil {
		e.reply(msg, protocol.JoinResponse{Accepted: false, Reason: "join handler is not configured"})
		return
	}

	e.reply(msg, e.handlers.OnJoin(request))
}

func (e *TableEndpoint) handleStart(msg *nats.Msg) {
	if msg.Reply == "" {
		return
	}

	var request protocol.StartRequest
	if err := unmarshalJSON(msg.Data, &request); err != nil {
		e.reply(msg, protocol.CommandResponse{Accepted: false, Reason: "invalid start request"})
		return
	}

	if e.handlers.OnStart == nil {
		e.reply(msg, protocol.CommandResponse{Accepted: false, Reason: "start handler is not configured"})
		return
	}

	e.reply(msg, e.handlers.OnStart(request))
}

func (e *TableEndpoint) handlePlay(msg *nats.Msg) {
	if msg.Reply == "" {
		return
	}

	var request protocol.PlayCardRequest
	if err := unmarshalJSON(msg.Data, &request); err != nil {
		e.reply(msg, protocol.CommandResponse{Accepted: false, Reason: "invalid play request"})
		return
	}

	if e.handlers.OnPlay == nil {
		e.reply(msg, protocol.CommandResponse{Accepted: false, Reason: "play handler is not configured"})
		return
	}

	e.reply(msg, e.handlers.OnPlay(request))
}

func (e *TableEndpoint) reply(msg *nats.Msg, payload any) {
	if msg.Reply == "" || e.nc == nil {
		return
	}

	data, err := marshalJSON(payload)
	if err != nil {
		return
	}

	_ = e.nc.Publish(msg.Reply, data)
}
