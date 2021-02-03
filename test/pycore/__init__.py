import logging
import os
import hashlib
from copy import deepcopy

log_level = os.environ.get('LOG_LEVEL', 'ERROR') # INFO,WARN,ERROR
if log_level:
    logging.basicConfig(level=log_level, format="%(asctime)s | %(levelname)s | %(filename)s:%(lineno)d | %(message)s")

from . import http_test
HttpTest = http_test.HttpTest
customHeaderPrefix = "X-OA-"

def init():
    global customHeaderPrefix
    server = os.environ.get('SERVER')
    if server:
        HttpTest.SERVER = server

    timeout = os.environ.get('TIMEOUT')
    if timeout:
        HttpTest.TIMEOUT = int(timeout)
    headerPrefix = os.environ.get('HEADER_PREFIX')
    if headerPrefix:
        customHeaderPrefix = headerPrefix

init()

import time
from threading import Lock
now = int(time.time()) - 1600000000
lock = Lock()
def random_tel():
    global now
    lock.acquire()
    try :
        tel = '130%08d' % (now)
        now += 1
    finally :
        lock.release()
    return tel

def random_username():
    global now
    username = 'user-%08d' % (now)
    now += 1
    return username

def CH(headerName):
    return customHeaderPrefix + headerName

class AccountTest(HttpTest):
    # 超级验证码
    SUPER_TEST_CODE = "0bce718389e18ba44fa98b9da51fc3e3"
    # 用户测试环境获取验证码
    SUPER_KEY = "91af98b3bd246347f8d6eea0573ef7e7"
    # 管理Token
    ADMIN_TOKEN = "2c36a5c195a4f66c1a09046af67126ed"

    def getDefaultHeaders(self):
        headers = {}
        headers[CH("Channel")] = "xiaomi" # app分发渠道
        headers[CH("Platform")] = "test"  # app平台:android/ios/h5/xxx
        headers[CH("Version")] = "0.1.290" # app的版本.
        headers[CH("DeviceID")] = "test-device-id" # app所在设备ID,应该根据唯一算法,生成一个唯一的ID.
        headers[CH("AppID")] = "open-account"
        return deepcopy(headers)

    def getAdminHeaders(self):
        headers = {}
        headers[CH("Channel")] = "h5" # app分发渠道
        headers[CH("Platform")] = "h5-test"  # app平台:android/ios/h5/xxx
        headers[CH("Version")] = "1.0.25" # app的版本.
        headers[CH("DeviceID")] = "test-web-device-id" # app所在设备ID,应该根据唯一算法,生成一个唯一的ID.
        headers[CH("AppID")] = "open-account"
        headers[CH("TOken")] = AccountTest.ADMIN_TOKEN

        return deepcopy(headers)

    def encodePassword(self, password):
        # 密码加密方式, 需要客户端统一就可以.
        return hashlib.sha1(password.encode("utf8")).hexdigest()