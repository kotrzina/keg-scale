#ifndef CONFIG_H
#define CONFIG_H

// Optional secrets.h (gitignored) can pre-seed the device so that the captive
// portal is not needed on first boot. See secrets_example.h.
#if defined(__has_include)
#  if __has_include("secrets.h")
#    include "secrets.h"
#  endif
#endif

// --- Backend defaults (used until overwritten from the captive portal) ---
#ifndef DEFAULT_BACKEND_URL
#define DEFAULT_BACKEND_URL "https://pub.kotrzina.cz/api/scale/push"
#endif

#ifndef DEFAULT_AUTH_TOKEN
#define DEFAULT_AUTH_TOKEN ""
#endif

// --- Pins (ESP32-DevKitV1, see docs/wiring.svg) ---
// Avoid GPIO6-11 (flash), GPIO0/2/5/12/15 as inputs (strapping), GPIO1/3
// (USB serial) and GPIO34-39 for outputs (input only).
#define SCALE_RX_PIN   16  // UART2 RX <- MAX3232 TTL output
#define DISPLAY_CLK    18  // TM1637 CLK
#define DISPLAY_DIO    19  // TM1637 DIO
#define LED_RED_PIN    25  // last request failed
#define LED_GREEN_PIN  26  // last request succeeded
#define LED_STATUS_PIN 2   // blue LED on the board, active HIGH

// --- Scale serial line ---
#define SCALE_BAUD 9600
// The scale streams continuously and an HTTPS request can block loop() for
// several seconds. A bigger buffer loses less while that happens; whatever
// does get lost is resynchronized by the parser.
#define SCALE_RX_BUFFER 2048

// --- Timing ---
// A changed weight is sent at most this often.
#define MEASURE_INTERVAL_MS 5000

// Heartbeat when the weight does not change. Any successful request counts,
// so a ping only goes out after this long without one. The backend closes
// the pub after 10 minutes without a message.
#define PING_INTERVAL_MS 60000

// Idle time between loop() iterations. Keep it short: the UART is drained
// every iteration and the fast blink needs ~100 ms granularity.
#define LOOP_DELAY_MS 50

// Task watchdog: at least one successful request must reach the server within
// this window, otherwise the device panics and reboots. Must comfortably
// exceed PING_INTERVAL_MS so that a couple of retries fit inside it.
#define WDT_TIMEOUT_MS 300000
#define USER_AGENT     "keg-scale/main"

// TM1637 brightness, 0-7 (the library ORs in the "display on" bit itself).
#define DISPLAY_BRIGHTNESS 7

// --- Serial logging ---
// Off by default so production builds carry no logging. Enable with
// `make compile-debug` / `make flash-debug`, or by uncommenting the line below.
//#define DEBUG_SERIAL

#ifdef DEBUG_SERIAL
#define LOG_BEGIN()  do { Serial.begin(115200); delay(500); } while (0)
#define LOGLN(msg)   Serial.println(msg)
#define LOGF(...)    Serial.printf(__VA_ARGS__)
#else
#define LOG_BEGIN()  do {} while (0)
#define LOGLN(msg)   do {} while (0)
#define LOGF(...)    do {} while (0)
#endif

// --- Default Wi-Fi networks (stored in NVS on first boot when empty) ---
// Define DEFAULT_NETWORKS_LIST in secrets.h as a comma separated list of
// {"ssid", "password"} pairs to pre-seed the device, e.g.:
//   #define DEFAULT_NETWORKS_LIST {"YourWifi", "your-password"}
struct WiFiCredential {
  const char* ssid;
  const char* password;
};

#ifdef DEFAULT_NETWORKS_LIST
static const WiFiCredential DEFAULT_NETWORKS[] = {DEFAULT_NETWORKS_LIST};
static const int DEFAULT_NETWORK_COUNT = sizeof(DEFAULT_NETWORKS) / sizeof(DEFAULT_NETWORKS[0]);
#else
static const WiFiCredential* DEFAULT_NETWORKS = nullptr;
static const int DEFAULT_NETWORK_COUNT = 0;
#endif

#endif
