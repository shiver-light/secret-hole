# 树洞后端 (Secret-Hole Backend)

极简匿名“树洞”：**阅后即焚** + **每日配额** + **一次回复 -> 临时好友（指数增长时长）**  
- 每人每天可**发送 3 次 / 领取 3 次**公共消息（独立计数）。  
- 消息含“阅读后持续时间”，**被读后按持续时间到期即焚**（服务器与客户端共同清理）。  
- 对一条公共消息可**一次回复**；当双方**同时在线（前台且活跃）**时，自动成为**临时好友**：
  - 初次 1 分钟，之后同一对用户再次成为临时好友，**有效期按 2、4、8… 指数增长**；
  - 临时好友**期间不限制私聊消息条数**；
  - 到期后：**清空双方会话消息**，但后台会**保留临时好友行与累计时长**用于成长统计（`streak` 保留）。

---

## 目录
- [环境与依赖](#环境与依赖)
- [本地启动](#本地启动)
- [运行 & 验证](#运行--验证)
- [API 速览](#api-速览)
- [数据模型要点](#数据模型要点)
- [定期清理任务](#定期清理任务)
- [许可](#许可)

---

## 环境与依赖

- Go **1.22+**
- PostgreSQL **16+**
- Docker / docker-compose（可选，本地最方便）
-（如在中国大陆网络）建议设置：
  ```bash
  go env -w GOPROXY=https://goproxy.cn,direct
  go env -w GOSUMDB=off
  ```

---

## 本地启动

1️⃣ 启动数据库

```bash
docker-compose up -d db
```

2️⃣ 启动 API

```bash
make run
# 等价于：
# ADDR=:8080 DATABASE_URL=postgres://postgres:postgres@localhost:5432/shudong?sslmode=disable go run ./cmd/api
```

3️⃣ 测试接口

```bash
# 登录/注册匿名设备
curl -s http://localhost:8080/v1/auth/anon \
  -H 'content-type: application/json' \
  -d '{"device_hash":"dev-abc"}'

# 发送消息
curl -s http://localhost:8080/v1/messages \
  -H 'X-User-ID: 1' -H 'content-type: application/json' \
  -d '{"body":"桃花仙人种桃树","read_duration":30}'

# 领取
curl -s -X POST http://localhost:8080/v1/messages/claim \
  -H 'X-User-ID: 2'
```
---

## 运行 & 验证

> 展示新功能：前台心跳、一次回复触发临时好友、临时好友期间 DM。

### 1️⃣ 心跳（前台活跃）

```bash
curl -s -X POST http://localhost:8080/v1/presence/heartbeat \
  -H 'X-User-ID: 1' -H 'content-type: application/json' \
  -d '{"is_foreground":true}'
```

### 2️⃣ 一次回复触发临时好友

```bash
curl -s -X POST http://localhost:8080/v1/messages/123/reply \
  -H 'X-User-ID: 1' -H 'content-type: application/json' \
  -d '{"body":"hi"}'
```

返回：

```json
{"ok":true,"active_until":"...","streak":1}
```

### 3️⃣ 临时好友期间私聊

```bash
curl -s -X POST http://localhost:8080/v1/dm/send \
  -H 'X-User-ID: 1' -H 'content-type: application/json' \
  -d '{"to":2,"body":"hello"}'
```

---

## API 速览

| Method | Endpoint                      | 说明                 |
| ------ | ----------------------------- | ------------------ |
| POST   | `/v1/auth/anon`               | 匿名注册/登录，返回 user_id |
| GET    | `/v1/me/quota`                | 获取当天配额             |
| POST   | `/v1/messages`                | 发送公共消息             |
| POST   | `/v1/messages/claim`          | 领取消息               |
| POST   | `/v1/messages/:id/ack-delete` | 阅后即焚确认删除           |
| POST   | `/v1/messages/:id/reply`      | 一次回复，触发临时好友        |
| POST   | `/v1/presence/heartbeat`      | 前台心跳（判断同时在线）       |
| POST   | `/v1/dm/send`                 | 临时好友私聊（有效期内不限条数）   |

---

## 数据模型要点

* **users**：`nickname`（≤7字）、`avatar_b64`、`gender_color`
* **daily_quota**：按用户每日独立计数
* **messages**：公共消息 + 一次回复字段
* **user_presence**：前台心跳
* **temp_friendships**：临时好友窗口 + streak 累计时长
* **direct_messages**：窗口期内私聊，过期清理

---

## 定期清理任务

### 公共消息（阅后即焚）

```sql
DELETE FROM messages
WHERE expires_at IS NOT NULL
  AND expires_at <= now()
LIMIT 5000;
```

### 临时好友过期清理

```sql
WITH expired AS (
  SELECT user_min, user_max, last_started_at, active_until
  FROM temp_friendships
  WHERE active_until IS NOT NULL AND active_until <= now()
)
UPDATE temp_friendships tf
SET total_active_seconds = tf.total_active_seconds +
    GREATEST(0, EXTRACT(EPOCH FROM (LEAST(tf.active_until, now()) - tf.last_started_at)))::BIGINT,
    last_started_at = NULL,
    active_until   = NULL
FROM expired e
WHERE tf.user_min=e.user_min AND tf.user_max=e.user_max;

DELETE FROM direct_messages dm
USING expired e
WHERE (dm.sender_id=e.user_min AND dm.recv_id=e.user_max)
   OR (dm.sender_id=e.user_max AND dm.recv_id=e.user_min);
```

---

## 许可

MIT License © 2025 Secret-Hole Authors
