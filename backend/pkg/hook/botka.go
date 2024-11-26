package hook

import (
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/kotrzina/keg-scale/pkg/config"
	"github.com/kotrzina/keg-scale/pkg/shops"
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

	IsOpen   bool
	OpenedAt time.Time

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
		client.RegisterEventHandler(w.helloHandler())
		client.RegisterEventHandler(w.pubHandler())
		client.RegisterEventHandler(w.kegHandler())
		client.RegisterEventHandler(w.pricesHandler())
		client.RegisterEventHandler(w.warehouseHandler())
		client.RegisterEventHandler(w.baracekHandler())
		client.RegisterEventHandler(w.maneoHandler())
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
				"/sklad - stav skladu" +
				"/baracek - ceník Baráček" +
				"/maneo - ceník Maneo"
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
		HandleFunc: func(from, _ string) error {
			reply := "Ahoj! Já jsem pan Botka. Napiš /help pro nápovědu."
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
		HandleFunc: func(from, _ string) error {
			b.mtx.RLock()
			defer b.mtx.RUnlock()
			var reply string
			if b.brain.IsOpen {
				reply = fmt.Sprintf("🍺 Hospoda je otevřená od %s.", utils.FormatTime(b.brain.OpenedAt))
			} else {
				reply = "😥 Hospoda je bohužel zavřená! Půjdeš otevřít?"
			}
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
		HandleFunc: func(from, _ string) error {
			b.mtx.RLock()
			defer b.mtx.RUnlock()
			var msg string

			if b.brain.ActiveKeg == 0 {
				msg = "Aktuálně nemáme naraženou žádnou bečku."
			} else {
				msg = fmt.Sprintf(
					"Máme naraženou %dl bečku a zbývá v ní %d %s. Naražena byla %s v %s.",
					b.brain.ActiveKeg,
					b.brain.BeerLeft,
					utils.FormatBeer(b.brain.BeerLeft),
					utils.FormatDateShort(b.brain.ActiveKegAt),
					utils.FormatTime(b.brain.ActiveKegAt),
				)
			}
			err := b.whatsapp.SendText(from, msg)
			return err
		},
	}
}

func (b *Botka) pricesHandler() wa.EventHandler {
	return wa.EventHandler{
		MatchFunc: func(msg string) bool {
			return strings.HasPrefix(b.sanitizeCommand(msg), "cenik")
		},
		HandleFunc: func(from, _ string) error {
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
			return strings.HasPrefix(b.sanitizeCommand(msg), "sklad")
		},
		HandleFunc: func(from, _ string) error {
			b.mtx.RLock()
			defer b.mtx.RUnlock()

			reply := fmt.Sprintf("Ve skladu máme celkem %d piv.", b.brain.WarehouseTotal)
			if b.brain.Warehouse[10] > 0 {
				reply += fmt.Sprintf("\n%d × 10l", b.brain.Warehouse[10])
			}
			if b.brain.Warehouse[15] > 0 {
				reply += fmt.Sprintf("\n%d × 15l", b.brain.Warehouse[15])
			}
			if b.brain.Warehouse[20] > 0 {
				reply += fmt.Sprintf("\n%d × 20l", b.brain.Warehouse[20])
			}
			if b.brain.Warehouse[30] > 0 {
				reply += fmt.Sprintf("\n%d × 30l", b.brain.Warehouse[30])
			}
			if b.brain.Warehouse[50] > 0 {
				reply += fmt.Sprintf("\n%d × 50l", b.brain.Warehouse[50])
			}

			err := b.whatsapp.SendText(from, reply)
			return err
		},
	}
}

func (b *Botka) baracekHandler() wa.EventHandler {
	return wa.EventHandler{
		MatchFunc: func(msg string) bool {
			return strings.HasPrefix(b.sanitizeCommand(msg), "baracek")
		},
		HandleFunc: func(from, _ string) error {
			baracek := shops.NewBaracek()
			urls := []string{
				"https://www.baracek.cz/sud-policka-otakar-11-10-l",
				"https://www.baracek.cz/sud-policka-otakar-11-15-l",
				"https://www.baracek.cz/sud-policka-otakar-11-30l",
				"https://www.baracek.cz/sud-policka-otakar-11-50-l",
				"https://www.baracek.cz/sud-bernard-svetly-11-15l-45obj",
				"https://www.baracek.cz/sud-bernard-svetly-11-30l-45-obj",
				"https://www.baracek.cz/sud-bernard-svetly-11-50l-45-obj",
				"https://www.baracek.cz/sud-bernard-svetly-12-20l-5obj",
				"https://www.baracek.cz/sud-bernard-svetly-12-50l-5obj",
				"https://www.baracek.cz/sud-plzen-12-keg-15l",
				"https://www.baracek.cz/sud-plzen-12-keg-30l",
				"https://www.baracek.cz/sud-plzen-12-keg-50l",
				"https://www.baracek.cz/sud-radegast-ryze-horka-12-15-l",
				"https://www.baracek.cz/sud-radegast-ryze-horka-12-30l",
				"https://www.baracek.cz/sud-radegast-ryze-horka-12-50l",
				"https://www.baracek.cz/sud-chotebor-12deg-premium-svlezak-30l",
				"https://www.baracek.cz/sud-chotebor-12deg-premium-svlezak-50l",
			}

			sb := strings.Builder{}
			sb.WriteString("*Ceník Baráček:*\n")
			for _, url := range urls {
				beer, err := baracek.GetBeer(url)
				if err != nil {
					b.logger.Errorf("could not get beer from Baracek (%s): %v", url, err)
					continue
				}

				stock := "❓"
				if beer.Stock == shops.StockTypeAvailable {
					stock = "🍺"
				}

				sb.WriteString(fmt.Sprintf("%s %s -> *%d* Kč\n", stock, strings.TrimPrefix(beer.Title, "sud "), beer.Price))
			}

			sb.WriteString("------\n🍺 - skladem\n❓ - neznámý stav skladu")

			err := b.whatsapp.SendText(from, sb.String())
			return err
		},
	}
}

func (b *Botka) maneoHandler() wa.EventHandler {
	return wa.EventHandler{
		MatchFunc: func(msg string) bool {
			return strings.HasPrefix(b.sanitizeCommand(msg), "maneo")
		},
		HandleFunc: func(from, _ string) error {
			maneo := shops.NewManeo()
			urls := []string{
				"https://www.eshop.maneo.cz/policka-11-otakar-10l-keg-ean27154-skup5Zn1ak1Zn1ak31.php",
				"https://www.eshop.maneo.cz/policka-11-otakar-15l-keg-sv-lezak-ean27155-skup5Zn1ak1Zn1ak31.php",
				"https://www.eshop.maneo.cz/policka-11-otakar-30l-keg-ean25145-skup5Zn1ak1Zn1ak31.php",
				"https://www.eshop.maneo.cz/policka-11-otakar-50l-keg-ean25147-skup5Zn1ak1Zn1ak31.php",
				"https://www.eshop.maneo.cz/bernard-11-svetly-lezak-15l-keg-ean256084-skup5Zn1ak1Zn1ak16.php",
				"https://www.eshop.maneo.cz/bernard-11-svetly-lezak-30l-keg-ean25132-skup5Zn1ak1Zn1ak16.php",
				"https://www.eshop.maneo.cz/bernard-11-svetly-lezak-50l-keg-ean25100-skup5Zn1ak1Zn1ak16.php",
				"https://www.eshop.maneo.cz/bernard-12-svetly-lezak-20l-keg-ean27802-skup5Zn1ak1Zn1ak16.php",
				"https://www.eshop.maneo.cz/bernard-12-svetly-lezak-30l-keg-ean25133-skup5Zn1ak1Zn1ak16.php",
				"https://www.eshop.maneo.cz/bernard-12-svetly-lezak-50l-keg-ean25111-skup5Zn1ak1Zn1ak16.php",
				"https://www.eshop.maneo.cz/plzen-12-15l-keg-ean24016-skup5Zn1ak1Zn1ak1.php",
				"https://www.eshop.maneo.cz/plzen-12-30l-keg-ean24072-skup5Zn1ak1Zn1ak1.php",
				"https://www.eshop.maneo.cz/plzen-12-50l-keg-ean24070-skup5Zn1ak1Zn1ak1.php",
				"https://www.eshop.maneo.cz/radegast-12-ryze-horka-15l-keg-ean25017-skup5Zn1ak1Zn1ak14.php",
				"https://www.eshop.maneo.cz/radegast-12-ryze-horka-30l-keg-ean24077-skup5Zn1ak1Zn1ak14.php",
				"https://www.eshop.maneo.cz/radegast-12-ryze-horka-50l-keg-ean24076-skup5Zn1ak1Zn1ak14.php",
				"https://www.eshop.maneo.cz/chotebor-sv-lezak-15l-keg-12-premium-ean25233-skup5Zn1ak1Zn1ak13.php",
				"https://www.eshop.maneo.cz/chotebor-12-sv-lezak-30l-keg-premium-ean25090-skup5Zn1ak1Zn1ak13.php",
				"https://www.eshop.maneo.cz/chotebor-12-sv-lezak-50l-keg-premium-ean25091-skup5Zn1ak1Zn1ak13.php",
				"https://www.eshop.maneo.cz/kocour-12-sv-lezak-20l-keg-ean244518-skup5Zn1ak1Zn1ak34.php",
				"https://www.eshop.maneo.cz/kocour-12-sv-lezak-30l-keg-ean25164-skup5Zn1ak1Zn1ak34.php",
				"https://www.eshop.maneo.cz/beskydsky-lezak-15l-keg-4-8-ean24399-skup5Zn1ak1Zn1ak41.php",
				"https://www.eshop.maneo.cz/beskydsky-lezak-30l-keg-4-8-ean24390-skup5Zn1ak1Zn1ak41.php",
				"https://www.eshop.maneo.cz/beskydsky-lezak-50l-keg-4-8-ean246480-skup5Zn1ak1Zn1ak41.php",
				"https://www.eshop.maneo.cz/valasek-12-sv-lezak-15l-keg-ean24441-skup5Zn1ak1Zn1ak47.php",
				"https://www.eshop.maneo.cz/valasek-12-sv-lezak-30l-keg-ean27177-skup5Zn1ak1Zn1ak47.php",
				"https://www.eshop.maneo.cz/jarosovska-11-jura-15l-keg-ean245529-skup5Zn1ak1Zn1ak56.php",
				"https://www.eshop.maneo.cz/jarosovska-11-jura-30l-keg-ean26320-skup5Zn1ak1Zn1ak56.php",
				"https://www.eshop.maneo.cz/jarosovska-12-matus-sv-lezak-30l-keg-ean26321-skup5Zn1ak1Zn1ak56.php",
				"https://www.eshop.maneo.cz/opice-11-sv-lezak-15l-keg-ean25460-skup5Zn1ak1Zn1ak87.php",
				"https://www.eshop.maneo.cz/opice-11-sv-lezak-30l-keg-ean25461-skup5Zn1ak1Zn1ak87.php",
				"https://www.eshop.maneo.cz/albert-salina-30l-svet-lezak-ean25602-skup5Zn1ak1Zn1ak90.php",
			}

			sb := strings.Builder{}
			sb.WriteString("*Ceník Maneo:*\n")
			for _, url := range urls {
				beer, err := maneo.GetBeer(url)
				if err != nil {
					b.logger.Errorf("could not get beer from Maneo (%s): %v", url, err)
					continue
				}

				stock := "❓"
				if beer.Stock == shops.StockTypeAvailable {
					stock = "🍺"
				}

				sb.WriteString(fmt.Sprintf(
					"%s %s -> *%d* Kč\n",
					stock,
					strings.TrimSpace(strings.TrimSuffix(beer.Title, " keg")),
					beer.Price,
				))
			}

			sb.WriteString("------\n🍺 - skladem\n❓ - neznámý stav skladu")

			err := b.whatsapp.SendText(from, sb.String())
			return err
		},
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
