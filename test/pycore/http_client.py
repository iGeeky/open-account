# coding=utf8
'''
http请求封装
'''

import ssl
import traceback
import httpx
# from urllib import request
from urllib.parse import urlencode
from past.builtins import basestring
import hashlib
import sys
import socket
import time
import json
import logging
log = logging.getLogger("httpclient")

class CaseInsensitiveDict(dict):
    @classmethod
    def _k(cls, key):
        return key.lower() if isinstance(key, basestring) else key

    def __init__(self, *args, **kwargs):
        super(CaseInsensitiveDict, self).__init__(*args, **kwargs)
        self._convert_keys()
    def __getitem__(self, key):
        return super(CaseInsensitiveDict, self).__getitem__(self.__class__._k(key))
    def __setitem__(self, key, value):
        super(CaseInsensitiveDict, self).__setitem__(self.__class__._k(key), value)
    def __delitem__(self, key):
        return super(CaseInsensitiveDict, self).__delitem__(self.__class__._k(key))
    def __contains__(self, key):
        return super(CaseInsensitiveDict, self).__contains__(self.__class__._k(key))
    def has_key(self, key):
        return super(CaseInsensitiveDict, self).has_key(self.__class__._k(key))
    def pop(self, key, *args, **kwargs):
        return super(CaseInsensitiveDict, self).pop(self.__class__._k(key), *args, **kwargs)
    def get(self, key, *args, **kwargs):
        return super(CaseInsensitiveDict, self).get(self.__class__._k(key), *args, **kwargs)
    def setdefault(self, key, *args, **kwargs):
        return super(CaseInsensitiveDict, self).setdefault(self.__class__._k(key), *args, **kwargs)
    def update(self, E={}, **F):
        super(CaseInsensitiveDict, self).update(self.__class__(E))
        super(CaseInsensitiveDict, self).update(self.__class__(**F))
    def _convert_keys(self):
        for k in list(self.keys()):
            v = super(CaseInsensitiveDict, self).pop(k)
            self.__setitem__(k, v)

def NewHeaders():
    return CaseInsensitiveDict()


def headerstr(headers):
    if not headers:
        return ""

    lines = []
    for key in headers:
        if key != "User-Agent":
            value = headers[key]
            lines.append("-H'" + str(key) + ": " + str(value) + "'")

    return ' '.join(lines)


def GetDebugStr(method, url, args, headers, body, timeout):
    if type(body) == dict:
        body = json.dumps(body)
    req_debug = ""
    if args:
        query = urlencode(args).encode(encoding='utf-8',errors='ignore')
        url = "%s?%s" % (url, query.decode("utf-8"))
    if method == "PUT" or method == "POST" or method == "DELETE":
        debug_body = ''
        content_type = headers.get("Content-Type")
        if content_type == None or content_type.startswith("text") or content_type == 'application/json':
            if len(body) < 2000:
                debug_body = body
            else:
                debug_body = body[0:1023]
        else:
            debug_body = "[[not text body: "  + str(content_type) + "]]"
        req_debug = "curl -v -k -X " + method + " " + headerstr(headers) + " '" + url + "' -d '" + debug_body + "' -o /dev/null"
    else:
        req_debug = "curl -v -k -X " + method + " " + headerstr(headers) + " '" + url + "' -o /dev/null"

    return req_debug

class Response(object):
    def __init__(self, data):
        self._data = data

    def __getattr__(self, attr):
        return self._data.get(attr, None)
    def __str__(self):
        return str(self._data)


def HttpReq(method, url, headers, args, body, timeout):
    timeout = timeout or 10
    headers = headers or NewHeaders()
    req_debug = GetDebugStr(method, url, args, headers, body, timeout)
    timeout_str = str(timeout)
    log.info("REQUEST [ %s ] timeout: %s", req_debug, timeout_str)

    res = Response({"status": 500, "body":  None, "headers": NewHeaders()})
    server_ip = ""
    begin = time.time()
    req = {}
    try:
        client:httpx.Client = httpx.Client(verify=False, timeout=timeout)
        if method == "PUT" or method == "POST" or method == "DELETE":
            resp:httpx.Response = client.request(method=method, url=url, data=body, params=args, headers=headers)
        else:
            resp:httpx.Response = client.request(method=method, url=url, params=args, headers=headers)
    except Exception as e:
        msg = 'REQUEST [ %s ] failed! err: %s' %(req_debug, e)
        log.error(msg)
        resp = httpx.Response(status_code=500, content='', headers={}, json={})
        res = Response({"status": 500, "body":  str(traceback.format_exc()), "headers": NewHeaders(), "rawResp": resp, "json": {}})
    else:
        status = resp.status_code
        body = resp.content
        res = Response({"status": status, "body": body, "headers": dict(resp.headers.items()), "rawResp": resp, "json": resp.json()})

    duration = time.time()-begin
    res.duration = duration

    if res.status >= 400:
        log.warning("FAIL REQUEST [ %s ] status: %s, duration: %.3f body: %s", req_debug, res.status, duration, res.body)
    else:
        log.info ("REQUEST [ %s ] status: %s, duration: %.3f", req_debug, res.status, duration)
    res.req_debug = req_debug

    return res

def HttpGet(url, headers, args, timeout):
    return HttpReq('GET', url, headers, args, None, timeout)

def HttpPost(url, headers, body, timeout):
    return HttpReq('POST', url, headers, None, body, timeout)


if __name__ == '__main__':
    res = HttpGet("http://api.test.com", {}, None, 10)
    print("headers: %s" % res.headers)
    print("body-length: %s" % (len(res.body)))
