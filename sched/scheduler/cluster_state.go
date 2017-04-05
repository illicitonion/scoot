package scheduler

import (
	"time"

	log "github.com/Sirupsen/logrus"

	"github.com/scootdev/scoot/cloud/cluster"
)

const noTask = ""
const maxLostDuration = time.Minute  // after which we remove a node from the cluster entirely
const maxFlakyDuration = time.Minute // after which we mark it not flaky and put it back in rotation.
var nilTime = time.Time{}

// clusterState maintains a cluster of nodes and information about what task is running on each node.
// nodeGroups is for node affinity where we want to remember which node last ran with what snapshot.
// TODO(jschiller): we may prefer to assert that updateCh never blips on a node so we can remove lost node concept.
type clusterState struct {
	updateCh       chan []cluster.NodeUpdate
	nodes          map[cluster.NodeId]*nodeState // All healthy nodes.
	suspendedNodes map[cluster.NodeId]*nodeState // All lost or flaky nodes, disjoint from 'nodes'.
	nodeGroups     map[string]*nodeGroup         //key is a snapshotId.
}

type nodeGroup struct {
	idle map[cluster.NodeId]*nodeState
	busy map[cluster.NodeId]*nodeState
}

func newNodeGroup() *nodeGroup {
	return &nodeGroup{idle: map[cluster.NodeId]*nodeState{}, busy: map[cluster.NodeId]*nodeState{}}
}

// The State of A Node in the Cluster
type nodeState struct {
	node        cluster.Node
	runningTask string
	snapshotId  string
	timeLost    time.Time // Time when node was marked lost, if set.
	timeFlaky   time.Time // Time when node was marked flaky, if set.
}

// This node was either reported lost by a NodeUpdate and we keep it around for a bit in case it revives,
// or it experienced connection related errors so we sideline it for a little while.
func (ns *nodeState) suspended() bool {
	return ns.timeLost != nilTime || ns.timeFlaky != nilTime
}

// Initializes a Node State for the specified Node
func newNodeState(node cluster.Node) *nodeState {
	return &nodeState{
		node:        node,
		runningTask: noTask,
		snapshotId:  "",
		timeLost:    nilTime,
		timeFlaky:   nilTime,
	}
}

// Creates a New State Distributor with the initial nodes, and which updates
// nodes added or removed based on the supplied channel.
func newClusterState(initial []cluster.Node, updateCh chan []cluster.NodeUpdate) *clusterState {
	nodes := make(map[cluster.NodeId]*nodeState)
	nodeGroups := map[string]*nodeGroup{"": newNodeGroup()}
	for _, n := range initial {
		nodes[n.Id()] = newNodeState(n)
		nodeGroups[""].idle[n.Id()] = nodes[n.Id()]
	}

	return &clusterState{
		updateCh:       updateCh,
		nodes:          nodes,
		suspendedNodes: map[cluster.NodeId]*nodeState{},
		nodeGroups:     nodeGroups,
	}
}

// Update ClusterState to reflect that a task has been scheduled on a particular node
// SnapshotId should be the value from the task definition associated with the given taskId.
// NOTE: taskId is not unique (and isn't currently required to be), but a jobId arg would fix that.
func (c *clusterState) taskScheduled(nodeId cluster.NodeId, taskId string, snapshotId string) {
	ns := c.nodes[nodeId]

	delete(c.nodeGroups[ns.snapshotId].idle, nodeId)
	if _, ok := c.nodeGroups[snapshotId]; !ok {
		c.nodeGroups[snapshotId] = newNodeGroup()
	}
	c.nodeGroups[snapshotId].busy[nodeId] = ns

	ns.snapshotId = snapshotId
	ns.runningTask = taskId
}

// Update ClusterState to reflect that a task has finished running on
// a particular node, whether successfully or unsuccessfully
func (c *clusterState) taskCompleted(nodeId cluster.NodeId, taskId string, flaky bool) {
	var ns *nodeState
	var ok bool
	if ns, ok = c.nodes[nodeId]; !ok {
		// This node was removed from the cluster already, check if it was moved to suspendedNodes.
		ns, ok = c.suspendedNodes[nodeId]
	}
	if ok {
		if flaky && !ns.suspended() {
			c.suspendedNodes[nodeId] = ns
			ns.timeFlaky = time.Now()
		}
		ns.runningTask = noTask
		delete(c.nodeGroups[ns.snapshotId].busy, nodeId)
		c.nodeGroups[ns.snapshotId].idle[nodeId] = ns
	}
}

func (c *clusterState) getNodeState(nodeId cluster.NodeId) (*nodeState, bool) {
	ns, ok := c.nodes[nodeId]
	return ns, ok
}

// upate cluster state to reflect added and removed nodes
func (c *clusterState) updateCluster() {
	select {
	case updates, ok := <-c.updateCh:
		if !ok {
			c.updateCh = nil
		}
		c.update(updates)
	default:
	}
}

// Processes nodes being added and removed from the cluster & updates the distributor state accordingly.
// Note, we don't expect there to be many updates after startup if the cluster is relatively stable.
func (c *clusterState) update(updates []cluster.NodeUpdate) {
	// Apply updates
	for _, update := range updates {
		switch update.UpdateType {
		case cluster.NodeAdded:
			if ns, ok := c.suspendedNodes[update.Id]; ok {
				// This node was suspended earlier (assuming id is unique to a running node instance), we can recover it now.
				ns.timeLost = nilTime
				c.nodes[update.Id] = ns
				delete(c.suspendedNodes, update.Id)
				log.Infof("Recovered suspended node %v (%v), now have %d healthy nodes, %d suspended nodes",
					update.Id, ns, len(c.nodes), len(c.suspendedNodes))
			} else if ns, ok := c.nodes[update.Id]; !ok {
				// This is a new unrecognized node, add it to the cluster.
				c.nodes[update.Id] = newNodeState(update.Node)
				c.nodeGroups[""].idle[update.Id] = c.nodes[update.Id]
				log.Infof("Added new node: %v (%+v), now have %d nodes", update.Id, update.Node, len(c.nodes))
			} else {
				// This node is already present, log this spurious add.
				log.Infof("Node already added!! %v (%v)", update.Id, ns)
			}
		case cluster.NodeRemoved:
			if ns, ok := c.suspendedNodes[update.Id]; ok {
				// Node already suspended, make sure it's now marked as lost and not flaky.
				log.Infof("Node already marked suspended (was flaky=%t) : %v (%v)", ns.timeFlaky != nilTime, update.Id, ns)
				ns.timeLost = time.Now()
				ns.timeFlaky = nilTime
			} else if ns, ok := c.nodes[update.Id]; ok {
				// This was a healthy node, mark it as lost now.
				ns.timeLost = time.Now()
				c.suspendedNodes[update.Id] = ns
				delete(c.nodes, update.Id)
				log.Infof("Removing node by marking as lost: %v (%v), now have %d nodes, %d suspended",
					update.Id, ns, len(c.nodes), len(c.suspendedNodes))
			} else {
				// We don't know about this node, log spurious remove.
				log.Infof("Cannot remove unknown node: %v", update.Id)
			}
		}
	}

	// Clean up lost nodes that haven't recovered in time, and put flaky nodes back in rotation.
	now := time.Now()
	for _, ns := range c.suspendedNodes {
		if ns.timeLost != nilTime && now.Sub(ns.timeLost) > maxLostDuration {
			log.Infof("Deleting lost node: %v (%v), now have %d healthy, %d suspended",
				ns.node.Id(), ns, len(c.nodes), len(c.suspendedNodes)-1)
			delete(c.suspendedNodes, ns.node.Id())
			delete(c.nodeGroups[ns.snapshotId].idle, ns.node.Id())
			delete(c.nodeGroups[ns.snapshotId].busy, ns.node.Id())
		} else if ns.timeFlaky != nilTime && now.Sub(ns.timeFlaky) > maxFlakyDuration {
			log.Infof("Reinstating flaky node: %v (%v), now have %d healthy, %d suspended",
				ns.node.Id(), ns, len(c.nodes), len(c.suspendedNodes))
			delete(c.suspendedNodes, ns.node.Id())
			c.nodes[ns.node.Id()] = ns
			ns.timeFlaky = nilTime
		}
	}
}
