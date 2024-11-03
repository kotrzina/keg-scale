CMD := arduino-cli
CORE := arduino:megaavr
FQBN := arduino:megaavr:uno2018
INTERFACE := $(shell $(CMD) board list | grep $(FQBN) | awk '{print $$1}')
SRCDIR := src

compile:
	$(CMD) compile -b $(FQBN) $(SRCDIR)

flush: compile
	$(CMD) upload -p $(INTERFACE) --fqbn $(FQBN) $(SRCDIR)

read:
	screen $(INTERFACE) 9600

deps:
	arduino-cli lib install WiFiNINA
	arduino-cli lib install ArduinoHttpClient
	arduino-cli lib install TM1637

install:
	$(CMD) core install $(CORE)

list:
	$(CMD) board list
