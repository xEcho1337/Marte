import sys
import json
import traceback

_exploit_fn = None

def exploit(fn):
    global _exploit_fn
    _exploit_fn = fn
    return fn
