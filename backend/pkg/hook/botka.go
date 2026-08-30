package hook

import (
	"context"
	"fmt"
	"regexp"
	"slices"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/dundee/qrpay"
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
		client.RegisterEventHandler(w.peopleHandler())
		client.RegisterEventHandler(w.thirstHandler())
		client.RegisterEventHandler(w.kegHandler())
		client.RegisterEventHandler(w.pricesHandler())
		client.RegisterEventHandler(w.qrPaymentHandler())
		client.RegisterEventHandler(w.bankHandler())
		client.RegisterEventHandler(w.warehouseHandler())
		client.RegisterEventHandler(w.resetHandler())

		client.RegisterEventHandler(w.secretHelpHandler())
		client.RegisterEventHandler(w.openHandler())
		client.RegisterEventHandler(w.cepHandler())
		client.RegisterEventHandler(w.pubTestingHandler())
		client.RegisterEventHandler(w.volleyballHandler())
		client.RegisterEventHandler(w.noMessageHandler())
		client.RegisterEventHandler(w.shoutHandler())

		client.RegisterEventHandler(w.aiHandler())
	}

	// send messages when the pub is open
	kegScale.RegisterEvent(scale.EventOpen, w.messageOpen)
	if len(conf.WhatsAppCustomMessages) > 0 {
		kegScale.RegisterEvent(scale.EventOpen, w.messageOpenCustom)
	}

	return w
}

func (b *Botka) ProvideWebHandlers() []wa.EventHandler {
	if b.config.Debug {
		return []wa.EventHandler{}
	}

	return []wa.EventHandler{
		b.helpHandler(),
		b.helloHandler(),
		b.pubHandler(),
		b.peopleHandler(),
		b.thirstHandler(),
		b.kegHandler(),
		b.pricesHandler(),
		// b.qrPaymentHandler(),
		b.bankHandler(),
		b.warehouseHandler(),
		// b.resetHandler(),
		b.secretHelpHandler(),
		b.openHandler(),
		b.cepHandler(),
		b.pubTestingHandler(),
		b.volleyballHandler(),
		b.noMessageHandler(),
		b.shoutHandler(),
		//b.aiHandler(), // web has own ai logic
	}
}

// nolint: govet // temporary
func (b *Botka) messageOpen(_ scale.EventType) error {
	msg, err := b.ai.GenerateGeneralOpenMessage()
	if err != nil {
		b.logger.Errorf("could not generate general open message: %v", err)

		// backup message
		data := b.scale.GetScale()
		msg = "Pivo! 🍺"
		if data.ActiveKeg > 0 {
			msg += fmt.Sprintf(
				"\nMáme naraženou %dl bečku a zbývá v ní %d %s.",
				data.ActiveKeg,
				data.BeersLeft,
				utils.FormatBeer(data.BeersLeft),
			)
		}
		if data.WarehouseBeerLeft > 0 {
			msg += fmt.Sprintf(
				"\nVe skladu máme %d %s.",
				data.WarehouseBeerLeft,
				utils.FormatBeer(data.WarehouseBeerLeft),
			)
		}
	}

	err = b.whatsapp.SendText(b.config.WhatsAppOpenJid, msg)
	if err != nil {
		return fmt.Errorf("could not send Botka message: %w", err)
	}

	return nil
}

func (b *Botka) messageOpenCustom(_ scale.EventType) error {
	for _, user := range b.config.WhatsAppCustomMessages {
		msg, err := b.ai.GenerateCustomOpenMessage(user.Name)
		if err != nil {
			return fmt.Errorf("could not generate custom open message: %w", err)
		}

		err = b.whatsapp.SendText(user.Phone, msg)
		if err != nil {
			return fmt.Errorf("could not send Botka open custom message: %w", err)
		}
	}

	return nil
}

func (b *Botka) helpHandler() wa.EventHandler {
	return wa.EventHandler{
		MatchFunc: func(msg string) bool {
			sanitized := b.sanitizeCommand(msg)

			if len(sanitized) > 10 {
				return false
			}

			return strings.HasPrefix(sanitized, "help") ||
				strings.HasPrefix(sanitized, "napoveda") ||
				strings.HasPrefix(sanitized, "pomoc")
		},
		HandleFunc: func(from, _ string) (string, error) {
			reply := "Příkazy: \n" +
				"/help - zobrazí nápovědu \n" +
				"/pub /hospoda - informace o hospodě \n" +
				"/zizen - pošle stamgastům zprávu, že bys dnes na jedno šel \n" +
				"/becka - informace o aktuální bečce \n" +
				"/cenik - ceník \n" +
				"/qr 275 - zaplať QR kódem \n" +
				"/banka - stav bankovního účtu \n" +
				"/sklad - stav skladu\n" +
				"/reset - Pan Botka zapomene všechno"

			return reply, nil
		},
	}
}

func (b *Botka) helloHandler() wa.EventHandler {
	return wa.EventHandler{
		MatchFunc: func(msg string) bool {
			sanitized := b.sanitizeCommand(msg)
			if len(sanitized) > 7 {
				return false
			}

			return strings.HasPrefix(sanitized, "hello") ||
				strings.HasPrefix(sanitized, "hi") ||
				strings.HasPrefix(sanitized, "ahoj") ||
				strings.HasPrefix(sanitized, "zdar") ||
				strings.HasPrefix(sanitized, "dorby") ||
				strings.HasPrefix(sanitized, "cau") ||
				strings.HasPrefix(sanitized, "cus")
		},
		HandleFunc: func(from, msg string) (string, error) {
			reply := "Ahoj! Já jsem Pan Botka. Napiš /help pro nápovědu."
			b.storeConversation(from, msg, reply)

			return reply, nil
		},
	}
}

func (b *Botka) pubHandler() wa.EventHandler {
	return wa.EventHandler{
		MatchFunc: func(msg string) bool {
			sanitized := b.sanitizeCommand(msg)

			if len(sanitized) > 8 {
				return false
			}

			return strings.HasPrefix(sanitized, "pub") ||
				strings.HasPrefix(sanitized, "hospoda")
		},
		HandleFunc: func(from, msg string) (string, error) {
			s := b.scale.GetScale()
			var reply string
			if s.Pub.IsOpen {
				reply = fmt.Sprintf("🍺 Hospoda je otevřená od %s.", s.Pub.OpenedAt)
			} else {
				reply = "😥 Hospoda je bohužel zavřená! Půjdeš otevřít?"
			}
			b.storeConversation(from, msg, reply)

			return reply, nil
		},
	}
}

func (b *Botka) thirstHandler() wa.EventHandler {
	return wa.EventHandler{
		MatchFunc: func(msg string) bool {
			sanitized := b.sanitizeCommand(msg)
			return strings.HasPrefix(sanitized, "zizen")
		},
		HandleFunc: func(from, msg string) (string, error) {
			// remove the command prefix
			sanitized := strings.TrimPrefix(b.sanitizeCommand(msg), "zizen")

			groupMsg, err := b.ai.GenerateRegularsMessage(sanitized)
			if err != nil {
				b.logger.Errorf("could not generate regulars message: %v", err)
			}

			err = b.whatsapp.SendText(b.config.WhatsAppRegularsJid, groupMsg)
			if err != nil {
				b.logger.Errorf("could not send regulars message: %v", err)
				return "Nemůžu poslat zprávu štamgastům, něco se pokazilo.", fmt.Errorf("could not send thirst message to regulars group chat: %w", err)
			}

			reply := "🙋🏻Ok, hned vygeneruji zprávu pro štamgasty."
			return reply, nil
		},
	}
}

// provides a list of people in the pub
func (b *Botka) peopleHandler() wa.EventHandler {
	return wa.EventHandler{
		MatchFunc: func(msg string) bool {
			sanitized := b.sanitizeCommand(msg)
			return strings.HasPrefix(sanitized, "stul")
		},
		HandleFunc: func(from, msg string) (string, error) {
			output := b.scale.GetScale()

			var names []string
			for _, bt := range output.BtDevices {
				if bt.Known {
					names = append(names, bt.Name)
				}
			}

			if len(names) == 0 {
				return "🪹Nikdo v hospodě není.", nil
			}

			sb := strings.Builder{}
			sb.WriteString("🍻Osazenstvo:\n")
			for _, name := range names {
				sb.WriteString(fmt.Sprintf("- %s\n", name))
			}

			return strings.TrimSpace(sb.String()), nil
		},
	}
}

func (b *Botka) kegHandler() wa.EventHandler {
	return wa.EventHandler{
		MatchFunc: func(msg string) bool {
			sanitized := b.sanitizeCommand(msg)

			if len(sanitized) > 6 {
				return false
			}

			return strings.HasPrefix(sanitized, "becka") ||
				strings.HasPrefix(sanitized, "keg")
		},
		HandleFunc: func(from, msg string) (string, error) {
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
			return reply, nil
		},
	}
}

func (b *Botka) pricesHandler() wa.EventHandler {
	return wa.EventHandler{
		MatchFunc: func(msg string) bool {
			return b.sanitizeCommand(msg) == "cenik"
		},
		HandleFunc: func(from, msg string) (string, error) {
			reply := "Ceník: \n" +
				"- Vše 25 Kč \n" +
				"- Víno 130 Kč"
			b.storeConversation(from, msg, reply)
			return reply, nil
		},
	}
}

func (b *Botka) qrPaymentHandler() wa.EventHandler {
	return wa.EventHandler{
		MatchFunc: func(msg string) bool {
			return len(msg) < 10 && strings.HasPrefix(b.sanitizeCommand(msg), "qr")
		},
		HandleFunc: func(from, msg string) (string, error) {
			errMsg := "Nepodařilo se vygenerovat QR kód"
			if b.config.FioIban == "" {
				return errMsg, fmt.Errorf("fio IBAN is not configured")
			}

			payment := qrpay.NewSpaydPayment()
			if err := payment.SetIBAN(b.config.FioIban); err != nil {
				return errMsg, fmt.Errorf("could not set IBAN: %w", err)
			}

			amount, err := parseAmountFromQrPaymentCommand(msg)
			if err == nil {
				// if amount is specified in the command, set it
				if err := payment.SetAmount(fmt.Sprintf("%d", amount)); err != nil {
					b.logger.Errorf("could not set payment amount: %s", err)
				}
			}

			img, err := qrpay.GetQRCodeImage(payment)
			if err != nil {
				return errMsg, fmt.Errorf("could not get QR Code: %w", err)
			}

			err = b.whatsapp.SendImage(from, "Zaplať QR kódem", img)
			if err != nil {
				return errMsg, fmt.Errorf("could not send image: %w", err)
			}

			b.storeConversation(from, msg, "Image with QR code for payment has been sent.")
			return "", nil
		},
	}
}

func (b *Botka) bankHandler() wa.EventHandler {
	return wa.EventHandler{
		MatchFunc: func(msg string) bool {
			return len(msg) < 8 && strings.HasPrefix(b.sanitizeCommand(msg), "bank")
		},
		HandleFunc: func(from, msg string) (string, error) {
			err := b.scale.BankRefresh(context.Background(), true)
			if err != nil {
				b.logger.Errorf("could not refresh bank data: %v", err)
				reply := "Něco se pokazilo při načítání dat z banky. Zkus to prosím znovu později."
				return reply, nil
			}

			s := b.scale.GetScale()

			sb := strings.Builder{}
			sb.WriteString(fmt.Sprintf("Stav účtu: %s Kč\n\n", s.BankBalance.Balance.String()))
			sb.WriteString("Poslední transakce:\n")
			slices.Reverse(s.BankTransactions)
			for _, t := range s.BankTransactions {
				sb.WriteString(fmt.Sprintf("- %s: %s Kč\n", t.AccountName, t.Amount.String()))
			}

			reply := strings.TrimSuffix(sb.String(), "\n")
			b.storeConversation(from, msg, reply)
			return reply, nil
		},
	}
}

func (b *Botka) warehouseHandler() wa.EventHandler {
	return wa.EventHandler{
		MatchFunc: func(msg string) bool {
			return b.sanitizeCommand(msg) == "sklad"
		},
		HandleFunc: func(from, msg string) (string, error) {
			s := b.scale.GetScale()
			reply := fmt.Sprintf("Ve skladu máme celkem %d piv.", s.WarehouseBeerLeft)
			for _, w := range s.Warehouse {
				if w.Amount > 0 {
					reply += fmt.Sprintf("\n%d × %dl", w.Amount, w.Keg)
				}
			}
			b.storeConversation(from, msg, reply)
			return reply, nil
		},
	}
}

func (b *Botka) resetHandler() wa.EventHandler {
	return wa.EventHandler{
		MatchFunc: func(msg string) bool {
			return strings.HasPrefix(b.sanitizeCommand(msg), "reset")
		},
		HandleFunc: func(from, _ string) (string, error) {
			err := b.storage.ResetConversation(from)
			reply := "Cože? O čem jsme to mluvili? 🤔"
			if err != nil {
				b.logger.Errorf("could not reset conversation: %v", err)
				reply = "Něco se pokazilo, zkuste to prosím znovu."
			} else {
				b.logger.Infof("conversation with %q has been reset", from)
			}

			return reply, nil
		},
	}
}

func (b *Botka) secretHelpHandler() wa.EventHandler {
	return wa.EventHandler{
		MatchFunc: func(msg string) bool {
			return checkSecretCommand(msg, b.config.Commands.Help)
		},
		HandleFunc: func(from, _ string) (string, error) {
			sb := strings.Builder{}

			sb.WriteString("*Příkazy:*\n")
			sb.WriteString(fmt.Sprintf("*!%s* - otevři hospodu\n", b.config.Commands.Open))
			sb.WriteString(fmt.Sprintf("*!%s* - dnes točíme tohle pivo\n", "cep")) // semi-secret command
			sb.WriteString(fmt.Sprintf("*!%s* - volejbal zpráva do skupiny hospoda\n", b.config.Commands.Volleyball))
			sb.WriteString(fmt.Sprintf("*!%s* - neposílej dnes zprávu o otevření hospody\n", b.config.Commands.NoMessage))
			sb.WriteString(fmt.Sprintf("*!%s ...* - zpráva do kanálu Hospoda\n", b.config.Commands.Shout))

			sb.WriteString("\nPříkaz musí být napsaný přesně tak, jak je zde uveden.")

			return sb.String(), nil
		},
	}
}

func (b *Botka) openHandler() wa.EventHandler {
	return wa.EventHandler{
		MatchFunc: func(msg string) bool {
			return checkSecretCommand(msg, b.config.Commands.Open)
		},
		HandleFunc: func(from, _ string) (string, error) {
			reply := "Jasnňačka! Otevírám hospodu. 🍻"
			if err := b.scale.ForceOpen(); err != nil {
				b.logger.Infof("could not open pub: %v", err)
				reply = "Něco se pokazilo, hospodu se nepodařilo otevřít. Zkus to prosím znovu později."
			}

			return reply, nil
		},
	}
}

func (b *Botka) cepHandler() wa.EventHandler {
	return wa.EventHandler{
		MatchFunc: func(msg string) bool {
			return strings.HasPrefix(b.sanitizeCommand(msg), "cep")
		},
		HandleFunc: func(from, msg string) (string, error) {
			beer := strings.TrimSpace(msg[4:]) // remove the command prefix

			if err := b.storage.SetTodayBeer(beer); err != nil {
				return "Nepodařilo se mi nastavit pivo na dnešek", fmt.Errorf("could not set today beer: %w", err)
			}

			reply := fmt.Sprintf("Ok, zmíním pivo: %s při otevření hospody.", beer)
			return reply, nil
		},
	}
}

// pubTestingHandler sets the pub testing flag in the storage
// hidden command - not listed in any help
func (b *Botka) pubTestingHandler() wa.EventHandler {
	return wa.EventHandler{
		MatchFunc: func(msg string) bool {
			_, ok := parsePubTestingCommand(b.sanitizeCommand(msg))
			return ok
		},
		HandleFunc: func(from, msg string) (string, error) {
			state, ok := parsePubTestingCommand(b.sanitizeCommand(msg))
			if !ok {
				return "Nerozumím. Použij *!pubtesting on* nebo *!pubtesting off*.", fmt.Errorf("could not parse pub testing command: %s", msg)
			}

			if err := b.storage.SetPubTesting(state); err != nil {
				return "Nepodařilo se mi nastavit pubTesting.", fmt.Errorf("could not set pub testing: %w", err)
			}

			b.logger.Infof("%s set pub testing to %t", from, state)
			if state {
				return "Ok, pubTesting zapnut. ✅", nil
			}

			return "Ok, pubTesting vypnut. 🚫", nil
		},
	}
}

func (b *Botka) volleyballHandler() wa.EventHandler {
	return wa.EventHandler{
		MatchFunc: func(msg string) bool {
			return checkSecretCommand(msg, b.config.Commands.Volleyball)
		},
		HandleFunc: func(from, _ string) (string, error) {
			msg, err := b.ai.GenerateVolleyballMessage()
			if err != nil {
				return "Nepodařilo se mi vygenerovat zprávu", fmt.Errorf("could not generate volleyball message: %w", err)
			}

			err = b.whatsapp.SendText(b.config.WhatsAppOpenJid, msg)
			if err != nil {
				return "Nepodařilo se mi odeslat zprávu do skupiny", fmt.Errorf("could not send volleyball message to group chat: %w", err)
			}

			reply := "Rozkaz kapitáne! 🏐🏐\n\nHned vygeneruji zprávu o volejbalu a pošlu ji do skupiny Hospoda."
			return reply, nil
		},
	}
}

func (b *Botka) noMessageHandler() wa.EventHandler {
	return wa.EventHandler{
		MatchFunc: func(msg string) bool {
			return checkSecretCommand(msg, b.config.Commands.NoMessage)
		},
		HandleFunc: func(from, _ string) (string, error) {
			b.scale.ResetOpenAt()
			b.logger.Infof("%s requested no message open", from)
			reply := "Rozumím, dneska na tajňačku!! 🤫🤫"
			return reply, nil
		},
	}
}

func (b *Botka) shoutHandler() wa.EventHandler {
	return wa.EventHandler{
		MatchFunc: func(msg string) bool {
			return strings.HasPrefix(msg, fmt.Sprintf("!%s", b.config.Commands.Shout))
		},
		HandleFunc: func(from, msg string) (string, error) {
			text := strings.TrimSpace(strings.TrimPrefix(msg, fmt.Sprintf("!%s", b.config.Commands.Shout)))
			if text == "" {
				return "Musíš něco napsat.", fmt.Errorf("no message provided for shout command")
			}

			if err := b.whatsapp.SendText(b.config.WhatsAppOpenJid, text); err != nil {
				return "Nepodařilo se mi poslat zprávu do skupiny Hospoda.", fmt.Errorf("could not send shout message to the group chat: %w", err)
			}

			b.logger.Infof("%s requested shout command", from)
			reply := "Ok, posílám zprávu do skupiny Hospoda."
			return reply, nil
		},
	}
}

func (b *Botka) aiHandler() wa.EventHandler {
	return wa.EventHandler{
		MatchFunc: func(_ string) bool {
			return true // always match as a backup command
		},
		HandleFunc: func(from, msg string) (string, error) {
			err := b.whatsapp.SetTyping(from, true)
			if err != nil {
				b.logger.Warnf("could not set typing: %v", err)
			}

			defer func() {
				err := b.whatsapp.SetTyping(from, false)
				if err != nil {
					b.logger.Warnf("could not unset typing: %v", err)
				}
			}()

			conversation, err := b.storage.GetConversation(from)
			if err != nil {
				return "Nemůžu ti odpovědět, protože se mi nepodařilo načíst konverzaci.", fmt.Errorf("could not get conversation: %w", err)
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

			// add the current message
			messages = append(messages, ai.ChatMessage{
				Text: msg,
				From: ai.Me,
			})

			response, err := b.ai.GetResponse(messages, ai.ModelQualityHigh)
			if err != nil {
				b.logger.Errorf("could not get response from AI: %v", err)
				response = ai.Response{
					Text: "Teď bohužel nedokážu odpovědět. Zkus to prosím později.",
					Cost: ai.Cost{
						Input:  0,
						Output: 0,
					},
				}
			}

			b.storeConversation(from, msg, response.Text)
			return response.Text, nil
		},
	}
}

func (b *Botka) storeConversation(id, question, answer string) {
	if id == "API" {
		return
	}

	now := time.Now()
	err := b.storage.AddConversationMessage(id, store.ConservationMessage{
		ID:      id,
		Message: question,
		At:      now,
		Author:  store.ConversationMessageAuthorUser,
	})
	if err != nil {
		b.logger.Errorf("could not add conversation message: %v", err)
	}

	err = b.storage.AddConversationMessage(id, store.ConservationMessage{
		ID:      id,
		Message: answer,
		At:      now,
		Author:  store.ConversationMessageAuthorBot,
	})
	if err != nil {
		b.logger.Errorf("could not add conversation message: %v", err)
	}
}

func (b *Botka) sanitizeCommand(command string) string {
	var err error
	c := strings.TrimPrefix(command, "/")
	c = strings.TrimPrefix(c, "!")
	c = strings.ToLower(strings.TrimSpace(c))
	c, err = diacritics.Remove(c)
	if err != nil {
		b.logger.Fatalf("could not remove diacritics: %v", err) // should never happen
	}

	return c
}

var reAmountQr = regexp.MustCompile(`/?[Qq][Rr] ([1-9][0-9]+).*`)

func parseAmountFromQrPaymentCommand(command string) (int, error) {
	matches := reAmountQr.FindStringSubmatch(command)
	if len(matches) < 2 {
		return 0, fmt.Errorf("could not parse amount from command: %s", command)
	}

	amount, err := strconv.Atoi(matches[1])
	if err != nil {
		return 0, fmt.Errorf("could not parse amount from command: %s", command)
	}

	return amount, nil
}

// parsePubTestingCommand parses the sanitized pubtesting command
// returns the requested state and true if the command is valid
func parsePubTestingCommand(sanitized string) (bool, bool) {
	switch strings.TrimSpace(sanitized) {
	case "pubtesting on":
		return true, true
	case "pubtesting off":
		return false, true
	default:
		return false, false
	}
}

func mapUser(author store.ConversationMessageAuthor) string {
	if author == store.ConversationMessageAuthorUser {
		return ai.Me
	}

	return "bot"
}

// checkSecretCommand checks if the message is a secret command
// secret commands are defined in the configuration
func checkSecretCommand(msg, command string) bool {
	if command == "" {
		return false // ignore if the command is not set
	}

	return strings.EqualFold(msg, fmt.Sprintf("!%s", command))
}
