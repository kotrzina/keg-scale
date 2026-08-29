CMD := arduino-cli
CORE := esp32:esp32
# CPUFreq=80 is the lowest frequency that still supports WiFi on the C3
# (40/20/10 MHz are not marked "(WiFi)" in `arduino-cli board details`).
# The debug build additionally routes Serial to the native USB CDC; without
# CDCOnBoot=cdc, Serial goes to UART0 on GPIO20/21 and `make monitor` is silent.
FQBN := esp32:esp32:esp32c3:CPUFreq=80
FQBN_DEBUG := esp32:esp32:esp32c3:CPUFreq=80,CDCOnBoot=cdc
SRCDIR := src
# Lazy (=), not immediate (:=): the ESP32-C3 native USB CDC port re-enumerates
# on reset, so it must be detected when the recipe runs, not when make parses.
INTERFACE = $(shell $(CMD) board list --format text | grep 'esp32' | grep -oE '/dev/cu\.[^ ]+' | head -1)

DEBUG_FLAGS := --build-property compiler.cpp.extra_flags=-DDEBUG_SERIAL

compile:
	$(CMD) compile -b $(FQBN) $(SRCDIR)

flash: compile
	@port="$(INTERFACE)"; \
	if [ -z "$$port" ]; then echo "No ESP32 board detected. Check the cable and run 'make list'."; exit 1; fi; \
	echo "Using $$port"; \
	$(CMD) upload -p "$$port" --fqbn $(FQBN) $(SRCDIR)

compile-debug:
	$(CMD) compile -b $(FQBN_DEBUG) $(DEBUG_FLAGS) $(SRCDIR)

flash-debug: compile-debug
	@port="$(INTERFACE)"; \
	if [ -z "$$port" ]; then echo "No ESP32 board detected. Check the cable and run 'make list'."; exit 1; fi; \
	echo "Using $$port"; \
	$(CMD) upload -p "$$port" --fqbn $(FQBN_DEBUG) $(SRCDIR)

monitor:
	@port="$(INTERFACE)"; \
	if [ -z "$$port" ]; then echo "No ESP32 board detected. Check the cable and run 'make list'."; exit 1; fi; \
	echo "Using $$port"; \
	$(CMD) monitor -p "$$port" -c baudrate=115200

deps:
	$(CMD) lib install "WiFiManager"

install:
	$(CMD) config add board_manager.additional_urls https://raw.githubusercontent.com/espressif/arduino-esp32/gh-pages/package_esp32_index.json
	$(CMD) core install $(CORE)

list:
	$(CMD) board list
