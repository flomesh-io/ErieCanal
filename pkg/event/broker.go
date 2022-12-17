/*
 * Copyright 2022 The flomesh.io Authors.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package event

import (
	"github.com/cskr/pubsub"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/util/workqueue"
	"k8s.io/klog/v2"
	"time"
)

type Broker struct {
	queue      workqueue.RateLimitingInterface
	messageBus *pubsub.PubSub
}

func NewBroker(stopCh <-chan struct{}) *Broker {
	b := &Broker{
		queue:      workqueue.NewRateLimitingQueue(workqueue.DefaultControllerRateLimiter()),
		messageBus: pubsub.New(0),
	}

	go b.runWorkqueueProcessor(stopCh)

	return b
}

func (b *Broker) GetMessageBus() *pubsub.PubSub {
	return b.messageBus
}

//func (b *Broker) GetQueue() workqueue.RateLimitingInterface {
//	return b.queue
//}

func (b *Broker) Enqueue(msg Message) {
	//msg, ok := obj.(Message)
	//if !ok {
	//	klog.Errorf("Received unexpected message %T, expected event.Message", obj)
	//	return
	//}

	b.queue.AddRateLimited(msg)
}

func (b *Broker) Unsub(pubSub *pubsub.PubSub, ch chan interface{}) {
	go pubSub.Unsub(ch)
	for range ch {
		// Drain channel until 'Unsub' results in a close on the subscribed channel
	}
}

func (b *Broker) runWorkqueueProcessor(stopCh <-chan struct{}) {
	go wait.Until(
		func() {
			for b.processNextItem() {
			}
		},
		time.Second,
		stopCh,
	)
}

func (b *Broker) processNextItem() bool {
	item, shutdown := b.queue.Get()
	if shutdown {
		return false
	}

	defer b.queue.Done(item)

	msg, ok := item.(Message)
	if !ok {
		b.queue.Forget(item)
		// Process next item in the queue
		return true
	}

	b.processEvent(msg)
	b.queue.Forget(item)

	return true
}

func (b *Broker) processEvent(msg Message) {
	klog.V(5).Infof("Processing event %v", msg)
	b.messageBus.Pub(msg, string(msg.Kind))
}
