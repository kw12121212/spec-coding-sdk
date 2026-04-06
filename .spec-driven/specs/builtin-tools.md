# builtin-tools

## ADDED Requirements

### Requirement: Supported built-in external tools

- 项目 MUST 提供一个内置外部工具管理能力，支持以下工具标识：`rg`、`gh`、`rtk`。
- 当调用方请求上述任一受支持工具时，系统 MUST 返回一个可执行文件路径，供后续调用方直接执行。
- 当调用方请求不受支持的工具标识时，系统 MUST 返回非 nil error，且错误信息 MUST 包含该工具标识。

### Requirement: Prefer system-installed executables

- 当受支持工具已在宿主系统可执行环境中可用时，系统 MUST 直接返回该可执行文件路径。
- 当系统已找到宿主安装的可执行文件时，系统 MUST NOT 再次下载或安装同一工具。
- 系统 MUST 让调用方能够区分解析结果来自宿主系统安装还是 SDK 托管安装。

### Requirement: Managed installation of prebuilt binaries

- 当受支持工具在宿主系统中不可用时，系统 MUST 尝试下载并安装该工具的受支持预编译二进制。
- 安装成功后，系统 MUST 返回新安装的可执行文件路径。
- 系统 MUST 将托管安装放置在 SDK 控制的目录中，而不是写入全局系统路径。
- 当相同工具的可复用托管安装已存在时，系统 MUST 复用该安装，而不是重复下载。
- 当当前操作系统或 CPU 架构不受支持时，系统 MUST 返回非 nil error，且错误信息 MUST 指出平台不受支持。
- 当下载、解压、写入或设置可执行权限失败时，系统 MUST 返回非 nil error，且 MUST NOT 报告安装成功。

### Requirement: Installation result integrity

- 当安装流程失败时，系统 MUST NOT 将不完整的可执行文件作为可用结果返回给后续调用方。
- 当同一工具被重复解析时，系统 MUST 返回一致的可执行文件路径，除非底层安装位置被外部删除或替换。
