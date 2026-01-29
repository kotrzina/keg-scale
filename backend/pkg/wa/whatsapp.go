package wa

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"sync"

	"github.com/kotrzina/keg-scale/pkg/config"
	"github.com/sirupsen/logrus"
	"github.com/skip2/go-qrcode"
	"go.mau.fi/whatsmeow"
	"go.mau.fi/whatsmeow/proto/waE2E"
	"go.mau.fi/whatsmeow/store"
	"go.mau.fi/whatsmeow/store/sqlstore"
	"go.mau.fi/whatsmeow/types"
	"go.mau.fi/whatsmeow/types/events"
	"google.golang.org/protobuf/proto"
)

type WhatsAppClient struct {
	client   *whatsmeow.Client
	ready    bool // true if client is ready to send messages
	store    *store.Device
	handlers []EventHandler
	qrCode   *string // QR code for pairing

	config   *config.Config
	ctx      context.Context
	readyMtx sync.RWMutex
	logger   *logrus.Logger
}

type EventMatchFunc func(msg string) bool

type EventHandleFunc func(from, msg string) (string, error) // from = sender ID

type EventHandler struct {
	MatchFunc  EventMatchFunc
	HandleFunc EventHandleFunc
}

func New(ctx context.Context, conf *config.Config, logger *logrus.Logger) *WhatsAppClient {
	customLogger := createLogger(logger)
	container, err := sqlstore.New(ctx, "postgres", conf.DBString, customLogger)
	if err != nil {
		logger.Fatalf("Failed to create container: %v", err)
	}

	// If you want multiple sessions, remember their JIDs and use .GetDevice(jid) or .GetAllDevices() instead.
	deviceStore, err := container.GetFirstDevice(ctx)
	if err != nil {
		logger.Fatalf("Failed to get device: %v", err)
	}

	emptyString := ""
	qrCode := &emptyString

	client := whatsmeow.NewClient(deviceStore, customLogger)

	if client.Store.ID == nil {
		logger.Infof("Not logged in, getting QR code")
		qrChan, err := client.GetQRChannel(ctx)
		if err != nil {
			logger.Fatalf("Failed to get QR channel: %v", err)
		}
		err = client.Connect()
		if err != nil {
			logger.Fatalf("Failed to connect: %v", err)
		}

		go func() {
			for evt := range qrChan {
				if evt.Event == "code" {
					logger.Infof("QR code: %s", evt.Code)
					*qrCode = evt.Code
				} else {
					logger.Infof("Login event: %s", evt.Event)
				}
			}

			logger.Infof("Logged in")
		}()
	} else {
		logger.Debugf("Already logged in")
		err = client.Connect()
		if err != nil {
			logger.Fatalf("Failed to connect: %v", err)
		}
	}

	err = client.SendPresence(ctx, types.PresenceAvailable)
	if err != nil {
		logger.Errorf("Failed to send presence: %v", err)
	}

	wa := &WhatsAppClient{
		client:   client,
		ready:    false,
		store:    deviceStore,
		handlers: []EventHandler{},
		qrCode:   qrCode,

		config:   conf,
		ctx:      ctx,
		readyMtx: sync.RWMutex{},
		logger:   logger,
	}

	client.AddEventHandler(wa.eventHandler)
	return wa
}

func (wa *WhatsAppClient) RegisterEventHandler(handler EventHandler) {
	wa.handlers = append(wa.handlers, handler)
}

func (wa *WhatsAppClient) eventHandler(evt interface{}) {
	wa.logger.Infof("Received WhatsApp event: %v", evt)

	//nolint: gocritic
	switch v := evt.(type) {
	case *events.Message:
		wa.handleIncomingMessage(v)
	}
}

func (wa *WhatsAppClient) handleIncomingMessage(msg *events.Message) {
	text := ""
	if msg.Message.GetExtendedTextMessage() != nil && msg.Message.GetExtendedTextMessage().Text != nil {
		text = msg.Message.GetExtendedTextMessage().GetText()
	}

	if msg.Message.GetConversation() != "" {
		text = msg.Message.GetConversation()
	}

	if text == "" {
		wa.logger.Warnf("Received unknown/empty message: %v", msg)
		return
	}

	from := fmt.Sprintf("%s@%s", msg.Info.Chat.User, msg.Info.Chat.Server) // we want to replay to the same chat
	wa.logger.Infof(
		"received message in chat %s@%s from %s@%s: %s",
		msg.Info.Chat.User,
		msg.Info.Chat.Server,
		msg.Info.Sender.User,
		msg.Info.Sender.Server,
		text,
	)

	// ignore messages from open WhatsApp chat
	if msg.Info.Chat.User == wa.config.WhatsAppOpenJid {
		wa.logger.Infof("Ignoring message from %q", wa.config.WhatsAppOpenJid)
		return
	}

	// ignore messages from regulars WhatsApp chat
	if msg.Info.Chat.User == wa.config.WhatsAppRegularsJid {
		wa.logger.Infof("Ignoring message from %q", wa.config.WhatsAppRegularsJid)
		return
	}

	if len(text) > 500 {
		wa.logger.Warnf("Message from %q is too long: %d", from, len(text))
		return // do not process long messages
	}

	for _, handler := range wa.handlers {
		if handler.MatchFunc(text) {
			reply, err := handler.HandleFunc(from, text)
			if err != nil {
				wa.logger.Errorf("Handler error: %v", err)
				if reply == "" {
					reply = "🥺Omlouvám se, ale něco se pokazilo. Zkuste to prosím později znovu."
				}
			}
			if reply != "" {
				if serr := wa.SendText(from, reply); serr != nil {
					wa.logger.Errorf("Failed to send reply: %v", serr)
				}
			}

			break // do not process other handlers
		}
	}
}

func (wa *WhatsAppClient) SetTyping(to string, typing bool) error {
	if !wa.IsReady() {
		return fmt.Errorf("WhatsAppClient is not ready")
	}

	state := types.ChatPresencePaused
	if typing {
		state = types.ChatPresenceComposing
	}

	err := wa.client.SendChatPresence(context.Background(), wa.buildJid(to), state, types.ChatPresenceMediaText)
	if err != nil {
		return fmt.Errorf("failed to send chat presence: %w", err)
	}

	return nil
}

func (wa *WhatsAppClient) SendText(to, text string) error {
	msg := &waE2E.Message{
		Conversation: proto.String(text),
	}
	return wa.send(to, msg)
}

type Location struct {
	Lat     float64
	Long    float64
	Name    string
	Address string
	Comment string
}

func (wa *WhatsAppClient) SendLocation(to string, loc Location) error {
	msg := &waE2E.Message{
		LocationMessage: &waE2E.LocationMessage{
			DegreesLatitude:  proto.Float64(loc.Lat),
			DegreesLongitude: proto.Float64(loc.Long),
			Name:             proto.String(loc.Name),
			Address:          proto.String(loc.Address),
			Comment:          proto.String(loc.Comment),
		},
	}
	return wa.send(to, msg)
}

func (wa *WhatsAppClient) SendImage(to, caption string, image []byte) error {
	if !wa.IsReady() {
		return fmt.Errorf("WhatsAppClient is not ready")
	}

	imgResp, err := wa.client.Upload(wa.ctx, image, whatsmeow.MediaImage)
	if err != nil {
		return fmt.Errorf("failed to upload image: %w", err)
	}
	wa.logger.Infof("Image uploaded: %s", imgResp.URL)

	imageMsg := &waE2E.ImageMessage{
		Caption:  proto.String(caption),
		Mimetype: proto.String(http.DetectContentType(image)),

		URL:           &imgResp.URL,
		DirectPath:    &imgResp.DirectPath,
		MediaKey:      imgResp.MediaKey,
		FileEncSHA256: imgResp.FileEncSHA256,
		FileSHA256:    imgResp.FileSHA256,
		FileLength:    &imgResp.FileLength,
	}
	msgResp, err := wa.client.SendMessage(wa.ctx, wa.buildJid(to), &waE2E.Message{
		ImageMessage: imageMsg,
	})
	if err != nil {
		return fmt.Errorf("failed to send image: %w", err)
	}

	wa.logger.Infof("Image sent: %s", msgResp.ID)
	return nil
}

func (wa *WhatsAppClient) Close() {
	wa.client.Disconnect()
}

func (wa *WhatsAppClient) buildJid(user string) types.JID {
	if wa.config.Debug {
		user = wa.config.WhatsAppOpenJid // it is set to my personal account
	}

	if strings.Contains(user, "@") {
		parts := strings.Split(user, "@")
		if len(parts) == 2 {
			return types.JID{
				User:   parts[0],
				Server: parts[1],
			}
		}

		wa.logger.Errorf("Invalid jid user: %s", user)
	}

	server := "s.whatsapp.net" // users
	if len(user) > 14 {        // longer that phone number
		server = "g.us" // groups
	}
	return types.JID{
		User:   user,
		Server: server,
	}
}

func (wa *WhatsAppClient) send(to string, msg *waE2E.Message) error {
	if !wa.IsReady() {
		return fmt.Errorf("WhatsAppClient is not ready")
	}

	resp, err := wa.client.SendMessage(wa.ctx, wa.buildJid(to), msg)
	if err != nil {
		return fmt.Errorf("failed to send message: %w", err)
	}

	wa.logger.Infof("Message sent to %s with ID %s", to, resp.ID)
	return err
}

func (wa *WhatsAppClient) MakeReady() {
	wa.readyMtx.Lock()
	defer wa.readyMtx.Unlock()

	wa.ready = true
	wa.logger.Infof("WhatsAppClient is ready")
}

func (wa *WhatsAppClient) IsReady() bool {
	if !wa.client.IsConnected() {
		wa.logger.Errorf("Not connected to WhatsAppClient")
	}

	wa.readyMtx.RLock()
	defer wa.readyMtx.RUnlock()

	if !wa.ready {
		wa.logger.Info("WhatsAppClient is not ready")
	}

	return wa.ready
}

func (wa *WhatsAppClient) QrCodeImageHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}

	auth := r.URL.Query().Get("auth")
	if auth != wa.config.AuthToken {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	if wa.qrCode == nil || *wa.qrCode == "" {
		http.Error(w, "QR code not available", http.StatusNotFound)
		return
	}

	png, err := qrcode.Encode(*wa.qrCode, qrcode.Highest, 1024)

	if err != nil {
		http.Error(w, "Failed to generate QR code", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "image/png")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(png)
}
