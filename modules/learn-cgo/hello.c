#include <stdio.h>

void hello_from_c() { printf("Hello from C\n"); }

void greet_user(const char *name) {
  printf("Hello, %s! (Greetings from C processing a Go string)\n", name);
}