# AI Coding 规范

## 编译规范

**必须使用 Makefile 命令进行编译：**

```bash
make build           # 编译所有服务 (web, consumer, cronjob, job)
```

**编译产物统一输出到 `bin/` 目录。**

如果 Makefile 中没有对应的构建目标，必须先添加，再使用。禁止直接使用 `go build` 命令编译到项目根目录或其他位置。

---

## 项目结构

```
├── cmd/                    # 服务入口
│   ├── web/                # Web API 服务 (gRPC + HTTP + Crontab)
│   ├── consumer/           # 独立 Asynq 消费者服务 (可弹性伸缩，Metrics HTTP :9090)
│   ├── cronjob/            # 定时任务服务
│   └── job/                # 一次性任务工具 (K8S Job)
├── internal/               # 内部业务代码
│   ├── conf/               # 配置定义 (Protobuf)
│   ├── crontab/            # 定时任务实现
│   ├── biz/                # 领域层
│   ├── data/               # 数据层
│   ├── infra/              # 基础设施层
│   └── service/            # 应用服务层
├── configs/                # 配置文件
└── bin/                    # 编译产物
```

---

## 依赖关系

允许的依赖方向：`api → service → biz → data → infra`

禁止反向依赖。

---

## 代码生成

修改 `.proto` 文件后，运行：

```bash
make config    # 生成 protobuf 代码 (conf.proto)
make api       # 生成 API protobuf 代码
make generate  # 生成 wire 依赖注入代码 (go generate ./...)
make all       # 以上全部执行
```
