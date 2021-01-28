import logging
import os
from copy import deepcopy

log_level = os.environ.get('LOG_LEVEL', 'ERROR') # INFO,WARN,ERROR
if log_level:
    logging.basicConfig(level=log_level, format="%(asctime)s | %(levelname)s | %(filename)s:%(lineno)d | %(message)s")

from . import http_test
HttpTest = http_test.HttpTest

def init():
    server = os.environ.get('SERVER')
    if server:
        HttpTest.SERVER = server

    timeout = os.environ.get('TIMEOUT')
    if timeout:
        HttpTest.TIMEOUT = int(timeout)

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

class AccountTest(HttpTest):
    SUPER_TEST_CODE = "0bce718389e18ba44fa98b9da51fc3e3"

    def getDefaultHeaders(self):
        headers = {}
        headers["X-OA-Channel"] = "xiaomi" # app分发渠道
        headers["X-OA-Platform"] = "test"  # app平台:android/ios/h5/xxx
        headers["X-OA-Version"] = "0.1.290" # app的版本.
        headers["X-OA-DeviceID"] = "test-device-id" # app所在设备ID,应该根据唯一算法,生成一个唯一的ID.
        headers["X-OA-AppID"] = "open-account"
        return deepcopy(headers)
