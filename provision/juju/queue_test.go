// Copyright 2013 tsuru authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package juju

import (
	"github.com/globocom/commandmocker"
	"github.com/globocom/tsuru/queue"
	"github.com/globocom/tsuru/testing"
	. "launchpad.net/gocheck"
	"sort"
	"strings"
)

func (s *ELBSuite) TestHandleMessageWithoutUnits(c *C) {
	instIds := make([]string, 3)
	for i := 0; i < len(instIds); i++ {
		id := s.server.NewInstance()
		defer s.server.RemoveInstance(id)
		instIds[i] = id
	}
	replace := []string{"i-00004444", "i-00004445", "i-00004450"}
	output := simpleCollectOutput
	for i, r := range replace {
		output = strings.Replace(output, r, instIds[i], 1)
	}
	tmpdir, err := commandmocker.Add("juju", output)
	c.Assert(err, IsNil)
	defer commandmocker.Remove(tmpdir)
	app := testing.NewFakeApp("symfonia", "python", 1)
	manager := ELBManager{}
	err = manager.Create(app)
	c.Assert(err, IsNil)
	defer manager.Destroy(app)
	handle(&queue.Message{
		Action: addUnitToLoadBalancer,
		Args:   []string{"symfonia"},
	})
	resp, err := s.client.DescribeLoadBalancers(app.GetName())
	c.Assert(err, IsNil)
	c.Assert(resp.LoadBalancerDescriptions, HasLen, 1)
	instances := resp.LoadBalancerDescriptions[0].Instances
	c.Assert(instances, HasLen, 3)
	ids := []string{instances[0].InstanceId, instances[1].InstanceId, instances[2].InstanceId}
	sort.Strings(ids)
	c.Assert(ids, DeepEquals, instIds)
}

func (s *ELBSuite) TestHandleMessageWithUnits(c *C) {
	id1 := s.server.NewInstance()
	id2 := s.server.NewInstance()
	defer s.server.RemoveInstance(id1)
	defer s.server.RemoveInstance(id2)
	app := testing.NewFakeApp("symfonia", "python", 1)
	manager := ELBManager{}
	err := manager.Create(app)
	c.Assert(err, IsNil)
	defer manager.Destroy(app)
	output := strings.Replace(simpleCollectOutput, "i-00004444", id1, -1)
	output = strings.Replace(output, "i-00004445", id2, -1)
	tmpdir, err := commandmocker.Add("juju", output)
	c.Assert(err, IsNil)
	defer commandmocker.Remove(tmpdir)
	msg := queue.Message{
		Action: addUnitToLoadBalancer,
		Args:   []string{"symfonia", "symfonia/0", "symfonia/1"},
	}
	handle(&msg)
	resp, err := s.client.DescribeLoadBalancers(app.GetName())
	c.Assert(err, IsNil)
	c.Assert(resp.LoadBalancerDescriptions, HasLen, 1)
	instances := resp.LoadBalancerDescriptions[0].Instances
	c.Assert(instances, HasLen, 2)
	ids := []string{instances[0].InstanceId, instances[1].InstanceId}
	sort.Strings(ids)
	c.Assert(ids, DeepEquals, []string{id1, id2})
	c.Assert(commandmocker.Ran(tmpdir), Equals, true)
}

func (s *ELBSuite) TestHandler(c *C) {
	var _ queue.Handler = handler
}
