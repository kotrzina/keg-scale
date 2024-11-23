package hook

import (
	"fmt"
	"strings"
	"sync"

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

	mtx    sync.Mutex
	logger *logrus.Logger
}

type BotkaBrain struct {
	Weight    float64
	BeerLeft  int
	ActiveKeg int

	Warehouse      map[int]int // keys 10, 15, 20, 30, 50
	WarehouseTotal int
}

func NewBotka(client *wa.WhatsAppClient, conf *config.Config, logger *logrus.Logger) *Botka {
	w := &Botka{
		whatsapp: client,
		brain:    &BotkaBrain{},
		config:   conf,

		mtx:    sync.Mutex{},
		logger: logger,
	}

	if !conf.Debug {
		// replies only on production
		client.RegisterEventHandler(w.helpHandler())
		client.RegisterEventHandler(w.pricesHandler())
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
			return strings.HasPrefix(sanitizeCommand(msg), "help")
		},
		Handler: func(from, _ string) error {
			reply := "Příkazy: \n" +
				"/help - zobrazí nápovědu \n" +
				"/cenik - ceník"
			err := b.whatsapp.SendText(from, reply)
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

func sanitizeCommand(command string) string {
	return strings.ToLower(strings.TrimSpace(strings.TrimPrefix(command, "/")))
}
