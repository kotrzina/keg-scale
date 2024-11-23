package hook

import (
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/kotrzina/keg-scale/pkg/config"
	"github.com/kotrzina/keg-scale/pkg/utils"
	"github.com/kotrzina/keg-scale/pkg/wa"
	"github.com/sirupsen/logrus"
)

// Botka is a struct that represents the Botka bot
// Mr. Botka is responsible for sending messages to the WhatsApp group
// also receives messages from the group and reacts to them
type Botka struct {
	whatsapp *wa.WhatsAppClient
	brain    *BotkaBrain
	config   *config.Config

	mtx    sync.RWMutex
	logger *logrus.Logger
}

type BotkaBrain struct {
	Weight      float64
	BeerLeft    int
	ActiveKeg   int
	ActiveKegAt time.Time

	Warehouse      map[int]int // keys 10, 15, 20, 30, 50
	WarehouseTotal int
}

func NewBotka(client *wa.WhatsAppClient, conf *config.Config, logger *logrus.Logger) *Botka {
	w := &Botka{
		whatsapp: client,
		brain:    &BotkaBrain{},
		config:   conf,

		mtx:    sync.RWMutex{},
		logger: logger,
	}

	if !conf.Debug {
		// replies only on production
		client.RegisterEventHandler(w.helpHandler())
		client.RegisterEventHandler(w.kegHandler())
		client.RegisterEventHandler(w.pricesHandler())
		client.RegisterEventHandler(w.warehouseHandler())
	}

	return w
}

// UpdateBotkaBrain updates the Botka's brain with the new data
// it is initialized by Scale with every significant change
func (b *Botka) UpdateBotkaBrain(bb *BotkaBrain) {
	b.mtx.Lock()
	defer b.mtx.Unlock()

	b.brain = bb
}

func (b *Botka) SendOpen() {
	go func() {
		b.mtx.RLock()
		defer b.mtx.RUnlock()

		msg := "Pivo! 🍺"

		if b.brain.ActiveKeg > 0 {
			msg += fmt.Sprintf(
				"\nMáme naraženou %dl bečku a zbývá v ní %d %s.",
				b.brain.ActiveKeg,
				b.brain.BeerLeft,
				utils.FormatBeer(b.brain.BeerLeft),
			)
		}

		if b.brain.WarehouseTotal > 0 {
			msg += fmt.Sprintf(
				"\nVe skladu máme %d %s.",
				b.brain.WarehouseTotal,
				utils.FormatBeer(b.brain.WarehouseTotal),
			)
		}

		err := b.whatsapp.SendText(b.config.WhatsAppOpenJid, msg)
		if err != nil {
			b.logger.Errorf("could not send Botka message: %v", err)
		}
	}()
}

func (b *Botka) helpHandler() wa.EventHandler {
	return wa.EventHandler{
		MatchFunc: func(msg string) bool {
			return strings.HasPrefix(sanitizeCommand(msg), "help") ||
				strings.HasPrefix(sanitizeCommand(msg), "napoveda") ||
				strings.HasPrefix(sanitizeCommand(msg), "pomoc")
		},
		Handler: func(from, _ string) error {
			reply := "Příkazy: \n" +
				"/help - zobrazí nápovědu \n" +
				"/becka - informace o aktuální bečce \n" +
				"/cenik - ceník \n" +
				"/sklad - stav skladu"
			err := b.whatsapp.SendText(from, reply)
			return err
		},
	}
}

func (b *Botka) kegHandler() wa.EventHandler {
	return wa.EventHandler{
		MatchFunc: func(msg string) bool {
			return strings.HasPrefix(sanitizeCommand(msg), "becka")
		},
		Handler: func(from, _ string) error {
			msg := fmt.Sprintf(
				"Máme naraženou %dl bečku a zbývá v ní %d %s. Naražena byla %s v %s.",
				b.brain.ActiveKeg,
				b.brain.BeerLeft,
				utils.FormatBeer(b.brain.BeerLeft),
				utils.FormatDateShort(b.brain.ActiveKegAt),
				utils.FormatTime(b.brain.ActiveKegAt),
			)
			err := b.whatsapp.SendText(from, msg)
			return err
		},
	}
}

func (b *Botka) pricesHandler() wa.EventHandler {
	return wa.EventHandler{
		MatchFunc: func(msg string) bool {
			return strings.HasPrefix(sanitizeCommand(msg), "cenik")
		},
		Handler: func(from, _ string) error {
			reply := "Ceník: \n" +
				"- Vše 25 Kč \n" +
				"- Víno 130 Kč"
			err := b.whatsapp.SendText(from, reply)
			return err
		},
	}
}

func (b *Botka) warehouseHandler() wa.EventHandler {
	return wa.EventHandler{
		MatchFunc: func(msg string) bool {
			return strings.HasPrefix(sanitizeCommand(msg), "sklad")
		},
		Handler: func(from, _ string) error {
			b.mtx.RLock()
			defer b.mtx.RUnlock()

			reply := fmt.Sprintf("Ve skladu máme celkem %d piv.", b.brain.WarehouseTotal)
			if b.brain.Warehouse[10] > 0 {
				reply += fmt.Sprintf("\n%d x 10l", b.brain.Warehouse[10])
			}
			if b.brain.Warehouse[15] > 0 {
				reply += fmt.Sprintf("\n%d x 15l", b.brain.Warehouse[15])
			}
			if b.brain.Warehouse[20] > 0 {
				reply += fmt.Sprintf("\n%d x 20l", b.brain.Warehouse[20])
			}
			if b.brain.Warehouse[30] > 0 {
				reply += fmt.Sprintf("\n%d x 30l", b.brain.Warehouse[30])
			}
			if b.brain.Warehouse[50] > 0 {
				reply += fmt.Sprintf("\n%d x 50l", b.brain.Warehouse[50])
			}

			err := b.whatsapp.SendText(from, reply)
			return err
		},
	}
}

func sanitizeCommand(command string) string {
	return strings.ToLower(strings.TrimSpace(strings.TrimPrefix(command, "/")))
}
