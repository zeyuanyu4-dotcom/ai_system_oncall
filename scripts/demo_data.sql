-- 智能OnCall问题处理与研发协同平台 - 演示数据
-- 注意：用户密码通过API注册生成，这里只创建项目和问题数据

USE oncall_platform;

-- 清理旧的演示数据（保留admin用户）
DELETE FROM issue_operation_logs WHERE issue_id IN (SELECT id FROM issues WHERE project_id IN (SELECT id FROM projects WHERE code IN ('user-center', 'order-system', 'data-platform')));
DELETE FROM issue_status_logs WHERE issue_id IN (SELECT id FROM issues WHERE project_id IN (SELECT id FROM projects WHERE code IN ('user-center', 'order-system', 'data-platform')));
DELETE FROM issue_comments WHERE issue_id IN (SELECT id FROM issues WHERE project_id IN (SELECT id FROM projects WHERE code IN ('user-center', 'order-system', 'data-platform')));
DELETE FROM issues WHERE project_id IN (SELECT id FROM projects WHERE code IN ('user-center', 'order-system', 'data-platform'));
DELETE FROM service_apis WHERE service_id IN (SELECT id FROM services WHERE project_id IN (SELECT id FROM projects WHERE code IN ('user-center', 'order-system', 'data-platform')));
DELETE FROM service_dependencies WHERE service_id IN (SELECT id FROM services WHERE project_id IN (SELECT id FROM projects WHERE code IN ('user-center', 'order-system', 'data-platform')));
DELETE FROM services WHERE project_id IN (SELECT id FROM projects WHERE code IN ('user-center', 'order-system', 'data-platform'));
DELETE FROM project_members WHERE project_id IN (SELECT id FROM projects WHERE code IN ('user-center', 'order-system', 'data-platform'));
DELETE FROM projects WHERE code IN ('user-center', 'order-system', 'data-platform');
DELETE FROM users WHERE email IN ('zhangsan@oncall.com', 'lisi@oncall.com', 'wangwu@oncall.com', 'zhaoliu@oncall.com');

-- 创建用户（密码都是 123456，bcrypt哈希）
-- 注意：实际密码哈希需要通过bcrypt生成，这里使用预计算的值
INSERT INTO `users` (`username`, `email`, `password_hash`, `role`, `status`) VALUES
('zhangsan', 'zhangsan@oncall.com', '$2a$10$N9qo8uLOickgx2ZMRZoMyeIjZRGdjGj/n3/ItB/XBG/eCknfIrqS6', 'project_admin', 1),
('lisi', 'lisi@oncall.com', '$2a$10$N9qo8uLOickgx2ZMRZoMyeIjZRGdjGj/n3/ItB/XBG/eCknfIrqS6', 'developer', 1),
('wangwu', 'wangwu@oncall.com', '$2a$10$N9qo8uLOickgx2ZMRZoMyeIjZRGdjGj/n3/ItB/XBG/eCknfIrqS6', 'tester', 1),
('zhaoliu', 'zhaoliu@oncall.com', '$2a$10$N9qo8uLOickgx2ZMRZoMyeIjZRGdjGj/n3/ItB/XBG/eCknfIrqS6', 'normal_user', 1);

-- 创建项目
INSERT INTO `projects` (`name`, `code`, `description`, `owner_id`, `status`, `created_by`) VALUES
('用户中心', 'user-center', '用户认证、用户管理、权限控制等核心功能', 1, 1, 1),
('订单系统', 'order-system', '订单处理、支付、退款等交易相关功能', 1, 1, 1),
('数据平台', 'data-platform', '数据分析、报表生成、数据可视化', 1, 1, 1);

-- 获取用户ID和项目ID的变量
SET @zhangsan_id = (SELECT id FROM users WHERE email = 'zhangsan@oncall.com');
SET @lisi_id = (SELECT id FROM users WHERE email = 'lisi@oncall.com');
SET @wangwu_id = (SELECT id FROM users WHERE email = 'wangwu@oncall.com');
SET @zhaoliu_id = (SELECT id FROM users WHERE email = 'zhaoliu@oncall.com');
SET @user_center_id = (SELECT id FROM projects WHERE code = 'user-center');
SET @order_system_id = (SELECT id FROM projects WHERE code = 'order-system');
SET @data_platform_id = (SELECT id FROM projects WHERE code = 'data-platform');

-- 创建项目成员
INSERT INTO `project_members` (`project_id`, `user_id`, `project_role`) VALUES
-- 用户中心项目成员
(@user_center_id, @zhangsan_id, 'admin'),
(@user_center_id, @lisi_id, 'developer'),
(@user_center_id, @wangwu_id, 'tester'),
(@user_center_id, @zhaoliu_id, 'member'),
-- 订单系统项目成员
(@order_system_id, @zhangsan_id, 'admin'),
(@order_system_id, @lisi_id, 'developer'),
-- 数据平台项目成员
(@data_platform_id, @zhangsan_id, 'admin'),
(@data_platform_id, @lisi_id, 'developer'),
(@data_platform_id, @wangwu_id, 'tester');

-- 创建服务
INSERT INTO `services` (`project_id`, `name`, `code`, `description`, `service_type`, `owner_id`, `language`, `status`) VALUES
-- 用户中心服务
(@user_center_id, '认证服务', 'auth-service', '用户登录、注册、Token管理', 'backend', @lisi_id, 'Go', 1),
(@user_center_id, '用户服务', 'user-service', '用户信息管理、权限管理', 'backend', @lisi_id, 'Go', 1),
(@user_center_id, '网关服务', 'gateway-service', 'API网关、路由转发、限流', 'gateway', @lisi_id, 'Go', 1),
-- 订单系统服务
(@order_system_id, '订单服务', 'order-service', '订单创建、查询、取消', 'backend', @lisi_id, 'Go', 1),
(@order_system_id, '支付服务', 'payment-service', '支付处理、退款', 'backend', @lisi_id, 'Go', 1),
-- 数据平台服务
(@data_platform_id, '报表服务', 'report-service', '数据报表生成', 'backend', @lisi_id, 'Python', 1),
(@data_platform_id, '同步任务', 'sync-job', '数据同步定时任务', 'job', @lisi_id, 'Go', 1);

-- 获取服务ID
SET @auth_service_id = (SELECT id FROM services WHERE code = 'auth-service');
SET @user_service_id = (SELECT id FROM services WHERE code = 'user-service');
SET @gateway_service_id = (SELECT id FROM services WHERE code = 'gateway-service');
SET @order_service_id = (SELECT id FROM services WHERE code = 'order-service');
SET @payment_service_id = (SELECT id FROM services WHERE code = 'payment-service');
SET @report_service_id = (SELECT id FROM services WHERE code = 'report-service');

-- 创建服务接口
INSERT INTO `service_apis` (`service_id`, `method`, `path`, `name`, `description`, `status`) VALUES
-- auth-service 接口
(@auth_service_id, 'POST', '/api/auth/login', '用户登录', '用户登录接口', 1),
(@auth_service_id, 'POST', '/api/auth/register', '用户注册', '用户注册接口', 1),
(@auth_service_id, 'POST', '/api/auth/logout', '用户登出', '用户登出接口', 1),
(@auth_service_id, 'GET', '/api/auth/me', '获取当前用户', '获取当前登录用户信息', 1),
-- user-service 接口
(@user_service_id, 'GET', '/api/users', '用户列表', '获取用户列表', 1),
(@user_service_id, 'GET', '/api/users/:id', '用户详情', '获取用户详情', 1),
(@user_service_id, 'PUT', '/api/users/:id', '更新用户', '更新用户信息', 1),
-- gateway-service 接口
(@gateway_service_id, 'GET', '/health', '健康检查', '服务健康检查', 1),
-- order-service 接口
(@order_service_id, 'POST', '/api/orders', '创建订单', '创建新订单', 1),
(@order_service_id, 'GET', '/api/orders/:id', '订单详情', '获取订单详情', 1),
(@order_service_id, 'GET', '/api/orders', '订单列表', '获取订单列表', 1);

-- 创建服务依赖关系
INSERT INTO `service_dependencies` (`service_id`, `depends_on_service_id`, `dependency_type`, `description`) VALUES
(@gateway_service_id, @auth_service_id, 'http', '网关转发认证请求'),
(@gateway_service_id, @user_service_id, 'http', '网关转发用户请求'),
(@gateway_service_id, @order_service_id, 'http', '网关转发订单请求'),
(@auth_service_id, @user_service_id, 'grpc', '认证服务调用用户服务'),
(@order_service_id, @payment_service_id, 'http', '订单服务调用支付服务'),
(@report_service_id, @order_service_id, 'http', '报表服务查询订单数据');

-- 创建问题单（完整链路演示）
INSERT INTO `issues` (`issue_no`, `title`, `description`, `project_id`, `service_id`, `issue_type`, `priority`, `environment`, `status`, `impact_scope`, `error_message`, `log_excerpt`, `creator_id`, `assignee_id`, `root_cause`, `solution`, `resolved_at`, `closed_at`, `created_at`) VALUES
('ISSUE-20260602-DEMO1', '测试环境登录接口返回500错误', '测试环境用户反馈登录接口返回500错误，无法正常登录系统。问题发生在今天上午10点左右，影响所有测试环境用户。', @user_center_id, @auth_service_id, 'api_error', 'P0', 'test', 'closed', '测试环境所有用户', 'Internal Server Error: Redis connection refused', '2026-06-02 10:15:23 ERROR auth-service: Redis connection failed: dial tcp 127.0.0.1:6379: connection refused', @wangwu_id, @lisi_id, '测试环境 auth-service 的 Redis 配置地址错误，配置文件中写的是生产环境的 Redis 地址，但测试环境没有该 Redis 实例。', '1. 修正测试环境 auth-service 的 Redis 配置地址\n2. 重新部署 auth-service\n3. 验证登录功能恢复正常', '2026-06-02 14:30:00', '2026-06-02 15:00:00', '2026-06-02 10:30:00');

-- 获取问题ID
SET @issue1_id = (SELECT id FROM issues WHERE issue_no = 'ISSUE-20260602-DEMO1');

-- 创建状态流转日志
INSERT INTO `issue_status_logs` (`issue_id`, `from_status`, `to_status`, `operator_id`, `reason`, `created_at`) VALUES
(@issue1_id, '', 'pending_analysis', @wangwu_id, '创建问题', '2026-06-02 10:30:00'),
(@issue1_id, 'pending_analysis', 'processing', @zhangsan_id, '分配给 auth-service 负责人处理', '2026-06-02 11:00:00'),
(@issue1_id, 'processing', 'pending_confirmation', @lisi_id, '问题已修复，等待用户确认', '2026-06-02 14:00:00'),
(@issue1_id, 'pending_confirmation', 'resolved', @wangwu_id, '用户确认问题已解决', '2026-06-02 14:30:00'),
(@issue1_id, 'resolved', 'closed', @zhangsan_id, '处理记录完整，关闭问题', '2026-06-02 15:00:00');

-- 创建评论和处理记录
INSERT INTO `issue_comments` (`issue_id`, `user_id`, `comment_type`, `content`, `visibility`, `created_at`) VALUES
(@issue1_id, @wangwu_id, 'comment', '补充信息：问题只发生在测试环境，生产环境登录正常。', 'public', '2026-06-02 10:35:00'),
(@issue1_id, @zhangsan_id, 'system', '问题已分配给李四处理', 'public', '2026-06-02 11:00:00'),
(@issue1_id, @lisi_id, 'process_record', '已查看登录接口请求日志，发现问题发生在调用 Redis 时。初步怀疑 Redis 连接配置问题。', 'public', '2026-06-02 11:30:00'),
(@issue1_id, @lisi_id, 'process_record', '已定位根因：测试环境 auth-service 的 Redis 配置地址错误。配置文件中写的是 prod-redis.internal:6379，但测试环境应该使用 test-redis.internal:6379。', 'public', '2026-06-02 12:00:00'),
(@issue1_id, @lisi_id, 'solution', '解决方案：\n1. 修正测试环境 auth-service 的 Redis 配置地址\n2. 重新部署 auth-service\n3. 验证登录功能恢复正常', 'public', '2026-06-02 13:00:00'),
(@issue1_id, @wangwu_id, 'comment', '已验证，登录功能恢复正常，问题已解决。', 'public', '2026-06-02 14:30:00');

-- 创建操作日志
INSERT INTO `issue_operation_logs` (`issue_id`, `operator_id`, `operation_type`, `operation_content`, `created_at`) VALUES
(@issue1_id, @wangwu_id, 'create_issue', '创建问题: 测试环境登录接口返回500错误', '2026-06-02 10:30:00'),
(@issue1_id, @zhangsan_id, 'assign_issue', '分配给李四处理', '2026-06-02 11:00:00'),
(@issue1_id, @zhangsan_id, 'change_status', '状态变更: pending_analysis -> processing', '2026-06-02 11:00:00'),
(@issue1_id, @lisi_id, 'add_comment', '添加处理记录', '2026-06-02 11:30:00'),
(@issue1_id, @lisi_id, 'add_comment', '添加处理记录', '2026-06-02 12:00:00'),
(@issue1_id, @lisi_id, 'update_issue', '填写根因和解决方案', '2026-06-02 13:00:00'),
(@issue1_id, @lisi_id, 'change_status', '状态变更: processing -> pending_confirmation', '2026-06-02 14:00:00'),
(@issue1_id, @wangwu_id, 'change_status', '状态变更: pending_confirmation -> resolved', '2026-06-02 14:30:00'),
(@issue1_id, @zhangsan_id, 'close_issue', '关闭问题', '2026-06-02 15:00:00');

-- 创建第二个问题（待处理状态，用于演示）
INSERT INTO `issues` (`issue_no`, `title`, `description`, `project_id`, `service_id`, `issue_type`, `priority`, `environment`, `status`, `impact_scope`, `error_message`, `creator_id`, `created_at`) VALUES
('ISSUE-20260602-DEMO2', '订单列表页面加载缓慢', '订单列表页面加载时间超过5秒，用户体验较差。数据量大约10万条订单。', @order_system_id, @order_service_id, 'performance_issue', 'P2', 'prod', 'pending_analysis', '所有用户', '', @zhaoliu_id, '2026-06-02 16:00:00');

-- 创建第三个问题（处理中状态）
INSERT INTO `issues` (`issue_no`, `title`, `description`, `project_id`, `service_id`, `issue_type`, `priority`, `environment`, `status`, `impact_scope`, `error_message`, `creator_id`, `assignee_id`, `created_at`) VALUES
('ISSUE-20260602-DEMO3', '报表数据与实际不符', '日报表中的订单金额统计与实际订单总额不一致，差异约5%。', @data_platform_id, @report_service_id, 'data_issue', 'P1', 'prod', 'processing', '运营团队', '', @wangwu_id, @lisi_id, '2026-06-02 14:00:00');

SELECT '演示数据创建完成' AS message;
