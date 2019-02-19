# scrcpy-go
原版 scrcpy 是 Genymobile 公司出品的 Android 设备显示与操纵开源工具，具备投屏、控制、截图等功能。scrcpy-go 是在其基础上制作的方便进行手机游戏的辅助工具，类似 TC-Games 出品的软件。

特别地，与 scrcpy 的主要不同点是：

* 使用 Golang 编写 client 端，而不是 C
* 增加了多点触控的支持（scrcpy 只支持单点触控）
* 去掉了一些游戏时不需要的功能
* 增加了 MacOS 上硬件解码的支持（VideoToolBox）

目前只支持在 MacOS 上运行（偷懒:-D），后面可以考虑支持 Windows 和 Linux。

### 依赖
1. Golang 环境
2. sdl2
3. sdl2_ttf
4. ffmpeg
5. Android adb 工具
6. yaml

### 环境设置
```bash
brew install sdl2 sdl2_ttf ffmpeg go
go get github.com/ClarkGuan/scrcpy-go
```