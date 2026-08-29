// Copy to secrets.h (gitignored) to pre-seed a device without using the
// captive portal. Everything here is optional — the portal can set all of it,
// and values saved from the portal always win over these defaults.
//
// Replace every placeholder below with your own values.
#ifndef SECRETS_H
#define SECRETS_H

#define DEFAULT_BACKEND_URL "https://example.com/api/scale/push"
#define DEFAULT_AUTH_TOKEN  "replace-with-your-auth-token"

// Comma separated {"ssid", "password"} pairs, tried in order via WiFiMulti.
#define DEFAULT_NETWORKS_LIST \
  {"YourWifi", "your-wifi-password"}, \
  {"YourBackupWifi", "your-backup-password"}

#endif
