## db的乱序客户端模拟
### 使用示例
```
Usage of ./tidb-muddled-client:
  -b string
        同一批次客户端的批次id
  -c string
        配置文件文件路径
  -id string
        节点id
  -p string
        伙伴id,exp: 1,2,3
```

因为两个client之间是完全无状态的，所以在没有三方协调进程的辅助下，需要通过数据做同步。
批次id就承担了这个作用，同时需要指定伙伴client的标识。

示例:
```
./tidb-muddled-client -c ./config/config.toml -id 1 -p 2 -b 9982
./tidb-muddled-client -c ./config/config2.toml -id 2 -p 1 -b 9982
```

### 代码结构
```
├── README.md
├── common                          公共库
│   ├── errors.go
│   ├── permutation.go
│   └── permutation_test.go
├── config                          配置目录
│   ├── config.toml
│   ├── config2.toml
│   ├── sql.file
│   └── sql2.file
├── coordinate                      任务协调器
│   ├── coordinate.go
│   └── db_coordinate.go            基于db的任务协调器
├── doc                             文档
│   └── sql.md
├── global.go
├── go.mod
├── go.sum
├── init.go
├── loader                          输入加载模块
│   ├── file_loader.go
│   └── loader.go
├── main.go
├── sql_cmd.go                      client的sql处理逻辑
├── sql_load.go                     client的sql加载逻辑
```
### demo编码思路
在没有任务调度的前提下，通过数据做状态同步和任务调度，是实现的主要思路。

在编码中，主要是把client输入，以及任务调度这两个地方，做了抽象:
1. sql加载可以通过不同的源进行扩展
2. 任务的调度是基于存储系统做的，在这个思路下，抽象了任务调度的基础接口

> 本例中实现的任务调度是基于db存储来做的
> 在基于存储做任务调度的思路下，可以实现其他的调度器，比如基于redis或者基于消息队列,单机可以考虑共享内存的方案

```
                                                                              +----------------------------+
                                                                              |                            |
                                                                        +-----+   Coordinater base on db   |
                                                                        |     |                            |
                                                                        |     +----------------------------+
                                             +-----------------+        |     +----------------------------+
                                             |                 |        |     |                            |
                                             |   Coordinater   +--------------+  Coordinater base on redis |
                                             |                 |        |     |                            |
                                             +----+-----+------+        |     +----------------------------+
                                                  ^     |               |
                                                  |     |               |     +----------------------------+
                                                  |     |               |     |                            |
                                     regist task  |     | check result  +-----+  Coordinate base one queue |
                                                  |     |                     |                            |
                                                  |     v                     +----------------------------+
             +-----------------+             +----+-----+------+
             |                 |   input     |                 |
             |    Sql Loader   +------------>+     Client      |
             |                 |             |                 |
             +-------+---------+             +-----------------+
                     |
      +--------------------------------+
      |              |                 |
      |              |                 |
+-----+------+  +----+-----+   +-------+------+
|            |  |          |   |              |
| FileLoader |  | DbLoader |   | OtherStorage |
|            |  |          |   |              |
+------------+  +----------+   +--------------+


```
