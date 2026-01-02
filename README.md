# AutoGLM-Go

[![Go Version](https://img.shields.io/badge/Go-1.23+-blue.svg)](https://golang.org)
[![License](https://img.shields.io/badge/License-MIT-green.svg)](LICENSE)

AutoGLM-Go 是 Open-AutoGLM 项目的 Go 语言重写版本，专注于 Android 设备的自动化操作。本项目使用 AI 模型来理解和执行手机操作任务，通过 ADB (Android Debug Bridge) 与 Android 设备进行交互。

> **注意**: 本项目是原 [Open-AutoGLM](https://github.com/zai-org/Open-AutoGLM) 项目的 Go 重写版本，与原项目的主要区别是：
> - 目前仅支持 Android 设备，不支持鸿蒙和 iOS 设备
> - 使用 Go 语言重写，提供更好的性能和更简单的部署
> - 保留了原项目的核心功能和 AI 驱动的自动化能力

## 功能特点

- 🤖 AI 驱动的手机自动化操作
- 📱 支持 Android 设备（通过 ADB）
- 🖼️ 屏幕截图和 UI 元素识别
- 📲 应用程序启动和操作
- 🌐 支持本地和远程设备连接
- 🛠️ 丰富的命令行工具

## 系统要求

- Go 1.23 或更高版本
- Android SDK Platform Tools (ADB)
- Android 设备或模拟器（已启用开发者选项和 USB 调试）
- OpenAI 兼容的 API 服务

## 安装步骤

### 1. 克隆仓库

```bash
git clone https://github.com/ZoroSpace/autoglm-go.git
cd autoglm-go
```

### 2. 安装依赖

```bash
go mod download
```

### 3. 编译项目

```bash
go build -o autoglm-go main.go
```

### 4. 设置 Android 设备

1. 在 Android 设备上启用"开发者选项"
2. 在"开发者选项"中启用"USB 调试"
3. 通过 USB 连接设备到计算机，或设置网络 ADB 连接

### 5. 验证 ADB 连接

```bash
adb devices
```

您应该能看到已连接的设备列表。

## 使用方法

### 基本用法

```bash
# 使用默认设置运行
./autoglm-go "打开微信并发送消息给张三"

# 指定模型 API
./autoglm-go --base-url http://localhost:8000/v1 --model autoglm-phone "打开设置并调整音量"

# 使用 API 密钥
./autoglm-go --apikey sk-xxxxx "打开抖音并搜索美食视频"

# 指定设备 ID
./autoglm-go --device-id emulator-5554 "打开淘宝搜索手机"
```

### 设备管理

```bash
# 列出已连接的设备
./autoglm-go --list-devices

# 连接到远程设备
./autoglm-go --connect 192.168.1.100:5555

# 断开远程设备连接
./autoglm-go --disconnect 192.168.1.100:5555

# 启用设备的 TCP/IP 调试模式
./autoglm-go --enable-tcpip 5555

# 获取设备 IP 地址
./autoglm-go --get-device-ip
```

### 应用程序支持

```bash
# 列出支持的应用程序
./autoglm-go --list-apps
```

## 配置选项

| 选项 | 环境变量 | 默认值 | 描述 |
|------|----------|--------|------|
| `--base-url` | `PHONE_AGENT_BASE_URL` | `http://localhost:8000/v1` | 模型 API 基础 URL |
| `--model` | `PHONE_AGENT_MODEL` | `autoglm-phone` | 模型名称 |
| `--apikey` | `PHONE_AGENT_API_KEY` | `EMPTY` | API 密钥 |
| `--max-steps` | `PHONE_AGENT_MAX_STEPS` | `100` | 每个任务的最大步数 |
| `--device-id` | `PHONE_AGENT_DEVICE_ID` | - | ADB 设备 ID |
| `--lang` | `PHONE_AGENT_LANG` | `cn` | 系统提示语言 (cn 或 en) |

## 支持的应用程序

本项目支持大量 Android 应用程序的自动化操作，包括但不限于：

- 社交应用：微信、QQ、微博、抖音等
- 购物应用：淘宝、京东、拼多多等
- 视频应用：哔哩哔哩、爱奇艺、腾讯视频等
- 音乐应用：网易云音乐、QQ音乐等
- 生活服务：支付宝、美团、饿了么等

使用 `--list-apps` 命令可以查看完整的应用程序列表。

## 工作原理

1. **截图获取**: 通过 ADB 获取设备当前屏幕截图
2. **UI 分析**: 使用 AI 模型分析屏幕内容和 UI 元素
3. **决策制定**: 根据任务和当前屏幕状态决定下一步操作
4. **操作执行**: 通过 ADB 执行点击、滑动、输入等操作
5. **循环迭代**: 重复上述过程直到任务完成

## 示例

```bash
# 社交应用操作
./autoglm-go "打开微信，给张三发消息说我今天晚点到"

# 购物应用操作
./autoglm-go "打开淘宝搜索iPhone 15并加入购物车"

# 视频应用操作
./autoglm-go "打开抖音搜索美食视频并点赞前三个"

# 系统设置操作
./autoglm-go "打开设置将屏幕亮度调整到50%"
```

## 开发

### 项目结构

```
autoglm-go/
├── main.go              # 主程序入口
├── constants/           # 常量定义
│   ├── apps.go         # 支持的应用程序包名
│   ├── device.go       # 设备相关常量
│   ├── i18n.go         # 国际化文本
│   └── prompt.go       # AI 提示词
├── phoneagent/          # 核心功能实现
│   ├── agent.go        # 代理主逻辑
│   ├── android/        # Android 设备实现
│   ├── definitions/    # 数据结构定义
│   ├── helper/         # 辅助函数
│   ├── interface.go    # 接口定义
│   └── llm/            # LLM 客户端
├── utils/              # 工具函数
└── scripts/            # 脚本文件
```

## 致谢

- [Open-AutoGLM](https://github.com/zai-org/Open-AutoGLM) - 原始项目
- [go-openai](https://github.com/sashabaranov/go-openai) - OpenAI Go 客户端
- [cobra](https://github.com/spf13/cobra) - Go CLI 框架