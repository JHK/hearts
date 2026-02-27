package natswire

import (
	"fmt"

	"github.com/JHK/hearts/internal/protocol"
)

func (p *ParticipantClient) StartRound() error {
	if p.playerID == "" {
		return fmt.Errorf("player id is required")
	}

	return p.StartRequest(protocol.StartRequest{PlayerID: p.playerID})
}

func (p *ParticipantClient) StartRequest(request protocol.StartRequest) error {
	var response protocol.CommandResponse
	if err := p.request(protocol.StartSubject(p.tableID), request, &response); err != nil {
		return err
	}
	if !response.Accepted {
		return fmt.Errorf("%s", response.Reason)
	}
	return nil
}

func (p *ParticipantClient) PlayCard(card string) error {
	if p.playerID == "" {
		return fmt.Errorf("player id is required")
	}

	return p.PlayRequest(protocol.PlayCardRequest{PlayerID: p.playerID, Card: card})
}

func (p *ParticipantClient) PlayRequest(request protocol.PlayCardRequest) error {
	var response protocol.CommandResponse
	if err := p.request(protocol.PlaySubject(p.tableID), request, &response); err != nil {
		return err
	}
	if !response.Accepted {
		return fmt.Errorf("%s", response.Reason)
	}
	return nil
}
