#include "httpConnection.h"

HttpConnection::HttpConnection()
{
}

void HttpConnection::setup()
{
    connect();

    client = new WiFiClient();
    client->setTimeout(5000);
}

void HttpConnection::connect()

{
    status = WiFi.status();
    if (status != WL_CONNECTED)
    {
        status = WiFi.begin(WIFI_SSID, WIFI_PASS);

        while (status != WL_CONNECTED)
        {
            status = WiFi.begin(WIFI_SSID, WIFI_PASS);
            delay(5000);
        }
    }
}

// Resolves BACKEND_HOST via public resolvers instead of the DNS the router hands out over DHCP.
// When resolving fails, the last good address is used, so a DNS hiccup doesn't stop the pushes.
bool HttpConnection::resolveBackend()
{
    WiFi.setDNS(IPAddress(1, 1, 1, 1), IPAddress(8, 8, 8, 8));

    IPAddress ip;
    for (int attempt = 0; attempt < 3; attempt++)
    {
        if (WiFi.hostByName(BACKEND_HOST, ip) == 1 && ip != IPAddress(0, 0, 0, 0))
        {
            backendIp = ip;
            Serial.print("DNS ");
            Serial.print(BACKEND_HOST);
            Serial.print(" -> ");
            Serial.println(backendIp);
            return true;
        }
        delay(1000);
    }

    Serial.print("DNS failed for ");
    Serial.print(BACKEND_HOST);
    if (backendIp == IPAddress(0, 0, 0, 0))
    {
        Serial.println(", no cached address");
        return false;
    }
    Serial.print(", using cached ");
    Serial.println(backendIp);
    return true;
}

void HttpConnection::push(String message, String value)
{
    connect();

    if (!resolveBackend())
    {
        lastCode = 0;
        lastResponse = 0;
        return;
    }
    HttpClient backend(*client, backendIp, BACKEND_PORT);

    // format:
    // messageType|messageId|rssi|value
    message.concat("|");
    message.concat(String(millis() / 1000));
    message.concat("|");
    long rssi = WiFi.RSSI();
    message.concat(String(rssi));
    message.concat("|");
    message.concat(String(value));
    
    backend.beginRequest();
    backend.setHttpResponseTimeout(5000);
    backend.setTimeout(5000);
    backend.post("/api/scale/push");
    backend.sendHeader(HTTP_HEADER_CONTENT_TYPE, "text/plain");
    backend.sendHeader(HTTP_HEADER_CONTENT_LENGTH, message.length());
    backend.sendHeader("Authorization", BACKEND_PASSWORD);
    backend.sendHeader("Host", BACKEND_HOST); // connected by IP, so HttpClient doesn't send it itself
    backend.write((const byte*)message.c_str(), message.length());
    int statusCode = backend.responseStatusCode();
    String response = backend.responseBody();
    backend.endRequest();

    lastCode = statusCode;
    lastResponse = response.toInt(); 
}

bool HttpConnection::success()
{
    return lastCode == 200;
}

int HttpConnection::getCode()
{
    return lastCode;
}

int HttpConnection::getBeersLeft()
{
    return lastResponse;
}
