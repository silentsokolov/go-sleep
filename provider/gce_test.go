package provider

import (
	"context"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"golang.org/x/oauth2/google"

	compute "google.golang.org/api/compute/v1"
)

var exampleGCEInstancesResponse = `
{
	"name": "test1",
	"networkInterfaces": [
	{
	"accessConfigs": [
			{
			"kind": "compute#accessConfig",
			"name": "External NAT",
			"natIP": "10.10.10.1",
			"type": "ONE_TO_ONE_NAT"
			}
		],
		"name": "nic0",
		"network": "default",
		"networkIP": "192.168.1.88"
	}
	],
	"status": "RUNNING"
	}
`

var exampleJWTFile = `
{
	"type": "service_account",
	"project_id": "my-project",
	"private_key_id": "1",
	"private_key": "-----BEGIN PRIVATE KEY-----\n-----END PRIVATE KEY-----\n",
	"client_email": "go-sleep@my-project.iam.gserviceaccount.com",
	"client_id": "1",
	"auth_uri": "https://accounts.google.com/o/oauth2/auth",
	"token_uri": "https://accounts.google.com/o/oauth2/token",
	"auth_provider_x509_cert_url": "https://www.googleapis.com/oauth2/v1/certs",
	"client_x509_cert_url": "https://www.googleapis.com/robot/v1/metadata/x509/go-sleep%40my-project.iam.gserviceaccount.com"
}
`

func initTestCGEServer(path string, resp string) *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(resp))
	}))
}

func initTestJWTFile() *os.File {
	content := []byte(exampleJWTFile)
	tmpfile, err := ioutil.TempFile("", "example")
	if err != nil {
		panic(err)
	}

	if _, err := tmpfile.Write(content); err != nil {
		panic(err)
	}

	if err := tmpfile.Close(); err != nil {
		panic(err)
	}

	return tmpfile
}

func TestNewGCE(t *testing.T) {
	tmpfile := initTestJWTFile()
	defer os.Remove(tmpfile.Name())

	projectID := "europe-west1-a"
	zone := "europe-west1-a"
	name := "instance-1"

	inst := NewGCE(tmpfile.Name(), projectID, zone, name, false)

	if inst.ProjectID != projectID {
		t.Errorf("NewGCE.ProjectID returned %+v, want %+v", inst.ProjectID, projectID)
	}

	if inst.Zone != zone {
		t.Errorf("NewGCE.Zone returned %+v, want %+v", inst.Zone, zone)
	}

	if inst.Name != name {
		t.Errorf("NewGCE.Name returned %+v, want %+v", inst.Name, name)
	}

	if inst.client == nil {
		t.Errorf("NewGCE.client not set")
	}

	if inst.computeService == nil {
		t.Errorf("NewGCE.computeService not set")
	}
}

func TestNormalizeGCEStatus(t *testing.T) {
	var statusTable = []struct {
		in  string
		out int
	}{
		{"PROVISIONING", StatusInstanceStarting},
		{"STAGING", StatusInstanceStarting},
		{"RUNNING", StatusInstanceRunning},
		{"STOPPING", StatusInstanceStopping},
		{"SUSPENDING", StatusInstanceStopping},
		{"SUSPENDED", StatusInstanceNotRun},
		{"TERMINATED", StatusInstanceNotRun},
		{"UNKNOWN", StatusInstanceNotAvailable},
	}

	for _, test := range statusTable {
		if s := normalizeGCEStatus(test.in); s != test.out {
			t.Errorf("normalizeGCEStatus is %v, want %v", s, test.out)
		}
	}
}

func TestGCE_String(t *testing.T) {
	inst := &GCE{
		ProjectID: "my-project",
		Zone:      "europe-west1-a",
		Name:      "instance-1",
	}
	s := "[GCE] Name: my-project-instance-1 in europe-west1-a"

	if s != inst.String() {
		t.Errorf("GCE.String returned %+v, want %+v", inst.String(), s)
	}
}

func TestGCE_Hash(t *testing.T) {
	inst := &GCE{
		ProjectID: "my-project",
		Zone:      "europe-west1-a",
		Name:      "instance-1",
	}
	s := "gce-my-project-europe-west1-a-instance-1"

	if s != inst.Hash() {
		t.Errorf("GCE.Hash returned %+v, want %+v", inst.Hash(), s)
	}
}

func TestGCE_GetStatus(t *testing.T) {
	server := initTestCGEServer("/", `{"status":"TERMINATED"}`)
	defer server.Close()

	if os.Getenv("TRAVIS") == "true" {
		t.Skip("no credentials on Travis")
	}

	ctx := context.TODO()
	client, err := google.DefaultClient(ctx, compute.ComputeScope)
	if err != nil {
		t.Fatal(err)
	}

	computeService, _ := compute.New(client)
	computeService.BasePath = server.URL

	inst := &GCE{
		ProjectID:      "my-project",
		Zone:           "europe-west1-a",
		Name:           "instance-1",
		computeService: computeService,
	}

	status, err := inst.GetStatus()
	if err != nil {
		t.Errorf("GCE.GetStatus returned unexpected error: %v", err)
	}

	if status != StatusInstanceNotRun {
		t.Errorf("GCE.GetStatus returned %+v, want %+v", status, StatusInstanceNotRun)
	}
}

func TestGCE_Start(t *testing.T) {
	server := initTestCGEServer("/", `{"status":"STAGING", "progress": 10}`)
	defer server.Close()

	if os.Getenv("TRAVIS") == "true" {
		t.Skip("no credentials on Travis")
	}

	ctx := context.TODO()
	client, err := google.DefaultClient(ctx, compute.ComputeScope)
	if err != nil {
		t.Fatal(err)
	}

	computeService, _ := compute.New(client)
	computeService.BasePath = server.URL

	inst := &GCE{
		ProjectID:      "my-project",
		Zone:           "europe-west1-a",
		Name:           "instance-1",
		computeService: computeService,
	}

	err = inst.Start()
	if err != nil {
		t.Errorf("GCE.Start returned unexpected error: %v", err)
	}
}

func TestGCE_Stop(t *testing.T) {
	server := initTestCGEServer("/", `{"status":"STOPPING", "progress": 10}`)
	defer server.Close()

	if os.Getenv("TRAVIS") == "true" {
		t.Skip("no credentials on Travis")
	}

	ctx := context.TODO()
	client, err := google.DefaultClient(ctx, compute.ComputeScope)
	if err != nil {
		t.Fatal(err)
	}

	computeService, _ := compute.New(client)
	computeService.BasePath = server.URL

	inst := &GCE{
		ProjectID:      "my-project",
		Zone:           "europe-west1-a",
		Name:           "instance-1",
		computeService: computeService,
	}

	err = inst.Stop()
	if err != nil {
		t.Errorf("GCE.Stop returned unexpected error: %v", err)
	}
}

func TestGCE_GetIP(t *testing.T) {
	server := initTestCGEServer("/", exampleGCEInstancesResponse)
	defer server.Close()

	if os.Getenv("TRAVIS") == "true" {
		t.Skip("no credentials on Travis")
	}

	ctx := context.TODO()
	client, err := google.DefaultClient(ctx, compute.ComputeScope)
	if err != nil {
		t.Fatal(err)
	}

	computeService, _ := compute.New(client)
	computeService.BasePath = server.URL

	inst := &GCE{
		ProjectID:      "my-project",
		Zone:           "europe-west1-a",
		Name:           "instance-1",
		computeService: computeService,
	}

	ip, err := inst.GetIP()
	if err != nil {
		t.Errorf("GCE.GetIP returned unexpected error: %v", err)
	}

	if ip != "10.10.10.1" {
		t.Errorf("GCE.GetIP returned %+v, want %+v", ip, "10.10.10.1")
	}
}

func TestGCE_GetIP_withInternalIP(t *testing.T) {
	server := initTestCGEServer("/", exampleGCEInstancesResponse)
	defer server.Close()

	if os.Getenv("TRAVIS") == "true" {
		t.Skip("no credentials on Travis")
	}

	ctx := context.TODO()
	client, err := google.DefaultClient(ctx, compute.ComputeScope)
	if err != nil {
		t.Fatal(err)
	}

	computeService, _ := compute.New(client)
	computeService.BasePath = server.URL

	inst := &GCE{
		ProjectID:      "my-project",
		Zone:           "europe-west1-a",
		Name:           "instance-1",
		UseInternalIP:  true,
		computeService: computeService,
	}

	ip, err := inst.GetIP()
	if err != nil {
		t.Errorf("GCE.GetIP (UseInternalIP) returned unexpected error: %v", err)
	}

	if ip != "192.168.1.88" {
		t.Errorf("GCE.GetIP (UseInternalIP) returned %+v, want %+v", ip, "192.168.1.88")
	}
}

func TestGetGoogleClient(t *testing.T) {
	tmpfile := initTestJWTFile()
	defer os.Remove(tmpfile.Name())

	_, err := getGoogleClient(tmpfile.Name(), compute.CloudPlatformScope, compute.ComputeScope)
	if err != nil {
		t.Errorf("getGoogleClient returned unexpected error: %v", err)
	}
}
