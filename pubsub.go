package pubsubmanager

import (
	"errors"
	"sync"

	events "github.com/kataras/go-events"
	"github.com/scylladb/go-set/strset"
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
	defer psm.lock.RUnlock()
	_, ok := psm.channels[channelid]
	return ok
}

//AddChannel 将频道注册到系统中允许监听
func (psm *PubSubManager) AddChannel(channelid string) {
	if !psm.ChannelInUse(channelid) {
		psm.lock.Lock()
		defer psm.lock.Unlock()
		psm.channels[channelid] = make(chan struct{})
	}
}

//CloseChannel 关闭频道操作.会将频道id先去除不再允许注册监听器,然后将频道id下的所有监听器去除
func (psm *PubSubManager) CloseChannel(channelid string) {
	if psm.ChannelInUse(channelid) {
		close(psm.channels[channelid])
		psm.lock.Lock()
		defer psm.lock.Unlock()
		delete(psm.channels, channelid)
	}
	psm.emmiter.RemoveAllListeners(events.EventName(channelid))
}

//Channels 查看现在被注册的频道有哪些
func (psm *PubSubManager) Channels() []string {
	result := []string{}
	psm.lock.RLock()
	defer psm.lock.RUnlock()
	for k := range psm.channels {
		result = append(result, k)
	}

	return result
}

//Send 指定频道发送消息
//@params msg interface{} 发送的消息
//@params channels ...string 发送去的频道
func (psm *PubSubManager) Send(msg interface{}, channels ...string) {
	chaset := strset.New(channels...)
	for _, channelid := range chaset.List() {
		if !psm.ChannelInUse(channelid) {
			psm.AddChannel(channelid)
		}
		e := events.EventName(channelid)
		psm.emmiter.Emit(e, msg)
	}
}

//SendWithDefault 指定频道发送消息,消息一并发送到default这个channel上
//@params msg interface{} 发送的消息
//@params channels ...string 发送去的频道
func (psm *PubSubManager) SendWithDefault(msg interface{}, channels ...string) {
	channels = append(channels, "default")
	psm.Send(msg, channels...)

}

//Publish 全频道广播消息
//@params msg interface{} 广播的消息
func (psm *PubSubManager) Publish(msg interface{}) {
	channels := []string{}
	psm.lock.RLock()
	defer psm.lock.RUnlock()
	for k := range psm.channels {
		channels = append(channels, k)
	}
	psm.Send(msg, channels...)
}

//PublishWithoutDefault 除了default这个channel外全频道广播
//@params msg interface{} 广播的消息
func (psm *PubSubManager) PublishWithoutDefault(msg interface{}) {
	channels := []string{}
	psm.lock.RLock()
	defer psm.lock.RUnlock()
	for k := range psm.channels {
		if k != "default" {
			channels = append(channels, k)
		}
	}
	psm.Send(msg, channels...)
}

type CloseListenerFunc func()

//RegistListener 注册频道的监听器
//@params channelid string 频道id
//@params size int 构造的消息channel的缓冲区大小,<=0时使用无缓冲channel
//@returns <-chan interface{} 消息channel
//@returns <-chan struct{} 关闭监听器的信号channel
//@returns CloseListenerFunc 关闭当前注册监听器的函数
//@returns error 当频道未被注册可以使用时会报错
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
//@params channelid string 频道id
//@returns <-chan struct{} 关闭频道所有监听器的信号channel
//@returns bool 当前频道是否可用
func (psm *PubSubManager) CloseNotify(channelid string) (ch <-chan struct{}, ok bool) {
	psm.lock.RLock()
	defer psm.lock.RUnlock()
	ch, ok = psm.channels[channelid]
	return
}

var PubSub = NewPubSubManager()
