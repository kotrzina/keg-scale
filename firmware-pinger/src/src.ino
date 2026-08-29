#include <Arduino.h>
#include <WiFi.h>
#include <WiFiMulti.h>
#include <WiFiClientSecure.h>
#include <WiFiManager.h>
#include <HTTPClient.h>
#include <Preferences.h>
#include <esp_task_wdt.h>
#include "config.h"

// --- Pin definitions ---
#define LED_PIN 8

// --- Constants ---
#define MAX_NETWORKS          20
#define WIFI_TIMEOUT_MS       15000
#define WIFI_RETRY_MS         10000
#define PORTAL_TIMEOUT_S      300
#define HTTP_TIMEOUT_MS       10000
#define DRD_WINDOW_MS         3000
#define LED_FAST_MS           200
#define LED_SLOW_MS           1000

static const char* AP_NAME = "Pinger-Setup";
static const char* PREFS_NS = "pinger";

// --- Globals ---
WiFiMulti wifiMulti;
Preferences prefs;

String backendUrl;
String authToken;

unsigned long lastPingTime = 0;
unsigned long lastWifiRetry = 0;
uint32_t pingCounter = 0;
bool lastPingOk = false;

// Portal parameters — globals so the WiFiManager save callback can reach them
WiFiManagerParameter* paramUrl = nullptr;
WiFiManagerParameter* paramToken = nullptr;

// ============================================================
// LED status indication
// ============================================================

void ledSolid(bool on) {
  // ESP32-C3 Super Mini LED is active LOW
  digitalWrite(LED_PIN, on ? LOW : HIGH);
}

void ledBlink(unsigned long intervalMs, unsigned long durationMs) {
  unsigned long start = millis();
  while (millis() - start < durationMs) {
    digitalWrite(LED_PIN, LOW);
    delay(intervalMs / 2);
    digitalWrite(LED_PIN, HIGH);
    delay(intervalMs / 2);
  }
}

// Non-blocking blink for use inside loop()
void ledBlinkOnce(unsigned long intervalMs) {
  static unsigned long lastToggle = 0;
  static bool state = false;
  if (millis() - lastToggle >= intervalMs / 2) {
    state = !state;
    digitalWrite(LED_PIN, state ? LOW : HIGH);
    lastToggle = millis();
  }
}

// ============================================================
// Backend configuration storage (Preferences / NVS)
// ============================================================

// Trims the value and lowercases the URL scheme, so that a phone keyboard
// autocapitalizing the portal field into "Https://..." still picks the TLS
// client, and a stray space does not silently break auth or URL parsing.
String normalizeUrl(String url) {
  url.trim();
  int sep = url.indexOf("://");
  if (sep > 0) {
    String scheme = url.substring(0, sep);
    scheme.toLowerCase();
    url = scheme + url.substring(sep);
  }
  return url;
}

void loadBackendConfig() {
  prefs.begin(PREFS_NS, true);
  backendUrl = normalizeUrl(prefs.getString("url", DEFAULT_BACKEND_URL));
  authToken = prefs.getString("token", DEFAULT_AUTH_TOKEN);
  authToken.trim();
  prefs.end();
  LOGF("Backend URL: %s\n", backendUrl.c_str());
  LOGF("Auth token: %s\n", authToken.length() > 0 ? "(set)" : "(empty)");
}

void saveBackendConfig(const char* url, const char* token) {
  String cleanUrl = normalizeUrl(String(url));
  String cleanToken = String(token);
  cleanToken.trim();

  prefs.begin(PREFS_NS, false);
  prefs.putString("url", cleanUrl);
  prefs.putString("token", cleanToken);
  prefs.end();
  backendUrl = cleanUrl;
  authToken = cleanToken;
  LOGF("  Saved backend URL: %s\n", cleanUrl.c_str());
  LOGF("  Saved auth token: %s\n", cleanToken.length() > 0 ? "(set)" : "(empty)");
}

// ============================================================
// WiFi credential storage (Preferences / NVS)
// ============================================================

int getStoredCount() {
  prefs.begin(PREFS_NS, true);
  int count = prefs.getInt("count", 0);
  prefs.end();
  return count;
}

void loadCredentials() {
  prefs.begin(PREFS_NS, true);
  int count = prefs.getInt("count", 0);
  for (int i = 0; i < count && i < MAX_NETWORKS; i++) {
    String ssidKey = "ssid" + String(i);
    String passKey = "pass" + String(i);
    String ssid = prefs.getString(ssidKey.c_str(), "");
    String pass = prefs.getString(passKey.c_str(), "");
    if (ssid.length() > 0) {
      wifiMulti.addAP(ssid.c_str(), pass.c_str());
      LOGF("  Loaded network %d: %s\n", i, ssid.c_str());
    }
  }
  prefs.end();
}

void saveCredential(const char* ssid, const char* password) {
  prefs.begin(PREFS_NS, false);
  int count = prefs.getInt("count", 0);

  // Check if this SSID already exists — update in place
  for (int i = 0; i < count && i < MAX_NETWORKS; i++) {
    String ssidKey = "ssid" + String(i);
    String stored = prefs.getString(ssidKey.c_str(), "");
    if (stored == ssid) {
      String passKey = "pass" + String(i);
      prefs.putString(passKey.c_str(), password);
      LOGF("  Updated existing network: %s (slot %d)\n", ssid, i);
      prefs.end();
      return;
    }
  }

  // Find the slot: next available, or oldest (circular buffer)
  int slot = count < MAX_NETWORKS ? count : (count % MAX_NETWORKS);
  String ssidKey = "ssid" + String(slot);
  String passKey = "pass" + String(slot);
  prefs.putString(ssidKey.c_str(), ssid);
  prefs.putString(passKey.c_str(), password);
  prefs.putInt("count", count + 1);

  LOGF("  Saved network: %s (slot %d)\n", ssid, slot);
  prefs.end();
}

// ============================================================
// Double Reset Detection
// ============================================================

bool checkDoubleReset() {
  prefs.begin(PREFS_NS, false);
  bool drdFlag = prefs.getBool("drd", false);

  if (drdFlag) {
    // Second reset within the window — open the config portal.
    // Stored networks are kept: re-entering a known SSID updates its password.
    prefs.putBool("drd", false);
    prefs.end();
    LOGLN("Double reset detected! Starting config portal...");
    return true;
  }

  // First reset — set flag, wait for window, then clear
  prefs.putBool("drd", true);
  prefs.end();

  LOGLN("Waiting for double-reset window (3s)...");
  ledBlink(LED_FAST_MS, DRD_WINDOW_MS);

  prefs.begin(PREFS_NS, false);
  prefs.putBool("drd", false);
  prefs.end();
  LOGLN("Single reset — proceeding normally");
  return false;
}

// ============================================================
// WiFi connection
// ============================================================

void initDefaultCredentials() {
  if (DEFAULT_NETWORK_COUNT == 0 || getStoredCount() > 0) {
    return;
  }
  LOGLN("NVS empty — loading default networks from secrets.h");
  for (int i = 0; i < DEFAULT_NETWORK_COUNT; i++) {
    saveCredential(DEFAULT_NETWORKS[i].ssid, DEFAULT_NETWORKS[i].password);
  }
}

void scanNetworks() {
  LOGLN("Scanning for WiFi networks...");
  int n = WiFi.scanNetworks();
  if (n == 0) {
    LOGLN("  No networks found!");
  } else {
    LOGF("  Found %d networks:\n", n);
    for (int i = 0; i < n; i++) {
      LOGF("    %2d: %-32s  RSSI: %d dBm  Ch: %d  %s\n",
                    i + 1,
                    WiFi.SSID(i).c_str(),
                    WiFi.RSSI(i),
                    WiFi.channel(i),
                    WiFi.encryptionType(i) == WIFI_AUTH_OPEN ? "open" : "encrypted");
    }
  }
  WiFi.scanDelete();
}

bool connectWiFi() {
  initDefaultCredentials();

  scanNetworks();

  LOGLN("Loading stored WiFi credentials...");
  loadCredentials();

  int count = getStoredCount();
  if (count == 0) {
    LOGLN("No stored credentials");
    return false;
  }

  LOGF("Connecting via WiFiMulti (%d stored networks, timeout %ds)...\n",
                count, WIFI_TIMEOUT_MS / 1000);
  unsigned long start = millis();
  int attempts = 0;
  while (millis() - start < WIFI_TIMEOUT_MS) {
    wl_status_t status = (wl_status_t)wifiMulti.run();
    attempts++;
    if (status == WL_CONNECTED) {
      LOGF("Connected to: %s (IP: %s, RSSI: %d dBm) after %d attempts\n",
                    WiFi.SSID().c_str(),
                    WiFi.localIP().toString().c_str(),
                    WiFi.RSSI(),
                    attempts);
      return true;
    }
    LOGF("  Attempt %d: status=%d elapsed=%lums\n", attempts, status, millis() - start);
    ledBlinkOnce(LED_FAST_MS);
    delay(500);
  }

  LOGF("WiFiMulti connection timed out after %d attempts\n", attempts);
  return false;
}

// ============================================================
// Captive portal — WiFi credentials + backend URL + auth token
// ============================================================

void onSaveParams() {
  if (paramUrl != nullptr && paramToken != nullptr) {
    saveBackendConfig(paramUrl->getValue(), paramToken->getValue());
  }
}

// Blocks until the portal is done, then reboots so setup() runs the normal
// connect flow with the freshly stored configuration.
void runConfigPortal() {
  LOGLN("Starting captive portal...");
  ledSolid(false);

  WiFiManager wm;
  wm.setConfigPortalTimeout(PORTAL_TIMEOUT_S);
  wm.setBreakAfterConfig(true);
  wm.setTitle("Kozel pinger");
  std::vector<const char*> menu = {"wifi", "param", "info", "sep", "restart", "exit"};
  wm.setMenu(menu);

  WiFiManagerParameter pUrl("burl", "Backend URL", backendUrl.c_str(), 160);
  WiFiManagerParameter pToken("btoken", "Auth token", authToken.c_str(), 128);
  wm.addParameter(&pUrl);
  wm.addParameter(&pToken);
  paramUrl = &pUrl;
  paramToken = &pToken;

  wm.setSaveParamsCallback(onSaveParams);
  wm.setAPCallback([](WiFiManager* /*wm*/) {
    LOGF("Captive portal active — connect to AP: %s\n", AP_NAME);
  });

  bool connected = wm.startConfigPortal(AP_NAME);

  if (connected) {
    String ssid = WiFi.SSID();
    if (ssid.length() == 0) {
      ssid = wm.getWiFiSSID();
    }
    String pass = wm.getWiFiPass();
    LOGF("Portal connected to: %s\n", ssid.c_str());
    if (ssid.length() > 0) {
      saveCredential(ssid.c_str(), pass.c_str());
    }
  } else {
    LOGLN("Portal closed without a WiFi connection");
  }

  // Make sure params entered on the WiFi page are persisted even if the
  // save callback did not fire for them.
  onSaveParams();

  paramUrl = nullptr;
  paramToken = nullptr;

  LOGLN("Restarting to apply configuration...");
  delay(1000);
  ESP.restart();
}

// ============================================================
// Watchdog
// ============================================================

// Subscribes loop() to the task watchdog. It is fed only after a successful
// request, so anything that stops the pings — a wedged WiFi stack, an
// unreachable backend, or a loop() blocked inside a library call — reboots the
// device once WDT_TIMEOUT_MS elapses.
void startWatchdog() {
  esp_task_wdt_config_t config = {};
  config.timeout_ms = WDT_TIMEOUT_MS;
  config.idle_core_mask = 0; // watch loop() only, not the idle task
  config.trigger_panic = true;

  // The IDF already starts the TWDT during boot, so init() reports
  // ESP_ERR_INVALID_STATE and the timeout has to be applied via reconfigure().
  esp_err_t err = esp_task_wdt_init(&config);
  if (err == ESP_ERR_INVALID_STATE) {
    err = esp_task_wdt_reconfigure(&config);
  }
  if (err != ESP_OK) {
    LOGF("Could not configure task watchdog: %d\n", err);
    return;
  }

  err = esp_task_wdt_add(NULL);
  if (err != ESP_OK) {
    LOGF("Could not subscribe loop to task watchdog: %d\n", err);
    return;
  }

  LOGF("Task watchdog armed (%d ms)\n", WDT_TIMEOUT_MS);
}

// ============================================================
// Ping
// ============================================================

// Sends "ping|<counter>|<rssi>|0" to the backend. The backend opens the pub
// on the first ping and closes it again once the pings stop.
bool sendPing() {
  if (backendUrl.length() == 0) {
    LOGLN("No backend URL configured — skipping ping");
    return false;
  }

  String body = "ping|";
  body += String(++pingCounter);
  body += "|";
  body += String(WiFi.RSSI());
  body += "|0";

  WiFiClientSecure secureClient;
  WiFiClient plainClient;
  HTTPClient http;

  bool begun;
  if (backendUrl.startsWith("https://")) {
    secureClient.setInsecure(); // skip certificate verification
    begun = http.begin(secureClient, backendUrl);
  } else {
    begun = http.begin(plainClient, backendUrl);
  }

  if (!begun) {
    LOGF("[%lu] Could not parse backend URL: %s\n", millis(), backendUrl.c_str());
    return false;
  }

  http.setTimeout(HTTP_TIMEOUT_MS);
  http.setUserAgent(USER_AGENT);
  http.addHeader("Content-Type", "text/plain");
  http.addHeader("Authorization", authToken);

  int httpCode = http.POST(body);
  String response = httpCode > 0 ? http.getString() : String("");
  http.end();

  if (httpCode == HTTP_CODE_OK) {
    LOGF("[%lu] Ping #%u ok (RSSI %d dBm, response: %s)\n",
         millis(), pingCounter, WiFi.RSSI(), response.c_str());
    return true;
  }

  LOGF("[%lu] Ping #%u failed: HTTP %d\n", millis(), pingCounter, httpCode);
  return false;
}

// ============================================================
// Arduino entry points
// ============================================================

void setup() {
  LOG_BEGIN();
  LOGLN("\n=== Kozel pinger starting ===");
  LOGF("Reset reason: %d\n", esp_reset_reason());

  pinMode(LED_PIN, OUTPUT);
  ledSolid(false);

  loadBackendConfig();

  bool doubleReset = checkDoubleReset();

  // Without a URL or a token the device could only ever get 401s, so make it
  // ask for configuration instead of rebooting forever.
  if (doubleReset || backendUrl.length() == 0 || authToken.length() == 0 || !connectWiFi()) {
    runConfigPortal(); // never returns — reboots
  }

  ledSolid(true);

  // Armed only here: connectWiFi() and the captive portal are far slower than
  // the watchdog window, and the portal path reboots instead of returning.
  startWatchdog();

  // Ping right away so the pub opens as soon as the device gets power
  lastPingTime = millis() - PING_INTERVAL_MS;
}

void updateLed() {
  if (WiFi.status() != WL_CONNECTED) {
    ledBlinkOnce(LED_FAST_MS); // fast blink — disconnected
  } else if (lastPingOk) {
    ledSolid(true); // solid — connected and the backend is answering
  } else {
    ledBlinkOnce(LED_SLOW_MS); // slow blink — connected but pings are failing
  }
}

void loop() {
  unsigned long now = millis();

  // 1. Reconnect if the WiFi dropped
  if (WiFi.status() != WL_CONNECTED && now - lastWifiRetry >= WIFI_RETRY_MS) {
    lastWifiRetry = now;
    wifiMulti.run();
  }

  // 2. Ping the backend
  if (now - lastPingTime >= PING_INTERVAL_MS) {
    lastPingTime = now;
    if (WiFi.status() == WL_CONNECTED) {
      lastPingOk = sendPing();
    } else {
      lastPingOk = false;
      LOGF("[%lu] WiFi disconnected — skipping ping\n", now);
    }

    // Feed the watchdog only on success, so that a wedged WiFi stack or an
    // unreachable backend eventually reboots the device.
    if (lastPingOk) {
      esp_task_wdt_reset();
    }
  }

  // 3. Status LED
  updateLed();
  delay(LOOP_DELAY_MS);
}
