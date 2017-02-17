package provider

import (
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/silentsokolov/go-sleep/log"
)

// EC2 ..
type EC2 struct {
	AccessKeyID     string
	SecretAccessKey string
	Region          string
	InstanceID      string
	UseInternalIP   bool
	session         *session.Session
	ec2Service      *ec2.EC2
}

// NewEC2 ..
func NewEC2(AccessKeyID, SecretAccessKey, Region, InstanceID string, UseInternalIP bool) *EC2 {
	session, err := getAWSSession(AccessKeyID, SecretAccessKey, Region)
	if err != nil {
		log.Fatalf("EC2 %s: Unable to session: %v", InstanceID, err)
	}

	return &EC2{
		AccessKeyID:     AccessKeyID,
		SecretAccessKey: SecretAccessKey,
		Region:          Region,
		InstanceID:      InstanceID,
		UseInternalIP:   UseInternalIP,
		session:         session,
		ec2Service:      ec2.New(session),
	}
}

// String ...
func (p *EC2) String() string {
	return fmt.Sprintf("[EC2] ID: %s in %s", p.InstanceID, p.Region)
}

// Hash ...
func (p *EC2) Hash() string {
	return fmt.Sprintf("ec2-%s-%s", p.InstanceID, p.Region)
}

// GetStatus ...
func (p *EC2) GetStatus() (int, error) {
	inst, err := p.getInstance()
	if err != nil {
		return StatusInstanceNotAvailable, err
	}

	return normalizeEC2Status(*inst.State.Name), nil
}

// GetIP ...
func (p *EC2) GetIP() (string, error) {
	inst, err := p.getInstance()
	if err != nil {
		return "", err
	}

	if p.UseInternalIP {
		return *inst.PrivateIpAddress, nil
	}

	return *inst.PublicIpAddress, nil
}

// Start ...
func (p *EC2) Start() error {
	params := &ec2.StartInstancesInput{
		InstanceIds: []*string{
			aws.String(p.InstanceID),
		},
	}

	_, err := p.ec2Service.StartInstances(params)
	if err != nil {
		return err
	}

	return nil
}

// Stop ...
func (p *EC2) Stop() error {
	params := &ec2.StopInstancesInput{
		InstanceIds: []*string{
			aws.String(p.InstanceID),
		},
	}

	_, err := p.ec2Service.StopInstances(params)
	if err != nil {
		return err
	}

	return nil
}

func (p *EC2) getInstance() (*ec2.Instance, error) {
	instances := make([]*ec2.Instance, 0)

	params := &ec2.DescribeInstancesInput{
		InstanceIds: []*string{
			aws.String(p.InstanceID),
		},
	}

	resp, err := p.ec2Service.DescribeInstances(params)
	if err != nil {
		return nil, err
	}

	for _, r := range resp.Reservations {
		for _, i := range r.Instances {
			instances = append(instances, i)
		}
	}

	if len(instances) < 1 {
		return nil, fmt.Errorf("EC2 instance %s not found", p.InstanceID)
	}

	return instances[0], nil
}

func normalizeEC2Status(originalStatus string) int {
	switch originalStatus {
	case "pending":
		return StatusInstanceStarting
	case "running":
		return StatusInstanceRunning
	case "shutting-down":
		return StatusInstanceStopping
	case "stopping":
		return StatusInstanceStopping
	case "stopped":
		return StatusInstanceNotRun
	case "terminated":
		return StatusInstanceNotRun
	default:
		return StatusInstanceNotAvailable
	}
}

func getAWSSession(AccessKeyID, SecretAccessKey, Region string) (*session.Session, error) {
	return session.NewSession(&aws.Config{
		Region:      aws.String(Region),
		Credentials: credentials.NewStaticCredentials(AccessKeyID, SecretAccessKey, ""),
	})
}
