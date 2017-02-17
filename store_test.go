package main

import (
	"errors"
	"fmt"
	"reflect"
	"testing"
	"time"

	"github.com/silentsokolov/go-sleep/provider"
)

type dummyProvider struct {
	DummyID       string
	UseInternalIP bool
}

func newDummyProvider(DummyID string, UseInternalIP bool) *dummyProvider {
	return &dummyProvider{
		DummyID: DummyID,
	}
}

func (p *dummyProvider) String() string {
	return fmt.Sprintf("[dummyProvider] ID: %s", p.DummyID)
}

func (p *dummyProvider) Hash() string {
	return fmt.Sprintf("dummy-%s", p.DummyID)
}

func (p *dummyProvider) GetStatus() (int, error) {
	return provider.StatusInstanceRunning, nil
}

func (p *dummyProvider) GetIP() (string, error) {
	return "www.example.org", nil
}

func (p *dummyProvider) Start() error {
	return nil
}

func (p *dummyProvider) Stop() error {
	return nil
}

func TestInstanceStore_Set(t *testing.T) {
	instance := &ComputeInstance{}
	store := NewInstanceStore()

	store.Set("key", instance)

	if inst, ok := store.values["key"]; !ok {
		if !reflect.DeepEqual(inst, instance) {
			t.Errorf("InstanceStore.Set unexpected instance %+v", inst)
		}
		t.Error("InstanceStore.Set not add instance to store")
	}
}

func TestInstanceStore_Get(t *testing.T) {
	instance := &ComputeInstance{}
	store := NewInstanceStore()

	store.Set("key", instance)

	if inst, ok := store.Get("key"); !ok {
		if !reflect.DeepEqual(inst, instance) {
			t.Errorf("InstanceStore.Get return unexpected instance %+v", inst)
		}
		t.Error("InstanceStore.Get return nothing")
	}
}

func TestComputeInstance_String(t *testing.T) {
	p := newDummyProvider("test", false)
	ci := NewComputeInstance(p, time.Duration(100)*time.Second)

	s := "Instance: [dummyProvider] ID: test, current status: 4"

	if ci.String() != s {
		t.Errorf("ComputeInstance.String returned %+v, want %+v", ci.String(), s)
	}
}

func TestComputeInstance_Hash(t *testing.T) {
	p := newDummyProvider("test", false)
	ci := NewComputeInstance(p, time.Duration(100)*time.Second)

	s := "dummy-test"

	if ci.Hash() != s {
		t.Errorf("ComputeInstance.Hash returned %+v, want %+v", ci.Hash(), s)
	}
}

func TestComputeInstance_GetStatus(t *testing.T) {
	p := newDummyProvider("test", false)
	ci := NewComputeInstance(p, time.Duration(100)*time.Second)

	providerStatus, _ := p.GetStatus()
	if ci.GetStatus() != providerStatus {
		t.Errorf("ComputeInstance.GetStatus returned %+v, want %+v", ci.GetStatus(), providerStatus)
	}
}

func TestComputeInstance_SetStatus(t *testing.T) {
	p := newDummyProvider("test", false)
	ci := NewComputeInstance(p, time.Duration(100)*time.Second)
	providerStatus, _ := p.GetStatus()

	if ci.currentStatus != providerStatus {
		t.Error("ComputeInstance.SetStatus init status not correct")
	}

	ci.SetStatus(provider.StatusInstanceNotRun)
	if ci.GetStatus() != provider.StatusInstanceNotRun {
		t.Errorf("ComputeInstance.SetStatus setted %+v, want %+v", ci.GetStatus(), provider.StatusInstanceNotRun)
	}
}

func TestComputeInstance_SetError(t *testing.T) {
	p := newDummyProvider("test", false)
	ci := NewComputeInstance(p, time.Duration(100)*time.Second)
	err := errors.New("test error")

	if ci.lastError != nil {
		t.Error("ComputeInstance.lastError not correct")
	}

	ci.SetError(err)
	if ci.lastError != err {
		t.Errorf("ComputeInstance.lastError setted %+v, want %+v", ci.lastError, err)
	}
}

func TestComputeInstance_Reset(t *testing.T) {
	p := newDummyProvider("test", false)
	ci := NewComputeInstance(p, time.Duration(100)*time.Second)
	ci.lastError = errors.New("test error")
	ci.IP = "127.0.0.1"
	ci.lastAccess = time.Now()
	ci.startRequest = time.Now()

	ci.Reset()

	if ci.lastError != nil {
		t.Error("ComputeInstance.Reset not clear lastError")
	}

	if ci.IP != "" {
		t.Error("ComputeInstance.Reset not clear IP")
	}

	if !ci.lastAccess.IsZero() {
		t.Error("ComputeInstance.Reset not clear lastAccess")
	}

	if !ci.startRequest.IsZero() {
		t.Error("ComputeInstance.Reset not clear startRequest")
	}
}
