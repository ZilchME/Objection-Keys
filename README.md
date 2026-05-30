# Objection Keys

给 macOS 键盘输入加上法庭系音效。

> **注意：** 项目主要由 Codex 辅助生成  
> **免责声明：** 本项目不是 Capcom 或《逆转裁判》官方项目，只是粉丝向小工具。

## 这是什么

一个能边打字边异议的工具。

[原版项目](https://github.com/Dreagonmon/input_sound)
的 macOS 原生 Go 实现，基于 CoreGraphics 构建，零运行时依赖。  

## 工作原理

应用通过 macOS CoreGraphics `CGEventTap` 捕获全局键盘事件，并根据按键类型实时播放对应音效。

## 系统要求

- macOS 12+（Apple Silicon / Intel）
- Go 1.24+（从源码构建时需要）

## 构建说明

### 从源码构建

```bash
# 克隆项目
git clone https://github.com:ZilchME/Objection-Keys.git
cd Objection-Keys

# 安装依赖
go mod download

# 构建
make build

# 构建 macOS 菜单栏应用
make build-app

# 交叉构建 Windows 托盘应用
make build-windows
````

或者手动构建：

```bash
CGO_ENABLED=1 go build -ldflags="-s -w" -o objection-keys ./cmd/objection-keys/
```

### 直接运行

```bash
./objection-keys
```

## 权限设置

macOS 需要**辅助功能权限**才能捕获全局键盘事件。

不授权的话，它只能在旁边看着你打字，不能替你播放庭审 BGM。

### 授予权限步骤

1. 首次运行应用，会显示友好的错误提示
2. 打开 **系统设置 → 隐私与安全性 → 辅助功能**
3. 点击 **+** 添加应用（导航到 `objection-keys` 文件所在位置）
4. 开启应用的开关
5. 重新运行应用

> **注意：** 应用无法通过代码自动请求或授予此权限，必须用户在系统设置中手动授权。

## 使用方法

```bash
# 在项目根目录下运行（需确保 sounds/ 目录存在）
./objection-keys

# 或从其他位置运行，将 sounds 目录复制过去即可
cp -r sounds/ /任意位置/
/path/to/objection-keys
```

直接运行二进制会启动菜单栏托盘图标：

```bash
./objection-keys
```

构建 `.app` 后可以从 Finder 打开 `Objection Keys.app`，或用命令启动：

```bash
open "Objection Keys.app"
```

托盘菜单支持暂停/恢复音效、检查辅助功能权限和退出应用。

### Windows

Windows 构建产物位于 `dist/windows/`：

```bash
make build-windows
```

把整个 `dist/windows/` 目录复制到 Windows 后运行 `objection-keys.exe` 即可。程序会在系统托盘显示图标，并从同目录的 `sounds/` 读取音效。

## WAV 格式

音效文件必须为 **8 位无符号 PCM**、**22050 Hz**、**立体声** 格式。

```bash
# 使用 ffmpeg 转换任意音频到目标格式
ffmpeg -i "输入文件" -c:a pcm_u8 -ar 22050 -ac 2 "输出文件.wav" -y

# 示例
ffmpeg -i "objection.mp3" -c:a pcm_u8 -ar 22050 -ac 2 "enter.wav" -y
ffmpeg -i "desk_slam.mp3" -c:a pcm_u8 -ar 22050 -ac 2 "space.wav" -y
ffmpeg -i "blip.mp3"      -c:a pcm_u8 -ar 22050 -ac 2 "alphabet_fast.wav" -y
ffmpeg -i "confirm.mp3"   -c:a pcm_u8 -ar 22050 -ac 2 "number.wav" -y
```

## 按键映射

| 按键                                    | 音效                                    |
| ------------------------------------- | ------------------------------------- |
| a–z                                   | alphabet_fast.wav / alphabet_slow.wav |
| 0–9, -, =, +, /, *, (, ), &, ^, %, $, @, !, &#96;, ~, [, ], {, }, ;, ', \, \|, ,, <, >, ? | number.wav |
| 空格                                   | space.wav                             |
| Enter / Return                        | enter.wav                             |
| Backspace / Delete                    | backspace.wav                         |
| Escape                                | esc.wav                               |

以下按键受支持但未映射音效，会被安静地忽略：

- 功能键（F1–F20）
- 修饰键（Ctrl、Shift、Alt、Command、Caps Lock）
- Tab、方向键和其他未映射按键

它们没有异议。
暂时。

## 常见问题

### 能不能自动申请权限？

不能。

macOS 不允许应用自己给自己开辅助功能权限。

### 会不会记录我的键盘输入？

不会。

应用只监听按键事件用于播放音效，不保存、不上传、不分析你的输入内容。
它只负责在旁边"哒哒哒"。

## 鸣谢

**原版项目：** [input_sound](https://github.com/Dreagonmon/input_sound)，作者 Dreagonmon

## 许可证

本项目采用 MIT 许可证 — 详见 [LICENSE](LICENSE) 文件。
