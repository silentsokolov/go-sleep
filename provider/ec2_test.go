package provider

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/awstesting/unit"
	"github.com/aws/aws-sdk-go/service/ec2"
)

var exampleDescribeInstancesResponse = `
<DescribeInstancesResponse xmlns="http://ec2.amazonaws.com/doc/2016-01-01/">
	<requestId>f215b40f-5a0c-4fe6-9624-657cd1f4ef6b</requestId>
	<reservationSet>
		<item>
			<reservationId>r-0</reservationId>
			<ownerId>901416387788</ownerId>
			<groupSet/>
			<instancesSet>
				<item>
					<instanceId>i-0</instanceId>
					<imageId>ami-0</imageId>
					<instanceState>
						<code>0</code>
						<name>stopped</name>
					</instanceState>
					<privateIpAddress>192.168.1.88</privateIpAddress>
                    <ipAddress>10.10.10.1</ipAddress>
				</item>
			</instancesSet>
			<requesterId>0</requesterId>
		</item>
	</reservationSet>
</DescribeInstancesResponse>`

func initTestEC2Server(path string, resp string) *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(resp))
	}))
}

func TestNewEC2(t *testing.T) {
	accessKeyID := "access"
	secretAccessKey := "secret"
	region := "us-west-2"
	instanceID := "instance-1"

	inst := NewEC2(accessKeyID, secretAccessKey, region, instanceID, false)

	if inst.AccessKeyID != accessKeyID {
		t.Errorf("NewEC2.AccessKeyID returned %+v, want %+v", inst.AccessKeyID, accessKeyID)
	}

	if inst.SecretAccessKey != secretAccessKey {
		t.Errorf("NewEC2.SecretAccessKey returned %+v, want %+v", inst.SecretAccessKey, secretAccessKey)
	}

	if inst.Region != region {
		t.Errorf("NewEC2.Region returned %+v, want %+v", inst.Region, region)
	}

	if inst.InstanceID != instanceID {
		t.Errorf("NewEC2.InstanceID returned %+v, want %+v", inst.InstanceID, instanceID)
	}

	if inst.ec2Service == nil {
		t.Errorf("NewEC2.ec2Service not set")
	}
}

func TestEC2_GetStatus(t *testing.T) {
	server := initTestEC2Server("/", exampleDescribeInstancesResponse)
	defer server.Close()

	svc := ec2.New(unit.Session, &aws.Config{Endpoint: aws.String(server.URL + "/")})

	inst := &EC2{
		AccessKeyID:     "TEST",
		SecretAccessKey: "TEST",
		Region:          "us-west-2",
		InstanceID:      "i-0",
		ec2Service:      svc,
	}

	status, err := inst.GetStatus()
	if err != nil {
		t.Errorf("EC2.GetStatus returned unexpected error: %v", err)
	}

	if status != StatusInstanceNotRun {
		t.Errorf("EC2.GetStatus returned %+v, want %+v", status, StatusInstanceNotRun)
	}
}

func TestNormalizeEC2Status(t *testing.T) {
	var statusTable = []struct {
		in  string
		out int
	}{
		{"pending", StatusInstanceStarting},
		{"running", StatusInstanceRunning},
		{"shutting-down", StatusInstanceStopping},
		{"stopping", StatusInstanceStopping},
		{"stopped", StatusInstanceNotRun},
		{"terminated", StatusInstanceNotRun},
		{"unknown", StatusInstanceNotAvailable},
	}

	for _, test := range statusTable {
		if s := normalizeEC2Status(test.in); s != test.out {
			t.Errorf("normalizeEC2Status is %v, want %v", s, test.out)
		}
	}
}

func TestEC2_String(t *testing.T) {
	inst := &EC2{
		AccessKeyID:     "TEST",
		SecretAccessKey: "TEST",
		Region:          "us-west-2",
		InstanceID:      "i-0",
	}
	s := "[EC2] ID: i-0 in us-west-2"

	if inst.String() != s {
		t.Errorf("GCE.String returned %+v, want %+v", inst.String(), s)
	}
}

func TestEC2_Hash(t *testing.T) {
	inst := &EC2{
		AccessKeyID:     "TEST",
		SecretAccessKey: "TEST",
		Region:          "us-west-2",
		InstanceID:      "i-0",
	}
	s := "ec2-i-0-us-west-2"

	if inst.Hash() != s {
		t.Errorf("GCE.Hash returned %+v, want %+v", inst.Hash(), s)
	}
}

func TestEC2_Start(t *testing.T) {
	server := initTestEC2Server("/", exampleDescribeInstancesResponse)
	defer server.Close()

	svc := ec2.New(unit.Session, &aws.Config{Endpoint: aws.String(server.URL + "/")})

	inst := &EC2{
		AccessKeyID:     "TEST",
		SecretAccessKey: "TEST",
		Region:          "us-west-2",
		InstanceID:      "i-0",
		ec2Service:      svc,
	}

	err := inst.Start()
	if err != nil {
		t.Errorf("EC2.Start returned unexpected error: %v", err)
	}
}

func TestEC2_Stop(t *testing.T) {
	server := initTestEC2Server("/", exampleDescribeInstancesResponse)
	defer server.Close()

	svc := ec2.New(unit.Session, &aws.Config{Endpoint: aws.String(server.URL + "/")})

	inst := &EC2{
		AccessKeyID:     "TEST",
		SecretAccessKey: "TEST",
		Region:          "us-west-2",
		InstanceID:      "i-0",
		ec2Service:      svc,
	}

	err := inst.Stop()
	if err != nil {
		t.Errorf("EC2.Stop returned unexpected error: %v", err)
	}
}

func TestEC2_GetIP(t *testing.T) {
	server := initTestEC2Server("/", exampleDescribeInstancesResponse)
	defer server.Close()

	svc := ec2.New(unit.Session, &aws.Config{Endpoint: aws.String(server.URL + "/")})

	inst := &EC2{
		AccessKeyID:     "TEST",
		SecretAccessKey: "TEST",
		Region:          "us-west-2",
		InstanceID:      "i-0",
		ec2Service:      svc,
	}

	ip, err := inst.GetIP()
	if err != nil {
		t.Errorf("EC2.GetIP returned unexpected error: %v", err)
	}

	if ip != "10.10.10.1" {
		t.Errorf("EC2.GetIP returned %+v, want %+v", ip, "10.10.10.1")
	}
}

func TestEC2_GetIP_withInternalIP(t *testing.T) {
	server := initTestEC2Server("/", exampleDescribeInstancesResponse)
	defer server.Close()

	svc := ec2.New(unit.Session, &aws.Config{Endpoint: aws.String(server.URL + "/")})

	inst := &EC2{
		AccessKeyID:     "TEST",
		SecretAccessKey: "TEST",
		Region:          "us-west-2",
		InstanceID:      "i-0",
		UseInternalIP:   true,
		ec2Service:      svc,
	}

	ip, err := inst.GetIP()
	if err != nil {
		t.Errorf("EC2.GetIP returned unexpected error: %v", err)
	}

	if ip != "192.168.1.88" {
		t.Errorf("EC2.GetIP returned %+v, want %+v", ip, "192.168.1.88")
	}
}
