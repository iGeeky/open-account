# coding=utf8
import time
import json
import unittest
from pycore import AccountTest, random_tel, random_username
from pycore.utils import get_ok_schema, get_fail_schema, get_user_login_schema
from pycore.schema_gen import set_schema_enums

def get_check_tel_exist_schema(enums={}):
    data_schema = {
        "type": "object",
        "properties": {
            "exist": {"type": "boolean"},
            "tel": { "type": "string" },
            "userType": {"type": "integer"}
        },
        "required": [ "exist", "tel", "userType" ]
    }
    set_schema_enums(data_schema['properties'], enums)
    schema = get_ok_schema(data_schema)
    return schema

def get_check_username_exist_schema(enums={}):
    data_schema = {
        "type": "object",
        "properties": {
            "exist": {"type": "boolean"},
            "username": { "type": "string" },
            "userType": {"type": "integer"}
        },
        "required": [ "exist", "username", "userType" ]
    }
    set_schema_enums(data_schema['properties'], enums)
    schema = get_ok_schema(data_schema)
    return schema

class TestAccount(AccountTest):
    TIMEOUT = 60

    def test_check_exist_tel_not_exist(self):
        headers = self.getDefaultHeaders()
        tel = random_tel()
        schema = get_check_tel_exist_schema(enums={"tel": tel, "exist": False})
        args = {
            "tel": tel,
        }
        res = self.http_get(url='/v1/account/user/check_exist/tel', headers=headers, args=args, status=200, schema=schema)

    def test_check_exist_username_not_exist(self):
        headers = self.getDefaultHeaders()
        username = random_username()
        schema = get_check_username_exist_schema(enums={"username": username, "exist": False})
        args = {
            "username": username,
        }
        res = self.http_get(url='/v1/account/user/check_exist/username', headers=headers, args=args, status=200, schema=schema)

    def test_check_exist_tel_exist(self):
        headers = self.getDefaultHeaders()
        tel = random_tel()
        body = {
            "tel": tel,
            "code": self.SUPER_TEST_CODE,
            "username": random_username(),
        }
        res = self.http_post(url='/v1/account/user/register', headers=headers, body=body, status=200)
        schema = get_check_tel_exist_schema(enums={"tel": tel, "exist": True})
        args = {
            "tel": tel,
        }
        res = self.http_get(url='/v1/account/user/check_exist/tel', headers=headers, args=args, status=200, schema=schema)

    def test_check_exist_username_exist(self):
        headers = self.getDefaultHeaders()
        username = random_username()
        body = {
            "tel": random_tel(),
            "code": self.SUPER_TEST_CODE,
            "username": username,
        }
        res = self.http_post(url='/v1/account/user/register', headers=headers, body=body, status=200)
        schema = get_check_username_exist_schema(enums={"username": username, "exist": True})
        args = {
            "username": username,
        }
        res = self.http_get(url='/v1/account/user/check_exist/username', headers=headers, args=args, status=200, schema=schema)

    def test_user_register_success(self):
        tel = random_tel()
        username = random_username()
        headers = self.getDefaultHeaders()
        body = {
            "tel": tel,
            "code": self.SUPER_TEST_CODE,
            "username": username,
            "userType": 1,
        }
        schema = get_user_login_schema()
        res = self.http_post(url='/v1/account/user/register', headers=headers, body=body, status=200, schema=schema)

    def test_user_register_failed_username_duplicate(self):
        headers = self.getDefaultHeaders()
        username = random_username()
        body = {
            "tel": random_tel(),
            "code": self.SUPER_TEST_CODE,
            "username": username,
            "userType": 1,
        }
        schema = get_user_login_schema()
        res = self.http_post(url='/v1/account/user/register', headers=headers, body=body, status=200, schema=schema)

        body = {
            "tel": random_tel(),
            "code": self.SUPER_TEST_CODE,
            "username": username,
            "userType": 1,
        }
        schema = get_fail_schema('ERR_USERNAME_REGISTED')
        res = self.http_post(url='/v1/account/user/register', headers=headers, body=body, status=200, schema=schema)

    def test_user_register_failed_tel_duplicate(self):
        headers = self.getDefaultHeaders()
        tel = random_tel()
        body = {
            "tel": tel,
            "code": self.SUPER_TEST_CODE,
            "username": random_username(),
            "userType": 1,
        }
        schema = get_user_login_schema()
        res = self.http_post(url='/v1/account/user/register', headers=headers, body=body, status=200, schema=schema)

        body = {
            "tel": tel,
            "code": self.SUPER_TEST_CODE,
            "username": random_username(),
            "userType": 1,
        }
        schema = get_fail_schema('ERR_TEL_REGISTED')
        res = self.http_post(url='/v1/account/user/register', headers=headers, body=body, status=200, schema=schema)


    def test_user_register_diffence_userType_success(self):
        tel = random_tel()
        username = random_username()
        headers = self.getDefaultHeaders()
        body = {
            "tel": tel,
            "code": self.SUPER_TEST_CODE,
            "username": username,
            "userType": 1,
        }
        schema = get_user_login_schema()
        res = self.http_post(url='/v1/account/user/register', headers=headers, body=body, status=200, schema=schema)

        body = {
            "tel": tel,
            "code": self.SUPER_TEST_CODE,
            "username": username,
            "userType": 2,
        }
        schema = get_user_login_schema()
        res = self.http_post(url='/v1/account/user/register', headers=headers, body=body, status=200, schema=schema)

    # 注册邀请码
    # 设置Profile
    # TOKEN 使用.
