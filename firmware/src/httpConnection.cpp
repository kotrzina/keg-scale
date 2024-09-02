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

void HttpConnection::sendValue(String value)
{
    connect();
    
    backend->beginRequest();
    backend->setHttpResponseTimeout(5000);
    backend->setTimeout(5000);
    backend->post("/api/scale/keg");
    backend->sendHeader(HTTP_HEADER_CONTENT_TYPE, "text/plain");
    backend->sendHeader(HTTP_HEADER_CONTENT_LENGTH, value.length());
    backend->sendHeader("Authorization", BACKEND_PASSWORD);
    backend->sendHeader("Host", BACKEND_HOST);
    backend->write((const byte*)value.c_str(), value.length());
    int statusCode = backend->responseStatusCode();
    backend->endRequest();
    
    lastCode = statusCode;
}

void HttpConnection::sendPing()
{
    String value = "ping";
    connect();
    
    backend->beginRequest();
    backend->setHttpResponseTimeout(5000);
    backend->setTimeout(5000);
    backend->post("/api/scale/ping");
    backend->sendHeader(HTTP_HEADER_CONTENT_TYPE, "text/plain");
    backend->sendHeader(HTTP_HEADER_CONTENT_LENGTH, value.length());
    backend->sendHeader("Authorization", BACKEND_PASSWORD);
    backend->sendHeader("Host", BACKEND_HOST);
    backend->write((const byte*)value.c_str(), value.length());
    int statusCode = backend->responseStatusCode();
    backend->endRequest();

    lastCode = statusCode;
}

bool HttpConnection::success()
{
    return lastCode == 200;
}

int HttpConnection::getCode()
{
    return lastCode;
}
