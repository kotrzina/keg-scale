package main

import (
	"fmt"
	"strconv"
	"strings"
)

const (
	PingMessageType = "ping"
	PushMessageType = "push"
)

type ScaleMessage struct {
	MessageType string
	MessageId   uint64 // arduino counter
	Rssi        float64
	Value       float64
}

// ParseScaleMessage parses a message from the scale
// String format: messageType|messageId|rssi|value
func ParseScaleMessage(message string) (ScaleMessage, error) {
	chunks := strings.Split(message, "|")
	if len(chunks) < 4 {
		return ScaleMessage{}, fmt.Errorf("invalid message format")
	}

	messageType := chunks[0]
	if messageType != PingMessageType && messageType != PushMessageType {
		return ScaleMessage{}, fmt.Errorf("invalid request type")
	}

	requestId, err := strconv.ParseUint(chunks[1], 10, 64)
	if err != nil {
		return ScaleMessage{}, fmt.Errorf("could not parse request id")
	}

	// rssi is in all messages
	rssi, err := strconv.ParseFloat(chunks[2], 64)
	if err != nil {
		return ScaleMessage{}, fmt.Errorf("could not parse rssi")
	}

	// value is only in push message
	value := 0.0
	if messageType == PushMessageType {
		value, err = strconv.ParseFloat(chunks[3], 64)
		if err != nil {
			return ScaleMessage{}, fmt.Errorf("could not parse value")
		}
	}

	return ScaleMessage{
		MessageId:   requestId,
		MessageType: messageType,
		Rssi:        rssi,
		Value:       value,
	}, nil
}
