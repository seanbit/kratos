# Kratos Project With SimpleDDD

## Init
```
make init
```
## Create a service
```
# Generate the source code of service by proto file
kratos proto server api/protos/web/xxx.proto -t internal/service
```
## Generate other auxiliary files by Makefile
```
# Generate all files
# Automated Initialization (wire)
make all
```
## Docker
```bash
# build
docker build -t <your-docker-image-name> .

# run
docker run --rm -p 8000:8000 -p 9000:9000 -v </path/to/your/configs>:/data/conf <your-docker-image-name>
```
## Debug
Program arguments:
-secret-file=secret.yaml
Working directory:
***/kratos-ddd/cmd/web

---
## Layered Standard

```azure
app/
├── api/ # Protobuf 定义 (Http/Rpc 服务、DTO、错误码)
│ ├── protos/ # Proto 主目录，子目录按业务层划分定义，导出代码在上级api目录下的同名子目录
│ │ └── web/ # （示例）这里是web相关的proto，可以增加portal、back等
├── internal/ # 业务代码（遵循 DDD 分层）
│ ├── biz/ # 领域层：实体、聚合、用例、接口抽象
│ ├── data/ # 数据层：Repo 实现（数据库、缓存、消息队列、Rpc调用）
│ │ └── dao/ # gen 生成的数据库 DAO
│ │ └── models/ # gen 生成的数据库 Models
│ │ └── mocks/ # 生成的mock代码
│ │ └── tests/ # 测试代码
│ │ └── xxx.go # repo实现、biz需要的外部依赖能力的service接口实现
│ ├── global/ # 全局配置访问、全局变量（env，serviceName等）
│ ├── infra/ # 基础设施层：DBProvider、RPC Client、第三方客户端、外部能力封装
│ ├── scripts/ # 内部脚本，dao-generate等
│ ├── service/ # 应用层：API <-> biz 的防腐层
│ ├── static/ # 静态资源目录（需要打入包内的）：html模版、文本信息等
│ └── conf/ # 配置文件目录（YAML、JSON 等）
└── pkg/ # 公共库 (工具方法、ecode、日志、middleware)
└── third_party/ # 第三方代码依赖（无法通过gomod引入的，或者需要定制化的三方代码）
```


## Layered Guid

### 1. `api/protos/xxx`
- 只定义 **protobuf 文件**：
    - `code.proto` → 定义错误码（`ErrorReason` 枚举）及error状态码（示例：INVALID_PARAMS = 400 [(errors.code) = 400];）
    - `constants.proto` → 定义公共常量、枚举等
    - `xxx.proto` → HTTP/RPC 接口，以及DTO message： Request/Response（对应HTTP）、Args/Reply（对应RPC） 
- **不包含任何业务逻辑**
- 通过`make api`命令输出生成代码到api/xxx目录下

---

### 2. `internal/biz/`
- **领域层（Domain）**
- 职责：
    - 定义 **实体（Entity）**、**聚合根（Aggregate）**
    - 定义 **用例（UseCase）**
    - 定义 **抽象接口**：
        - 存储能力：`XxxRepo`
        - 外部依赖能力：`XxxService`
- 原则：只定义接口，不关心实现（依赖倒置）。

---

### 3. `internal/data/`
- **数据层（Repository 实现）**
- 职责：
    - 持久化（数据库、缓存、MQ）
    - 封装外部能力提供给Biz（转换外部数据给到Biz所需的数据）
    - 实现Biz的 `XxxRepo` 接口
    - 实现Biz的 `XxxService` 接口
- **依赖 infra 提供的 Client 完成内部/外部调用**。

---

### 4. `internal/infra/`
- **基础设施层（Infrastructure）**
- 职责：
    - 封装所有 **外部系统能力**（DB、RPC、HTTP SDK、第三方服务）
    - 命名规范：`XxxClient`（如 `PaymentClient`）
    - 提供稳定的调用接口给 data 层使用
- 原则：不暴露给 biz 层，所有调用必须通过 data。

---

### 5. `internal/service/`
- **应用层（Application Service）**
- 职责：
    - API 防腐层（api ↔ biz）
    - DTO ↔ 领域对象转换
    - 调用 biz 用例进行编排
- 命名：`XxxService`（如 `UserService`、`OrderService`）。
- 文件初始化使用kratos proto server命令（示例：kratos proto server api/protos/xxx/xxx.proto -t internal/service）

---

### 6. `pkg/`
- **公共库**
- 存放可复用的工具方法、通用中间件、统一错误码处理等。
- 原则：不依赖 biz/data/service，只能是“纯工具”。

### 7. `third_party/`
- **第三方依赖**
- 存放需要深度集成的三方代码等（不通过gomod引入的）。
- 有定制化需求改造后的三方代码。


## 依赖关系规范

✅ 允许的依赖方向：

- `api` → `service` → `biz` → `data` → `infra`


🚫 不允许的反向依赖，例如：
- biz 调用 service
- service 直接调用 infra
- biz 直接依赖 gen 生成的 model/DAO

**示例：调用外部支付服务**

- `biz/` 定义接口：`PaymentService`
- `infra/` 实现 RPC 客户端：`PaymentClient`
- `data/` 实现 `PaymentService`，内部调用 `PaymentClient`
- `service/` 编排 API → biz → data

依赖链示例：

- api.OrderHttp → service.OrderService → biz.OrderUseCase → data.PaymentService → infra.PaymentClient

---
