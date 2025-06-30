# Simple test Makefile
CC := gcc
CFLAGS := -Wall -O2

.PHONY: all clean

# Build all targets
all: build test

# Build the application
build: main.o utils.o
	$(CC) $(CFLAGS) -o app main.o utils.o

main.o: main.c
	$(CC) $(CFLAGS) -c main.c

utils.o: utils.c utils.h
	$(CC) $(CFLAGS) -c utils.c

# Run tests
test:
	./test.sh

# Clean build artifacts
clean:
	rm -f *.o app