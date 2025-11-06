# 树洞后端


## 本地启动（需要 Go 1.22+、PostgreSQL 16+）


1. 启动数据库
```bash
docker compose up -d db
```
2. 启动 API
```bash
make run
```
3. 用 curl 测试
```bash
# 登录/注册匿名设备
curl -s http://localhost:8080/v1/auth/anon -d '{"device_hash":"dev-abc"}' -H 'content-type: application/json'
# 假设返回 user_id = 1


# 发一条
curl -s http://localhost:8080/v1/messages \
-H 'X-User-ID: 1' -H 'content-type: application/json' \
-d '{"body":"你好，世界","read_duration":30}'


# 领取
curl -s http://localhost:8080/v1/messages/claim -H 'X-User-ID: 2'

### 定期清理任务
每分钟执行一次（示例 crontab）：
```sql
DELETE FROM messages WHERE expires_at IS NOT NULL AND expires_at <= now() LIMIT 5000;
```
> 生产建议用队列/asynq 或后台 worker；并配合 `VACUUM (AUTO)`。