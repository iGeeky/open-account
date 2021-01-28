# coding=utf8
import time
import json
import unittest
from pycore import AccountTest, random_tel, random_username
from pycore.utils import get_ok_schema, get_fail_schema, get_user_login_schema, get_userinfo_schema
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

    def test_user_register_with_invite_code(self):
        tel = random_tel()
        username = random_username()
        headers = self.getDefaultHeaders()
        inviteCode = "test-invite-code"
        body = {
            "tel": tel,
            "code": self.SUPER_TEST_CODE,
            "username": username,
            "userType": 1,
            "inviteCode": inviteCode,
        }
        schema = get_user_login_schema(enums={"regInviteCode": inviteCode})
        res = self.http_post(url='/v1/account/user/register', headers=headers, body=body, status=200, schema=schema)


    def test_user_get_userinfo_failed_token_missing(self):
        tel = random_tel()
        username = random_username()
        headers = self.getDefaultHeaders()
        inviteCode = "test-invite-code"
        body = {
            "tel": tel,
            "code": self.SUPER_TEST_CODE,
            "username": username,
            "userType": 1,
            "inviteCode": inviteCode,
        }
        schema = get_user_login_schema(enums={"regInviteCode": inviteCode})
        res = self.http_post(url='/v1/account/user/register', headers=headers, body=body, status=200, schema=schema)
        schema = get_fail_schema('ERR_TOKEN_INVALID')
        res = self.http_get(url='/v1/account/user/userinfo', headers=headers, status=401, schema=schema)

    def test_user_get_userinfo_failed_token_invlid(self):
        tel = random_tel()
        username = random_username()
        headers = self.getDefaultHeaders()
        inviteCode = "test-invite-code"
        body = {
            "tel": tel,
            "code": self.SUPER_TEST_CODE,
            "username": username,
            "userType": 1,
            "inviteCode": inviteCode,
        }
        schema = get_user_login_schema(enums={"regInviteCode": inviteCode})
        res = self.http_post(url='/v1/account/user/register', headers=headers, body=body, status=200, schema=schema)
        token = res.json["data"]["token"]

        headers["X-OA-Token"] = "TOKEN-INVALID"
        schema = get_fail_schema('ERR_TOKEN_INVALID')
        res = self.http_get(url='/v1/account/user/userinfo', headers=headers, status=401, schema=schema)

    def test_user_get_set_userinfo_success(self):
        tel = random_tel()
        username = random_username()
        headers = self.getDefaultHeaders()
        inviteCode = "test-invite-code"
        body = {
            "tel": tel,
            "code": self.SUPER_TEST_CODE,
            "username": username,
            "userType": 1,
            "inviteCode": inviteCode,
            "profile": {"id": "test01", "city": "sz"} # 字段需要在服务器端定义
        }
        schema = get_user_login_schema(enums={"regInviteCode": inviteCode})
        res = self.http_post(url='/v1/account/user/register', headers=headers, body=body, status=200, schema=schema)
        token = res.json["data"]["token"]
        headers["X-OA-Token"] = token
        schema = get_userinfo_schema(enums={"tel": tel, "username": username, "regInviteCode": inviteCode})
        res = self.http_get(url='/v1/account/user/userinfo', headers=headers, status=200, schema=schema)

        # 设置用户信息
        body = {
            "avatar": "https://avatars.githubusercontent.com/u/57409417?s=460&u=1fff104b9010d84a47ecf12c5abce96436fc7b8b&v=4",
            "nickname": "igeeky",
            "sex": 1,
            "birthday": "2021-01-18",
        }
        schema = get_userinfo_schema(enums=body)
        res = self.http_put(url='/v1/account/user/userinfo', headers=headers, body=body, status=200, schema=schema)


    def test_user_logout(self):
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
        token = res.json["data"]["token"]
        headers["X-OA-Token"] = token
        schema = get_ok_schema()
        body = {}
        res = self.http_post(url='/v1/account/user/logout', headers=headers, body=body, status=200, schema=schema)

        # 第二次logout失败.
        schema = get_fail_schema('ERR_TOKEN_EXPIRED')
        res = self.http_post(url='/v1/account/user/logout', headers=headers, body=body, status=401, schema=schema)

