#ifndef _SWITCH_H
#define _SWITCH_H
#include <stdbool.h>
#include <stdlib.h>

typedef bool (*callbackfunc)(int* ranks, bool* ops, int length);

static bool callbackfuncwrap(callbackfunc cb, int* ranks, bool* ops, int len) {
    return cb(ranks, ops, len);
}
#endif