import base64 as stdlib_base64

def b64encode(data):
    '''encodes data with base64 and returns a string
    (the stdlib version returns bytes)'''
    return stdlib_base64.b64encode(data).decode()

b64decode = stdlib_base64.b64decode

def urlsafe_b64encode(data):
    '''encodes data with URL-safe base64,
    and returns a string without padding
    (stdlib version returns bytes with padding)'''
    return stdlib_base64.urlsafe_b64encode(data).rstrip(b'=').decode()