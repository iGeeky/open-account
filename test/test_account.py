# coding=utf8
import time
import json
import unittest
from pycore import AccountTest, random_tel, random_username, CH
from pycore.utils import get_ok_schema, get_fail_schema, get_user_login_schema, get_userinfo_schema, get_sms_code_schema
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

def get_invite_code_settable_schema(settable=True):
    data_schema = {
        "type": "object",
        "properties": {
            "settable": {"type": "boolean"},
        },
        "required": [ "settable" ]
    }
    enums = {"settable": settable}
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

    def userRegister(self, tel, username, userType=1, password=None, inviteCode=None, 
                    profile=None, headers=None, login_enums={}):
        if not headers:
            headers = self.getDefaultHeaders()
        body = {
            "tel": tel,
            "code": self.SUPER_TEST_CODE,
            "username": username,
            "userType": userType,
        }
        if password:
            body["password"] = password
        if inviteCode:
            body["inviteCode"] = inviteCode
        if profile:
            body["profile"] = profile

        schema = get_user_login_schema(enums=login_enums)
        res = self.http_post(url='/v1/account/user/register', headers=headers, body=body, status=200, schema=schema)
        data = res.json["data"]
        token = data["token"]
        userInfo = data["userInfo"]
        return token, userInfo

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
        self.userRegister(tel,username)

    def test_user_register_failed_username_duplicate(self):
        headers = self.getDefaultHeaders()
        username = random_username()
        self.userRegister(random_tel(),username)

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
        username = random_username()
        self.userRegister(tel,username)

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
        self.userRegister(tel,username, userType=1)
        self.userRegister(tel,username, userType=2)

    def test_user_register_with_invite_code(self):
        tel = random_tel()
        username = random_username()
        inviteCode = "test-invite-code"
        self.userRegister(tel,username, inviteCode=inviteCode, login_enums={"regInviteCode": inviteCode })

    def test_user_get_userinfo_failed_token_missing(self):
        headers = self.getDefaultHeaders()
        schema = get_fail_schema('ERR_TOKEN_INVALID')
        res = self.http_get(url='/v1/account/user/userinfo', headers=headers, status=401, schema=schema)

    def test_user_get_userinfo_failed_token_invlid(self):
        headers = self.getDefaultHeaders()
        headers[CH("Token")] = "TOKEN-INVALID"
        schema = get_fail_schema('ERR_TOKEN_INVALID')
        res = self.http_get(url='/v1/account/user/userinfo', headers=headers, status=401, schema=schema)

    def test_user_get_set_userinfo_success(self):
        tel = random_tel()
        username = random_username()
        inviteCode = "test-invite-code"
        profile = {"id": "test01", "city": "sz"} # 字段需要在服务器端定义
        token, _ = self.userRegister(tel, username, inviteCode=inviteCode, profile=profile, login_enums={"regInviteCode": inviteCode})

        headers = self.getDefaultHeaders()
        headers[CH("Token")] = token
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
        token, _ = self.userRegister(tel, username)

        headers = self.getDefaultHeaders()
        headers[CH("Token")] = token
        schema = get_ok_schema()
        body = {}
        res = self.http_post(url='/v1/account/user/logout', headers=headers, body=body, status=200, schema=schema)

        # 第二次logout失败.
        schema = get_fail_schema('ERR_TOKEN_EXPIRED')
        res = self.http_post(url='/v1/account/user/logout', headers=headers, body=body, status=401, schema=schema)

    def test_user_register_empty_password(self):
        tel = random_tel()
        username = random_username()
        headers = self.getDefaultHeaders()
        emptyPassword = self.encodePassword("empty-password")
        self.userRegister(tel, username)

        body = {
            "tel": tel,
            "password": emptyPassword,
        }
        schema = get_fail_schema('ERR_PASSWORD_ERR')
        res = self.http_post(url="/v1/account/user/login", headers=headers, body=body, status=200, schema=schema)

    def test_user_register_login_success(self):
        tel = random_tel()
        username = random_username()
        headers = self.getDefaultHeaders()
        password = self.encodePassword("password")
        self.userRegister(tel, username, password=password)


        schema = get_fail_schema('ERR_ARGS_INVALID')
        body = {
            "password": password,
        }
        res = self.http_post(url="/v1/account/user/login", headers=headers, body=body, status=400, schema=schema)

        # 手机号登录
        schema = get_user_login_schema()
        body = {
            "tel": tel,
            "password": password,
        }
        res = self.http_post(url="/v1/account/user/login", headers=headers, body=body, status=200, schema=schema)
        # 用户名登录
        body = {
            "username": username,
            "password": password,
        }
        res = self.http_post(url="/v1/account/user/login", headers=headers, body=body, status=200, schema=schema)


    def test_user_register_login_failed(self):
        tel = random_tel()
        username = random_username()
        headers = self.getDefaultHeaders()
        password = self.encodePassword("password")
        errPassword = self.encodePassword("err-password")
        self.userRegister(tel, username, password=password)

        body = {
            "tel": tel,
            "password": errPassword,
        }
        schema = get_fail_schema('ERR_PASSWORD_ERR')
        res = self.http_post(url="/v1/account/user/login", headers=headers, body=body, status=200, schema=schema)

    def test_user_change_pwd_failed_old_pwd_error(self):
        tel = random_tel()
        username = random_username()
        oldPassword = self.encodePassword("old-password")
        newPassword = self.encodePassword("new-password")
        errPassword = self.encodePassword("err-password")

        token, _ = self.userRegister(tel, username, password=oldPassword)

        headers = self.getDefaultHeaders()
        headers[CH("Token")] = token

        body = {
            "oldPassword": errPassword,
            "password": newPassword,
        }
        schema = get_fail_schema('ERR_PASSWORD_ERR')
        res = self.http_put(url="/v1/account/user/password", headers=headers, body=body, status=200, schema=schema)

    def test_user_change_pwd_success(self):
        tel = random_tel()
        username = random_username()
        oldPassword = self.encodePassword("old-password")
        newPassword = self.encodePassword("new-password")
        token, _ = self.userRegister(tel, username, password=oldPassword)

        headers = self.getDefaultHeaders()
        headers[CH("Token")] = token

        # 修改密码
        body = {
            "oldPassword": oldPassword,
            "password": newPassword,
        }
        schema = get_ok_schema()
        res = self.http_put(url="/v1/account/user/password", headers=headers, body=body, status=200, schema=schema)

        # 使用旧密码登录失败
        body = {
            "tel": tel,
            "password": oldPassword,
        }
        schema = get_fail_schema('ERR_PASSWORD_ERR')
        res = self.http_post(url="/v1/account/user/login", headers=headers, body=body, status=200, schema=schema)

        # 使用新密码登录成功
        body = {
            "tel": tel,
            "password": newPassword,
        }
        schema = get_user_login_schema(enums={})
        res = self.http_post(url="/v1/account/user/login", headers=headers, body=body, status=200, schema=schema)

    def test_user_reset_pwd_failed_sms_code_invalid(self):
        tel = random_tel()
        username = random_username()
        oldPassword = self.encodePassword("old-password")
        newPassword = self.encodePassword("new-password")
        token, _ = self.userRegister(tel, username, password=oldPassword)

        headers = self.getDefaultHeaders()
        headers[CH("Token")] = token

        # 发送验证码
        body = {
            "tel": tel,
            "bizType": "login",
        }
        schema = get_ok_schema()
        res = self.http_post(url='/v1/account/user/sms/send', headers=headers, body=body, status=200, schema=schema)

        # 通过后门接口, 查询验证码.
        args = {
            "bizType": "login",
            "tel": tel,
            "key": AccountTest.SUPER_KEY
        }
        schema = get_sms_code_schema()
        res = self.http_get(url='/v1/man/account/sms/get/code', args=args, status=200, schema=schema)
        code = res.json["data"]["code"]

        # 重置密码, 验证码类型错误
        body = {
            "tel": tel,
            "code": code,
            "password": newPassword,
        }
        schema = get_fail_schema('ERR_CODE_INVALID')
        res = self.http_put(url="/v1/account/user/password/reset", headers=headers, body=body, status=200, schema=schema)

    def test_user_reset_pwd_failed_tel_not_exist(self):
        tel = random_tel()
        username = random_username()
        headers = self.getDefaultHeaders()
        oldPassword = self.encodePassword("old-password")
        newPassword = self.encodePassword("new-password")

        # 发送验证码
        body = {
            "tel": tel,
            "bizType": "resetPwd",
        }
        schema = get_ok_schema()
        res = self.http_post(url='/v1/account/user/sms/send', headers=headers, body=body, status=200, schema=schema)

        # 通过后门接口, 查询验证码.
        args = {
            "bizType": "resetPwd",
            "tel": tel,
            "key": AccountTest.SUPER_KEY
        }
        schema = get_sms_code_schema()
        res = self.http_get(url='/v1/man/account/sms/get/code', args=args, status=200, schema=schema)
        code = res.json["data"]["code"]

        # 重置密码, 手机号不存在
        body = {
            "tel": tel,
            "code": code,
            "password": newPassword,
        }
        schema = get_fail_schema('ERR_TEL_NOT_EXIST')
        res = self.http_put(url="/v1/account/user/password/reset", headers=headers, body=body, status=200, schema=schema)

    def test_user_reset_pwd_success(self):
        tel = random_tel()
        username = random_username()
        oldPassword = self.encodePassword("old-password")
        newPassword = self.encodePassword("new-password")
        token, _ = self.userRegister(tel, username, password=oldPassword)

        headers = self.getDefaultHeaders()
        headers[CH("Token")] = token

        # 发送验证码
        body = {
            "tel": tel,
            "bizType": "resetPwd",
        }
        schema = get_ok_schema()
        res = self.http_post(url='/v1/account/user/sms/send', headers=headers, body=body, status=200, schema=schema)

        # 通过后门接口, 查询验证码.
        args = {
            "bizType": "resetPwd",
            "tel": tel,
            "key": AccountTest.SUPER_KEY
        }
        schema = get_sms_code_schema()
        res = self.http_get(url='/v1/man/account/sms/get/code', args=args, status=200, schema=schema)
        code = res.json["data"]["code"]

        # 重置密码
        body = {
            "tel": tel,
            "code": code,
            "password": newPassword,
        }
        schema = get_ok_schema()
        res = self.http_put(url="/v1/account/user/password/reset", headers=headers, body=body, status=200, schema=schema)

        # 使用旧密码登录失败
        body = {
            "tel": tel,
            "password": oldPassword,
        }
        schema = get_fail_schema('ERR_PASSWORD_ERR')
        res = self.http_post(url="/v1/account/user/login", headers=headers, body=body, status=200, schema=schema)

        # 使用新密码登录成功
        body = {
            "tel": tel,
            "password": newPassword,
        }
        schema = get_user_login_schema(enums={})
        res = self.http_post(url="/v1/account/user/login", headers=headers, body=body, status=200, schema=schema)

    def test_user_set_invite_code(self):
        tel = random_tel()
        username = random_username()
        inv_tel = random_tel()
        inv_username = random_username()

        oldPassword = self.encodePassword("old-password")
        newPassword = self.encodePassword("new-password")
        # 邀请的用户
        token1, userInfo = self.userRegister(inv_tel, inv_username)
        inviteCode = userInfo["inviteCode"]
        headers = self.getDefaultHeaders()
        headers[CH("Token")] = token1
        # 设置自己的邀请码
        body = { "inviteCode": inviteCode }
        schema = get_fail_schema('ERR_INVITE_CODE_INVALID')
        res = self.http_put(url="/v1/account/user/invite_code", headers=headers, body=body, status=200, schema=schema)

        # 被邀请的用户
        token2, _ = self.userRegister(tel, username, password=oldPassword)
        headers = self.getDefaultHeaders()
        headers[CH("Token")] = token2

        # 检查是否可设置邀请码
        schema = get_invite_code_settable_schema(True)
        self.http_get(url="/v1/account/user/invite_code/settable", headers=headers, status=200, schema=schema)

        # 设置不存在的邀请码
        body = { "inviteCode": 'not-exist-invite-code' }
        schema = get_fail_schema('ERR_INVITE_CODE_INVALID')
        res = self.http_put(url="/v1/account/user/invite_code", headers=headers, body=body, status=200, schema=schema)

        # 设置成功
        body = { "inviteCode": inviteCode}
        schema = get_ok_schema()
        res = self.http_put(url="/v1/account/user/invite_code", headers=headers, body=body, status=200, schema=schema)

        # 重复设置成功
        body = { "inviteCode": inviteCode}
        schema = get_ok_schema()
        res = self.http_put(url="/v1/account/user/invite_code", headers=headers, body=body, status=200, schema=schema)


    def test_user_set_username(self):
        tel = random_tel()
        token, _ = self.userRegister(tel, "")
        headers = self.getDefaultHeaders()
        headers[CH("Token")] = token

        # 设置用户名
        username = random_username()
        body = { "username": username}
        schema = get_ok_schema()
        self.http_put(url="/v1/account/user/username", headers=headers, body=body, status=200, schema=schema)

        # 再次用户名,出错
        schema = get_fail_schema('ERR_USER_HAVE_USERNAME')
        self.http_put(url="/v1/account/user/username", headers=headers, body=body, status=200, schema=schema)


    def test_user_manager_reset_pwd_success(self):
        tel = random_tel()
        username = random_username()
        oldPassword = self.encodePassword("old-password")
        newPassword = self.encodePassword("new-password")
        token, _ = self.userRegister(tel, username, password=oldPassword)

        headers = self.getDefaultHeaders()
        headers[CH("Token")] = token

        # 后台重置密码
        body = {
            "tel": tel,
            "password": newPassword,
        }
        schema = get_ok_schema()
        res = self.http_put(url="/v1/man/account/user/password/reset", headers=self.getAdminHeaders(), body=body, status=200, schema=schema)

        # 使用旧密码登录失败
        body = {
            "tel": tel,
            "password": oldPassword,
        }
        schema = get_fail_schema('ERR_PASSWORD_ERR')
        res = self.http_post(url="/v1/account/user/login", headers=headers, body=body, status=200, schema=schema)

        # 使用新密码登录成功
        body = {
            "tel": tel,
            "password": newPassword,
        }
        schema = get_user_login_schema(enums={})
        res = self.http_post(url="/v1/account/user/login", headers=headers, body=body, status=200, schema=schema)

    def test_user_manager_deregister(self):
        tel = random_tel()
        username = random_username()
        password = self.encodePassword("password")
        token, userInfo = self.userRegister(tel, username, password=password)

        headers = self.getDefaultHeaders()
        headers[CH("Token")] = token

        # 后台注销注册用户
        body = {
            "id": userInfo['id']
        }
        schema = get_ok_schema()
        res = self.http_delete(url="/v1/man/account/user/deregister", headers=self.getAdminHeaders(), body=body, status=200, schema=schema)

        schema = get_fail_schema('ERR_TOKEN_EXPIRED')
        res = self.http_get(url='/v1/account/user/userinfo', headers=headers, status=401, schema=schema)

    def test_user_manager_lock_user(self):
        tel = random_tel()
        username = random_username()
        password = self.encodePassword("password")
        token, userInfo = self.userRegister(tel, username, password=password)

        headers = self.getDefaultHeaders()
        headers[CH("Token")] = token

        # 后台注销注册用户
        UserStatusDisabled = -1
        body = {
            "id": userInfo['id'],
            "status": UserStatusDisabled,
        }
        schema = get_ok_schema()
        res = self.http_put(url="/v1/man/account/user/status", headers=self.getAdminHeaders(), body=body, status=200, schema=schema)

        # 需要token的操作, 提示用户被锁定
        schema = get_fail_schema('ERR_USER_IS_LOCKED')
        res = self.http_get(url='/v1/account/user/userinfo', headers=headers, status=200, schema=schema)

        # 重新登录,提示用户被锁定
        body = {
            "tel": tel,
            "password": password,
        }
        schema = get_fail_schema('ERR_USER_IS_LOCKED')
        res = self.http_post(url="/v1/account/user/login", headers=headers, body=body, status=200, schema=schema)