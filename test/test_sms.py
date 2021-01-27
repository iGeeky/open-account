# coding=utf8
import time
import json
import unittest
from pycore import AccountTest, random_tel
from pycore.utils import get_ok_schema, get_fail_schema, get_user_login_schema

def get_sms_code_schema():
    data_schema = {
        "type": "object",
        "properties": {
            "code": {
                "type": "string"
            }
        },
        "required": [
            "code"
        ]
    }
    schema = get_ok_schema(data_schema)
    return schema

class TestSMS(AccountTest):
    TIMEOUT = 60
    tel = random_tel()
    def test_sms_send(self):
        body = {
            "tel": self.tel,
        }
        headers = self.getDefaultHeaders()
        schema = get_ok_schema()
        res = self.http_post(url='/v1/account/user/sms/send', headers=headers, body=body, status=200, schema=get_ok_schema())
        # print(res.json)

    def test_sms_login_ok(self):
        headers = self.getDefaultHeaders()
        # 发送验证码
        body = {
            "tel": self.tel,
        }
        schema = get_sms_code_schema()
        res = self.http_post(url='/v1/account/user/sms/send', headers=headers, body=body, status=200, schema=get_ok_schema())
        # 通过后门接口, 查询验证码.
        args = {
            "bizType": "login",
            "tel": self.tel,
            "key": "91af98b3bd246347f8d6eea0573ef7e7"
        }
        schema = get_sms_code_schema()
        res = self.http_get(url='/v1/account/sms/get/code', args=args, status=200, schema=schema)
        code = res.json["data"]["code"]

        # 验证码登录.
        body = {
            "tel": self.tel,
            "code": code,
        }
        schema = get_user_login_schema()
        res = self.http_post(url='/v1/account/user/sms/login', headers=headers, body=body, status=200, schema=schema)

    def test_sms_login_failed_code_invalid(self):
        headers = self.getDefaultHeaders()
        body = {
            "tel": self.tel,
            "code": "000000",
        }
        schema = get_fail_schema('ERR_CODE_INVALID')
        res = self.http_post(url='/v1/account/user/sms/login', headers=headers, body=body, status=200, schema=schema)
        # print("body: %s" % (res.json))
