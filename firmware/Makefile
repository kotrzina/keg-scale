CMD := arduino-cli
CORE := esp32:esp32
# ESP32-DevKitV1 (ESP32-D0WD-V3, CH9102 USB-UART). Serial goes through the
# USB converter, so the debug build needs no extra board options.
FQBN := esp32:esp32:esp32
SRCDIR := src
# Lazy (=), not immediate (:=): the port has to be detected when the recipe
# runs. The CH9102 shows up as a generic USB serial device without a board
# name: /dev/ttyACM* or /dev/ttyUSB* on Linux, /dev/cu.wchusbserial* or
# /dev/cu.usbserial* on macOS.
INTERFACE = $(shell $(CMD) board list --format text | grep -oE '/dev/(ttyACM|ttyUSB|cu\.wchusbserial|cu\.usbserial|cu\.usbmodem|cu\.SLAB_USBtoUART)[^ ]*' | head -1)

DEBUG_FLAGS := --build-property compiler.cpp.extra_flags=-DDEBUG_SERIAL

compile:
	$(CMD) compile -b $(FQBN) $(SRCDIR)

flash: compile
	@port="$(INTERFACE)"; \
	if [ -z "$$port" ]; then echo "No ESP32 board detected. Check the cable and run 'make list'."; exit 1; fi; \
	echo "Using $$port"; \
	$(CMD) upload -p "$$port" --fqbn $(FQBN) $(SRCDIR)

compile-debug:
	$(CMD) compile -b $(FQBN) $(DEBUG_FLAGS) $(SRCDIR)

flash-debug: compile-debug
	@port="$(INTERFACE)"; \
	if [ -z "$$port" ]; then echo "No ESP32 board detected. Check the cable and run 'make list'."; exit 1; fi; \
	echo "Using $$port"; \
	$(CMD) upload -p "$$port" --fqbn $(FQBN) $(DEBUG_FLAGS) $(SRCDIR)

monitor:
	@port="$(INTERFACE)"; \
	if [ -z "$$port" ]; then echo "No ESP32 board detected. Check the cable and run 'make list'."; exit 1; fi; \
	echo "Using $$port"; \
	$(CMD) monitor -p "$$port" -c baudrate=115200

# Host-side unit tests of the scale protocol parser (needs only g++).
test:
	@mkdir -p build
	g++ -std=c++17 -Wall -Wextra -Werror -o build/scale_parser_test test/scale_parser_test.cpp
	./build/scale_parser_test

# Regenerate the wiring diagram from docs/wiring.py.
wiring:
	python3 docs/wiring.py docs/wiring.svg

deps:
	$(CMD) lib install "WiFiManager"
	$(CMD) lib install "TM1637"

install:
	$(CMD) config add board_manager.additional_urls https://raw.githubusercontent.com/espressif/arduino-esp32/gh-pages/package_esp32_index.json
	$(CMD) core install $(CORE)

list:
	$(CMD) board list

.PHONY: compile flash compile-debug flash-debug monitor test wiring deps install list
