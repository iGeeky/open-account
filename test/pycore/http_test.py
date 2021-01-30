# coding=utf8

import unittest
from . import http_client as http
from . import schema_gen
import json
import time
import re
import sys
import jsonschema as jschema


"""
You can use these ANSI escape codes:

Black        0;30     Dark Gray     1;30
Red          0;31     Light Red     1;31
Green        0;32     Light Green   1;32
Brown/Orange 0;33     Yellow        1;33
Blue         0;34     Light Blue    1;34
Purple       0;35     Light Purple  1;35
Cyan         0;36     Light Cyan    1;36
Light Gray   0;37     White         1;37
And then use them like this in your script:

#    .---------- constant part!
#    vvvv vvvv-- the code from above
RED='\033[0;31m'
NC='\033[0m' # No Color
printf "I ${RED}love${NC} Stack Overflow\n"
"""
RED_ = "\033[0;31m"
GREEN_ = "\033[0;32m"
YELLOW_ = "\033[1;33m"

NC_ = "\033[0m"

def RED(msg):
    return RED_ + msg + NC_

def GREEN(msg):
    return GREEN_ + msg + NC_

def YELLOW(msg):
    return YELLOW_ + msg + NC_


def is_json(myjson):
    try:
        json_object = json.loads(myjson)
    except ValueError as e:
        return False
    return True

class HttpTest(unittest.TestCase):
    SERVER = "http://127.0.0.1:2021"
    TIMEOUT = 300.0
    # 执行总时间
    total_time = 0.0
    # 执行总次数
    total_num = 0

    def __init__(self, methodName='runTest'):
        unittest.TestCase.__init__(self, methodName)
        opts = {}
        self.longMessage = True
        self.server = self.SERVER # 服务器地址
        self.timeout = self.TIMEOUT
        # 记录当前正在执行的测试结果
        self.current_res = {}

    @classmethod
    def tearDownClass(cls):
        total_num = HttpTest.total_num
        if total_num < 1:
            total_num = 1
        print("\n%s: Seconds(%.3f)/Requests(%s): %.3f" % (cls.__name__, HttpTest.total_time, HttpTest.total_num, HttpTest.total_time/total_num))

    # 出错处理的回调.
    def onFailed(self, test_name, errmsgs, **kwargs):
        pass

    def run(self, result=None):
        errors = len(result.errors)
        failures = len(result.failures)

        result = unittest.TestCase.run(self, result)
        if not result.wasSuccessful():
            test_name = result.getDescription(self)
            errmsgs = []
            if len(result.errors) > errors:
                errmsgs.append("------------------------ ERROR DETAIL ------------------\n")
                for test, err in result.errors[errors:]:
                    errmsgs.append(err)
            if len(result.failures) > failures:
                errmsgs.append("----------------- FAIL  DETAIL ------------------\n")
                for test, err in result.failures[failures:]:
                    errmsgs.append(err)
            if len(errmsgs) > 0:
                self.onFailed(test_name, errmsgs, result=result, res=self.current_res)

        return result

    def check_response(self, res, status, schema, match, notMatch, kwargs):
        showSchema = kwargs.get("showSchema", False)
        body = ""
        if res.body:
            if isinstance(res.body, str):
                body = res.body
            else:
                body = str(res.body, encoding = "utf-8")
        if showSchema:
            deep = kwargs.get("deep", 4)
            method = res.rawResp.request.method
            path = res.rawResp.url
            print('/************** request: [%s %s] ****************/' % (method, path))
            if(body):
                if is_json(body):
                    print('/***************** schema of data(deep: %d) ******************/' %(deep))
                    schema_debug = schema_gen.auto_schema(json.loads(body), deep=deep)
                    print(json.dumps(schema_debug, indent=4))
                else:
                    print("Couldn't show schema, because the response body is not a valid json.")
            else:
                schema_debug = schema_gen.auto_schema(body, deep=deep)
                print(json.dumps(schema_debug, indent=4))

        if not status:
            status = 200
        msg = RED("expect status (%s), but res.status (%s)" % (status, res.status))
        self.assertEqual(res.status, status, msg=msg)

        if match:
            if type(match) == list:
                for pattern in match:
                    self.assertRegex(body, pattern, msg=RED("response text not matched regex: %s" %(pattern)))
            else:
                pattern= match
                self.assertRegex(body, pattern, msg=RED("response text not matched regex: %s" %(pattern)))
        if notMatch:
            if type(notMatch) == list:
                for pattern in notMatch:
                    self.assertNotRegex(body, pattern, msg=RED("response text matched regex: %s" %(pattern)))
            else:
                pattern= notMatch
                self.assertNotRegex(body, pattern, msg=RED("response text matched regex: %s" %(pattern)))

        if schema:
            schemaEx = None
            try:
                body = json.loads(body)
                jschema.validate(body, schema)
            except jschema.exceptions.ValidationError as ex:
                schemaEx = ex
            except ValueError as ex:
                schemaEx = ex
            if schemaEx:
                self.fail(RED("request [" + res.req_debug + "]'s respone body is invalid:\n" + str(schemaEx)))


    def http(self, **kwargs):
        """
        ### request options:

        - method: http method
        - url: request url
        - headers: request headers
        - args: request args for get request
        - body: request body for post/put/delete request.
        - status: expect response status, default is 200.
        - schema: expect response body schema for restful response
        - match: regex pattern for matching response body.
        - notMatch: regex pattern for not matching response body.
        """
        method = kwargs.get('method', 'GET')
        url = kwargs.get('url', '')
        headers = kwargs.get('headers', {})
        args = kwargs.get('args', {})
        body = kwargs.get('body')
        if body:
            if isinstance(body, dict):
                body = json.dumps(body)
        if url.startswith('/'):
            full_url = "%s%s" % (self.server , url)
        else:
            full_url = url
        res = http.HttpReq(method, full_url, headers, args, body, self.timeout)
        self.current_res = res
        status = kwargs.get("status", None)
        schema = kwargs.get('schema')
        match = kwargs.get('match')
        notMatch = kwargs.get('notMatch')
        statTime = kwargs.get('statTime', True)
        if statTime:
            HttpTest.total_time += res.duration
            HttpTest.total_num += 1

        self.check_response(res, status, schema, match, notMatch, kwargs)
        return res

    def http_get(self, **kwargs):
        """
        ### request options:

        - url: request url
        - headers: request headers
        - args: request args for get request
        - status: expect response status, default is 200.
        - schema: expect response body schema for restful response
        - match: regex pattern for matching response body.
        - notMatch: regex pattern for not matching response body.
        """
        return self.http(method='GET', **kwargs)

    def http_post(self, **kwargs):
        """
        ### request options:

        - url: request url
        - headers: request headers
        - body: request body for post/put/delete request
        - status: expect response status, default is 200.
        - schema: expect response body schema for restful response
        - match: regex pattern for matching response body.
        - notMatch: regex pattern for not matching response body.
        """
        return self.http(method='POST', **kwargs)

    def http_put(self, **kwargs):
        """
        ### request options:

        - url: request url
        - headers: request headers
        - body: request body for post/put/delete request
        - status: expect response status, default is 200.
        - schema: expect response body schema for restful response
        - match: regex pattern for matching response body.
        - notMatch: regex pattern for not matching response body.
        """
        return self.http(method='PUT', **kwargs)

    def http_delete(self, **kwargs):
        """
        ### request options:

        - url: request url
        - headers: request headers
        - body: request body for post/put/delete request
        - status: expect response status, default is 200.
        - schema: expect response body schema for restful response
        - match: regex pattern for matching response body.
        - notMatch: regex pattern for not matching response body.
        """
        return self.http(method='DELETE', **kwargs)

