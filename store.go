package main

import (
	"fmt"
	"sync"
	"time"

	"github.com/silentsokolov/go-sleep/log"
	"github.com/silentsokolov/go-sleep/provider"
)

// ComputeInstance ...
type ComputeInstance struct {
	sync.RWMutex
	currentStatus provider.StatusInstance
	sleepAfter    time.Duration
	Provider      provider.Provider
	IP            string
	statusChan    chan provider.StatusInstance
	stopChan      chan bool
	lastAccess    time.Time
	lastError     error
	HTTPHealth    bool
	startRequest  time.Time
}

// NewComputeInstance ...
func NewComputeInstance(p provider.Provider, sleepAfter time.Duration) *ComputeInstance {
	status, err := p.Status()
	if err != nil {
		log.Fatal(err)
	}

	instance := &ComputeInstance{
		currentStatus: status,
		sleepAfter:    sleepAfter,
		Provider:      p,
		statusChan:    make(chan provider.StatusInstance, 5),
		stopChan:      make(chan bool),
	}

	if status == provider.StatusInstanceRunning {
		if instance.IP, err = p.IP(); err != nil {
			log.Fatal(err)
		}
		instance.SetLastAccess()
		instance.SetHTTPHealth()
	}

	return instance
}

func (instance *ComputeInstance) String() string {
	return fmt.Sprintf("Instance: %s, current status: %s", instance.Provider.String(), instance.currentStatus)
}

// Hash ...
func (instance *ComputeInstance) Hash() string {
	return instance.Provider.Hash()
}

func (instance *ComputeInstance) startMonitor(wg *sync.WaitGroup) {
	wg.Add(1)

	go func() {
		defer wg.Done()
		for {
			select {
			case status := <-instance.statusChan:
				log.Printf("Change status for %s", instance.Provider)

				if instance.Status() != status {
					switch status {
					case provider.StatusInstanceStarting:
						log.Printf("Starting %s", instance)
						if err := instance.Provider.Start(); err != nil {
							instance.SetError(err)
							instance.SetStatus(provider.StatusInstanceError)
						} else {
							instance.startRequest = time.Now()
							instance.SetStatus(provider.StatusInstanceStarting)
						}
					case provider.StatusInstanceStopping:
						log.Printf("Stopping %s", instance)
						if err := instance.Provider.Stop(); err != nil {
							log.Printf("Stopping %s raise error: %s", instance, err)
						} else {
							instance.SetStatus(provider.StatusInstanceStopping)
						}
					}
				}
			case <-time.After(1 * time.Minute):
				log.Printf("Check status for %s", instance.Provider)
				providerStatus, err := instance.Provider.Status()

				if err != nil {
					log.Printf("Get status %s raise error: %s", instance, err)
					break
				}

				if providerStatus != instance.currentStatus {
					switch providerStatus {
					case provider.StatusInstanceRunning:
						if instance.IP, err = instance.Provider.IP(); err != nil {
							instance.SetError(err)
							instance.SetStatus(provider.StatusInstanceError)
							break
						} else {
							instance.SetLastAccess()
						}
					case provider.StatusInstanceNotRun:
						instance.Reset()
					}

					instance.SetStatus(providerStatus)
				} else if !instance.lastAccess.IsZero() && providerStatus == provider.StatusInstanceRunning {
					duration := time.Since(instance.lastAccess)
					if duration.Seconds() >= instance.sleepAfter.Seconds() {
						instance.Stop()
					}
				}
			case <-instance.stopChan:
				log.Printf("Stop monitor %s", instance.Provider)
				return
			}
		}
	}()
}

func (instance *ComputeInstance) stopMonitor() {
	go func() {
		instance.stopChan <- true
	}()
}

// Status ...
func (instance *ComputeInstance) Status() provider.StatusInstance {
	instance.RLock()
	defer instance.RUnlock()
	return instance.currentStatus
}

// SetStatus ...
func (instance *ComputeInstance) SetStatus(s provider.StatusInstance) {
	instance.Lock()
	defer instance.Unlock()
	instance.currentStatus = s
}

// SetHTTPHealth ...
func (instance *ComputeInstance) SetHTTPHealth() {
	instance.Lock()
	defer instance.Unlock()
	instance.HTTPHealth = true
}

// SetLastAccess ...
func (instance *ComputeInstance) SetLastAccess() {
	instance.Lock()
	defer instance.Unlock()
	instance.lastAccess = time.Now()
}

// SetError ...
func (instance *ComputeInstance) SetError(err error) {
	instance.Lock()
	defer instance.Unlock()
	instance.lastError = err
}

// Reset ...
func (instance *ComputeInstance) Reset() {
	instance.Lock()
	defer instance.Unlock()
	instance.IP = ""
	instance.lastAccess = time.Time{}
	instance.lastError = nil
	instance.startRequest = time.Time{}
	instance.HTTPHealth = false
}

// Start ...
func (instance *ComputeInstance) Start() {
	instance.statusChan <- provider.StatusInstanceStarting
}

// Stop ...
func (instance *ComputeInstance) Stop() {
	instance.statusChan <- provider.StatusInstanceStopping
}

// InstanceStore ...
type InstanceStore struct {
	sync.RWMutex
	wg     *sync.WaitGroup
	values map[string]*ComputeInstance
}

// NewInstanceStore ...
func NewInstanceStore() *InstanceStore {
	return &InstanceStore{
		wg:     &sync.WaitGroup{},
		values: make(map[string]*ComputeInstance),
	}
}

// Set ...
func (store *InstanceStore) Set(key string, instance *ComputeInstance) error {
	store.Lock()
	defer store.Unlock()

	store.values[key] = instance
	instance.startMonitor(store.wg)

	return nil
}

// Get ...
func (store *InstanceStore) Get(k string) (*ComputeInstance, bool) {
	store.RLock()
	defer store.RUnlock()

	if instance, ok := store.values[k]; ok {
		return instance, true
	}
	return nil, false
}

// Close ...
func (store *InstanceStore) Close() {
	for _, i := range store.values {
		i.stopMonitor()
	}
	store.wg.Wait()
}
