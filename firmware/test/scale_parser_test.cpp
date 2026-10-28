// Host-side tests for ScaleParser. Run with `make test`.
#include <cstdio>
#include <string>
#include <vector>

#include "../src/scale_parser.h"

static int failures = 0;

// Feeds the whole stream and returns every reading the parser reported.
static std::vector<long> parse(const std::string& stream) {
  ScaleParser p;
  std::vector<long> out;
  for (char c : stream) {
    if (p.feed(static_cast<uint8_t>(c))) {
      out.push_back(p.value());
    }
  }
  return out;
}

static void expect(const char* name, const std::string& stream, const std::vector<long>& want) {
  std::vector<long> got = parse(stream);
  if (got == want) {
    std::printf("ok   %s\n", name);
    return;
  }
  failures++;
  std::printf("FAIL %s: got [", name);
  for (long v : got) std::printf(" %ld", v);
  std::printf(" ], want [");
  for (long v : want) std::printf(" %ld", v);
  std::printf(" ]\n");
}

static std::string frame(const char* body) {
  // body = sign + "ddd.dd" + unit, followed by the trailing bytes the scale sends
  return std::string(":W") + body + "g  \r\n";
}

int main() {
  std::string f12 = frame(" 012.34k");
  std::string fNeg = frame("-  1.50k");
  std::string fZero = frame("   0.00k");

  expect("single frame is not enough", f12, {});
  expect("two identical frames", f12 + f12, {12340});
  expect("four copies, last corrupted", f12 + f12 + f12 + ":W 012.3", {12340, 12340});
  expect("negative value", fNeg + fNeg, {-1500});
  expect("zero with space padding", fZero + fZero, {0});
  expect("max value", frame(" 999.99k") + frame(" 999.99k"), {999990});
  expect("value change needs confirmation", f12 + f12 + fNeg + fNeg, {12340, -1500});
  expect("different consecutive frames", f12 + fNeg + f12 + fNeg, {});
  expect("wrong unit is dropped", frame(" 012.34l") + frame(" 012.34l"), {});
  expect("garbage digit is dropped", frame(" 0x2.34k") + frame(" 0x2.34k"), {});
  expect("leading noise", std::string("\x00\xff::W:") + f12 + f12, {12340});
  expect("header inside a broken frame", std::string(":W 01:W 012.34kg") + f12, {12340});
  expect("spliced frame never confirms", std::string(":W 012") + ".99k" + f12, {});

  if (failures > 0) {
    std::printf("%d test(s) failed\n", failures);
    return 1;
  }
  std::printf("all tests passed\n");
  return 0;
}
