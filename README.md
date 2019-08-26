# Music-Get

[![Build Status](https://travis-ci.org/winterssy/music-get.svg?branch=master)](https://travis-ci.org/winterssy/music-get)
[![golang.org](https://img.shields.io/badge/golang-1.12-blue.svg)](https://golang.org)
[![Latest Release](https://img.shields.io/github/release/winterssy/music-get.svg)](https://github.com/winterssy/music-get/releases)
[![License GPL-3.0](https://img.shields.io/github/license/winterssy/music-get.svg)](/LICENSE)

[网易云音乐](https://music.163.com) | [QQ音乐](https://y.qq.com) 下载助手，支持一键下载单曲/专辑/歌单以及歌手热门歌曲，并自动更新音乐标签。

![music-get](/screenshots/music-get.gif)

>本项目仅供学习研究使用。如侵犯你的权益，请 [联系作者](mailto:winterssy@foxmail.com) 删除。

## 下载安装

你可以前往 [Releases](https://github.com/winterssy/music-get/releases) 标签下载程序的最新版本，或者克隆项目源码自行编译。

## 如何使用？

直接将音乐地址作为命令行参数传入即可，如：

- 下载单曲：
```
$ music-get https://music.163.com/#/song?id=553310243
$ music-get https://y.qq.com/n/yqq/song/002Zkt5S2z8JZx.html
```

- 下载专辑：
```
$ music-get https://music.163.com/#/album?id=38373053
$ music-get https://y.qq.com/n/yqq/album/002fRO0N4FftzY.html
```

- 下载歌单：
```
$ music-get https://music.163.com/#/playlist?id=156934569
$ music-get https://y.qq.com/n/yqq/playsquare/5474239760.html
```

- 下载歌手热门歌曲：
```
$ music-get https://music.163.com/#/artist?id=13193
$ music-get https://y.qq.com/n/yqq/singer/000Sp0Bz4JXH0o.html
```

命令选项：
- `-br`：优先下载音质，可选128/192/320，默认128。
- `-o`：下载保存目录，默认为 `/home/用户名/Music-Get`  （Windows为 `C:\\Users\\用户名\\Music-Get` ）。
- `-f`：是否覆盖已下载的音乐，默认跳过。
- `-n`：并发下载任务数，最大值16，默认1，即单任务下载。
- `-h`：获取命令帮助。

**注：** 命令选项必须先于其它命令行参数输入。

## 配置文件

程序的配置文件位于 `/home/用户名/music-get.json`（Windows为 `C:\\Users\\用户名\\music-get.json` ），用于本地存储cookies以及配置默认下载的比特率（最近一次使用的值，优先级低于 `-br` 指令）。**请勿对该文件进行任何修改！**

## 运行截图

- 单任务下载：

![](/screenshots/single-download.png)

- 多任务同时下载：

![](/screenshots/concurrent-download.png)

- 自动更新音乐标签（效果预览）：

![](/screenshots/tag-updated.png)

## FAQ

- 为什么网易云音乐需要登录？

  > 因为网易云音乐反爬，不登录会被服务端识别成欺诈而无法下载。程序会存储cookie到本地，但如果cookie失效了你需要再次登录，一般是每两周需要重新登录一次。目前仅支持手机登录方式。

- 是否支持一键下载网易云音乐『我喜欢的音乐』列表？

  > 支持。它本质上是一个歌单。

- 为什么指定了 `-br=320` 下载的却是128kbps？

  > 这只是在请求上优先保证，实际上下载的比特率由服务器返回的数据决定。

- 是否有支持其它音乐平台的计划？

  > 目前暂无，但开发者可以fork本项目的源码自行实现，只须实现 `MusicRequest` 接口即可。同时，欢迎PR。

- 下载失败的原因？

  > 网络状态不佳导致响应超时；触发了服务端的反爬机制（下调并发下载任务数/隔一段时间再试）；音乐提供商变更了API（这种情况下请提issue反馈）。

## 致谢

- 网易云音乐 Node.js API：[Binaryify/NeteaseCloudMusicApi](https://github.com/Binaryify/NeteaseCloudMusicApi)


## License

GPLv3.
