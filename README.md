# IOT KEG SCALE

```
!!! The project is in the early development stage !!!
```

![workflow](https://github.com/kotrzina/keg-scale/actions/workflows/go.yaml/badge.svg)
![workflow](https://github.com/kotrzina/keg-scale/actions/workflows/docker-publish.yml/badge.svg)

## Equipment:

- Scale - LESAK PS-B, 150kg/50g, 350x350mm
- Arduino UNO
- RS232 to TTL, module with MAX3232
- Clock display TM1637

## How it supposed to work:

1. Scale communicate with Arduino using SerialPort - it continuously sends a current value.
1. Arduino does not process all messages -> actual value is processed in Arduino every ~5 seconds.
1. Arduino sends actual value over HTTP to Backend service written in GoLang.
1. Backend processes values and expose them as metrics for Prometheus.
1. Mr. Botka informs about significat event using WhatsApp

## Components:

- `firmware` - Firmware for Arduino
- `backend` - Backend service written in GoLang
- `frontend` - Frontend service written in React

!! Backend and frontend are packed together in a single Docker container.
