#include "httpConnection.h"

HttpConnection::HttpConnection()
{
}

void HttpConnection::setup()
{
    connect();

    client = new WiFiClient();
    client->setTimeout(5000);
    backend = new HttpClient(*client, BACKEND_IP, BACKEND_PORT);   
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

void HttpConnection::push(String message, String value)
{
    connect();

    // format:
    // messageType|messageId|rssi|value
    message.concat("|");
    message.concat(String(millis() / 1000));
    message.concat("|");
    long rssi = WiFi.RSSI();
    message.concat(String(rssi));
    message.concat("|");
    message.concat(String(value));
    
    backend->beginRequest();
    backend->setHttpResponseTimeout(5000);
    backend->setTimeout(5000);
    backend->post("/api/scale/push");
    backend->sendHeader(HTTP_HEADER_CONTENT_TYPE, "text/plain");
    backend->sendHeader(HTTP_HEADER_CONTENT_LENGTH, message.length());
    backend->sendHeader("Authorization", BACKEND_PASSWORD);
    backend->sendHeader("Host", BACKEND_HOST);
    backend->write((const byte*)message.c_str(), message.length());
    int statusCode = backend->responseStatusCode();
    String response = backend->responseBody();
    backend->endRequest();

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
