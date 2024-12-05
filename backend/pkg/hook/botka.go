package hook

import (
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/kotrzina/keg-scale/pkg/ai"
	"github.com/kotrzina/keg-scale/pkg/config"
	"github.com/kotrzina/keg-scale/pkg/scale"
	"github.com/kotrzina/keg-scale/pkg/store"
	"github.com/kotrzina/keg-scale/pkg/utils"
	"github.com/kotrzina/keg-scale/pkg/wa"
	"github.com/kozaktomas/diacritics"
	"github.com/sirupsen/logrus"
)

// Botka is a struct that represents the Botka bot
// Mr. Botka is responsible for sending messages to the WhatsApp group
// also receives messages from the group and reacts to them
type Botka struct {
	whatsapp *wa.WhatsAppClient
	scale    *scale.Scale
	ai       *ai.Ai
	config   *config.Config
	storage  store.Storage

	mtx    sync.RWMutex
	logger *logrus.Logger
}

type BotkaBrain struct {
	Weight      float64
	BeerLeft    int
	ActiveKeg   int
	ActiveKegAt time.Time

	IsOpen   bool
	OpenedAt time.Time

	Warehouse      map[int]int // keys 10, 15, 20, 30, 50
	WarehouseTotal int
}

func NewBotka(
	client *wa.WhatsAppClient,
	kegScale *scale.Scale,
	intelligence *ai.Ai,
	conf *config.Config,
	storage store.Storage,
	logger *logrus.Logger,
) *Botka {
	w := &Botka{
		whatsapp: client,
		scale:    kegScale,
		ai:       intelligence,
		config:   conf,
		storage:  storage,

		mtx:    sync.RWMutex{},
		logger: logger,
	}

	if !conf.Debug {
		// replies only on production
		client.RegisterEventHandler(w.helpHandler())
		client.RegisterEventHandler(w.helloHandler())
		client.RegisterEventHandler(w.pubHandler())
		client.RegisterEventHandler(w.kegHandler())
		client.RegisterEventHandler(w.pricesHandler())
		client.RegisterEventHandler(w.warehouseHandler())
		client.RegisterEventHandler(w.aiHandler())
	}

	return w
}

func (b *Botka) helpHandler() wa.EventHandler {
	return wa.EventHandler{
		MatchFunc: func(msg string) bool {
			sanitized := b.sanitizeCommand(msg)
			return strings.HasPrefix(sanitized, "help") ||
				strings.HasPrefix(sanitized, "napoveda") ||
				strings.HasPrefix(sanitized, "pomoc")
		},
		HandleFunc: func(from, _ string) error {
			reply := "Příkazy: \n" +
				"/help - zobrazí nápovědu \n" +
				"/pub /hospoda - informace o hospodě \n" +
				"/becka - informace o aktuální bečce \n" +
				"/cenik - ceník \n" +
				"/sklad - stav skladu\n"
			err := b.whatsapp.SendText(from, reply)
			return err
		},
	}
}

func (b *Botka) helloHandler() wa.EventHandler {
	return wa.EventHandler{
		MatchFunc: func(msg string) bool {
			sanitized := b.sanitizeCommand(msg)
			return strings.HasPrefix(sanitized, "hello") ||
				strings.HasPrefix(sanitized, "hi") ||
				strings.HasPrefix(sanitized, "ahoj") ||
				strings.HasPrefix(sanitized, "zdar") ||
				strings.HasPrefix(sanitized, "dorby") ||
				strings.HasPrefix(sanitized, "cau") ||
				strings.HasPrefix(sanitized, "cus")
		},
		HandleFunc: func(from, msg string) error {
			reply := "Ahoj! Já jsem Pan Botka. Napiš /help pro nápovědu."
			b.storeConversation(from, msg, reply)
			err := b.whatsapp.SendText(from, reply)
			return err
		},
	}
}

func (b *Botka) pubHandler() wa.EventHandler {
	return wa.EventHandler{
		MatchFunc: func(msg string) bool {
			sanitized := b.sanitizeCommand(msg)
			return strings.HasPrefix(sanitized, "pub") ||
				strings.HasPrefix(sanitized, "hospoda")
		},
		HandleFunc: func(from, msg string) error {
			s := b.scale.GetScale()
			var reply string
			if s.Pub.IsOpen {
				reply = fmt.Sprintf("🍺 Hospoda je otevřená od %s.", s.Pub.OpenedAt)
			} else {
				reply = "😥 Hospoda je bohužel zavřená! Půjdeš otevřít?"
			}
			b.storeConversation(from, msg, reply)
			err := b.whatsapp.SendText(from, reply)
			return err
		},
	}
}

func (b *Botka) kegHandler() wa.EventHandler {
	return wa.EventHandler{
		MatchFunc: func(msg string) bool {
			sanitized := b.sanitizeCommand(msg)
			return strings.HasPrefix(sanitized, "becka") ||
				strings.HasPrefix(sanitized, "keg")
		},
		HandleFunc: func(from, msg string) error {
			s := b.scale.GetScale()
			var reply string
			if s.ActiveKeg == 0 {
				reply = "Aktuálně nemáme naraženou žádnou bečku."
			} else {
				reply = fmt.Sprintf(
					"Máme naraženou %dl bečku a zbývá v ní %d %s. Naražena byla %s v %s.",
					s.ActiveKeg,
					s.BeersLeft,
					utils.FormatBeer(s.BeersLeft),
					utils.FormatDateShort(s.ActiveKegAt),
					utils.FormatTime(s.ActiveKegAt),
				)
			}
			b.storeConversation(from, msg, reply)
			err := b.whatsapp.SendText(from, reply)
			return err
		},
	}
}

func (b *Botka) pricesHandler() wa.EventHandler {
	return wa.EventHandler{
		MatchFunc: func(msg string) bool {
			return strings.HasPrefix(b.sanitizeCommand(msg), "cenik")
		},
		HandleFunc: func(from, msg string) error {
			reply := "Ceník: \n" +
				"- Vše 25 Kč \n" +
				"- Víno 130 Kč"
			b.storeConversation(from, msg, reply)
			err := b.whatsapp.SendText(from, reply)
			return err
		},
	}
}

func (b *Botka) warehouseHandler() wa.EventHandler {
	return wa.EventHandler{
		MatchFunc: func(msg string) bool {
			return strings.HasPrefix(b.sanitizeCommand(msg), "sklad")
		},
		HandleFunc: func(from, msg string) error {
			s := b.scale.GetScale()
			reply := fmt.Sprintf("Ve skladu máme celkem %d piv.", s.WarehouseBeerLeft)
			if s.Warehouse[0].Amount > 0 {
				reply += fmt.Sprintf("\n%d × 10l", s.Warehouse[0].Amount)
			}
			if s.Warehouse[1].Amount > 0 {
				reply += fmt.Sprintf("\n%d × 15l", s.Warehouse[1].Amount)
			}
			if s.Warehouse[2].Amount > 0 {
				reply += fmt.Sprintf("\n%d × 20l", s.Warehouse[2].Amount)
			}
			if s.Warehouse[3].Amount > 0 {
				reply += fmt.Sprintf("\n%d × 30l", s.Warehouse[3].Amount)
			}
			if s.Warehouse[4].Amount > 0 {
				reply += fmt.Sprintf("\n%d × 50l", s.Warehouse[4].Amount)
			}

			b.storeConversation(from, msg, reply)
			err := b.whatsapp.SendText(from, reply)
			return err
		},
	}
}

func (b *Botka) aiHandler() wa.EventHandler {
	return wa.EventHandler{
		MatchFunc: func(_ string) bool {
			return true // always match as a backup command
		},
		HandleFunc: func(from, msg string) error {
			conversation, err := b.storage.GetConversation(from)
			if err != nil {
				return fmt.Errorf("could not get conversation: %w", err)
			}

			var messages []ai.ChatMessage
			count := 0
			for _, message := range conversation {
				if time.Since(message.At) < 12*time.Hour { // ignore message sent more than 12 hours ago
					// we need to make sure that first message will be from user
					if count == 0 && message.Author == store.ConversationMessageAuthorBot {
						continue
					}

					messages = append(messages, ai.ChatMessage{
						Text: message.Message,
						From: mapUser(message.Author),
					})

					count++
				}
			}

			response, err := b.ai.GetResponse(messages)
			if err != nil {
				b.logger.Errorf("could not get response from AI: %v", err)
				return err
			}

			b.storeConversation(from, msg, response.Text)
			return b.whatsapp.SendText(from, response.Text)
		},
	}
}

func (b *Botka) storeConversation(ID, question, answer string) {
	now := time.Now()
	err := b.storage.AddConversationMessage(ID, store.ConservationMessage{
		ID:      ID,
		Message: question,
		At:      now,
		Author:  store.ConversationMessageAuthorUser,
	})
	if err != nil {
		b.logger.Errorf("could not add conversation message: %v", err)
	}

	err = b.storage.AddConversationMessage(ID, store.ConservationMessage{
		ID:      ID,
		Message: answer,
		At:      now,
		Author:  store.ConversationMessageAuthorBot,
	})
	if err != nil {
		b.logger.Errorf("could not add conversation message: %v", err)
	}
}

func (b *Botka) sanitizeCommand(command string) string {
	c := strings.ToLower(strings.TrimSpace(strings.TrimPrefix(command, "/")))
	c, err := diacritics.Remove(c)
	if err != nil {
		b.logger.Fatalf("could not remove diacritics: %v", err) // should never happen
	}

	return c
}

func mapUser(author store.ConversationMessageAuthor) string {
	if author == store.ConversationMessageAuthorUser {
		return ai.Me
	}

	return "bot"
}
