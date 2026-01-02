package main

/*
#cgo LDFLAGS: -lm -L/Users/manik/Documents/workspace/go/projects/learn_go/modules/learn-cgo -lcommon
#include <math.h>
#include <hello.h>
*/
import "C"
import (
	"log"
)

func main() {
	log.Println("learn_cgo")
	x := 4.0
	result := float64(C.double(C.sqrt(C.double(x))))
	log.Println(result)
	C.hello_from_c()
	C.hello_from_cpp()
}
