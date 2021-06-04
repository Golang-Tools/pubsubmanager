package pubsubmanager

import (
	"errors"
	"sync"

	events "github.com/kataras/go-events"
)

//PubSubManager 广播器
type PubSubManager struct {
	channels map[string]chan struct{}
	lock     *sync.RWMutex
	emmiter  events.EventEmmiter
}

//NewPubSubManager 创建广播器
func NewPubSubManager() *PubSubManager {
	psm := new(PubSubManager)
	psm.emmiter = events.New()
	psm.channels = map[string]chan struct{}{
		"default": make(chan struct{}),
	}
	psm.lock = &sync.RWMutex{}
	return psm
}

//ChannelInUse 检测channeldid是否在被使用
func (psm *PubSubManager) ChannelInUse(channelid string) bool {
	psm.lock.RLock()
	_, ok := psm.channels[channelid]
	psm.lock.RUnlock()
	return ok
}

//AddChannel 将频道注册到系统中允许监听
func (psm *PubSubManager) AddChannel(channelid string) {
	if !psm.ChannelInUse(channelid) {
		psm.lock.Lock()
		psm.channels[channelid] = make(chan struct{})
		psm.lock.Unlock()
	}
}

//CloseChannel 关闭频道操作.会将频道id先去除不再允许注册监听器,然后将频道id下的所有监听器去除
func (psm *PubSubManager) CloseChannel(channelid string) {
	if psm.ChannelInUse(channelid) {
		close(psm.channels[channelid])
		psm.lock.Lock()
		delete(psm.channels, channelid)
		psm.lock.Unlock()
	}
	psm.emmiter.RemoveAllListeners(events.EventName(channelid))
}

//Channels 查看现在被注册的频道有哪些
func (psm *PubSubManager) Channels() []string {
	result := []string{}
	psm.lock.RLock()
	for k, _ := range psm.channels {
		result = append(result, k)
	}
	psm.lock.RUnlock()
	return result
}

//Publish 指定频道广播消息
//@Param msg interface{} 广播的消息
//@Param channels ...string 广播的频道
func (psm *PubSubManager) Publish(msg interface{}, channels ...string) {
	for _, channelid := range channels {
		if !psm.ChannelInUse(channelid) {
			psm.AddChannel(channelid)
		}
		e := events.EventName(channelid)
		psm.emmiter.Emit(e, msg)
	}
}

//PublishWithDefault 指定频道广播消息,消息一并广播到default这个channel上
//@Param msg interface{} 广播的消息
//@Param channels ...string 广播的频道
func (psm *PubSubManager) PublishWithDefault(msg interface{}, channels ...string) {
	channels = append(channels, "default")
	psm.Publish(msg, channels...)
}

type CloseListenerFunc func()

//RegistListener 注册频道的监听器
//@Param channelid string 频道id
//@Param size int 构造的消息channel的缓冲区大小,<=0时使用无缓冲channel
//@Returns <-chan interface{} 消息channel
//@Returns <-chan struct{} 关闭监听器的信号channel
//@Returns CloseListenerFunc 关闭当前注册监听器的函数
//@Returns error 当频道未被注册可以使用时会报错
func (psm *PubSubManager) RegistListener(channelid string, size int) (<-chan interface{}, <-chan struct{}, CloseListenerFunc, error) {
	if !psm.ChannelInUse(channelid) {
		return nil, nil, nil, errors.New("channel not in use now")
	}
	var ssech chan interface{}
	if size > 0 {
		ssech = make(chan interface{}, size)
	} else {
		ssech = make(chan interface{})
	}
	closech := make(chan struct{})
	e := events.EventName(channelid)
	l := events.Listener(func(payload ...interface{}) {
		for _, pl := range payload {
			ssech <- pl
		}
	})
	psm.emmiter.AddListener(e, l)
	return ssech, closech, func() {
		psm.emmiter.RemoveListener(e, l)
		close(closech)
	}, nil
}

//CloseNotify 关闭指定频道的信号提醒
//@Param channelid string 频道id
//@Returns <-chan struct{} 关闭频道所有监听器的信号channel
//@Returns bool 当前频道是否可用
func (psm *PubSubManager) CloseNotify(channelid string) (ch <-chan struct{}, ok bool) {
	psm.lock.RLock()
	ch, ok = psm.channels[channelid]
	psm.lock.RUnlock()
	return
}

var PubSub = NewPubSubManager()
