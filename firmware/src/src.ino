#include "httpConnection.h"
#include <SoftwareSerial.h>


SoftwareSerial scale(2, 3); // RX, TX are not switched
HttpConnection conn;

void setup()
{
    Serial.begin(9600);
    scale.begin(9600);
    
    pinMode(LED_BUILTIN, OUTPUT);
    pinMode(8, OUTPUT); // error - red light
    pinMode(9, OUTPUT); // success - green light

    digitalWrite(8, HIGH);
    digitalWrite(9, HIGH);
    delay(150);
    digitalWrite(8, LOW);
    digitalWrite(9, LOW);
    delay(150);
    digitalWrite(8, HIGH);
    digitalWrite(9, HIGH);
    delay(150);
    digitalWrite(8, LOW);
    digitalWrite(9, LOW);

    conn.setup();
    conn.push("ping", "");
    if (conn.success()) {
        digitalWrite(9, HIGH);
        delay(150);
        digitalWrite(9, LOW);
        delay(150);
        digitalWrite(9, HIGH);
        delay(150);
        digitalWrite(9, LOW);
        delay(150);
    }
}

int value = 1; // imposible random value - always send value after restart
unsigned long lastMeasurement = 0;
unsigned long lastPing = 0;

void loop()
{
    long ts = millis() / 1000;

    // send ping every minute
    if (lastPing + 60 < ts) {
        conn.push("ping", "");
        signal(conn.success());
        lastPing = ts;
    }

    // we want to read value from the scale every 5 seconds
    // in reality it's like 6 seconds
    // but it's fine because we are going to process the message every 30 seconds on backend
    if (lastMeasurement + 5 < ts) {
        readScale();
        lastMeasurement = ts;
    }
    
    while (scale.available()) {
        // if we are not reading we still want to process messages sent from the scale
        // and ignore them ofc
        scale.read();
    }

    delay(500);
}

// data are sent as ascii characters
// sequence bytes:
// [0] - :
// [1] - W
// [2] - minus sign (0x2D) or space
// [3] - first digin - hundreds
// [4] - second digin - tens
// [5] - third digin - units 
// [6] - decimal separator - always the same
// [7] - tenths
// [8] - hundredths
// [9] - first letter of unit - preferably kg - so it's k
// [10-14] - some other shit we don't care about

// scale sends the sequence 4 times in row for some unknown reason
// the last sequence is somehow corrupted but we can detect it using k in 9th byte - it has to be there
// the we can skip the rest
void readScale() {
    char c;     // current character
    char prev;  // character from the last iteration
    bool reading = true;

    while (scale.available() && reading) {
        c = scale.read();

        if (prev == 0x3a && c == 0x57) { // begining of the sequence
            long n = 0;
            int sign = scale.read();                // minus sign

            n += calcValue(scale.read()) * 100000;  // hundreds
            n += calcValue(scale.read()) * 10000;   // tens
            n += calcValue(scale.read()) * 1000;    // units
            scale.read();                           // separator - ignore
            n += calcValue(scale.read()) * 100;     // tenths
            n += calcValue(scale.read()) * 10;      // hundredths
            if (sign == 0x2D) { 
                n *= -1;
            }
            int p = scale.read();                   // first letter of unit

            if (p == 0x6b) { // there needs to be "k" - no idea why
                if (n != value) { // the value is not the same as in the last measurement
                    conn.push("push", String(n)); // send value to backend
                    if (conn.success()) {
                        value = n; // update global value
                        signalSuccess();
                    } else {
                        signalError();
                    }
                    
                }
                reading = false;
            }
        }

        prev = c;
    }
}

// values are sent as characters
// the number is represented in ascii as number 48-57
// for space we return 0
// for other values we want to return read int value (not ascii value)
long calcValue(long c) {
    if (c == 0x20) {
        return 0;
    }

    return c - 48;
}

void signal(bool success) {
    if (success) {
        signalSuccess();
    } else {
        signalError();
    }
}

void signalError() {
    digitalWrite(8, HIGH);
    digitalWrite(9, LOW);
}

void signalSuccess() {
    digitalWrite(9, HIGH);
    digitalWrite(8, LOW);
}

