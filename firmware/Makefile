CMD := arduino-cli
CORE := arduino:megaavr
FQBN := arduino:megaavr:uno2018
INTERFACE := $(shell $(CMD) board list | grep $(FQBN) | awk '{print $$1}')
SKETCH := src

compile:
	$(CMD) compile -b $(FQBN) $(SKETCH)

flush: compile
	$(CMD) upload -p $(INTERFACE) --fqbn $(FQBN) $(SKETCH)

read:
	screen $(INTERFACE) 9600

deps:
	arduino-cli lib install WiFiNINA
	arduino-cli lib install ArduinoHttpClient

install:
	$(CMD) core install $(CORE)

list:
	$(CMD) board list
