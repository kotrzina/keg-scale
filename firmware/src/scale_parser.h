#ifndef SCALE_PARSER_H
#define SCALE_PARSER_H

#include <stdint.h>

// Parser for the ASCII stream the LESAK PS-B scale sends over RS232.
//
// Plain C++ without any Arduino dependency, so it can be unit tested on the
// host (see test/scale_parser_test.cpp, `make test`).
//
// A frame looks like this (byte offsets):
//   [0]    ':'
//   [1]    'W'
//   [2]    sign - '-' (0x2D) for negative values, anything else for positive
//   [3-5]  hundreds, tens, units of kg - ASCII digits, leading zeros are spaces
//   [6]    decimal separator - ignored
//   [7-8]  tenths, hundredths of kg
//   [9]    first letter of the unit - always 'k'
//   [10-]  trailing bytes we do not care about
//
// The scale repeats every frame several times in a row and the last copy is
// usually corrupted; requiring the 'k' and valid digits drops those. On top of
// that a value is only reported once two consecutive frames agree, so a frame
// glued together from two halves (e.g. after a UART buffer overflow while the
// device was busy with HTTP) can never be reported as a reading.
//
// Values are in grams with a 10 g resolution (12.34 kg -> 12340), which is the
// unit the backend expects.
class ScaleParser {
 public:
  // Feeds one received byte. Returns true when this byte completed a frame
  // whose value matches the previous frame; value() then holds the reading.
  bool feed(uint8_t c) {
    if (pos_ == 0) {
      // waiting for the ":W" header
      if (prev_ == ':' && c == 'W') {
        pos_ = 1;
        negative_ = false;
        acc_ = 0;
      }
      prev_ = c;
      return false;
    }

    prev_ = c;
    int field = pos_++;  // 1 = sign, 2-4 = integer digits, 5 = separator, 6-7 = decimals, 8 = unit

    if (field == 1) {
      negative_ = (c == '-');
      return false;
    }

    if (field == 5) {
      return false;  // decimal separator
    }

    if (field == 8) {
      pos_ = 0;
      if (c != 'k') {
        return false;  // corrupted frame
      }
      long v = negative_ ? -acc_ : acc_;
      bool confirmed = hasCandidate_ && candidate_ == v;
      candidate_ = v;
      hasCandidate_ = true;
      if (confirmed) {
        value_ = v;
      }
      return confirmed;
    }

    // digit fields 2-4 and 6-7
    int digit;
    if (c == ' ') {
      digit = 0;
    } else if (c >= '0' && c <= '9') {
      digit = c - '0';
    } else {
      // Not a digit: the frame is broken. Restart, but keep the byte as
      // "previous" so a ':' here can still start the next header.
      pos_ = 0;
      return false;
    }
    acc_ = acc_ * 10 + digit;
    if (field == 7) {
      acc_ *= 10;  // hundredths of kg -> grams
    }
    return false;
  }

  // Last confirmed reading in grams.
  long value() const { return value_; }

 private:
  int pos_ = 0;  // 0 = looking for the header, otherwise the next field
  uint8_t prev_ = 0;
  bool negative_ = false;
  long acc_ = 0;

  bool hasCandidate_ = false;
  long candidate_ = 0;
  long value_ = 0;
};

#endif
