# GoMusical

独立音乐人 / 播客创作者的数字货架、高保真流媒体与赞助履约系统。

## 1. 如何启动

```bash
docker compose up --build -d
```

浏览器打开 `http://localhost:29471`。首次启动种子曲转码约 10–20 秒。

## 2. 使用说明

听众浏览货架、试听前 30 秒、在波形上写时间点乐评、赞助后解锁全曲与无损下载。创作者登录工作室管理授权档位、上传与转码、阅读乐评收件箱。管理员查看风控审计。

## 3. 服务列表及API说明

| 服务 | 地址 |
|---|---|
| Web | http://localhost:29471 |
| API（调试） | http://localhost:29472 |
| Postgres | localhost:29473 |
| Redis | localhost:29474 |

接口详见 `docs/API.md`。健康检查：`GET /api/health`。

## 4. 测试账号

| 角色 | 邮箱 | 密码 |
|---|---|---|
| 创作者 | creator@gomusical.local | Creator123! |
| 听众 | listener@gomusical.local | Listener123! |
| 管理员 | admin@gomusical.local | Admin123! |

## 5. 题目内容

用 Go 实现个人独立音乐人数字货架：Wavesurfer 波形播放（倍速 + 时间点乐评）、三档授权、不暴露源路径的 HLS 流与 HMAC 签名无损下载、粉丝赞助履约。

## 6. 项目结构

```
backend/              Go API + FFmpeg worker
frontend-user/         听众 SPA 构建入口
frontend-creator/      创作者页面源码分区
frontend-admin/        管理端源码分区
docs/                  需求 / 路线图 / API / 审核
tests/                 冒烟与 E2E
```

## 7. API 模拟与切换指南

支付抽象为 `payment.Provider`。

- **Mock（默认，合法）**：`PAYMENT_MODE=mock`。`MOCK_PAY_BEHAVIOR` 可选 `success` / `fail` / `timeout` / `delayed`。QA 与 CI 必须走 Mock，预期支出 ¥0。
- **Real 骨架**：`PAYMENT_MODE=real` 且设置 `PAYMENT_REAL_KEY`。密钥缺失时进程告警并**降级 Mock**，不会假装已经接通支付宝/微信/Stripe。真实渠道契约见 `docs/.meta/api_contracts.md`，状态 **UNVERIFIED**。

音频存储为本地卷 `mediadata`，接口 `storage.Local` 可替换对象存储（V1）。
