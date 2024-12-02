package wa

import (
	"context"
	"fmt"
	"net/http"
	"os"

	"github.com/kotrzina/keg-scale/pkg/config"
	"github.com/sirupsen/logrus"
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
	store    *store.Device
	handlers []EventHandler

	config *config.Config
	ctx    context.Context
	logger *logrus.Logger
}

type EventHandler struct {
	MatchFunc  func(msg string) bool
	HandleFunc func(from, msg string) error // from = sender ID
}

func New(ctx context.Context, conf *config.Config, logger *logrus.Logger) *WhatsAppClient {
	customLogger := createLogger(logger)
	container, err := sqlstore.New("postgres", conf.DBString, customLogger)
	if err != nil {
		logger.Fatalf("Failed to create container: %v", err)
	}

	// If you want multiple sessions, remember their JIDs and use .GetDevice(jid) or .GetAllDevices() instead.
	deviceStore, err := container.GetFirstDevice()
	if err != nil {
		logger.Fatalf("Failed to get device: %v", err)
	}

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

	wa := &WhatsAppClient{
		client:   client,
		store:    deviceStore,
		handlers: []EventHandler{},

		config: conf,
		ctx:    ctx,
		logger: logger,
	}

	client.AddEventHandler(wa.eventHandler)
	return wa
}

func (wa *WhatsAppClient) RegisterEventHandler(handler EventHandler) {
	wa.handlers = append(wa.handlers, handler)
}

func (wa *WhatsAppClient) eventHandler(evt interface{}) {
	switch v := evt.(type) {
	case *events.Message:
		wa.handleIncomingMessage(v)
	}
}

func (wa *WhatsAppClient) handleIncomingMessage(msg *events.Message) {
	if msg.Message.Conversation == nil {
		return
	}
	text := *msg.Message.Conversation
	from := msg.Info.MessageSource.Chat.User // we want to replay to the same chat

	wa.logger.Infof(
		"received message in chat %s@%s from %s@%s: %s",
		msg.Info.MessageSource.Chat.User,
		msg.Info.MessageSource.Chat.Server,
		msg.Info.MessageSource.Sender.User,
		msg.Info.MessageSource.Sender.Server,
		text,
	)

	if len(text) > 20 {
		return // do not process long messages
	}

	for _, handler := range wa.handlers {
		if handler.MatchFunc(text) {
			if err := handler.HandleFunc(from, text); err != nil {
				wa.logger.Errorf("Failed from handle message: %v", err)
			}
			continue // do not process other handlers
		}
	}
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

func (wa *WhatsAppClient) SendImage(to, caption, imagePath string) error {
	if !wa.client.IsConnected() {
		wa.logger.Errorf("Not connected to WhatsAppClient")
		return fmt.Errorf("not connected to WhatsAppClient")
	}

	imageBytes, err := os.ReadFile(imagePath)
	if err != nil {
		return fmt.Errorf("failed to read image: %w", err)
	}

	imgResp, err := wa.client.Upload(wa.ctx, imageBytes, whatsmeow.MediaImage)
	if err != nil {
		return fmt.Errorf("failed to upload image: %w", err)
	}
	wa.logger.Infof("Image uploaded: %s", imgResp.URL)

	imageMsg := &waE2E.ImageMessage{
		Caption:  proto.String(caption),
		Mimetype: proto.String(http.DetectContentType(imageBytes)),

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
	if !wa.client.IsConnected() {
		wa.logger.Errorf("Not connected to WhatsAppClient")
		return fmt.Errorf("not connected to WhatsAppClient")
	}

	resp, err := wa.client.SendMessage(wa.ctx, wa.buildJid(to), msg)
	if err != nil {
		return fmt.Errorf("failed to send message: %w", err)
	}

	wa.logger.Infof("Message sent: %s", resp.ID)
	return err
}
