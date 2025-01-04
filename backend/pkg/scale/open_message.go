package scale

import (
	"fmt"
	"time"

	"github.com/kotrzina/keg-scale/pkg/utils"
)

// shouldSendOpen applies the rules for sending a message when the pub is open
// we don't want to spam the group with messages
// it could happen for example when the scale is restarted or lost Wi-Fi connection for a while
func (s *Scale) shouldSendOpen() bool {
	// send message only once in 12 hours
	if time.Since(s.pub.openedAt) < 12*time.Hour {
		return false
	}

	// send message only if the pub was closed for at least 3 hours
	if time.Since(s.pub.closedAt) < 3*time.Hour {
		return false
	}

	return true
}

// sendWhatsAppOpen sends a message to the WhatsApp group when the pub is open.
// The call is asynchronous
func (s *Scale) sendWhatsAppOpen() {
	type openingMessageData struct {
		Weight      float64
		BeerLeft    int
		ActiveKeg   int
		ActiveKegAt time.Time

		IsOpen   bool
		OpenedAt time.Time

		Warehouse      map[int]int // keys 10, 15, 20, 30, 50
		WarehouseTotal int
	}

	params := openingMessageData{
		Weight:      s.weight,
		BeerLeft:    s.beersLeft,
		ActiveKeg:   s.activeKeg,
		ActiveKegAt: s.activeKegAt,

		IsOpen:   s.pub.isOpen,
		OpenedAt: s.pub.openedAt,

		Warehouse: map[int]int{
			10: s.warehouse[0],
			15: s.warehouse[1],
			20: s.warehouse[2],
			30: s.warehouse[3],
			50: s.warehouse[4],
		},
		WarehouseTotal: GetWarehouseBeersLeft(s.warehouse),
	}

	go func(data openingMessageData) {
		msg := "Pivo! 🍺"

		if data.ActiveKeg > 0 {
			msg += fmt.Sprintf(
				"\nMáme naraženou %dl bečku a zbývá v ní %d %s.",
				data.ActiveKeg,
				data.BeerLeft,
				utils.FormatBeer(data.BeerLeft),
			)
		}

		if data.WarehouseTotal > 0 {
			msg += fmt.Sprintf(
				"\nVe skladu máme %d %s.",
				data.WarehouseTotal,
				utils.FormatBeer(data.WarehouseTotal),
			)
		}

		err := s.whatsapp.SendText(s.config.WhatsAppOpenJid, msg)
		if err != nil {
			s.logger.Errorf("could not send Botka message: %v", err)
		}
	}(params)
}
