#!/usr/bin/env python3
"""Generates docs/wiring.svg, the wiring diagram of the keg scale.

Run `make wiring` after changing it. Plain Python 3, no dependencies.
Pin assignment must match src/config.h.
"""

import sys

W, H = 1280, 860

BG = "#f7f6f2"
INK = "#1f2328"
MUTED = "#5b636e"
PCB = "#1d2733"
PCB_EDGE = "#0f151c"
PIN = "#c9a227"
METAL = "#c9ced4"
BLUE_PCB = "#1f5fa8"

C_3V3 = "#e03a3a"
C_GND = "#2a2f36"
C_RX = "#7b4fd0"
C_CLK = "#e09a00"
C_DIO = "#14998a"
C_LED_R = "#d81b60"
C_LED_G = "#2e9d4f"
C_5V = "#f07c1f"

SANS = "Inter, 'Segoe UI', Helvetica, Arial, sans-serif"
MONO = "'JetBrains Mono', 'SFMono-Regular', Menlo, Consolas, monospace"

# ESP32-DevKitV1 (30 pins), drawn with the USB connector on the left and the
# antenna on the right. Lists go left to right.
TOP_ROW = ["VIN", "GND", "D13", "D12", "D14", "D27", "D26", "D25",
           "D33", "D32", "D35", "D34", "VN", "VP", "EN"]
BOTTOM_ROW = ["3V3", "GND", "D15", "D2", "D4", "D16", "D17", "D5",
              "D18", "D19", "D21", "RX0", "TX0", "D22", "D23"]

BOARD_X, BOARD_Y, BOARD_W, BOARD_H = 340, 330, 680, 140
PIN0_X, PITCH = 400, 40
TOP_PIN_Y, BOTTOM_PIN_Y = BOARD_Y + 16, BOARD_Y + BOARD_H - 16

out = []


def add(s):
    out.append(s)


def pin_x(row, name):
    return PIN0_X + row.index(name) * PITCH


def text(x, y, s, size=14, weight=400, fill=INK, anchor="start", family=SANS, extra=""):
    add(f'<text x="{x}" y="{y}" font-family="{family}" font-size="{size}" '
        f'font-weight="{weight}" fill="{fill}" text-anchor="{anchor}"{extra}>{s}</text>')


def wire(points, color, width=4):
    d = "M" + " L".join(f"{x},{y}" for x, y in points)
    add(f'<path d="{d}" fill="none" stroke="{color}" stroke-width="{width}" '
        f'stroke-linecap="round" stroke-linejoin="round"/>')


def hline_with_hops(x1, x2, y, hops, color, width=4, r=7):
    """Horizontal wire from x1 to x2 that jumps over vertical wires at `hops`."""
    lo, hi = min(x1, x2), max(x1, x2)
    d = f"M{lo},{y}"
    for hx in sorted(h for h in hops if lo < h < hi):
        d += f" L{hx - r},{y} A{r},{r} 0 0 1 {hx + r},{y}"
    d += f" L{hi},{y}"
    add(f'<path d="{d}" fill="none" stroke="{color}" stroke-width="{width}" stroke-linecap="round"/>')


def dot(x, y, color):
    add(f'<circle cx="{x}" cy="{y}" r="6" fill="{color}" stroke="{BG}" stroke-width="2"/>')


def badge(x, y, s, color):
    """Small rounded label sitting on a wire."""
    w = 9 * len(s) + 14
    add(f'<rect x="{x - w / 2}" y="{y - 11}" width="{w}" height="22" rx="11" fill="{color}"/>')
    text(x, y + 5, s, size=12, weight=700, fill="#ffffff", anchor="middle", family=MONO)


# ---------------------------------------------------------------- background
add(f'<svg xmlns="http://www.w3.org/2000/svg" width="{W}" height="{H}" viewBox="0 0 {W} {H}">')
add(f'<title>Keg scale wiring: ESP32-DevKitV1, MAX3232, TM1637</title>')
add(f'<rect width="{W}" height="{H}" rx="18" fill="{BG}"/>')
text(40, 58, "Keg scale — wiring", size=30, weight=700)
text(40, 86, "ESP32-DevKitV1 (30 pin) · MAX3232 RS232↔TTL · TM1637 4-digit display · LESAK PS-B scale",
     size=15, fill=MUTED)

# ---------------------------------------------------------------- wires
# Drawn first so that boards and modules sit on top of the wire ends.

# Pin positions
x_top_gnd = pin_x(TOP_ROW, "GND")
x_d26 = pin_x(TOP_ROW, "D26")
x_d25 = pin_x(TOP_ROW, "D25")
x_3v3 = pin_x(BOTTOM_ROW, "3V3")
x_bot_gnd = pin_x(BOTTOM_ROW, "GND")
x_d16 = pin_x(BOTTOM_ROW, "D16")
x_d18 = pin_x(BOTTOM_ROW, "D18")
x_d19 = pin_x(BOTTOM_ROW, "D19")

# MAX3232 module (bottom left), TTL header on its top edge
MAX_X, MAX_Y, MAX_W, MAX_H = 150, 640, 250, 130
max_pins = {"VCC": 220, "GND": 260, "TXD": 300, "RXD": 340}
# TM1637 display (bottom right), header on its top edge: CLK DIO VCC GND
DISP_X, DISP_Y, DISP_W, DISP_H = 680, 640, 300, 150
disp_pins = {"CLK": x_d18, "DIO": x_d19, "VCC": 800, "GND": 840}

RAIL_3V3_Y, RAIL_GND_Y = 525, 555
RX_Y = 590

# Signal wires (vertical runs that the power rails hop over)
wire([(x_d16, BOTTOM_PIN_Y), (x_d16, RX_Y), (max_pins["RXD"], RX_Y), (max_pins["RXD"], MAX_Y)], C_RX)
wire([(x_d18, BOTTOM_PIN_Y), (x_d18, DISP_Y)], C_CLK)
wire([(x_d19, BOTTOM_PIN_Y), (x_d19, DISP_Y)], C_DIO)

# 3V3 rail
wire([(x_3v3, BOTTOM_PIN_Y), (x_3v3, RAIL_3V3_Y)], C_3V3)
hline_with_hops(max_pins["VCC"], disp_pins["VCC"], RAIL_3V3_Y, [x_bot_gnd, x_d16, x_d18, x_d19], C_3V3)
wire([(max_pins["VCC"], RAIL_3V3_Y), (max_pins["VCC"], MAX_Y)], C_3V3)
wire([(disp_pins["VCC"], RAIL_3V3_Y), (disp_pins["VCC"], DISP_Y)], C_3V3)
dot(x_3v3, RAIL_3V3_Y, C_3V3)

# GND rail
wire([(x_bot_gnd, BOTTOM_PIN_Y), (x_bot_gnd, RAIL_GND_Y)], C_GND)
hline_with_hops(max_pins["GND"], disp_pins["GND"], RAIL_GND_Y,
                [x_d16, x_d18, x_d19, disp_pins["VCC"]], C_GND)
wire([(max_pins["GND"], RAIL_GND_Y), (max_pins["GND"], MAX_Y)], C_GND)
wire([(disp_pins["GND"], RAIL_GND_Y), (disp_pins["GND"], DISP_Y)], C_GND)
dot(x_bot_gnd, RAIL_GND_Y, C_GND)

badge(x_3v3 - 70, RAIL_3V3_Y, "3.3 V", C_3V3)
badge(660, RAIL_GND_Y, "GND", C_GND)

# LEDs (top). Anode via 220 Ω from the GPIO, cathode to GND.
LED_G_X, LED_R_X = 560, 780
LED_TURN_Y = 290
GND_BUS_Y = 130
wire([(x_d26, TOP_PIN_Y), (x_d26, LED_TURN_Y), (LED_G_X, LED_TURN_Y), (LED_G_X, 268)], C_LED_G)
wire([(x_d25, TOP_PIN_Y), (x_d25, LED_TURN_Y), (LED_R_X, LED_TURN_Y), (LED_R_X, 268)], C_LED_R)
for x in (LED_G_X, LED_R_X):
    wire([(x, 220), (x, 198)], C_LED_G if x == LED_G_X else C_LED_R)
    wire([(x, 150), (x, GND_BUS_Y)], C_GND)
wire([(LED_R_X, GND_BUS_Y), (x_top_gnd, GND_BUS_Y), (x_top_gnd, TOP_PIN_Y)], C_GND)
dot(LED_G_X, GND_BUS_Y, C_GND)
badge(x_top_gnd + 60, GND_BUS_Y, "GND", C_GND)

# Scale -> MAX3232 RS232 cable
CABLE = "#8a929c"
add(f'<path d="M150,300 L150,330 Q150,345 135,345 L105,345 Q90,345 90,360 L90,690 Q90,705 105,705 L122,705" '
    f'fill="none" stroke="{CABLE}" stroke-width="10" stroke-linecap="round" stroke-linejoin="round"/>')
text(78, 520, "RS232 cable (DB9) — unchanged", size=13, fill=MUTED, anchor="middle",
     extra=' transform="rotate(-90 78 520)"')

# USB-C power
wire([(BOARD_X - 4, 400), (262, 400)], C_5V, width=8)

# ---------------------------------------------------------------- ESP32 board
add(f'<rect x="{BOARD_X}" y="{BOARD_Y}" width="{BOARD_W}" height="{BOARD_H}" rx="10" '
    f'fill="{PCB}" stroke="{PCB_EDGE}" stroke-width="2"/>')
# ESP-WROOM-32 can + antenna
add(f'<rect x="{BOARD_X + 470}" y="{BOARD_Y + 44}" width="160" height="52" rx="4" fill="{METAL}" stroke="#9aa1a9"/>')
text(BOARD_X + 550, BOARD_Y + 67, "ESP32-WROOM-32", size=12, weight=700, fill="#3a414a", anchor="middle", family=MONO)
text(BOARD_X + 550, BOARD_Y + 84, "D0WD-V3", size=11, fill="#59616b", anchor="middle", family=MONO)
add(f'<path d="M{BOARD_X + 642},{BOARD_Y + 46} h16 v10 h-12 v10 h12 v10 h-12 v10 h12 v10 h-16" '
    f'fill="none" stroke="{PIN}" stroke-width="2.5"/>')
# USB-C
add(f'<rect x="{BOARD_X - 14}" y="388" width="34" height="24" rx="8" fill="{METAL}" stroke="#8e959d"/>')
# CH9102 + buttons
add(f'<rect x="{BOARD_X + 54}" y="{BOARD_Y + 46}" width="30" height="30" rx="3" fill="#0b0f14"/>')
text(BOARD_X + 69, BOARD_Y + 88, "CH9102", size=9, fill="#8b95a1", anchor="middle", family=MONO)
for bx, label in ((BOARD_X + 120, "EN"), (BOARD_X + 120 + 40, "BOOT")):
    add(f'<rect x="{bx}" y="{BOARD_Y + 50}" width="22" height="22" rx="3" fill="#d6d9dd"/>')
    add(f'<circle cx="{bx + 11}" cy="{BOARD_Y + 61}" r="6" fill="#30363d"/>')
    text(bx + 11, BOARD_Y + 88, label, size=9, fill="#8b95a1", anchor="middle", family=MONO)
# on-board blue LED (GPIO2)
add(f'<rect x="{BOARD_X + 222}" y="{BOARD_Y + 57}" width="12" height="8" rx="2" fill="#3b82f6"/>')
add(f'<circle cx="{BOARD_X + 228}" cy="{BOARD_Y + 61}" r="11" fill="#3b82f6" opacity="0.25"/>')
text(BOARD_X + 228, BOARD_Y + 88, "LED · D2", size=9, fill="#8b95a1", anchor="middle", family=MONO)
text(BOARD_X + 350, BOARD_Y + 76, "ESP32-DevKitV1", size=16, weight=700, fill="#e6edf3", anchor="middle")

used = {"GND", "D25", "D26", "3V3", "D16", "D18", "D19"}
for row, py, label_dy in ((TOP_ROW, TOP_PIN_Y, 22), (BOTTOM_ROW, BOTTOM_PIN_Y, -14)):
    for name in row:
        x = pin_x(row, name)
        add(f'<circle cx="{x}" cy="{py}" r="7" fill="{PIN}" stroke="#7d6416" stroke-width="1.5"/>')
        add(f'<circle cx="{x}" cy="{py}" r="2.6" fill="{PCB_EDGE}"/>')
        hot = name in used
        text(x, py + label_dy, name, size=10, weight=700 if hot else 400,
             fill="#ffffff" if hot else "#7f8b98", anchor="middle", family=MONO)

# ---------------------------------------------------------------- USB power supply
add(f'<rect x="170" y="372" width="96" height="56" rx="10" fill="#ffffff" stroke="#c4c8cd" stroke-width="2"/>')
text(218, 397, "5 V", size=16, weight=700, anchor="middle")
text(218, 415, "USB-C PSU", size=11, fill=MUTED, anchor="middle")

# ---------------------------------------------------------------- scale
add(f'<rect x="40" y="150" width="230" height="150" rx="14" fill="#ffffff" stroke="#c4c8cd" stroke-width="2"/>')
add(f'<rect x="58" y="168" width="194" height="68" rx="8" fill="#e9ecef"/>')
add(f'<rect x="78" y="182" width="120" height="40" rx="5" fill="#26313c"/>')
text(138, 209, "42.50", size=20, weight=700, fill="#7ee787", anchor="middle", family=MONO)
text(206, 209, "kg", size=12, fill=MUTED, family=MONO)
text(58, 262, "LESAK PS-B", size=16, weight=700)
text(58, 282, "150 kg / 50 g · RS232 out", size=12, fill=MUTED)
add(f'<path d="M136,300 h28 l-4,10 h-20 z" fill="{METAL}" stroke="#8e959d"/>')

# ---------------------------------------------------------------- MAX3232
add(f'<rect x="{MAX_X}" y="{MAX_Y}" width="{MAX_W}" height="{MAX_H}" rx="8" fill="{BLUE_PCB}" stroke="#123e70" stroke-width="2"/>')
# DB9 female on the left
add(f'<path d="M{MAX_X - 28},{MAX_Y + 40} h34 v90 h-34 l-6,-12 v-66 z" fill="{METAL}" stroke="#8e959d" stroke-width="1.5"/>')
for i in range(5):
    add(f'<circle cx="{MAX_X - 14}" cy="{MAX_Y + 52 + i * 16}" r="2.6" fill="#5c636b"/>')
add(f'<rect x="{MAX_X + 70}" y="{MAX_Y + 52}" width="44" height="30" rx="2" fill="#0b0f14"/>')
text(MAX_X + 92, MAX_Y + 96, "MAX3232", size=10, fill="#cfe0f5", anchor="middle", family=MONO)
text(MAX_X + 130, MAX_Y + 118, "RS232 ↔ TTL", size=13, weight=700, fill="#ffffff", anchor="middle")
for name, x in max_pins.items():
    add(f'<rect x="{x - 8}" y="{MAX_Y - 6}" width="16" height="16" rx="2" fill="#151a20"/>')
    add(f'<circle cx="{x}" cy="{MAX_Y + 2}" r="4" fill="{PIN}"/>')
    text(x, MAX_Y + 30, name, size=10, weight=700, fill="#ffffff", anchor="middle", family=MONO)
text(max_pins["TXD"], MAX_Y + 44, "n/c", size=9, fill="#9cc0e8", anchor="middle", family=MONO)

# ---------------------------------------------------------------- TM1637
add(f'<rect x="{DISP_X}" y="{DISP_Y}" width="{DISP_W}" height="{DISP_H}" rx="8" fill="{BLUE_PCB}" stroke="#123e70" stroke-width="2"/>')
for name, x in disp_pins.items():
    add(f'<rect x="{x - 8}" y="{DISP_Y - 6}" width="16" height="16" rx="2" fill="#151a20"/>')
    add(f'<circle cx="{x}" cy="{DISP_Y + 2}" r="4" fill="{PIN}"/>')
    text(x, DISP_Y + 28, name, size=10, weight=700, fill="#ffffff", anchor="middle", family=MONO)
add(f'<rect x="{DISP_X + 30}" y="{DISP_Y + 42}" width="{DISP_W - 60}" height="92" rx="4" fill="#111418"/>')


def seven_seg(x, y, digit, on="#ff3b30", off="#2a1d1f"):
    """Draws a 7-segment digit with its top-left corner at x, y (36x64)."""
    segs = {
        "a": (x + 6, y, 24, 6), "d": (x + 6, y + 58, 24, 6), "g": (x + 6, y + 29, 24, 6),
        "f": (x, y + 5, 6, 25), "b": (x + 30, y + 5, 6, 25),
        "e": (x, y + 34, 6, 25), "c": (x + 30, y + 34, 6, 25),
    }
    lit = {"4": "fgbc", "2": "abged", " ": ""}[digit]
    for k, (sx, sy, sw, sh) in segs.items():
        fill = on if k in lit else off
        add(f'<rect x="{sx}" y="{sy}" width="{sw}" height="{sh}" rx="2.5" fill="{fill}" '
            f'transform="skewX(-6)" transform-origin="{x + 18} {y + 32}"/>')


for i, ch in enumerate("  42"):
    seven_seg(DISP_X + 58 + i * 50, DISP_Y + 56, ch)
text(DISP_X + DISP_W / 2, DISP_Y + DISP_H + 22, "TM1637 — beers left in the keg", size=13, fill=MUTED, anchor="middle")
text(MAX_X + MAX_W / 2, MAX_Y + MAX_H + 22, "TTL side: 3.3 V, not 5 V!", size=13, weight=700, fill=C_3V3, anchor="middle")


# ---------------------------------------------------------------- LEDs
def led(x, color, label):
    # resistor
    add(f'<rect x="{x - 10}" y="220" width="20" height="48" rx="6" fill="#e8d9b5" stroke="#b49a5e"/>')
    for i, c in enumerate(("#d1342f", "#d1342f", "#5a3a1a")):
        add(f'<rect x="{x - 10}" y="{230 + i * 9}" width="20" height="4" fill="{c}"/>')
    text(x + 18, 249, "220 Ω", size=12, fill=MUTED, family=MONO)
    # LED body
    add(f'<path d="M{x - 14},198 v-30 a14,14 0 0 1 28,0 v30 z" fill="{color}" opacity="0.9"/>')
    add(f'<rect x="{x - 17}" y="194" width="34" height="6" rx="2" fill="{color}"/>')
    add(f'<circle cx="{x}" cy="172" r="22" fill="{color}" opacity="0.15"/>')
    text(x - 26, 180, label, size=13, weight=700, fill=INK, anchor="end")


led(LED_G_X, "#34c759", "green · OK")
led(LED_R_X, "#ff3b30", "red · error")
text(LED_R_X + 26, 166, "long leg (anode) → resistor", size=11, fill=MUTED)
text(LED_R_X + 26, 182, "short leg (cathode) → GND", size=11, fill=MUTED)

# ---------------------------------------------------------------- legend
LX, LY = 1060, 150
add(f'<rect x="{LX}" y="{LY}" width="190" height="460" rx="12" fill="#ffffff" stroke="#d8dbe0"/>')
text(LX + 18, LY + 32, "Connections", size=16, weight=700)
rows = [
    (C_RX, "D16", "MAX3232 RXD"),
    (C_CLK, "D18", "TM1637 CLK"),
    (C_DIO, "D19", "TM1637 DIO"),
    (C_LED_G, "D26", "green LED"),
    (C_LED_R, "D25", "red LED"),
    (C_3V3, "3V3", "VCC of both"),
    (C_GND, "GND", "GND of all"),
    (C_5V, "USB", "5 V power"),
]
for i, (c, pin, what) in enumerate(rows):
    y = LY + 64 + i * 32
    add(f'<rect x="{LX + 18}" y="{y - 5}" width="26" height="6" rx="3" fill="{c}"/>')
    text(LX + 54, y + 2, pin, size=12, weight=700, family=MONO)
    text(LX + 92, y + 2, what, size=12, fill=MUTED)
y = LY + 64 + len(rows) * 32 + 4
add(f'<line x1="{LX + 18}" y1="{y - 14}" x2="{LX + 172}" y2="{y - 14}" stroke="#e3e6ea"/>')
for i, line in enumerate(["ESP32 pins are 3.3 V only.",
                          "Power MAX3232 and the",
                          "display from 3V3, never 5 V.",
                          "",
                          "MAX3232 TXD stays",
                          "unconnected (the scale",
                          "only sends)."]):
    text(LX + 18, y + 6 + i * 17, line, size=12, fill=INK if i < 3 else MUTED, weight=700 if i < 3 else 400)

add("</svg>")

path = sys.argv[1] if len(sys.argv) > 1 else "wiring.svg"
with open(path, "w", encoding="utf-8") as f:
    f.write("\n".join(out) + "\n")
