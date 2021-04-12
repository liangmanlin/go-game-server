# 使用golang开发的框架

> ****
## How?

- 参考 [main.go](main.go)
  
## Features

### `kernel`使用channel 和 goroutine 模拟的`Actor`模式
 * 使用channel 模拟的消息队列
 * `kernel.Context`内部实现了一个链表，用于发起call的时候，由于自身channel满，导致双向阻塞
   * 后续需要改成环形队列，减少一些垃圾的产生


### 乞丐版otp框架，核心package：`kernel`
 * 提供了简略版的`supervisor`
 * 提供类型`gen_server`的进程回调
 * 内置`logger`模块
 * 支持名字注册（local）
 * 支持进程`link`
 * 内置全局定时器`timer`
 * 内置简略控制`web api`


### 框架提供一种热更新方案
 * 详情请参考`hotfix`相关代码
 * how to use?
   * 参考`hotfix/20210119/20210119.go`编写更新代码，编译通过后
    ```bash
      ##本地网络执行网络请求，浏览器中也可以
          
      curl "http://127.0.0.1:3000/reload?mod=20210119"
    ```
 * 其中`reload`的逻辑由`main.go`中的`func reload(w http.ResponseWriter, req *http.Request)`提供


### 内置`mysql`支持
 * 引用如下代码开始使用
    ```go
        import (
        "github.com/liangmanlin/gootp/db"
        )
  
        tabSlice := []*db.TabDef{
            {Name: "account", DataStruct: &global.Account{}, Pkey: []string{"Account"}, Keys: []string{"AgentID"}},
        }
        db.Env.DBConfig = db.Config{Host: "127.0.0.1", Port: 3306, User: "root", PWD: "123456"}
  
        db.Start(tabSlice, "dbName", "logDbName")
    ```

### 内置一种定长协议
 
* 定义文件：`global/pb_def.go`
    ```go
      type LoginTosLogin struct { // router:LoginLogin
        Account string
      }

      type LoginTocLogin struct {
        ID   int32
        Name string
        F    float32
        M    map[int32]int32
        S    []string
      }
    ```
 
* 核心实现package：`github.com/liangmanlin/gootp/gate/pb`

* 辅助工具：`tool/pbBuild`

* `router`是工具生成的，仅仅提供路由到`player`包的函数
    
* 目前可以到处lua代码直接使用(仅支持lua5.3及以上版本)
   * 使用如下脚本到处lua代码
    ```bat
        %% 假设在game目录
        tool\bin\pbBuild.exe -f ./global/pb_def.go -c client
    ```

* 项目内 **require** 所有导出文件；接着可以如下：
    ```lua
        local function test()
            local tos = LoginTosLogin()
            local sendBuf = tos.encode()
            -- 然后你可以同网络发送这个sendBuf到服务器

            -- 假设接收到服务器发来的数据
            local recBuff
            ---@type LoginTocLogin
            local proto,protoID = ProtoDecode(recBuff)
            -- proto即为解析到的对象
            local a = proto.F
        end
    ```


### 框架内置一个网关：`gate`

* 通过如下方式启动
    ```go
      import "github.com/liangmanlin/gootp/gate"
  
      gate.Start(flag string, handler *kernel.Actor, port int, opt ...interface{})
      
    ```

* 建议结合内置的定长协议使用
    

### 2021.1.28 添加分布式多节点支持
* 目前不支持在**windows**中运行多节点
    * `Pid`支持跨节点传输
      * 你可以往另外一个节点发送一个本地pid，然后在那个节点上调用`call`和`cast`
    * 向一个远程节点发送数据需要先定义协议,例如：
    ```go
        import "game/node"
        // 目前尚未找到一种可以直接发送对象的办法，所以这个是一种妥协
        // 即使使用grpc，也是要定义一个proto，所以依然选择使用内置协议，减少一层转换
        nodeProto := []interface{}{&global.TestCall{},&global.StrArgs{},&global.RpcStrResult{}}
        Cookie := "6d27544c07937e4a7fab8123291cc4df"
        node.Start(global.Root, "game@127.0.0.1", Cookie, nodeProto)
    ```
* 支持`RpcCall`
     * 目前是单线程执行的，如果执行的函数会阻塞，尽量不要使用
     * 需要自行调用`node.RpcRegister(fun string,function RpcFunc)`注册
    
### 2021.2.23 添加excel配置文件支持

* 具体使用参考 Excel/Effect.xlsm 配置文件
    * 如需导出配置，请双击 `Excel/_打包单个配置.bat`
    
* 支持热更新配置
    ```bash
    ## 以热更Effect配置为例
    curl "http://127.0.0.1:3000/reload_config?name=Effect"
    ```

### 2021.4.1 抽离大部分公共代码到一个代码仓库，命名为 `gootp`

- 同时移除原来使用一个函数对象结构体的做法，直接使用包的全局变量代替导出函数，
  这样也可以达到热更新的目的，并且代码编写会轻松一点

## Problem
  * 为了热更新，框架使用起来比较繁琐，期待有人提出更好的优化方案
    * 为了热更新，框架使用大量全局变量、函数指针，会降低函数的执行效率
  * `gate` 的实现是直接使用net包实现的，为了异步接收数据，开启了一个goroutine，后续是优化方向


## Future
  * `db`添加缓存