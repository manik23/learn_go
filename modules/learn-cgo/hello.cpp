#include <iostream>

using namespace std;

extern "C"
{
    void hello_from_cpp()
    {
        cout << "hello from c++" << endl;
    }
}