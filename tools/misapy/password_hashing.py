import sys
from base64 import b64encode

try:
    # https://argon2-cffi.readthedocs.io
    import argon2 as argon2_cffi
except ImportError:
    sys.exit('cannot import argon2 (pip install argon2-cffi)')

def hash_password(password, salt_base64, iterations=1, memory=1024, parallelism=1):
    hash = argon2_cffi.low_level.hash_secret_raw(
        'password'.encode(),
        salt=salt_base64.encode(),
        time_cost=iterations,
        memory_cost=memory,
        parallelism=parallelism,
        hash_len=16,
        type=argon2_cffi.low_level.Type.I
    )

    return {
        'params' : {
            'memory': 1024,
            'iterations': 1,
            'parallelism': 1,
            'salt_base64': salt_base64,
        },
        'hash_base64': b64encode(hash).decode(),
    }

