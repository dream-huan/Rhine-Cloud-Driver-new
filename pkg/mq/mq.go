package mq

import (
	"sync"
)

type Message struct {
	// 消息标识
	Event string
	// 消息内容
	Content interface{}
}

type MessageQueue interface {
	// 消息订阅
	Subscribe()
	// 取消订阅
	Unsubscribe()
	// 发布消息
	Publish()
	// 订阅回调函数
	SubscribeCallback()
}

type inMemoryMQ struct {
	Topics    map[string][]chan Message
	Callbacks map[string][]func(message Message)
	sync.RWMutex
}

var GlobalMQ = NewMQ()

func NewMQ() *inMemoryMQ {
	return &inMemoryMQ{
		Topics:    make(map[string][]chan Message),
		Callbacks: make(map[string][]func(Message)),
	}
}

func MessageSend(subscribeChan []chan Message, message Message) {
	for i := 0; i < len(subscribeChan); i++ {
		subscribeChan[i] <- message
	}
}

func (mq *inMemoryMQ) Publish(topic string, message Message) {
	mq.RLock()
	subscribeChan, isChannelExist := mq.Topics[topic]
	subscribeCallback, isCallbackExist := mq.Callbacks[topic]
	mq.RUnlock()
	if isChannelExist {
		go MessageSend(subscribeChan, message)
	}
	if isCallbackExist {
		for i := 0; i < len(subscribeCallback); i++ {
			go subscribeCallback[i](message)
		}
	}
}

func (mq *inMemoryMQ) Subscribe(topic string, bufferSize int64) <-chan Message {
	ch := make(chan Message, bufferSize)
	mq.Lock()
	defer mq.Unlock()
	mq.Topics[topic] = append(mq.Topics[topic], ch)
	return ch
}

func (mq *inMemoryMQ) Unsubscribe(topic string, ch <-chan Message) {
	mq.Lock()
	defer mq.Unlock()
	subscribes, isExist := mq.Topics[topic]
	if isExist {
		return
	}
	var newSubscribers []chan Message
	if len(subscribes)-1 != 0 {
		newSubscribers = make([]chan Message, len(subscribes)-1)
	} else {
		delete(mq.Topics, topic)
		return
	}
	for i := 0; i < len(subscribes); i++ {
		if ch != subscribes[i] {
			newSubscribers = append(newSubscribers, subscribes[i])
		}
	}
}

func (mq *inMemoryMQ) SubscribeCallback(topic string, callbackFunc func(Message)) {
	mq.Lock()
	defer mq.Unlock()
	mq.Callbacks[topic] = append(mq.Callbacks[topic], callbackFunc)
}

func (mq *inMemoryMQ) CheckStatus(topic string) bool {
	mq.RLock()
	defer mq.RUnlock()
	return mq.Topics[topic] != nil
}
