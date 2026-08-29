# Pinger

ESP32-C3 firmware that does nothing but tell the backend "the pub has power".

The device is powered from the pub's mains. When the pub opens, it gets power,
boots, and starts sending a ping every 30 s. When the pub closes, the power goes
away and the pings stop.

## How it works

Every 30 s the device sends:

```
POST <backend URL>
Authorization: <auth token>
Content-Type: text/plain

ping|<counter>|<rssi>|0
```

Default URL is `https://pub.kotrzina.cz/api/scale/push` — the same endpoint the
scale uses. On the backend side:

- The first ping calls `scale.Ping()`, which opens the pub and dispatches the
  WhatsApp open message. `shouldSendOpen()` prevents spam (max once per 12 h, and
  only if the pub was closed for at least 3 h), so pinging every 30 s is safe.
- `Recheck()` closes the pub again once no ping arrives for 10 minutes
  (`okLimit`), which happens automatically when the device loses power.

The first ping is sent immediately after boot, so the pub opens as soon as the
device is powered on.

## Hardware

- [ESP32-C3 Super Mini](https://www.laskakit.cz/esp32-c3-super-mini-wifi-bluetooth-modul/)
- Any 5 V USB power supply plugged into a socket that is only live when the pub is open

No sensors, no relay — the presence of power *is* the signal.

## Configuration

Everything is configured from a captive portal, nothing has to be compiled in.

1. Get the device into the portal — see [Forcing the portal](#forcing-the-portal).
2. Connect to the `Pinger-Setup` access point.
3. The portal offers:
   - **WiFi** — pick a network and enter the password
   - **Setup** — `Backend URL` and `Auth token`

   Both `http://` and `https://` URLs work, with an optional port
   (`http://192.168.1.10:8080/api/scale/push`). HTTPS uses `setInsecure()`, so
   the certificate is not verified. The scheme is lowercased and the values are
   trimmed on save.
4. Save. The device reboots and starts pinging.

Configuration lives in NVS (`Preferences`, namespace `pinger`):

- up to 20 Wi-Fi networks, tried via `WiFiMulti`; re-entering a known SSID
  updates its password instead of adding a duplicate
- `url` — backend endpoint
- `token` — value sent in the `Authorization` header

### Forcing the portal

**Double reset within 3 seconds.** Press RESET, wait for the blue LED to start
blinking fast, and press RESET again while it is still blinking. The fast blink
*is* the window (`DRD_WINDOW_MS`, 3 s). The AP comes up and the LED goes dark.

**Power cycling works the same way**, which is what you need for a device wired
into a socket with no reachable button: cut the power, restore it, then cut and
restore it again within those three seconds.

`checkDoubleReset()` writes a `drd` flag to NVS on boot, blinks for three
seconds and then clears it. A boot that finds the flag still set opens the
portal instead of pinging — and because the flag lives in NVS, losing power
preserves it exactly like a reset does.

The portal also opens **on its own** when there are no stored networks, no
backend URL or no auth token, or when Wi-Fi fails to connect within 15 s.

Two things worth knowing:

- Stored networks are **not** wiped, so this is also the way to change just the
  URL or the token. Re-entering a known SSID updates its password rather than
  adding a duplicate.
- The portal times out after `PORTAL_TIMEOUT_S` (300 s) and the device reboots.
  Nothing is changed if you did not save — and it reopens immediately if the
  URL or token is still missing.

There is no way to force the portal remotely; it needs physical access to the
button or the power.

### Optional pre-seeding

Copy `src/secrets_example.h` to `src/secrets.h` (gitignored) to bake in a
default URL, token and Wi-Fi list, so a fresh device needs no portal at all.
NVS values always win once something has been saved from the portal.

## LED (GPIO8, active LOW)

| Pattern | Meaning |
|---|---|
| Fast blink (200 ms) | Double-reset window, connecting, or Wi-Fi lost |
| Slow blink (1 s) | Wi-Fi connected but pings are failing |
| Solid on | Connected and the backend is answering |
| Off | Captive portal is running |

There is only one controllable LED and it is single colour — the pattern carries
the meaning, not the colour. The second, red LED on the board is hardwired to
3V3 and only says "this thing has power".

Reading it:

- The first three seconds after **every** boot are a fast blink. That is the
  double-reset window, not a fault. If it goes solid afterwards, all is well.
- Fast blink that never stops — cannot join Wi-Fi.
- Slow blink — Wi-Fi is fine, the backend is not answering: wrong token, wrong
  URL, or the server is down.
- A slow blink interrupted by that three-second fast burst every five minutes is
  the watchdog rebooting a device whose pings never succeed.

## Watchdog

`loop()` is subscribed to the ESP-IDF task watchdog, and `esp_task_wdt_reset()`
is called **only after a ping the server answered with 200**. If no successful
request gets through within `WDT_TIMEOUT_MS` (5 minutes), the watchdog panics and
the device reboots, which re-runs the whole connect flow.

Feeding on success rather than every iteration is what makes one mechanism cover
every failure mode: a wedged WiFi stack, an unreachable or misconfigured
backend, and a `loop()` blocked inside a library call all look the same to it.

The watchdog is armed at the end of `setup()` — `connectWiFi()` (15 s) and the
captive portal (300 s) are slower than the window, and the portal path reboots
instead of returning anyway.

## Build

```bash
make install   # install ESP32 core + board manager URL (once)
make deps      # install WiFiManager
make compile
make flash
```

Serial logging is compiled out by default. For a chatty build:

```bash
make flash-debug
make monitor   # 115200 baud
```

`make compile-debug` / `make flash-debug` pass `-DDEBUG_SERIAL`; the same can be
had by uncommenting `#define DEBUG_SERIAL` in `src/config.h`. Without it
`Serial.begin()` is never called and every log statement compiles to nothing.
