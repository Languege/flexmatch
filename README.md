[亚马逊FlexMatch](https://docs.aws.amazon.com/zh_cn/gamelift/latest/flexmatchguide/match-intro.html) 服务的go语言实现

### 安装
```shell
cd $GOPATH/src
mkdir github.com/Languege
cd github.com/Languege
git clone github.com/Languege/flexmatch
```

### 运行
```shell
cd flexmatch/service/match
bash start.sh
```

### 使用
#### 1. 对局服务器实现`open.FlexMatchGame`服务
用于向对局服务创建唯一游戏对局信息

#### 2. 创建对局配置
对局配置用于团队、匹配规则、事件通知等设置，具体参见`open.MatchmakingConfiguration`协议说明。
样例位于中台项目match应用`TestCreateMatchmakingConfiguration`测试用例下

#### 3. 快速开始
##### 3.1. 启动对局demo battle服务
启动脚本service/battle/start.sh

##### 3.2 启动匹配服务
启动脚本service/match/start.sh

##### 3.3 测试对局并消费匹配事件
测试用例`TestMatchEventConsume`位于service/match/api/test下，由于battle、match服务均使用的main-local.yaml, 测试`TestMatchEventConsume`时可指定配置配置为conf/main-local.yaml, 或临时修改文件名为main-local

#### 3.4 游戏对各个事件进行处理
考虑到实际使用过程中，部分信息匹配服务无法获知，游戏的推送交于游戏自定义。测试用例`TestMatchEventConsume`中TODO部分交于游戏实现。

考虑到消息推送可能存在丢失的情景，客户端应设置保底时长，超过阈值时对票据状态进行轮询，
对应的匹配服务接口`DescribeMatchmaking`

### 性能
在未对接游戏对局创建、kafka事件写入时，采用单一分值距离匹配、自动接受匹配规则，5v5单人匹配mac m1性能如下


| 对局数 | 平均对局完成耗时ns |
| ------ | ------ |
| `50` | `20695223` |
| `200` | `5586288` |
|`2000`|`1338885`|

测试用例`TestMatchmaking_TicketInput`位于service/match/entities/matchmaking_test.go中, 参数N设置对局数

### 其他
本服务为亚马逊匹配服务的实现，变动如下
1. 取消最小最大团队玩家人数规则、采用固定人数对局;

产品特性可参考亚马逊[匹配文档](https://docs.aws.amazon.com/zh_cn/gamelift/latest/flexmatchguide/match-client.html#match-client-track)




