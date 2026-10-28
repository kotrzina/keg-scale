# IOT KEG SCALE

```
!!! The project is in the early development stage !!!
```

![workflow](https://github.com/kotrzina/keg-scale/actions/workflows/go.yaml/badge.svg)
![workflow](https://github.com/kotrzina/keg-scale/actions/workflows/docker-publish.yml/badge.svg)

## Equipment:

- Scale - LESAK PS-B, 150kg/50g, 350x350mm
- ESP32-DevKitV1 (30 pin)
- RS232 to TTL, module with MAX3232
- Clock display TM1637

Wiring and setup: [firmware/README.md](firmware/README.md).

## How it supposed to work:

1. Scale communicates with the ESP32 using a serial port - it continuously sends the current value.
1. ESP32 reads every message, but sends a changed value to the backend at most every ~5 seconds.
1. ESP32 sends the value over HTTPS to the Backend service written in GoLang.
1. Backend processes values and expose them as metrics for Prometheus.
1. Mr. Botka informs about significat event using WhatsApp

## Components:

- `firmware` - Firmware for the scale (ESP32)
- `firmware-pinger` - Firmware for the pub power pinger (ESP32-C3)
- `backend` - Backend service written in GoLang
- `frontend` - Frontend service written in React

!! Backend and frontend are packed together in a single Docker container.
