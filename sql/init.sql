-- ============================================================
-- 基于联盟链的高校毕业生技能成果可信存证系统
-- MySQL 初始化脚本（CentOS7 / MySQL 5.7+）
--
-- !!! 警告 !!!
-- 本脚本使用 DROP TABLE IF EXISTS，会清空已有数据（含只增不改的 credentials 表）。
-- 仅用于首次建库或明确要重置的演示环境。
-- 已录入数据的环境请勿重复执行；如需修复演示数据请用 fix_demo_data.sql。
-- ============================================================

CREATE DATABASE IF NOT EXISTS skill_credential
  DEFAULT CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;

USE skill_credential;

-- ------------------------------------------------------------
-- 1. 用户表：四类角色，注册时自动生成链账户
-- ------------------------------------------------------------
DROP TABLE IF EXISTS users;
CREATE TABLE users (
  id              BIGINT       NOT NULL AUTO_INCREMENT COMMENT '主键',
  username        VARCHAR(64)  NOT NULL COMMENT '登录名',
  password        VARCHAR(255) NOT NULL COMMENT 'bcrypt 密码哈希',
  real_name       VARCHAR(64)  NOT NULL DEFAULT '' COMMENT '真实姓名',
  student_no      VARCHAR(32)  NOT NULL DEFAULT '' COMMENT '学号(学生角色必填)',
  role            ENUM('student','school_admin','company','employer') NOT NULL COMMENT '角色',
  org_name        VARCHAR(128) NOT NULL DEFAULT '' COMMENT '所属学校/企业名称',
  eth_address     VARCHAR(42)  NOT NULL DEFAULT '' COMMENT '自动生成的链账户地址',
  eth_private_key VARCHAR(256) NOT NULL DEFAULT '' COMMENT '链账户私钥(已弃用：后端不再写入，仅保留列兼容旧数据)',
  status          TINYINT      NOT NULL DEFAULT 1 COMMENT '1正常 0禁用',
  created_at      DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '注册时间',
  PRIMARY KEY (id),
  UNIQUE KEY uk_username (username),
  KEY idx_role (role),
  KEY idx_student_no (student_no)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='用户表';

-- ------------------------------------------------------------
-- 2. 成果表：证书/竞赛/实训评价/项目成果
-- ------------------------------------------------------------
DROP TABLE IF EXISTS achievements;
CREATE TABLE achievements (
  id          BIGINT       NOT NULL AUTO_INCREMENT COMMENT '主键',
  student_id  BIGINT       NOT NULL COMMENT '学生用户ID',
  title       VARCHAR(128) NOT NULL COMMENT '成果名称',
  ach_type    ENUM('certificate','competition','internship','project') NOT NULL COMMENT '成果类型',
  description TEXT         COMMENT '成果描述',
  issuer      VARCHAR(128) NOT NULL DEFAULT '' COMMENT '颁发/评价单位',
  file_hash   VARCHAR(64)  NOT NULL DEFAULT '' COMMENT '附件文件 SHA-256',
  file_path   VARCHAR(255) NOT NULL DEFAULT '' COMMENT '附件存储路径或IPFS CID',
  status      ENUM('draft','submitted','onchain') NOT NULL DEFAULT 'draft' COMMENT '状态',
  created_by  BIGINT       NOT NULL COMMENT '录入人ID',
  created_at  DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '录入时间',
  PRIMARY KEY (id),
  KEY idx_student (student_id),
  KEY idx_type (ach_type),
  KEY idx_status (status)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='技能成果表';

-- ------------------------------------------------------------
-- 3. 存证凭证表：只可新增，不可删除修改（对应链上不可篡改）
-- ------------------------------------------------------------
DROP TABLE IF EXISTS credentials;
CREATE TABLE credentials (
  id            BIGINT       NOT NULL AUTO_INCREMENT COMMENT '主键',
  credential_no VARCHAR(64)  NOT NULL COMMENT '唯一存证编号(链上返回)',
  achievement_id BIGINT      NOT NULL COMMENT '成果ID',
  student_id    BIGINT       NOT NULL COMMENT '学生ID',
  student_name  VARCHAR(64)  NOT NULL DEFAULT '' COMMENT '学生姓名快照',
  student_no    VARCHAR(32)  NOT NULL DEFAULT '' COMMENT '学号快照',
  ach_type      VARCHAR(32)  NOT NULL DEFAULT '' COMMENT '成果类型快照',
  title         VARCHAR(128) NOT NULL DEFAULT '' COMMENT '成果名称快照',
  chain_tx_hash VARCHAR(66)  NOT NULL DEFAULT '' COMMENT '上链交易哈希',
  chain_block_no BIGINT      NOT NULL DEFAULT 0 COMMENT '区块高度',
  file_hash     VARCHAR(64)  NOT NULL COMMENT '上链时的文件哈希',
  issuer_addr   VARCHAR(42)  NOT NULL DEFAULT '' COMMENT '上链节点地址',
  onchain_time  DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '上链时间',
  verify_count  INT          NOT NULL DEFAULT 0 COMMENT '被核验次数',
  status        TINYINT      NOT NULL DEFAULT 1 COMMENT '1有效',
  PRIMARY KEY (id),
  UNIQUE KEY uk_credential_no (credential_no),
  KEY idx_student (student_id),
  KEY idx_student_no (student_no),
  KEY idx_ach_type (ach_type),
  KEY idx_onchain_time (onchain_time)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='存证凭证表(只增不改删)';

-- ------------------------------------------------------------
-- 4. 核验记录表：外部企业每次核验留痕
-- ------------------------------------------------------------
DROP TABLE IF EXISTS verify_logs;
CREATE TABLE verify_logs (
  id            BIGINT      NOT NULL AUTO_INCREMENT COMMENT '主键',
  credential_no VARCHAR(64) NOT NULL COMMENT '被核验凭证编号',
  employer_id   BIGINT      NOT NULL DEFAULT 0 COMMENT '核验企业用户ID(0=游客核验)',
  employer_name VARCHAR(128) NOT NULL DEFAULT '' COMMENT '核验企业名称',
  input_hash    VARCHAR(64) NOT NULL DEFAULT '' COMMENT '传入的文件哈希',
  chain_result  VARCHAR(16) NOT NULL DEFAULT '' COMMENT '链上核验结果 pass/fail',
  created_at    DATETIME    NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '核验时间',
  PRIMARY KEY (id),
  KEY idx_credential_no (credential_no),
  KEY idx_created_at (created_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='核验记录表';

-- ------------------------------------------------------------
-- 5. 系统日志表：全部操作留痕，日志哈希批量上链
-- ------------------------------------------------------------
DROP TABLE IF EXISTS sys_logs;
CREATE TABLE sys_logs (
  id          BIGINT       NOT NULL AUTO_INCREMENT COMMENT '主键',
  user_id     BIGINT       NOT NULL DEFAULT 0 COMMENT '操作人ID',
  username    VARCHAR(64)  NOT NULL DEFAULT '' COMMENT '操作人用户名',
  action      VARCHAR(64)  NOT NULL COMMENT '操作类型',
  detail      TEXT         COMMENT '操作详情',
  log_hash    VARCHAR(64)  NOT NULL DEFAULT '' COMMENT '本条日志SHA-256',
  batch_hash  VARCHAR(80)  NOT NULL DEFAULT '' COMMENT '批次摘要/链上交易哈希(上链后回填)',
  onchain     TINYINT      NOT NULL DEFAULT 0 COMMENT '0未上链 1已上链',
  created_at  DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '操作时间',
  PRIMARY KEY (id),
  KEY idx_user (user_id),
  KEY idx_action (action),
  KEY idx_onchain (onchain),
  KEY idx_created_at (created_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='系统操作日志表';

-- ------------------------------------------------------------
-- 初始账号（演示用固定账号）
-- 口令由后端注册接口注册时生成 bcrypt 哈希，不在本文件中明文记录
-- ------------------------------------------------------------
-- 说明：以下为演示用固定账号，password 字段请用后端 /api/auth/register 注册生成
-- 这里仅插入角色占位，联调时通过接口注册真实账号
