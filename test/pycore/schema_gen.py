# coding=utf8



def get_type(value):
    t = type(value)
    if t == dict:
        t = 'object'
    elif t == list:
        t = 'array'
    elif value == None:
        t = 'null'
    elif t == str:
        t = 'string'
    elif t == int:
        t = 'integer'
    elif t == float:
        t = 'number'
    elif t == bool:
        t = 'boolean'
    else:
        t = 'unknow'
    return t


def generate_schema(field, value, **opts):
    t = get_type(value)
    schema = { "type": t }
    opts = opts or {}
    enums = opts.get("enums", False)
    forceEnumFields = opts.get("forceEnumFields", {})
    deep = opts.get("deep", 10)
    curLevel = opts.get("curLevel", 0)

    level = curLevel + 1
    opts["curLevel"] = level

    if t == 'object':
        if level <= deep:
            properties = {}
            required = []
            subFields = value.keys()
            for subField in subFields:
                childValue = value[subField]
                properties[subField] = generate_schema(subField, childValue, **opts.copy())
                required.append(subField)
            schema["properties"] = properties
            schema["required"] = required
    elif t == 'array':
        if level <= deep and len(value) > 0:
            schema["items"] = generate_schema(None, value[0], **opts.copy())
    elif t == 'number' or t == 'float' or t == 'string' or t == 'integer' or t == 'boolean':
        if enums or (field and forceEnumFields and forceEnumFields[field]):
            schema["enum"] = [value]
    elif t == 'null': # null的不自动生成,指定null容易有出错的情况.
        del(schema["type"])
    else:
        raise BaseException('UnKnown type:%s, value:%s' % (t, value))
    return schema

def auto_schema(value, **opts):
    return generate_schema(None, value, **opts)

def set_schema_enums(schema, enums):
    for field in enums:
        field_schema = schema.get(field)
        if field_schema:
            enum_value = enums[field]
            if type(enum_value) != list:
                enum_value = [enum_value]
            field_schema["enum"] = enum_value
