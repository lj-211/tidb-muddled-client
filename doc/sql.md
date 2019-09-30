# db
## 批次锁
```
CREATE TABLE `client_test`.`batch_lock` (
  `id` int(11) unsigned NOT NULL AUTO_INCREMENT,
  `batch_id` varchar(128) CHARACTER SET utf8mb4 COLLATE utf8mb4_bin NOT NULL DEFAULT '' COMMENT '批次id',
  PRIMARY KEY (`id`),
  UNIQUE KEY `idx_batch_id` (`batch_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci AUTO_INCREMENT=7208 COMMENT='批次锁';
```
## 批次节点信息表
```
CREATE TABLE `client_test`.`coordinate_info` (
  `id` int(11) unsigned NOT NULL AUTO_INCREMENT,
  `created_at` int(11) unsigned NOT NULL DEFAULT '0' COMMENT '申请时间',
  `updated_at` int(11) unsigned NOT NULL DEFAULT '0' COMMENT '修改时间',
  `deleted_at` datetime DEFAULT NULL COMMENT '删除时间',
  `batch_id` varchar(128) CHARACTER SET utf8mb4 COLLATE utf8mb4_bin NOT NULL DEFAULT '' COMMENT '批次id',
  `node_id` varchar(128) CHARACTER SET utf8mb4 COLLATE utf8mb4_bin NOT NULL DEFAULT '' COMMENT '节点id',
  `task_cnt` int(11) unsigned NOT NULL DEFAULT '0' COMMENT '任务数量',
  `done_task_cnt` int(11) unsigned NOT NULL DEFAULT '0' COMMENT '已完成任务数量',
  `status` int(11) unsigned NOT NULL DEFAULT '0' COMMENT '状态',
  PRIMARY KEY (`id`),
  UNIQUE KEY `idx_batch_id_node_id` (`batch_id`,`node_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci AUTO_INCREMENT=7208 COMMENT='批次节点信息表'
```

## 任务信息表
```
CREATE TABLE `client_test`.`cmd_info` (
  `id` int(11) unsigned NOT NULL AUTO_INCREMENT,
  `created_at` int(11) unsigned NOT NULL DEFAULT '0' COMMENT '申请时间',
  `updated_at` int(11) unsigned NOT NULL DEFAULT '0' COMMENT '修改时间',
  `deleted_at` datetime DEFAULT NULL COMMENT '删除时间',
  `batch_id` varchar(128) CHARACTER SET utf8mb4 COLLATE utf8mb4_bin NOT NULL DEFAULT '' COMMENT '批次id',
  `node_id` varchar(128) CHARACTER SET utf8mb4 COLLATE utf8mb4_bin NOT NULL DEFAULT '' COMMENT '节点id',
  `sql` varchar(512) CHARACTER SET utf8mb4 COLLATE utf8mb4_bin NOT NULL DEFAULT '' COMMENT 'sql语句',
  PRIMARY KEY (`id`),
  KEY `idx_batch_id_node_id` (`batch_id`,`node_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci AUTO_INCREMENT=0 COMMENT='任务信息表';
```

## 批次任务表
```
CREATE TABLE `client_test`.`cmd_order` (
  `id` int(11) unsigned NOT NULL AUTO_INCREMENT,
  `created_at` int(11) unsigned NOT NULL DEFAULT '0' COMMENT '申请时间',
  `updated_at` int(11) unsigned NOT NULL DEFAULT '0' COMMENT '修改时间',
  `deleted_at` datetime DEFAULT NULL COMMENT '删除时间',
  `batch_id` varchar(128) CHARACTER SET utf8mb4 COLLATE utf8mb4_bin NOT NULL DEFAULT '' COMMENT '批次id',
  `node_id` varchar(128) CHARACTER SET utf8mb4 COLLATE utf8mb4_bin NOT NULL DEFAULT '' COMMENT '节点id',
  `cmd_id` int(11) unsigned NOT NULL DEFAULT '0' COMMENT '任务id',
  `is_done` int(11) unsigned NOT NULL DEFAULT '0' COMMENT '是否完成 0 未完成 1 已完成',
  PRIMARY KEY (`id`),
  KEY `idx_batch_id_node_id` (`batch_id`,`node_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci AUTO_INCREMENT=7270 COMMENT='任务顺序表';
```

## 测试表
```
CREATE TABLE `test` (
  `id` int(11) unsigned NOT NULL AUTO_INCREMENT,
  `test_id` int(11) unsigned NOT NULL DEFAULT '0' COMMENT 'test_id',
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci AUTO_INCREMENT=1001 COMMENT='测试表';
```
