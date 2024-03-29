CREATE TABLE `test_case_log` (
  `id` bigint(20) unsigned NOT NULL AUTO_INCREMENT comment '自增ID',
  `host` varchar(100) DEFAULT '' comment '服务域名' ,
  `method` varchar(10) not null default '' comment '接口请求方法',
  `path` varchar(200) not null default '' comment '接口path',
  `request_id` varchar(64) not null default '' comment '请求ID',
  `request_header` text not null default '' comment '请求头',
  `request_query` text not null default '' comment '请求查询',
  `request_body` text not null default '' comment '请求体',
  `response_header` text not null default '' comment '响应头',
  `response_body` text not null default '' comment '响应体',
  `test_case_id` varchar(120) not null  DEFAULT "" comment '测试用例ID',
  `description` varchar(120) not null  DEFAULT "" comment '测试用例描述',
  `test_case_input` text not null default '' comment '测试用例期望输入',
  `test_case_output` text not null default '' comment '测试用例期望输出',
  `created_at` datetime not null DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `updated_at` datetime not null DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '修改时间',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_request_id` (`request_id`) USING BTREE,
  KEY `idx_test_case_id` (`test_case_id`)
) ENGINE=InnoDB AUTO_INCREMENT=1 DEFAULT CHARSET=utf8mb4;