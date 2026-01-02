#if !defined(learn_cgo)
#define learn_cgo

/*
# Compile the C source file into an object file with Position Independent Code (PIC)
gcc -fPIC -c hello.c -o hello.o

# Link the object file into a shared library file (conventionally starting with 'lib')
gcc -shared -o libhello.so hello.o
# (On macOS, the output file would typically be libhello.dylib)



#cgo LDFLAGS: -lm -L. -lhello
#include <math.h>
#include "hello.h"


Linker Recognition (-l flag): The primary reason for this convention is how compilers and linkers work.
When you use the GCC linker flag -lhello in your Cgo directive, the linker automatically assumes the filename you want is libhello.so (or libhello.a for static libraries).
It automatically adds the lib prefix and the appropriate extension when searching its paths.


Naming Across Different Systems
The convention is standard across platforms, although the file extensions differ:
Linux: libname.so (and often includes version numbers, e.g., libname.so.1.0.1).
macOS (Darwin/Apple): libname.dylib.
Windows: Uses a different naming system. Shared libraries are name.dll files, and they are typically accompanied by a separate "import library" file named libname.lib used only during the link phase.



*/

void hello_from_c();

/*
extern "C" specifies C-style linkage and calling conventions to allow interoperability between C and C++ code .
Used in C++, ignored by C compilers (often wrapped in #ifdef __cplusplus).

g++ -c hello.cpp -o helloCpp.o
create a single static lib

# 'r' replaces existing files, 'c' creates the archive if it doesn't exist, 's' writes an object-file index.
ar rcs libcommonlib.a cpp_part.o c_part.o


*/
#ifdef __cplusplus
extern "C"
{
#endif __cplusplus

    void hello_from_cpp();

#ifdef __cplusplus
}
#endif __cplusplus

#endif // learn_cgo
