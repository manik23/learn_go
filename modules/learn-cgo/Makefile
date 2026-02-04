# Makefile for learn-cgo

# Detect OS
UNAME_S := $(shell uname -s)
ifeq ($(UNAME_S),Darwin)
	LIB_EXT := dylib
else
	LIB_EXT := so
endif

LIB_NAME := libcommon.$(LIB_EXT)

all: build-lib run

# 1. Compile C and C++ objects
# 2. Link them into a shared library
# 3. Build/Run the Go application
build-lib:
	gcc -fPIC -c hello.c -o hello.o
	g++ -fPIC -c hello.cpp -o helloCpp.o
	g++ -shared -o $(LIB_NAME) hello.o helloCpp.o

run:
	go run example_cgo.go

clean:
	rm -f *.o *.so *.dylib *.a
