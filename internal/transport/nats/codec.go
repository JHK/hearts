package natswire

import (
	"encoding/json"
	"fmt"

	"github.com/JHK/hearts/internal/protocol"
)

func marshalJSON(payload any) ([]byte, error) {
	data, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("marshal json: %w", err)
	}
	return data, nil
}

func unmarshalJSON(raw []byte, out any) error {
	if err := json.Unmarshal(raw, out); err != nil {
		return fmt.Errorf("unmarshal json: %w", err)
	}
	return nil
}

func decodeEvent(raw []byte) (protocol.Event, error) {
	var event protocol.Event
	if err := unmarshalJSON(raw, &event); err != nil {
		return protocol.Event{}, err
	}
	return event, nil
}

func decodeEventPayload[T any](event protocol.Event) (T, error) {
	var payload T
	if err := unmarshalJSON(event.Data, &payload); err != nil {
		return payload, err
	}
	return payload, nil
}
