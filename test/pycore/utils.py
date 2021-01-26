

def get_ok_schema(data_schema={"type": "object" }):
    schema = {
        "type": "object",
        "properties": {
            "data": data_schema,
            "ok": { "type": "boolean", "enum": [True] },
            "reason": { "type": "string", "enum": [""]}
        },
        "required": [ "ok","reason", "data" ]
    }
    return schema

def get_fail_schema(reason=""):
    schema = {
        "type": "object",
        "properties": {
            "data": {"type": "object"},
            "ok": { "type": "boolean", "enum": [False] },
            "reason": { "type": "string", "enum": [reason]}
        },
        "required": [ "ok","reason" ]
    }
    return schema

def get_userinfo_detail_schema():
    schema = {
        "type": "object",
        "properties": {
            "id": { "type": "integer" },
            "uid": { "type": "string"},
            "tel": { "type": "string"},
            "nickname": { "type": "string"},
            "avatar": { "type": "string" },
            "sex": { "type": "integer"},
            "birthday": { "type": "string" },
            "userType": { "type": "integer"},
            "regInviteCode": { "type": "string" },
            "inviteCode": { "type": "string"},
            "createTime": { "type": "integer"}
        },
        "required": [ "id", "uid", "tel", "nickname", "avatar", "sex", "birthday", "userType", "regInviteCode", "inviteCode", "createTime"]
    }
    return schema


def get_user_login_schema():
    data_schema = {
        "type": "object",
        "properties": {
            "token": { "type": "string" },
            "userInfo": get_userinfo_detail_schema()
        },
        "required": [ "token", "userInfo" ]
    }
    schema = get_ok_schema(data_schema)
    return schema
