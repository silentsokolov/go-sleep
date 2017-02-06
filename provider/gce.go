package provider

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"

	compute "google.golang.org/api/compute/v1"
)

// GCE ..
type GCE struct {
	JWTPath        string
	ProjectID      string
	Zone           string
	Name           string
	UseInternalIP  bool
	client         *http.Client
	computeService *compute.Service
}

// NewGCE ..
func NewGCE(JWTPath, ProjectID, Zone, Name string, UseInternalIP bool) *GCE {
	client, err := getGoogleClient(JWTPath, compute.CloudPlatformScope, compute.ComputeScope)
	if err != nil {
		log.Fatalf("GCE  %s: Unable to create HTTP client: %v", Name, err)
	}

	computeService, err := compute.New(client)
	if err != nil {
		log.Fatalf("GCE  %s: Unable to create Compute service: %v", Name, err)
	}

	return &GCE{
		JWTPath:        JWTPath,
		ProjectID:      ProjectID,
		Zone:           Zone,
		Name:           Name,
		UseInternalIP:  UseInternalIP,
		client:         client,
		computeService: computeService,
	}
}

// String ...
func (p *GCE) String() string {
	return fmt.Sprintf("[GCE] Name: %s-%s in %s", p.ProjectID, p.Name, p.Zone)
}

// Hash ...
func (p *GCE) Hash() string {
	return fmt.Sprintf("gce-%s-%s-%s", p.ProjectID, p.Zone, p.Name)
}

// GetStatus ...
func (p *GCE) GetStatus() (int, error) {
	inst, err := p.computeService.Instances.Get(p.ProjectID, p.Zone, p.Name).Do()
	if err != nil {
		return StatusInstanceNotAvailable, err
	}

	return normalizeGCEStatus(inst.Status), nil
}

// GetIP ...
func (p *GCE) GetIP() (string, error) {
	inst, err := p.computeService.Instances.Get(p.ProjectID, p.Zone, p.Name).Do()
	if err != nil {
		return "", err
	}
	if p.UseInternalIP {
		return inst.NetworkInterfaces[0].NetworkIP, nil
	}

	return inst.NetworkInterfaces[0].AccessConfigs[0].NatIP, nil
}

// Start ...
func (p *GCE) Start() error {
	_, err := p.computeService.Instances.Start(p.ProjectID, p.Zone, p.Name).Do()
	if err != nil {
		return err
	}
	return nil
}

// Stop ...
func (p *GCE) Stop() error {
	_, err := p.computeService.Instances.Stop(p.ProjectID, p.Zone, p.Name).Do()
	if err != nil {
		return err
	}
	return nil
}

func normalizeGCEStatus(originalStatus string) int {
	switch originalStatus {
	case "PROVISIONING":
		return StatusInstanceStarting
	case "STAGING":
		return StatusInstanceStarting
	case "RUNNING":
		return StatusInstanceRunning
	case "STOPPING":
		return StatusInstanceStopping
	case "SUSPENDING":
		return StatusInstanceStopping
	case "SUSPENDED":
		return StatusInstanceNotRun
	case "TERMINATED":
		return StatusInstanceNotRun
	default:
		return StatusInstanceNotAvailable
	}
}

func getGoogleClient(JWTpath string, scope ...string) (*http.Client, error) {
	data, err := ioutil.ReadFile(JWTpath)

	if err != nil {
		return nil, err
	}

	conf, err := google.JWTConfigFromJSON(data, scope...)
	if err != nil {
		return nil, err
	}

	return conf.Client(oauth2.NoContext), nil
}
