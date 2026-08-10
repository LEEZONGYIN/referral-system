-- Referral System MySQL schema used by this project.
-- MySQL 8+

CREATE TABLE users (
    id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '用户ID',
    name VARCHAR(64) NOT NULL COMMENT '姓名/昵称',
    email VARCHAR(128) DEFAULT NULL COMMENT '邮箱',
    phone VARCHAR(32) DEFAULT NULL COMMENT '手机号',
    referral_code VARCHAR(32) NOT NULL COMMENT '用户邀请码',
    status TINYINT NOT NULL DEFAULT 1 COMMENT '状态：1=正常，0=禁用',
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
    PRIMARY KEY (id),
    UNIQUE KEY uk_users_email (email),
    UNIQUE KEY uk_users_phone (phone),
    UNIQUE KEY uk_users_referral_code (referral_code)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='用户表';

CREATE TABLE referral_rules (
    id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '规则ID',
    rule_code VARCHAR(64) NOT NULL COMMENT '规则编码',
    reward_amount BIGINT NOT NULL COMMENT '奖励积分',
    trigger_event VARCHAR(32) NOT NULL COMMENT '触发事件，如 REGISTERED / FIRST_ORDER',
    status TINYINT NOT NULL DEFAULT 1 COMMENT '状态：1=启用，0=停用',
    effective_from DATETIME DEFAULT NULL COMMENT '生效开始时间',
    effective_to DATETIME DEFAULT NULL COMMENT '生效结束时间',
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
    PRIMARY KEY (id),
    UNIQUE KEY uk_referral_rules_rule_code (rule_code),
    KEY idx_referral_rules_status_time (status, effective_from, effective_to)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='邀请奖励规则表';

CREATE TABLE referral_relations (
    id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '关系ID',
    inviter_user_id BIGINT UNSIGNED NOT NULL COMMENT '邀请人ID',
    invitee_user_id BIGINT UNSIGNED NOT NULL COMMENT '被邀请人ID',
    referral_code VARCHAR(32) NOT NULL COMMENT '使用的邀请码',
    rule_id BIGINT UNSIGNED DEFAULT NULL COMMENT '命中的规则ID',
    status TINYINT NOT NULL DEFAULT 0 COMMENT '状态：0=PENDING,1=CONFIRMED,2=REWARDED,3=REVOKED',
    qualified_at DATETIME DEFAULT NULL COMMENT '满足奖励条件时间',
    rewarded_at DATETIME DEFAULT NULL COMMENT '发奖时间',
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
    PRIMARY KEY (id),
    UNIQUE KEY uk_referral_relations_invitee (invitee_user_id),
    UNIQUE KEY uk_referral_relations_inviter_invitee (inviter_user_id, invitee_user_id),
    KEY idx_referral_relations_inviter_status_time (inviter_user_id, status, created_at),
    KEY idx_referral_relations_referral_code (referral_code),
    KEY idx_referral_relations_status_time (status, created_at),
    CONSTRAINT fk_referral_relations_inviter FOREIGN KEY (inviter_user_id) REFERENCES users(id),
    CONSTRAINT fk_referral_relations_invitee FOREIGN KEY (invitee_user_id) REFERENCES users(id),
    CONSTRAINT fk_referral_relations_rule FOREIGN KEY (rule_id) REFERENCES referral_rules(id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='邀请关系表';

CREATE TABLE referral_events (
    id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '事件ID',
    relation_id BIGINT UNSIGNED DEFAULT NULL COMMENT '邀请关系ID',
    inviter_user_id BIGINT UNSIGNED DEFAULT NULL COMMENT '邀请人ID',
    invitee_user_id BIGINT UNSIGNED DEFAULT NULL COMMENT '被邀请人ID',
    event_type VARCHAR(32) NOT NULL COMMENT '事件类型：INVITE_CREATED/LINK_CLICKED/REGISTERED/QUALIFIED/REWARDED/REWARD_FAILED',
    idempotency_key VARCHAR(128) NOT NULL COMMENT '幂等键',
    payload JSON DEFAULT NULL COMMENT '事件负载',
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    PRIMARY KEY (id),
    UNIQUE KEY uk_referral_events_idempotency_key (idempotency_key),
    KEY idx_referral_events_relation_time (relation_id, created_at),
    KEY idx_referral_events_invitee_time (invitee_user_id, created_at),
    KEY idx_referral_events_type_time (event_type, created_at),
    CONSTRAINT fk_referral_events_relation FOREIGN KEY (relation_id) REFERENCES referral_relations(id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='邀请事件表';

CREATE TABLE credit_accounts (
    user_id BIGINT UNSIGNED NOT NULL COMMENT '用户ID',
    balance BIGINT NOT NULL DEFAULT 0 COMMENT '可用积分余额',
    frozen_balance BIGINT NOT NULL DEFAULT 0 COMMENT '冻结积分',
    version BIGINT NOT NULL DEFAULT 0 COMMENT '乐观锁版本号',
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
    PRIMARY KEY (user_id),
    CONSTRAINT fk_credit_accounts_user FOREIGN KEY (user_id) REFERENCES users(id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='积分账户表';

CREATE TABLE credit_ledger (
    id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '流水ID',
    user_id BIGINT UNSIGNED NOT NULL COMMENT '用户ID',
    biz_type VARCHAR(32) NOT NULL COMMENT '业务类型：REFERRAL_REWARD/REFERRAL_REVERSAL/ADJUSTMENT',
    biz_id VARCHAR(64) NOT NULL COMMENT '业务单号',
    direction TINYINT NOT NULL COMMENT '方向：1=入账，2=出账',
    amount BIGINT NOT NULL COMMENT '积分变动值',
    before_balance BIGINT NOT NULL COMMENT '变动前余额',
    after_balance BIGINT NOT NULL COMMENT '变动后余额',
    idempotency_key VARCHAR(128) NOT NULL COMMENT '幂等键',
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    PRIMARY KEY (id),
    UNIQUE KEY uk_credit_ledger_idempotency_key (idempotency_key),
    UNIQUE KEY uk_credit_ledger_biz (biz_type, biz_id),
    KEY idx_credit_ledger_user_time (user_id, created_at),
    KEY idx_credit_ledger_user_biz (user_id, biz_type, created_at),
    CONSTRAINT fk_credit_ledger_user FOREIGN KEY (user_id) REFERENCES users(id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='积分流水表';

CREATE TABLE referral_stats_daily (
    id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '统计ID',
    stat_date DATE NOT NULL COMMENT '统计日期',
    user_id BIGINT UNSIGNED NOT NULL COMMENT '用户ID',
    invite_count INT NOT NULL DEFAULT 0 COMMENT '邀请数',
    qualified_count INT NOT NULL DEFAULT 0 COMMENT '达标数',
    reward_amount_total BIGINT NOT NULL DEFAULT 0 COMMENT '累计奖励积分',
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
    PRIMARY KEY (id),
    UNIQUE KEY uk_referral_stats_daily_user_date (user_id, stat_date),
    KEY idx_referral_stats_daily_date (stat_date),
    KEY idx_referral_stats_daily_user (user_id),
    CONSTRAINT fk_referral_stats_daily_user FOREIGN KEY (user_id) REFERENCES users(id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='邀请日统计表';
