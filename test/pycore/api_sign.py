#!/usr/bin/evn python
# -*- coding: UTF-8 -*-

import re
import hashlib
import logging
from . import utils
#from urllib import quote

def is_encoded(str_):
    regex = r"%[0-9A-Fa-f]{2}"
    if re.search(regex, str_):
        return True

    return False


def uri_encode_internal(arg, encodeSlash):
    if not arg:
        return arg

    if is_encoded(arg):
        return arg

    if encodeSlash == None:
        encodeSlash = True

    chars = []
    for ch in arg:
        if (ch >= 'A' and ch <= 'Z') or (ch >= 'a' and ch <= 'z') or (ch >= '0' and ch <= '9') or ch == '_' or ch == '-' or ch == '~' or ch == '.':
            chars.append(ch)
        elif ch == '/':
            if encodeSlash:
                chars.append("%2F")
            else:
                chars.append(ch)
        else:
            chars.append("%%%02X" % ord(ch))

    return  ''.join(chars)

def URI_ENCODE(uri):
    return uri_encode_internal(uri, False)

def uri_encode(uri):
    return uri_encode_internal(uri, True)


def CreateCanonicalArgs(args):
    if args == None:
        return ""

    keys = list(args.keys())
    keys.sort()

    key_values = []
    for key in keys:
        value = args[key]
        # print("key :", key, " ", type(key), " value:", type(value))
        if type(value) == bool:
            value = ""

        if type(value) == list:
            value.sort()
            for value_sub in value:
                key_values.append(uri_encode(key) + "=" + uri_encode(value_sub))
        else:
            key_values.append(uri_encode(key) + "=" + uri_encode(value))
    return "&".join(key_values)

all_normal_headers = {
    "host": True,
    "date": True,
}

def CreateCanonicalHeaders(headers):
    t = headers
    headers_lower = {}
    signed_headers = []
    header_values = []
    prefix = utils.customHeaderPrefix.lower()
    for k,v in t.items():
        k = k.lower()
        if type(v) == list:
            v.sort()
            v = ",".join(v)

        if k != utils.CH("Sign").lower() and (all_normal_headers.get(k) or k.startswith(prefix)):
            signed_headers.append(k)
            headers_lower[k] = v

    signed_headers.sort()

    for k in signed_headers:
        header_values.append(k + ":" + headers_lower[k].strip())

    return "\n".join(header_values), ";".join(signed_headers)


def CreateSignStr(uri, args, headers, Sha1Body, app_key):
    strs = []
    CanonicalURI = URI_ENCODE(uri)
    CanonicalArgs = CreateCanonicalArgs(args)
    CanonicalHeaders, SignedHeaders = CreateCanonicalHeaders(headers)

    strs.append(CanonicalURI)
    strs.append(CanonicalArgs)
    strs.append(CanonicalHeaders)
    strs.append(SignedHeaders)
    strs.append(Sha1Body)
    if app_key:
        strs.append(app_key)

    return "\n".join(strs)


def sign(uri, args, headers, body, app_key):
    body = "" if body == None else body
    if type(body) == str:
        body = body.encode("utf-8")
    Sha1Body = hashlib.sha1(body).hexdigest()
    SignStr = CreateSignStr(uri, args, headers, Sha1Body, app_key)
    signature = hashlib.sha1(SignStr.encode("utf-8")).hexdigest()
    return signature, SignStr