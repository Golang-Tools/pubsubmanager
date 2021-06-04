# pubsubmanager

这个项目的设计的用途是给sse和websocket维护推送数据.

本项目只是`github.com/kataras/go-events`的扩展,解决go-events的一部分问题

1. 只能注册回调函数不能注册事件的问题.因此造成发送消息必须在有注册回调函数后.这里引入字典来维护已经声明过的事件,只有声明过的事件(频道)才可以注册回调函数.而发送消息时如果频道未注册会自己注册频道.

2. 无法知道发送端是否结束发送的问题.此处在注册事件(频道)时它的value是一个信号channel,专门用于监听发送端是否结束发送.

3. 将注册的回调函数获取的结果放入channel,只是注册回调函数很多场景下难以使用,这边定义了`RegistListener(channelid string, size int) (<-chan interface{}, <-chan struct{}, CloseListenerFunc, error)`方法用于将回调函数中获取到的消息传出来让其他结构可以消费.
