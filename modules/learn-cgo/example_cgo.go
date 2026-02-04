package main

/*
#cgo LDFLAGS: -lm -L/Users/manik/Documents/workspace/go/projects/learn_go/modules/learn-cgo -lcommon
#include <math.h>
#include <stdlib.h>
#include "hello.h"
*/
import "C"

import (
	"log"
	"unsafe"
)

func main() {
	log.Println("--- learn_cgo ---")

	// 1. Basic C Math implementation
	x := 4.0
	result := float64(C.sqrt(C.double(x)))
	log.Printf("C.sqrt(%v) = %v\n", x, result)

	// 2. Call simple C function
	C.hello_from_c()

	// 3. Call C++ function (via C wrapper/linkage)
	C.hello_from_cpp()

	// 4. Data Exchange: Passing a string from Go to C
	// C.CString allocates memory on the C heap, so we MUST free it manually.
	name := "Antigravity"
	cName := C.CString(name)
	defer C.free(unsafe.Pointer(cName))

	C.greet_user(cName)
}
