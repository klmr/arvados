// Copyright (C) The Arvados Authors. All rights reserved.
//
// SPDX-License-Identifier: AGPL-3.0

package worker

import (
	"io"
	"time"

	"git.curoverse.com/arvados.git/lib/cloud"
	"git.curoverse.com/arvados.git/lib/dispatchcloud/test"
	"git.curoverse.com/arvados.git/sdk/go/arvados"
	"github.com/sirupsen/logrus"
	check "gopkg.in/check.v1"
)

const GiB arvados.ByteSize = 1 << 30

var _ = check.Suite(&PoolSuite{})

type lessChecker struct {
	*check.CheckerInfo
}

func (*lessChecker) Check(params []interface{}, names []string) (result bool, error string) {
	return params[0].(int) < params[1].(int), ""
}

var less = &lessChecker{&check.CheckerInfo{Name: "less", Params: []string{"obtained", "expected"}}}

type PoolSuite struct{}

func (suite *PoolSuite) SetUpSuite(c *check.C) {
	logrus.StandardLogger().SetLevel(logrus.DebugLevel)
}

func (suite *PoolSuite) TestStartContainer(c *check.C) {
	// TODO: use an instanceSet stub with an SSH server
}

func (suite *PoolSuite) TestVerifyHostKey(c *check.C) {
	// TODO: use an instanceSet stub with an SSH server
}

func (suite *PoolSuite) TestCreateUnallocShutdown(c *check.C) {
	lameInstanceSet := &test.LameInstanceSet{Hold: make(chan bool)}
	type1 := arvados.InstanceType{Name: "a1s", ProviderType: "a1.small", VCPUs: 1, RAM: 1 * GiB, Price: .01}
	type2 := arvados.InstanceType{Name: "a2m", ProviderType: "a2.medium", VCPUs: 2, RAM: 2 * GiB, Price: .02}
	type3 := arvados.InstanceType{Name: "a2l", ProviderType: "a2.large", VCPUs: 4, RAM: 4 * GiB, Price: .04}
	pool := &Pool{
		logger:      logrus.StandardLogger(),
		newExecutor: func(cloud.Instance) Executor { return &stubExecutor{} },
		instanceSet: lameInstanceSet,
		instanceTypes: arvados.InstanceTypeMap{
			type1.Name: type1,
			type2.Name: type2,
			type3.Name: type3,
		},
	}
	notify := pool.Subscribe()
	defer pool.Unsubscribe(notify)
	notify2 := pool.Subscribe()
	defer pool.Unsubscribe(notify2)

	c.Check(pool.Unallocated()[type1], check.Equals, 0)
	c.Check(pool.Unallocated()[type2], check.Equals, 0)
	c.Check(pool.Unallocated()[type3], check.Equals, 0)
	pool.Create(type2)
	pool.Create(type1)
	pool.Create(type2)
	pool.Create(type3)
	c.Check(pool.Unallocated()[type1], check.Equals, 1)
	c.Check(pool.Unallocated()[type2], check.Equals, 2)
	c.Check(pool.Unallocated()[type3], check.Equals, 1)

	// Unblock the pending Create calls.
	go lameInstanceSet.Release(4)

	// Wait for each instance to either return from its Create
	// call, or show up in a poll.
	suite.wait(c, pool, notify, func() bool {
		pool.mtx.RLock()
		defer pool.mtx.RUnlock()
		return len(pool.workers) == 4
	})

	// Place type3 node on admin-hold
	for _, instv := range pool.Instances() {
		if instv.ArvadosInstanceType == type3.Name {
			pool.SetIdleBehavior(instv.Instance, IdleBehaviorHold)
			break
		}
	}
	c.Check(pool.Shutdown(type3), check.Equals, false)
	suite.wait(c, pool, notify, func() bool {
		return pool.Unallocated()[type3] == 0
	})

	// Shutdown both type2 nodes
	c.Check(pool.Shutdown(type2), check.Equals, true)
	suite.wait(c, pool, notify, func() bool {
		return pool.Unallocated()[type1] == 1 && pool.Unallocated()[type2] == 1
	})
	c.Check(pool.Shutdown(type2), check.Equals, true)
	suite.wait(c, pool, notify, func() bool {
		return pool.Unallocated()[type1] == 1 && pool.Unallocated()[type2] == 0
	})
	c.Check(pool.Shutdown(type2), check.Equals, false)
	for {
		// Consume any waiting notifications to ensure the
		// next one we get is from Shutdown.
		select {
		case <-notify:
			continue
		default:
		}
		break
	}

	// Shutdown type1 node
	c.Check(pool.Shutdown(type1), check.Equals, true)
	suite.wait(c, pool, notify, func() bool {
		return pool.Unallocated()[type1] == 0 && pool.Unallocated()[type2] == 0 && pool.Unallocated()[type3] == 0
	})
	select {
	case <-notify2:
	case <-time.After(time.Second):
		c.Error("notify did not receive")
	}

	// Place type3 node on admin-drain so it shuts down right away
	for _, instv := range pool.Instances() {
		if instv.ArvadosInstanceType == type3.Name {
			pool.SetIdleBehavior(instv.Instance, IdleBehaviorDrain)
			break
		}
	}
	suite.wait(c, pool, notify, func() bool {
		return pool.Unallocated()[type3] == 0
	})

	go lameInstanceSet.Release(4) // unblock Destroy calls

	suite.wait(c, pool, notify, func() bool {
		pool.getInstancesAndSync()
		return len(pool.Instances()) == 0
	})
}

func (suite *PoolSuite) wait(c *check.C, pool *Pool, notify <-chan struct{}, ready func() bool) {
	timeout := time.NewTimer(time.Second).C
	for !ready() {
		select {
		case <-notify:
			continue
		case <-timeout:
		}
		break
	}
	c.Check(ready(), check.Equals, true)
}

type stubExecutor struct{}

func (*stubExecutor) SetTarget(cloud.ExecutorTarget) {}

func (*stubExecutor) Execute(env map[string]string, cmd string, stdin io.Reader) ([]byte, []byte, error) {
	return nil, nil, nil
}

func (*stubExecutor) Close() {}
