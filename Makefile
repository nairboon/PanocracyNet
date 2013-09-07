

all: native

native:
	$(MAKE) -C native/src

simulation:
	gd -o simulation/sim simulation

.PHONY: native simulation
