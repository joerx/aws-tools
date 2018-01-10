package main

import (
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ec2"
)

// Return the value of a single key from a list of AWS resource tags
func (t ResourceTags) valueOf(key string) (value string) {
	for _, tag := range t {
		if tag.Key == key {
			value = tag.Value
		}
	}
	return
}

// NewSession creates a new session with shared config state
func NewSession() *session.Session {
	return session.Must(session.NewSessionWithOptions(session.Options{
		SharedConfigState: session.SharedConfigEnable,
	}))
}

// Create new EC2 service instance using shared state
func NewEC2() *ec2.EC2 {
	sess := NewSession()
	return ec2.New(sess)
}

func describeVolumes(client *ec2.EC2) ([]*ebsVolume, error) {
	params := ec2.DescribeVolumesInput{}
	info, err := client.DescribeVolumes(&params)
	if err != nil {
		return nil, err
	}

	log.Printf("Got %d volumes", len(info.Volumes))

	// map results into ebsVolume structs
	result := make([]*ebsVolume, 0, len(info.Volumes))
	ids := make([]*string, 0, len(result))
	instanceIds := make([]*string, 0, len(result))

	for _, ev := range info.Volumes {
		ids = append(ids, ev.VolumeId)
		if len(ev.Attachments) > 0 {
			instanceIds = append(instanceIds, ev.Attachments[0].InstanceId)
		}
	}

	// Note: assume the below could be parallelized and would hence be a use case for goroutines
	// Leaving this as an exercise for future /me

	// Find all tags for all volumes we found (map later)
	tags, err := describeTags(client, ids)
	if err != nil {
		return nil, err
	}

	// Find all instances for all volumes we foudn (map later)
	instances, err := describeInstances(client, instanceIds)
	if err != nil {
		return nil, err
	}

	for _, ev := range info.Volumes {
		vol := &ebsVolume{
			id:    *ev.VolumeId,
			zone:  *ev.AvailabilityZone,
			size:  *ev.Size,
			class: *ev.VolumeType,
			state: *ev.State,
		}

		// Resolve attachment information to EC2 instance, if any
		if len(ev.Attachments) > 0 {
			instanceID := *ev.Attachments[0].InstanceId
			vol.attachment = &ebsAttachment{
				instanceID: instanceID,
				device:     *ev.Attachments[0].Device,
			}
			if instances[instanceID] != nil {
				vol.attachment.instance = instances[instanceID]
			}
		}

		// Resolve tags for this volume, if any
		if tags[*ev.VolumeId] != nil {
			vol.tags = tags[*ev.VolumeId]
		} else {
			vol.tags = ResourceTags{}
		}

		result = append(result, vol)
	}
	return result, nil
}

func min(one int, two int) int {
	if one > two {
		return two
	}
	return one
}

// Get tags for a given list of resourceIds, works for at least volume ids and ec2 instance ids
func describeTags(client *ec2.EC2, resourceIDs []*string) (map[string]ResourceTags, error) {
	start := 0
	upto := 200
	result := make(map[string]ResourceTags)

	// need to process in chunks since only up to 200 ids are allowed
	for start <= len(resourceIDs) {
		ids := resourceIDs[start:min(upto, len(resourceIDs))]
		upto += 200
		start += 200

		input := &ec2.DescribeTagsInput{
			Filters: []*ec2.Filter{
				{
					Name:   aws.String("resource-id"),
					Values: ids,
				},
			},
		}

		tags, err := client.DescribeTags(input)
		if err != nil {
			return nil, err
		}

		for _, tag := range tags.Tags {
			rid := *tag.ResourceId
			rt := &ResourceTag{Key: *tag.Key, Value: *tag.Value}
			if result[rid] == nil {
				result[rid] = ResourceTags{rt}
			} else {
				result[rid] = append(result[rid], rt)
			}
		}
	}

	return result, nil
}

func describeInstances(client *ec2.EC2, instanceIDs []*string) (map[string]*ec2Instance, error) {
	start := 0
	upto := 200
	result := make(map[string]*ec2Instance)

	fmt.Println(len(instanceIDs), "instances")

	// Another interesting scenario for goroutines - get all batches in parallel
	for start <= len(instanceIDs) {
		if upto >= len(instanceIDs) {
			upto = len(instanceIDs) - 1
		}

		input := &ec2.DescribeInstancesInput{
			Filters: []*ec2.Filter{
				{
					Name:   aws.String("instance-id"),
					Values: instanceIDs[start:upto],
				},
			},
		}

		response, err := client.DescribeInstances(input)
		if err != nil {
			return nil, err
		}

		for _, res := range response.Reservations {
			for _, in := range res.Instances {
				if result[*in.InstanceId] == nil {
					result[*in.InstanceId] = &ec2Instance{
						id: *in.InstanceId,
					}
				}
			}
		}

		start += upto
		upto += upto
	}

	tags, err := describeTags(client, instanceIDs)
	if err != nil {
		return nil, err
	}

	for instanceID, instance := range result {
		if tags[instanceID] != nil {
			instance.tags = tags[instanceID]
			instance.name = tags[instanceID].valueOf("Name")
		}
	}

	return result, nil
}
