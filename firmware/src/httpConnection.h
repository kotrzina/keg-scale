#ifndef http_connection_h
#define http_connection_h

#include <WiFiNINA.h>
#include <ArduinoHttpClient.h>
#include "secrets.h"

class HttpConnection
{
public:
    HttpConnection();
    void setup();
    void connect();

    void push(String message, String value);
    
    bool success();
    int getCode();
    int getBeersLeft();

private:
    int status = WL_IDLE_STATUS; // wifi status
    WiFiSSLClient *sslClient;
    WiFiClient *client;
    HttpClient *backend;

    int lastCode = 0;
    int lastResponse = 0;
};

#endif // http_connection_h