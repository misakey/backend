import os
import re
import json
from collections import namedtuple
from math import ceil

import msgpack

from cryptography.hazmat.primitives import (
    hashes,
    hmac,
    serialization,
)
from cryptography.hazmat.primitives.ciphers import (
    Cipher, algorithms, modes,
)
import cryptography.hazmat.primitives.asymmetric.rsa as rsa
from cryptography.hazmat.primitives.asymmetric.rsa import (
    RSAPrivateKey, RSAPublicNumbers,
)
from cryptography.hazmat.primitives.asymmetric import padding

from ..utils.base64 import urlsafe_b64encode, urlsafe_b64decode

ALGORITHM_NAME = 'com.misakey.aes-rsa-enc'

RSA_MODULUS_LENGTH = 3072
RSA_PUBLIC_EXPONENT = 65537
RSA_PADDING = padding.OAEP(
    mgf=padding.MGF1(algorithm=hashes.SHA256()),
    algorithm=hashes.SHA256(),
    label=None
)

AES_CTR_KEY_SIZE = 32 # in bytes
AES_CTR_M = 64 # in bits
AES_CTR_NONCE_SIZE = (128 - AES_CTR_M)//8 # in bytes

HMAC_KEY_SIZE = 16 # in bytes

# used internally only
_Cryptogram = namedtuple('_Cryptogram', ['ciphertext', 'nonce', 'auth_tag', 'wrapped_key'])

KeyPair = namedtuple('KeyPair', ['secret_key', 'public_key'])

def _load_public_key(recipient_public_key):
    match = re.match(r'^com.misakey.aes-rsa-enc:([a-zA-Z0-9-_]+)$', recipient_public_key)
    if match:
        rsa_modulus = match.groups()[0]
        return RSAPublicNumbers(
            e=RSA_PUBLIC_EXPONENT,
            n=int.from_bytes(urlsafe_b64decode(rsa_modulus), byteorder='big')
        ).public_key()
    else:
        raise ValueError('Malformed recipient public key')

def generate_key_pair():
    secret_key = rsa.generate_private_key(
        RSA_PUBLIC_EXPONENT,
        RSA_MODULUS_LENGTH,
    )
    public_key = secret_key.public_key()

    encoded_sk = urlsafe_b64encode(secret_key.private_bytes(
        serialization.Encoding.DER,
        serialization.PrivateFormat.PKCS8,
        serialization.NoEncryption(),
    ))

    modulus = public_key.public_numbers().n
    encoded_pk = (
        ALGORITHM_NAME
        +':'
        +urlsafe_b64encode(modulus.to_bytes(
            ceil(ceil(RSA_MODULUS_LENGTH)/8),
            'big',
        ))
    )

    return KeyPair(
        secret_key=encoded_sk,
        public_key=encoded_pk,
    )

def _generate_symmetric_key():
    return os.urandom(
        max(
            AES_CTR_KEY_SIZE,
            HMAC_KEY_SIZE,
        )
    )

def _derive_key(base_key, label: str, key_size: int):
    # HKDF in our case boils down to HMAC
    h = hmac.HMAC(base_key, hashes.SHA256())
    h.update(label.encode())
    derived_key = h.finalize()[:key_size]
    return derived_key

def _symmetric_encrypt(plaintext, key):
    aes_key = _derive_key(
        base_key=key,
        label='encryption',
        key_size=AES_CTR_KEY_SIZE,
    )
    nonce = os.urandom(AES_CTR_NONCE_SIZE)

    # NIST SP 800-38a
    # (https://csrc.nist.gov/publications/detail/sp/800-38a/final, defines CTR mode)
    # suggests (in Appendix B.2)
    # to initialize the counter block with the nonce concatenated with `m` zeros,
    # `m` being the number of bits being incremented by the counter (defined in Appendix B.1).
    # We follow this recommendation,
    # but the "cryptography" library insteads expects a "nonce"
    # as long as the block size.
    # (see https://cryptography.io/en/latest/hazmat/primitives/symmetric-encryption.html#cryptography.hazmat.primitives.ciphers.modes.CTR)
    # So our "nonce" is not the same as their "nonce".
    init_counter_block = nonce + bytes(AES_CTR_M//8)

    cipher = Cipher(
        algorithms.AES(aes_key),
        modes.CTR(init_counter_block),
    )
    encryptor = cipher.encryptor()
    ciphertext = encryptor.update(plaintext) + encryptor.finalize()

    mac_key =_derive_key(
        base_key=key,
        label='mac',
        key_size=HMAC_KEY_SIZE,
    )

    h = hmac.HMAC(mac_key, hashes.SHA256())
    h.update(ciphertext)
    auth_tag = h.finalize()

    return (ciphertext, nonce, auth_tag)

def _symmetric_decrypt(ciphertext, nonce, auth_tag, key):
    mac_key =_derive_key(
        base_key=key,
        label='mac',
        key_size=HMAC_KEY_SIZE,
    )

    h = hmac.HMAC(mac_key, hashes.SHA256())
    h.update(ciphertext)
    h.verify(auth_tag)

    aes_key = _derive_key(
        base_key=key,
        label='encryption',
        key_size=AES_CTR_KEY_SIZE,
    )

    # see comment in "_symmetric_encrypt" function
    init_counter_block = nonce + bytes(AES_CTR_M//8)

    cipher = Cipher(
        algorithms.AES(aes_key),
        modes.CTR(init_counter_block),
    )
    decryptor = cipher.decryptor()
    plaintext = decryptor.update(ciphertext) + decryptor.finalize()

    return plaintext

def encrypt_message(message: bytes, public_key: str):
    pubkey = _load_public_key(public_key)

    symmetric_key = _generate_symmetric_key()

    ciphertext, nonce, auth_tag = _symmetric_encrypt(message, symmetric_key)

    wrapped_key = pubkey.encrypt(symmetric_key, RSA_PADDING)

    return urlsafe_b64encode(
        msgpack.packb({
            'ciphertext': ciphertext,
            'nonce': nonce,
            'auth_tag': auth_tag,
            'wrapped_key': wrapped_key,
        })
    )

def decrypt_message(cryptogram: str, secret_key: str):
    cgm = _Cryptogram(
        **msgpack.unpackb(urlsafe_b64decode(cryptogram))
    )

    actual_secret_key: RSAPrivateKey = serialization.load_der_private_key(
        data=urlsafe_b64decode(secret_key),
        password=None,
    )
    
    symmetric_key = actual_secret_key.decrypt(
        cgm.wrapped_key,
        RSA_PADDING,
    )

    plaintext = _symmetric_decrypt(cgm.ciphertext, cgm.nonce, cgm.auth_tag, symmetric_key)

    return plaintext

def encrypt_file(file_content: bytes, file_name, public_key):
    file_key = _generate_symmetric_key()

    encrypted_file, nonce, auth_tag = _symmetric_encrypt(file_content, file_key)

    file_encryption_metadata = {
        'encryption' : {
            'algorithm': ALGORITHM_NAME,
            'key': urlsafe_b64encode(file_key),
            'nonce': urlsafe_b64encode(nonce),
            'authTag': urlsafe_b64encode(auth_tag),
        },
        'fileName': file_name,
        'fileSize': len(file_content),
        # TODO see about "fileType" field (use the extension?)
    }

    encrypted_message_content = encrypt_message(
        json.dumps(file_encryption_metadata).encode(),
        public_key,
    )

    return (encrypted_file, encrypted_message_content)

def decrypt_file(encrypted_file: bytes, encrypted_message_content: str, secret_key):
    file_encryption_metadata = json.loads(
        decrypt_message(encrypted_message_content, secret_key)
    )

    assert file_encryption_metadata['encryption']['algorithm'] == ALGORITHM_NAME

    file = _symmetric_decrypt(
        encrypted_file,
        urlsafe_b64decode(file_encryption_metadata['encryption']['nonce']),
        urlsafe_b64decode(file_encryption_metadata['encryption']['authTag']),
        urlsafe_b64decode(file_encryption_metadata['encryption']['key']),
    )

    return file
