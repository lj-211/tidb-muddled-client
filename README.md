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

### demo设计思路

