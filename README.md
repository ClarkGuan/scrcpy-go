# scrcpy-go
原版 [scrcpy](https://github.com/Genymobile/scrcpy) 是 Genymobile 公司出品的 Android 设备显示与操纵开源工具，具备投屏、控制、截图等功能。scrcpy-go 是在其基础上制作的方便进行手机游戏的辅助工具，类似 TC-Games 出品的软件。

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

### 构建
```bash
brew install sdl2 sdl2_ttf ffmpeg go
go get github.com/ClarkGuan/scrcpy-go
```

### 使用说明
```bash
scrcpy-go -log {日志等级} -bitrate {H.264 码率} -port {adb 端口号} -cfg {settings.yml 配置文件路径}
```

一般情况下，直接双击 `scrcpy-go` 即可；如果想要查看日志信息可以使用 `scrcpy-go -log 4` 查看具体日志输出。

选项默认值：
* log: 0
* bitrate: 8000000
* port: 27183
* cfg: scrcpy-go 所在目录下 res/settings.yml

### 配置文件
[res/settings.yml](res/settings.yml) 是默认的配置文件所在位置。其内容是作者在玩刺激战场时配置的数值，可以根据自身机型和爱好自定义配置（而且不局限于射击类手游）。

#### 属性说明
1. code：对应 SDL 内键盘映射的[字符串值](https://wiki.libsdl.org/SDL_Keycode?highlight=%28%5CbCategoryEnum%5Cb%29%7C%28CategoryKeyboard%29)。特别地，以 SCRCPY_ 开头的是作者自定义的常量值，为了完成一些特定的功能（与射击类游戏相关），具体细节可以参看代码实现。另，SDL 中不存在使用字符串反查鼠标按键的功能，所以将鼠标按键映射的字符串都是作者自定义的（BUTTON_LEFT、BUTTON_MIDDLE、BUTTON_RIGHT、BUTTON_X1、BUTTON_X2）。
2. point：屏幕坐标映射。
3. macro：宏定义，可以是一系列坐标点事件。
4. delay：宏定义中，不同点击事件之间的时间间隔。
5. type：可选值有 ctrl 和 mouse 两种，表示是否需要同时按下 ctrl 键或者是否是鼠标按键事件。
6. show_pointer：是否切换[鼠标状态](https://wiki.libsdl.org/SDL_SetRelativeMouseMode?highlight=%28%5CbCategoryMouse%5Cb%29%7C%28CategoryEnum%29%7C%28CategoryStruct%29)。
7. comment：注释。

#### 特殊功能
1. ctrl + h：点击 Home
2. ctrl + b：点击 Back
3. ctrl + m：点击 Menu
4. ctrl + p：点击 Power
5. ctrl + s：点击 App Switch
6. ctrl + ;：音量放大
7. ctrl + '：音量缩小
8. ctrl + x：切换鼠标状态

### 后续可能的计划
1. 重构代码。因为该工具只是个人爱好而作，能用即可，代码无层次无章法。后续可能进行少许重构，调整一些代码结构，以求层次鲜明（勉强能看）。
2. 支持 Windows 和 Linux，可能加入对应平台的硬解支持。
3. 使用原生 UI 框架支持而非 SDL。

### 代码版权
代码随便 copy 随便用，但请记得署名一下~ :-D

