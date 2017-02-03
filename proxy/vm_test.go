// Copyright (c) 2016 Intel Corporation
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package main

import (
	"net"
	"os"
	"sync"
	"testing"
	"time"

	"github.com/containers/virtcontainers/hyperstart/mock"

	hyper "github.com/hyperhq/runv/hyperstart/api/json"
	"github.com/stretchr/testify/assert"
)

// vmRig maintains a test environment for vm objects
type vmRig struct {
	t  *testing.T
	wg sync.WaitGroup

	// hyperstart mocking
	Hyperstart      *mock.Hyperstart
	ctlPath, ioPath string

	// fd leak detection
	detector          *FdLeakDetector
	startFds, stopFds *FdSnapshot
}

func newVMRig(t *testing.T) *vmRig {
	return &vmRig{
		t: t,
	}
}

func (rig *vmRig) Start() {
	var err error

	rig.startFds, err = rig.detector.Snapshot()
	assert.Nil(rig.t, err)

	// Start hyperstart go routine
	rig.Hyperstart = mock.NewHyperstart(rig.t)
	rig.Hyperstart.Start()

	// Explicitly send READY message from hyperstart mock
	rig.wg.Add(1)
	go func() {
		rig.Hyperstart.SendMessage(int(hyper.INIT_READY), []byte{})
		rig.wg.Done()
	}()

}

func (rig *vmRig) Stop() {
	var err error

	rig.Hyperstart.Stop()

	rig.wg.Wait()

	// We shouldn't have leaked a fd between the beginning of Start() and
	// the end of Stop().
	rig.stopFds, err = rig.detector.Snapshot()
	assert.Nil(rig.t, err)

	assert.True(rig.t,
		rig.detector.Compare(os.Stdout, rig.startFds, rig.stopFds))
}

const testVM = "testVM"

// CreateVM creates a vm instance that is connected to the rig's Hyperstart
// mock object.
func (rig *vmRig) CreateVM() *vm {
	ctlSocketPath, ioSocketPath := rig.Hyperstart.GetSocketPaths()

	vm := newVM(testVM, ctlSocketPath, ioSocketPath)
	assert.NotNil(rig.t, vm)

	err := vm.Connect()
	assert.Nil(rig.t, err)

	return vm
}

// CreateIOConn returns a connection and an ioBase sequence number that can be
// used to send and receive IO data to/from hyperstart through the vm object.
// This function will allocate n consecutive sequence numbers.
func (rig *vmRig) CreateIOConn(vm *vm, n int) (net.Conn, uint64) {
	// this mocks a shim connected to the proxy, we'll give the proxy end
	// of the socketpair() to allocateIO while returning the shim end the
	// for the test code  to use.
	shimConn, proxyConn, err := Socketpair()
	assert.Nil(rig.t, err)

	seq := vm.AllocateIo(n, 1, proxyConn)

	return shimConn, seq
}

// TestShimGone tests what happens when a shim goes away and we still have I/O
// data we'd like to write to it. This happens when the shim is forcefully
// killed by SIGKILL, for instance (docker kill)
func TestShimGone(t *testing.T) {
	rig := NewVMRig(t)
	rig.Start()

	vm := rig.CreateVM()
	shim, seq := rig.CreateIOConn(vm, 1)

	// Simulate the shim going away.
	shim.Close()

	// Simulate the VM process sending data for a shim that is no longer there.
	rig.Hyperstart.SendIoString(seq, "foo\n")

	// Wait until the vm goroutine has read the IO data from hyperstart
	// sent in the line above.
	// Because that goroutine should cleanup vm.ioSessions on error, we
	// use that condition for the wait. There isn't really an easier to
	// synchronise here.
	for i := 0; i < 2000; i++ {
		vm.Lock()
		l := len(vm.ioSessions)
		vm.Unlock()
		if l == 0 {
			break
		}
		time.Sleep(1 * time.Millisecond)
	}

	// Close the qemu/hyperstart sockets by hand. We don't use vm.Close() because
	// that one will close the proxy end of the shim connection for us, which is
	// what we want to ensure the hyper->client goroutine does for us.
	vm.hyperHandler.CloseSockets()
	vm.wg.Wait()

	// If all goes well, the socket that was connected to the shim should
	// be closed and so the fd leak test in Stop() should pass

	rig.Stop()
}
