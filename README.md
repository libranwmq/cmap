# cmap: concurrent map, 并发安全 map.

本仓库是重构了[concurrent-map](https://github.com/orcaman/concurrent-map);
由于原仓库已经很久没有维护了，所以重构了一遍，并添加了一些新的功能。
仓库会由大家共建维护，欢迎大家提交 PR。

#### 主要修改点：
1. 支持增加默认分片数量，由原来的固定 32 个分片，改为可配置，不设置默认也为 32 个分片。
2. 支持整形（int32/int64）作为 key。去掉默认 string 类型 New() 默认函数。
3. 通过接口的方式对外提供功能。
4. 使用 mockery 工具对泛型接口进行 mock。
5. 添加了 examples 文件夹，提供了一些使用示例。