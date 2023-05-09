// etcd
// @author LanguageY++2013 2023/5/8 23:49
// @company soulgame
package etcd

import (
	"context"
	"github.com/Languege/flexmatch/common/logger"
	"github.com/coreos/etcd/clientv3"
	"sync"
	"time"
)

const timeToLive int64 = 10

var (
	TimeToLive = timeToLive
)

type ServicePublisher struct {
	parentCtx   context.Context
	client      *clientv3.Client
	lease       clientv3.LeaseID
	leaseCancel func()
	done        chan struct{}
	doneOnce    sync.Once
	key         string
}

//NewServicePublisher 新建服务发布
func NewServicePublisher(ctx context.Context, client *clientv3.Client) (p *ServicePublisher, err error) {
	p = &ServicePublisher{
		client:    client,
		parentCtx: ctx,
		done:      make(chan struct{}),
	}

	err = p.KeepAlive()
	return
}

//KeepAlive 授权
func (p *ServicePublisher) KeepAlive() (err error) {
	resp, err := p.client.Grant(p.parentCtx, TimeToLive)
	if err != nil {
		return
	}

	p.lease = resp.ID

	return p.KeepAliveAsync()
}

func (p *ServicePublisher) KeepAliveAsync() (err error) {
	var (
		ctx context.Context
		ch  <-chan *clientv3.LeaseKeepAliveResponse
	)
	ctx, p.leaseCancel = context.WithCancel(p.parentCtx)
	ch, err = p.client.KeepAlive(ctx, p.lease)
	if err != nil {
		return
	}

	go func() {
		defer func() {
			if r := recover(); r != nil {
				logger.Errorf("%v", r)
			}
		}()

		for {
			select {
			case _, ok := <-ch:
				if !ok {
					p.revoke()
					logger.Infof("lease %d close due to keep alive chan interrupt", p.lease)
					if err := p.KeepAlive(); err != nil {
						logger.Errorf("KeepAlive err %v", err)
					}

					return
				}
			case <-p.done:
				p.revoke()
				return
			}
		}
	}()

	return
}

//revoke 撤销租约
func (p *ServicePublisher) revoke() (err error) {
	p.leaseCancel()
	time.Sleep(time.Second)
	_, err = p.client.Revoke(context.TODO(), p.lease)
	return
}

//Stop 停止
func(p *ServicePublisher) Stop() {
	p.doneOnce.Do(func() {
		close(p.done)
	})
}

func(p *ServicePublisher) Publish(sd *ServiceDescriptor) (err error){
	_, err = p.client.Put(p.parentCtx, sd.Key(), sd.Value(), clientv3.WithLease(p.lease))
	return
}