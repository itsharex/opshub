CREATE TABLE IF NOT EXISTS `k8s_user_kube_configs` (
  `id` BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY COMMENT '主键ID',
  `cluster_id` BIGINT UNSIGNED NOT NULL COMMENT '集群ID',
  `user_id` BIGINT UNSIGNED NOT NULL COMMENT '平台用户ID',
  `service_account` VARCHAR(255) NOT NULL COMMENT 'K8s ServiceAccount名称',
  `namespace` VARCHAR(255) DEFAULT 'default' COMMENT '命名空间',
  `is_active` TINYINT(1) DEFAULT 1 COMMENT '是否激活（1=激活，0=已吊销）',
  `created_by` BIGINT UNSIGNED NOT NULL COMMENT '创建人ID',
  `created_at` DATETIME DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `revoked_at` DATETIME DEFAULT NULL COMMENT '吊销时间',
  UNIQUE KEY `uk_cluster_user_sa` (`cluster_id`, `user_id`, `service_account`),
  KEY `idx_cluster_id` (`cluster_id`),
  KEY `idx_user_id` (`user_id`),
  KEY `idx_service_account` (`service_account`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='用户K8s凭据表';
