package main

// Simplified struct representing and EBS volume and the EC2 instance it is attached to
type ebsVolume struct {
	id         string
	zone       string
	class      string
	tags       ResourceTags
	size       int64
	state      string
	attachment *ebsAttachment
}

// Attachment of an EBS volume to an EC2 instance
type ebsAttachment struct {
	instanceID string
	device     string
	instance   *ec2Instance
}

// ResourceTags represent a list of AWS resource tags
type ResourceTags []*ResourceTag

// ResourceTag represents a single AWS resource tag key-value pair
type ResourceTag struct {
	Key   string
	Value string
}

// Simplified struct representing an EC2 instance
type ec2Instance struct {
	id   string
	tags ResourceTags
	name string
}
