-- ============================================================
-- 修复演示数据中文乱码
--
-- 背景：早期用 SQL 导入演示数据时，连接字符集不是 utf8mb4，
--       中文被写成了字面的 '?' 字符（0x3F），原始内容已丢失。
--
-- 用法（在 CentOS7 虚拟机上执行）：
--   mysql -uroot -p skill_credential < fix_demo_data.sql
--
-- 说明：本文件所有中文均使用 _utf8mb4 X'..' 十六进制字面量，
--       不依赖终端字符集，可安全传输与执行。
--
-- 注意：以下中文为按演示场景补齐的取值，如需改为你原来的姓名/单位，
--       请修改本文件后重新执行（或直接在库中 UPDATE）。
-- ============================================================

SET NAMES utf8mb4;
USE skill_credential;

START TRANSACTION;

-- ------------------------------------------------------------
-- 1. 用户表：真实姓名 / 所属单位
-- ------------------------------------------------------------
UPDATE users SET real_name = _utf8mb4 X'E69D8EE5BBBAE59BBD', org_name = _utf8mb4 X'E69F90E8818CE4B89AE68A80E69CAFE5ADA6E999A2' WHERE username = _utf8mb4 X'61646D696E3031';
UPDATE users SET real_name = _utf8mb4 X'E5BCA0E4BC9F' WHERE username = _utf8mb4 X'64656D6F5F73747564656E74';
UPDATE users SET real_name = _utf8mb4 X'E78E8BE5BCBA', org_name = _utf8mb4 X'E58D8EE4B8BAE68A80E69CAFE69C89E99990E585ACE58FB8' WHERE username = _utf8mb4 X'636F6D70616E793031';
UPDATE users SET real_name = _utf8mb4 X'E99988E99D99', org_name = _utf8mb4 X'E5B9BFE5B79EE4BABAE6898DE99B86E59BA2' WHERE username = _utf8mb4 X'656D706C6F7965723031';

-- ------------------------------------------------------------
-- 2. 成果表：成果名称 / 颁发单位 / 描述
-- ------------------------------------------------------------
UPDATE achievements
   SET title       = _utf8mb4 X'E585A8E59BBDE8818CE4B89AE999A2E6A0A1E68A80E883BDE5A4A7E8B59BE4B880E7AD89E5A596',
       issuer      = _utf8mb4 X'E69599E882B2E983A8',
       description = _utf8mb4 X'E585A8E59BBDE8818CE4B89AE999A2E6A0A1E68A80E883BDE5A4A7E8B59BE4B880E7AD89E5A596E8AF81E4B9A6EFBC8CE7BB8FE999A2E6A0A1E6A0B8E9AA8CE5908EE4B88AE993BEE5AD98E8AF81E38082'
 WHERE id = 1;

-- ------------------------------------------------------------
-- 3. 存证表：学生姓名与成果名称是上链时的快照字段
--    链上交易哈希、区块高度不受影响，仅修正展示用快照。
--    注意：快照与链上原始元数据可能不一致，仅用于演示展示。
-- ------------------------------------------------------------
UPDATE credentials
   SET student_name = _utf8mb4 X'E5BCA0E4BC9F',
       title        = _utf8mb4 X'E585A8E59BBDE8818CE4B89AE999A2E6A0A1E68A80E883BDE5A4A7E8B59BE4B880E7AD89E5A596'
 WHERE credential_no = _utf8mb4 X'43432D32303236303932373132303331362D413344304233';

-- ------------------------------------------------------------
-- 4. 操作日志：详情中的成果名称
--    本表 log_hash = SHA256(username|action|detail|timestamp)。
--    改写 detail 后必须同步重算 log_hash，并把 onchain 置 0、清空 batch_hash，
--    让后台 worker 重新批量上链，否则链上锚点与内容不再对应。
-- ------------------------------------------------------------
UPDATE sys_logs
   SET detail = REPLACE(detail, _utf8mb4 X'7469746C653D3F3F3F3F3F3F3F3F3F3F3F3F3F', _utf8mb4 X'7469746C653DE585A8E59BBDE8818CE4B89AE999A2E6A0A1E68A80E883BDE5A4A7E8B59BE4B880E7AD89E5A596'),
       log_hash = SHA2(CONCAT(username,'|',action,'|',
                    REPLACE(detail, _utf8mb4 X'7469746C653D3F3F3F3F3F3F3F3F3F3F3F3F3F', _utf8mb4 X'7469746C653DE585A8E59BBDE8818CE4B89AE999A2E6A0A1E68A80E883BDE5A4A7E8B59BE4B880E7AD89E5A596'),
                    '|', UNIX_TIMESTAMP(created_at)), 256),
       onchain = 0,
       batch_hash = ''
 WHERE detail LIKE _utf8mb4 X'257469746C653D3F3F3F3F3F3F3F3F3F3F3F3F3F25';

COMMIT;

-- ------------------------------------------------------------
-- 5. 结果校验
--    注意：早期乱码写入的其他日志（REGISTER/LOGIN 等）detail 中仍可能含 '?'，
--    这些行未在本脚本修复范围（原始中文已丢失）。
--    bad_log_rows 期望值：仅 CREATE_ACHIEVEMENT 相关行被修复；
--    其余历史日志的 '?' 属已知残留，不影响核验主流程。
-- ------------------------------------------------------------
SELECT id, username, real_name, org_name FROM users ORDER BY id;
SELECT id, student_id, title, issuer, status FROM achievements;
SELECT credential_no, student_name, title, chain_block_no FROM credentials;
SELECT COUNT(*) AS bad_user_rows    FROM users        WHERE real_name LIKE '%\?%' OR org_name LIKE '%\?%';
SELECT COUNT(*) AS bad_ach_rows     FROM achievements WHERE title LIKE '%\?%' OR issuer LIKE '%\?%';
SELECT COUNT(*) AS bad_cred_rows    FROM credentials  WHERE student_name LIKE '%\?%' OR title LIKE '%\?%';
SELECT COUNT(*) AS bad_log_rows     FROM sys_logs     WHERE detail LIKE '%\?%';
SELECT COUNT(*) AS pending_reanchor FROM sys_logs     WHERE onchain = 0;
