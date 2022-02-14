# 0.0.2

## 修改接口

1. 原来的`Publish`和`PublishWithDefault`接口改为`Send`和`SendWithDefault`接口

## 新增接口

+ 新增`Publish`和`PublishWithoutDefault`接口,含义为向当前所有注册中的chan(或者排除default外)发送消息

## 行为修改

+ 消息发布时对channel进行去重
+ 修改注释避免和swag冲突

# 0.0.1

项目创
