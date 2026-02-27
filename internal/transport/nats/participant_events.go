package natswire

import (
	"github.com/JHK/hearts/internal/protocol"
)

func (p *ParticipantClient) handlePublicEvent(raw []byte) {
	event, err := decodeEvent(raw)
	if err != nil {
		p.emitDecodeError(err)
		return
	}

	handlers := p.snapshotHandlers()
	if handlers == nil {
		return
	}

	switch event.Type {
	case protocol.EventPlayerJoined:
		payload, err := decodeEventPayload[protocol.PlayerJoinedData](event)
		if err != nil {
			p.emitDecodeError(err)
			return
		}
		if handlers.OnPlayerJoined != nil {
			handlers.OnPlayerJoined(payload)
		}
	case protocol.EventGameStarted:
		if handlers.OnGameStarted != nil {
			handlers.OnGameStarted()
		}
	case protocol.EventTurnChanged:
		payload, err := decodeEventPayload[protocol.TurnChangedData](event)
		if err != nil {
			p.emitDecodeError(err)
			return
		}
		if handlers.OnTurnChanged != nil {
			handlers.OnTurnChanged(payload)
		}
	case protocol.EventCardPlayed:
		payload, err := decodeEventPayload[protocol.CardPlayedData](event)
		if err != nil {
			p.emitDecodeError(err)
			return
		}
		if handlers.OnCardPlayed != nil {
			handlers.OnCardPlayed(payload)
		}
	case protocol.EventTrickCompleted:
		payload, err := decodeEventPayload[protocol.TrickCompletedData](event)
		if err != nil {
			p.emitDecodeError(err)
			return
		}
		if handlers.OnTrickCompleted != nil {
			handlers.OnTrickCompleted(payload)
		}
	case protocol.EventRoundCompleted:
		payload, err := decodeEventPayload[protocol.RoundCompletedData](event)
		if err != nil {
			p.emitDecodeError(err)
			return
		}
		if handlers.OnRoundCompleted != nil {
			handlers.OnRoundCompleted(payload)
		}
	}
}

func (p *ParticipantClient) handlePrivateEvent(subject string, raw []byte) {
	playerID, ok := p.playerIDFromPrivateSubject(subject)
	if !ok {
		return
	}

	event, err := decodeEvent(raw)
	if err != nil {
		p.emitDecodeError(err)
		return
	}

	handlers := p.snapshotHandlers()
	if handlers == nil {
		return
	}

	switch event.Type {
	case protocol.EventHandUpdated:
		payload, err := decodeEventPayload[protocol.HandUpdatedData](event)
		if err != nil {
			p.emitDecodeError(err)
			return
		}
		if handlers.OnHandUpdated != nil {
			handlers.OnHandUpdated(playerID, payload)
		}
	case protocol.EventYourTurn:
		payload, err := decodeEventPayload[protocol.YourTurnData](event)
		if err != nil {
			p.emitDecodeError(err)
			return
		}
		if handlers.OnYourTurn != nil {
			handlers.OnYourTurn(playerID, payload)
		}
	}
}
