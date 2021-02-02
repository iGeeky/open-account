/**
//手动执行脚本,创建数据库,账号.
create database `openaccount` CHARACTER SET utf8mb4;
use openaccount;
CREATE USER 'openaccount'@'127.0.0.1' IDENTIFIED BY '123456';
CREATE USER 'openaccount'@'localhost' IDENTIFIED BY '123456';
CREATE USER 'openaccount'@'%' IDENTIFIED BY '123456';

grant DELETE,EXECUTE,INSERT,SELECT,UPDATE on openaccount.* to 'openaccount'@'127.0.0.1';
grant DELETE,EXECUTE,INSERT,SELECT,UPDATE on openaccount.* to 'openaccount'@'localhost';
grant DELETE,EXECUTE,INSERT,SELECT,UPDATE on openaccount.* to 'openaccount'@'%';

FLUSH PRIVILEGES;
**/

use openaccount;

create table if not exists `user` (
    `id` bigint not null primary key auto_increment comment '用户唯一ID',
    `uid` varchar(64) null comment '用户UID,非数字',
    `user_type` tinyint(4) default 1 NOT NULL COMMENT '用户类型(1:普通用户)',
    `username` varchar(256) null comment '用户名',
    `tel` varchar(16) null comment '用户手机号',
    `password` varchar(64) null comment '用户密码',
    `nickname` varchar(256) null comment '用户昵称',
    `avatar` varchar(512) null comment '用户头像地址',
    `sex` tinyint default 0 comment '用户性别, 0: 未知, 1: 男, 2: 女, 3:其它',
    `birthday` varchar(64) null comment '生日: yyyy-mm-dd',
    `reg_invite_code` varchar(64) null comment '注册输入的邀请码(别人的)',
    `invite_code` varchar(64) null comment '我的邀请码',
    `status` tinyint default 0 comment '帐号状态：0，正常',
    `level` smallint default 0 comment '帐号等级',
    `channel` varchar(128) null comment "用户渠道",
    `platform` varchar(64) null comment "APP/客户端平台: ios/android/h5/xxx",
    `version` varchar(128) null comment "客户端版本: 1.120.999",
    `device_id` varchar(128) null comment "设备id",
    `ip` varchar(32) null comment '用户IP',
    `create_time` int not null comment '创建时间',
    `update_time` int not null comment '修改时间',
    `profile` JSON null comment '概况,配置信息等',
    UNIQUE KEY(`uid`),
    UNIQUE KEY(`user_type`, `tel`),
    KEY(`user_type`, `username`),
    KEY(`invite_code`),
    KEY(`create_time`)
)ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE utf8mb4_bin comment '用户表';

create table if not exists user_login_log (
    `id` bigint not null primary key auto_increment comment '自动增长id',
    `user_id` bigint not null,
    `device_id` varchar(128) null comment "设备id",
    `login_ip` varchar(32) null comment '用户登录IP',
    `country_code` varchar(32) default null,
    `city_name` varchar(32) default null,
    `channel` varchar(128) null comment "用户渠道",
    `platform` varchar(64) null comment "APP/客户端平台: ios/android/h5/xxx",
    `version` varchar(128) null comment "客户端版本: 1.120.999",
    create_time int not null,
    KEY(user_id),
    KEY(login_ip),
    KEY(create_time)
)ENGINE=MyISAM DEFAULT CHARSET=utf8mb4 COLLATE utf8mb4_bin comment '用户登录日志表';
