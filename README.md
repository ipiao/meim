#### meim
> 这并不是一个完整的im,甚至不是一个完整的程序。只是提供了socket服务的一些功能模块，实际应用中需要自己组装。
    
1. 固定模块 
- server
    - 提供了 MakeListener，可以注册不同的listener从而满足不同的服务要求
- client
    - 简单的client内容，以及相关client的读写操作
- plugin
    - server与client读写过程中的中间操作，在这里实现真正的业务逻辑
    
 2. 其他
 - router，middleware， example
    - 这些都是在socket业务中会使用到的一些基础工具，只是作为组件放在那里，若不适用，可自定义