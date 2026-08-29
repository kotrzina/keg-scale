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

// --- Timing ---
#define PING_INTERVAL_MS 30000

// Idle time between loop() iterations. delay() yields to the FreeRTOS idle
// task, so a longer value means less pointless spinning. Do not raise it far
// above LED_FAST_MS/2 (100 ms) or the fast blink stretches out until it is
// no longer distinguishable from the slow one.
#define LOOP_DELAY_MS 100

// Task watchdog: at least one successful request must reach the server within
// this window, otherwise the device panics and reboots. Must comfortably
// exceed PING_INTERVAL_MS so that a couple of retries fit inside it.
#define WDT_TIMEOUT_MS 300000
#define USER_AGENT       "pinger/main"

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
