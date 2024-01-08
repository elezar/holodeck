/*
 * Copyright (c) 2023, NVIDIA CORPORATION.  All rights reserved.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package aws

import (
	"context"
	"os"

	"github.com/NVIDIA/holodeck/api/holodeck/v1alpha1"
	"github.com/NVIDIA/holodeck/pkg/jyaml"
	"sigs.k8s.io/yaml"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/aws/aws-sdk-go-v2/service/route53"
	"github.com/aws/aws-sdk-go/aws"
)

const (
	// Name of this builder provider
	Name                             = "aws"
	VpcID                     string = "vpc-id"
	SubnetID                  string = "subnet-id"
	InternetGwID              string = "internet-gateway-id"
	InternetGatewayAttachment string = "internet-gateway-attachment-vpc-id"
	RouteTable                string = "route-table-id"
	SecurityGroupID           string = "security-group-id"
	InstanceID                string = "instance-id"
	PublicDnsName             string = "public-dns-name"
)

var (
	description string = "Holodeck managed AWS Cloud Provider"
)

var (
	yes        = true
	no         = false
	tcp string = "tcp"

	k8s443        int32 = 443
	k8s6443       int32 = 6443
	minMaxCount   int32 = 1
	storageSizeGB int32 = 64
)

type AWS struct {
	Vpcid                     string
	Subnetid                  string
	InternetGwid              string
	InternetGatewayAttachment string
	RouteTable                string
	SecurityGroupid           string
	Instanceid                string
	PublicDnsName             string
}

type Client struct {
	Tags      []types.Tag
	ec2       *ec2.Client
	r53       *route53.Client
	cachePath string

	*v1alpha1.Environment
}

func New(env v1alpha1.Environment, cachePath string) (*Client, error) {
	// Create an AWS session and configure the EC2 client
	cfg, err := config.LoadDefaultConfig(context.TODO(), config.WithRegion(env.Spec.Region))
	if err != nil {
		return nil, err
	}

	client := ec2.NewFromConfig(cfg)
	r53 := route53.NewFromConfig(cfg)
	c := &Client{
		[]types.Tag{
			{Key: aws.String("Product"), Value: aws.String("Cloud Native")},
			{Key: aws.String("Name"), Value: aws.String("devel")},
			{Key: aws.String("Project"), Value: aws.String("holodeck")},
			{Key: aws.String("Environment"), Value: aws.String("cicd")},
		},
		client,
		r53,
		cachePath,
		&env,
	}

	return c, nil
}

// Name returns the name of the builder provisioner
func (a *Client) Name() string { return Name }

// unmarsalCache unmarshals the cache file into the AWS struct
func (a *Client) unmarsalCache() (*AWS, error) {
	env, err := jyaml.UnmarshalFromFile[v1alpha1.Environment](a.cachePath)
	if err != nil {
		return nil, err
	}

	aws := &AWS{}

	for _, p := range env.Status.Properties {
		switch p.Name {
		case VpcID:
			aws.Vpcid = p.Value
		case SubnetID:
			aws.Subnetid = p.Value
		case InternetGwID:
			aws.InternetGwid = p.Value
		case InternetGatewayAttachment:
			aws.InternetGatewayAttachment = p.Value
		case RouteTable:
			aws.RouteTable = p.Value
		case SecurityGroupID:
			aws.SecurityGroupid = p.Value
		case InstanceID:
			aws.Instanceid = p.Value
		case PublicDnsName:
			aws.PublicDnsName = p.Value
		default:
			// Ignore non AWS infra properties
			continue
		}
	}

	return aws, nil
}

// dump writes the AWS struct to the cache file
func (a *Client) dumpCache(aws *AWS) error {
	env := a.Environment.DeepCopy()
	env.Status.Properties = []v1alpha1.Properties{
		{Name: VpcID, Value: aws.Vpcid},
		{Name: SubnetID, Value: aws.Subnetid},
		{Name: InternetGwID, Value: aws.InternetGwid},
		{Name: InternetGatewayAttachment, Value: aws.InternetGatewayAttachment},
		{Name: RouteTable, Value: aws.RouteTable},
		{Name: SecurityGroupID, Value: aws.SecurityGroupid},
		{Name: InstanceID, Value: aws.Instanceid},
		{Name: PublicDnsName, Value: aws.PublicDnsName},
	}

	data, err := yaml.Marshal(env)
	if err != nil {
		return err
	}

	err = os.WriteFile(a.cachePath, data, 0644)
	if err != nil {
		return err
	}

	// Update the environment object with the new properties
	a.Environment = env

	return nil
}
