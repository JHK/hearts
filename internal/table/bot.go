package table

import (
	"fmt"
	"github.com/JHK/hearts/internal/protocol"
)

const botIDPrefix = "bot-"

func botName(seat int) string {
	return fmt.Sprintf("BOT %d", seat+1)
}

func (s *Service) fillBotsLocked() {
	for len(s.players) < seatsPerTable {
		seat := len(s.players)
		bot := &player{
			ID:   s.nextBotIDLocked(),
			Name: botName(seat),
			Seat: seat,
		}

		s.players = append(s.players, bot)
		s.playerByID[bot.ID] = bot

		s.publishEvent(protocol.EventPlayerJoined, protocol.PlayerJoinedData{
			Player: protocolPlayerInfo(bot),
		})
	}
}

func (s *Service) nextBotIDLocked() string {
	for i := 1; ; i++ {
		id := fmt.Sprintf("%s%d", botIDPrefix, i)
		if _, exists := s.playerByID[id]; !exists {
			return id
		}
	}
}
