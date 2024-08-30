#include "httpConnection.h"

HttpConnection::HttpConnection()
{
}

void HttpConnection::setup(char *ssid, char *pass, char *backendHost, int backendPort, char *backendPassword)
{
    settings.ssid = ssid;
    settings.pass = pass;
    settings.backendHost = backendHost;
    settings.backendPort = backendPort;
    settings.backendPassword = backendPassword;

    connect();

    client = new WiFiClient();
    client->setTimeout(5000);
    sslClient = new WiFiSSLClient();
    backend = new HttpClient(*client, settings.backendHost, settings.backendPort);   
}

void HttpConnection::connect()

{
    status = WiFi.status();
    if (status != WL_CONNECTED)
    {
        status = WiFi.begin(settings.ssid, settings.pass);

        while (status != WL_CONNECTED)
        {
            status = WiFi.begin(settings.ssid, settings.pass);
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
    backend->sendHeader("Authorization", settings.backendPassword);
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
    backend->sendHeader("Authorization", settings.backendPassword);
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
