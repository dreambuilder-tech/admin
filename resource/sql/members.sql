-- 用户表查询示例

-- 分页查询用户列表（无筛选条件）
SELECT
    `id`,
    `account`,
    `area_code`,
    `phone`,
    `nickname`,
    `avatar`,
    `email`,
    `lang`,
    `status`,
    `last_login_ip`,
    `login_times`,
    `last_login_at`,
    `register_ip`,
    `created_at`,
    `updated_at`
FROM `member`.`members`
ORDER BY `id` DESC
LIMIT 20 OFFSET 0;

-- 分页查询用户列表（带筛选条件）
SELECT
    `id`,
    `account`,
    `area_code`,
    `phone`,
    `nickname`,
    `avatar`,
    `email`,
    `lang`,
    `status`,
    `last_login_ip`,
    `login_times`,
    `last_login_at`,
    `register_ip`,
    `created_at`,
    `updated_at`
FROM `member`.`members`
WHERE `account` LIKE '%user%'
  AND `phone` LIKE '%138%'
ORDER BY `id` DESC
LIMIT 20 OFFSET 0;

-- 查询总记录数（无筛选条件）
SELECT COUNT(*) FROM `member`.`members`;

-- 查询总记录数（带筛选条件）
SELECT COUNT(*) FROM `member`.`members`
WHERE `account` LIKE '%user%'
  AND `phone` LIKE '%138%';

