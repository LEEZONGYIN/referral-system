# Referral System

一个 Referral 系统参考实现，包含邀请归因、积分发放、账本流水、历史查询和基础 HTTP 接口。

## 目录

- [如何在本地运行项目](#如何在本地运行项目)
- [配置说明](#配置说明)
- [项目的整体设计思路](#项目的整体设计思路)
- [开发过程中作出的设计取舍](#开发过程中作出的设计取舍)
- [接口说明](#接口说明)
- [项目结构](#项目结构)
- [后续可继续增强的方向](#后续可继续增强的方向)
- [常见问题](#常见问题)

## 如何在本地运行项目

### 1. 准备 MySQL

项目默认使用 MySQL 8+。你可以先创建数据库，例如：

```sql
CREATE DATABASE referral_system DEFAULT CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;
```

然后执行建表脚本：

- `internal/model/sql/referral.sql`

### 2. 配置文件

项目通过 `config.yaml` 读取运行配置，启动前请先确认文件内容正确。

### 3. 启动服务

在项目根目录执行：

```bash
go run .
```

默认监听 `config.yaml` 中配置的地址，通常是 `:8080`。

### 4. 验证服务

```bash
curl http://localhost:8080/healthz
```

### 5. 调用核心接口

注册并绑定邀请关系：

```bash
curl -X POST http://localhost:8080/api/v1/referrals/register \
  -H 'Content-Type: application/json' \
  -d '{
    "name": "Bob",
    "email": "bob@example.com",
    "phone": "13800000000",
    "referral_code": "ABC123",
    "idempotency_key": "register-bob-001"
  }'
```

发放奖励：

```bash
curl -X POST http://localhost:8080/api/v1/referrals/reward \
  -H 'Content-Type: application/json' \
  -d '{
    "relation_id": 1001,
    "biz_id": "reg-2001",
    "amount": 100,
    "idempotency_key": "reward-1001"
  }'
```

## 跑通一次完整流程

如果你想在本地快速走通一次闭环，推荐按下面顺序操作。

### 1. 初始化数据库

```sql
CREATE DATABASE IF NOT EXISTS referral_system
DEFAULT CHARACTER SET utf8mb4
COLLATE utf8mb4_unicode_ci;
```

然后执行建表脚本：

- `internal/model/sql/referral.sql`

### 2. 插入最小初始化数据

先插入一个邀请人，系统会基于这个用户的邀请码完成后续归因：

```sql
INSERT INTO users (name, email, phone, referral_code, status)
VALUES ('Alice', 'alice@example.com', '13800000001', 'ABC123', 1);
```

可选再插入一条默认奖励规则，方便语义更完整：

```sql
INSERT INTO referral_rules (
    rule_code,
    reward_amount,
    trigger_event,
    status,
    effective_from,
    effective_to
) VALUES (
    'DEFAULT_REGISTER_REWARD',
    100,
    'REGISTERED',
    1,
    NOW(),
    NULL
);
```

### 3. 配置本地 MySQL DSN

如果你的本地 MySQL 是空密码，可以这样配置：

```yaml
server:
  addr: ":8080"
mysql:
  dsn: "root@tcp(127.0.0.1:3306)/referral_system?parseTime=true&charset=utf8mb4&loc=Local"
```

如果你后续切换成有密码账号，只需要把 `root` 后面补上 `:password` 即可：

```yaml
mysql:
  dsn: "root:your_password@tcp(127.0.0.1:3306)/referral_system?parseTime=true&charset=utf8mb4&loc=Local"
```

### 4. 启动服务

```bash
go run .
```

### 5. 发起被邀请人注册

```bash
curl -X POST http://localhost:8080/api/v1/referrals/register \
  -H 'Content-Type: application/json' \
  -d '{
    "name": "Bob",
    "email": "bob@example.com",
    "phone": "13800000002",
    "referral_code": "ABC123",
    "idempotency_key": "register-bob-001"
  }'
```

接口会返回类似结果：

```json
{
  "relation_id": 1,
  "invitee_user_id": 2
}
```

### 6. 给邀请人发奖励

```bash
curl -X POST http://localhost:8080/api/v1/referrals/reward \
  -H 'Content-Type: application/json' \
  -d '{
    "relation_id": 1,
    "biz_id": "reg-bob-001",
    "amount": 100,
    "idempotency_key": "reward-rel-1"
  }'
```

### 7. 查询结果验证

查询邀请历史：

```bash
curl "http://localhost:8080/api/v1/referrals/history?user_id=1&limit=20&offset=0"
```

查询 Dashboard：

```bash
curl "http://localhost:8080/api/v1/referrals/dashboard?user_id=1"
```

查询积分余额：

```bash
curl "http://localhost:8080/api/v1/credits/balance?user_id=1"
```

查询积分流水：

```bash
curl "http://localhost:8080/api/v1/credits/ledger?user_id=1&limit=20&offset=0"
```

### 8. 预期数据变化

完成后，数据库里会出现这些变化：

- `users`：新增被邀请人 `Bob`
- `referral_relations`：新增 `Alice -> Bob` 的邀请关系
- `referral_events`：新增注册和奖励事件
- `credit_accounts`：`Alice` 的余额增加 `100`
- `credit_ledger`：新增一条积分入账流水

如果你只想做最小验证，重点看这 3 个点即可：

1. `register` 接口返回了 `relation_id`
2. `reward` 接口返回 `ok`
3. `credit_ledger` 中出现了积分流水

## 配置说明

项目使用 `config.yaml` 管理运行配置：

```yaml
server:
  addr: ":8080"
mysql:
  dsn: "root@tcp(127.0.0.1:3306)/referral_system?parseTime=true&charset=utf8mb4&loc=Local"
```

如果你的 MySQL 有密码，把上面的 `root@...` 改成 `root:password@...` 即可。

### 字段说明

- `server.addr`
  - HTTP 服务监听地址
- `mysql.dsn`
  - MySQL 连接串

## 项目整体设计思路

这个系统采用的是**分层架构 + 领域拆分 + 事务编排**的方式。

### 1. 分层职责

- `main.go`
  - 负责启动、优雅退出、读取配置
- `internal/app`
  - 负责应用装配、数据库初始化、HTTP 路由挂载
- `internal/service`
  - 负责业务编排、事务控制、幂等处理、状态流转
- `internal/repository`
  - 定义数据访问接口，隔离业务层与持久化层
- `internal/repository/mysql`
  - 提供 MySQL 实现
- `internal/model`
  - 定义领域模型、枚举常量、数据结构
- `internal/txctx`
  - 在 `context.Context` 中传递事务对象

### 2. 核心业务模型

系统围绕三条主线展开：

- 邀请归因：谁邀请了谁
- 积分账务：发了多少积分、是否重复发放
- 统计展示：历史、看板、余额、流水

其中：

- `referral_relations` 负责“业务事实”
- `credit_ledger` 负责“账务事实”
- `credit_accounts` 负责“余额快照”
- `referral_events` 负责“过程审计”
- `referral_stats_daily` 负责“查询性能”

### 3. 主链路设计

注册链路：

1. 校验邀请参数
2. 根据邀请码找到邀请人
3. 创建被邀请人
4. 创建邀请关系
5. 记录注册事件
6. 提交事务

奖励链路：

1. 校验奖励参数
2. 检查幂等
3. 查询邀请关系
4. 确保账户存在
5. 用乐观锁更新余额
6. 写入积分流水
7. 更新邀请关系状态
8. 记录奖励事件
9. 提交事务

## 开发过程中作出的设计取舍

### 1. 选择“模块化单体”而不是一开始就拆微服务

**原因：**

- 这个题目本质上更关注业务正确性和数据一致性
- 邀请、发奖、账本、统计之间有较强的事务关联
- 直接拆微服务会引入更多分布式一致性成本

**取舍：**

- 第一版优先保证可读性、可维护性和正确性
- 通过清晰分层为未来拆分预留接口边界

### 2. 选择“账本 + 账户快照”而不是只存余额

**原因：**

- 只存余额无法审计
- 只存流水查询会慢

**取舍：**

- `credit_accounts` 提供快速读取
- `credit_ledger` 提供完整审计和重算能力

### 3. 选择“唯一索引 + 幂等键”双重防重

**原因：**

- 只靠代码判断挡不住并发
- MQ 重试、接口重放都可能导致重复执行

**取舍：**

- 数据库唯一约束作为最终兜底
- 服务层幂等键负责业务层防重

### 4. 选择“事务内写主表 + 异步/聚合查询”思路

**原因：**

- 交易链路必须准确
- Dashboard 和历史查询属于读多写少场景

**取舍：**

- 主链路尽量轻
- 聚合统计交给统计表或缓存

### 5. 选择 MySQL 原生 SQL 而不是 ORM 优先

**原因：**

- 账本、幂等、事务和唯一约束场景下，SQL 更直观
- 面试题更适合展示明确的数据控制能力

**取舍：**

- 开发量稍多，但行为更可控
- 方便解释索引、锁、事务和一致性策略

### 6. 选择将事务对象放进 context

**原因：**

- Service 层希望统一编排事务
- Repository 层希望自动识别当前是否处于事务中

**取舍：**

- 实现上增加一点基础设施代码
- 但能显著减少事务传递的样板代码

## 接口说明

### 健康检查

```bash
GET /healthz
```

返回示例：

```json
{"status":"ok"}
```

### 注册并绑定邀请关系

```bash
POST /api/v1/referrals/register
```

请求示例：

```json
{
  "name": "Bob",
  "email": "bob@example.com",
  "phone": "13800000000",
  "referral_code": "ABC123",
  "idempotency_key": "register-bob-001"
}
```

### 发放奖励

```bash
POST /api/v1/referrals/reward
```

请求示例：

```json
{
  "relation_id": 1001,
  "biz_id": "reg-2001",
  "amount": 100,
  "idempotency_key": "reward-1001"
}
```

### 查询接口

以下接口已经实现查询逻辑：

- `GET /api/v1/referrals/history?user_id=1&limit=20&offset=0`
- `GET /api/v1/referrals/dashboard?user_id=1`
- `GET /api/v1/credits/balance?user_id=1`
- `GET /api/v1/credits/ledger?user_id=1&limit=20&offset=0`

## 项目结构

```text
config.yaml
main.go
internal/
  app/
  model/
  repository/
  service/
  txctx/
```

## 当前已经具备的能力

- `GET /healthz`
- `POST /api/v1/referrals/register`
- `POST /api/v1/referrals/reward`
- `GET /api/v1/referrals/history`
- `GET /api/v1/referrals/dashboard`
- `GET /api/v1/credits/balance`
- `GET /api/v1/credits/ledger`
- MySQL 仓储实现
- 事务编排
- 幂等控制
- 基础测试骨架
- `config.yaml` 配置加载

## 后续可继续增强的方向

- 增加输入校验与统一错误码
- 增加完整测试用例
- 接入日志系统
- 增加风控与反作弊策略
- 增加统计异步任务与缓存层
- 将配置解析替换为成熟的 YAML 解析库

## 常见问题

### 1. 为什么用 `config.yaml` 而不是环境变量直接启动？

`config.yaml` 更适合本地开发和配置沉淀，启动时只需要读取一个文件即可。后续如果上云，也可以再叠加环境变量覆盖。

### 2. 为什么有些查询接口是占位实现？

当前查询接口已经接入 Service 层和 Repository 层的调用链，后续只需要补齐更复杂的分页字段、过滤条件和聚合统计即可。

### 3. 为什么不用 ORM 一把梭？

这类题目更适合展示对事务、索引、幂等和账本的控制能力。原生 SQL 更直观，也更容易解释设计取舍。

### 4. 为什么同时有 `credit_accounts` 和 `credit_ledger`？

前者是余额快照，用于快速查询；后者是审计流水，用于追踪、重算和对账。两者一起才完整。

### 5. 为什么要把事务放进 `context`？

这样 Service 层可以统一编排事务，Repository 层无需显式传递 `*sql.Tx`，只需要根据上下文自动感知是否在事务中。

## 说明

这是一套偏“面试可讲、可落地、可扩展”的参考实现，重点是展示 Referral 场景中的：

- 归因一致性
- 账本审计
- 幂等控制
- 查询优化
- 架构分层
