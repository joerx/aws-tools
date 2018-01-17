package main

import (
	"log"
	"strings"

	"github.com/aws/aws-sdk-go/service/route53"
)

var r53Client *route53.Route53

// Route53Record represents a simplified AWS Route53 resource record
type Route53Record struct {
	IsAlias bool
	Value   []string
	Name    string
	Type    string
	ZoneID  string
}

// Route53Zone represents a single AWS R53 hosted zone
type Route53Zone struct {
	Name   string
	ZoneID string
}

// ExportZone exports the records from a Route53 zone into a rowset so it can be exported as CSV
func ExportZone(zoneIDs ...string) (*RowSet, error) {
	records, err := ListZonesRecords(zoneIDs...)

	if err != nil {
		return nil, err
	}

	headers := []string{"ZoneID", "Name", "Type", "Value"}
	rows := make([][]string, len(records))
	for i, record := range records {
		rows[i] = []string{
			record.ZoneID,
			record.Name,
			record.Type,
			record.FormatValue(),
		}
	}

	data := &RowSet{
		Rows:    rows,
		SortCol: 0,
		Headers: headers,
	}

	return data, nil
}

// getRoute53Client creates a new route53 api client
func getRoute53Client() *route53.Route53 {
	if r53Client == nil {
		sess := NewSession()
		r53Client = route53.New(sess)
	}
	return r53Client
}

// ListZones lists all zones visible to the current IAM user
func ListZones() ([]*Route53Zone, error) {
	client := getRoute53Client()
	allZones := make([]*Route53Zone, 0, 100)

	cb := func(response *route53.ListHostedZonesOutput, lastPage bool) bool {
		for _, r := range response.HostedZones {
			zone := &Route53Zone{
				Name:   *r.Name,
				ZoneID: *r.Id,
			}
			allZones = append(allZones, zone)
		}
		return true
	}
	input := &route53.ListHostedZonesInput{} // defaults

	if err := client.ListHostedZonesPages(input, cb); err != nil {
		return nil, err
	}

	return allZones, nil
}

// ListZonesRecords returns all resource records for multiple zones
func ListZonesRecords(zoneIDs ...string) ([]*Route53Record, error) {
	allRecords := []*Route53Record{}
	for _, zoneID := range zoneIDs {
		records, err := ListZoneRecords(zoneID)
		if err != nil {
			return nil, err
		}
		allRecords = append(allRecords, records...)
	}
	return allRecords, nil
}

// ListZoneRecords returns all resource records for a single zone
func ListZoneRecords(zoneID string) ([]*Route53Record, error) {
	client := getRoute53Client()
	input := &route53.ListResourceRecordSetsInput{
		HostedZoneId: &zoneID,
	}

	allRecords := []*route53.ResourceRecordSet{}

	cb := func(response *route53.ListResourceRecordSetsOutput, lastPage bool) bool {
		allRecords = append(allRecords, response.ResourceRecordSets...)
		return true
	}

	if err := client.ListResourceRecordSetsPages(input, cb); err != nil {
		return nil, err
	}

	log.Printf("Found %d records for zone %s", len(allRecords), zoneID)

	result := make([]*Route53Record, len(allRecords))
	for i, rs := range allRecords {
		IsAlias := rs.AliasTarget != nil
		var Value []string

		if IsAlias {
			Value = []string{*rs.AliasTarget.DNSName}
		} else {
			Value = recordValues(rs)
		}

		result[i] = &Route53Record{
			Name:    *rs.Name,
			Type:    *rs.Type,
			IsAlias: IsAlias,
			Value:   Value,
			ZoneID:  zoneID,
		}
	}

	return result, nil
}

// recordValues all values for a resource record as string slice
func recordValues(rs *route53.ResourceRecordSet) []string {
	strs := make([]string, len(rs.ResourceRecords))
	for i, r := range rs.ResourceRecords {
		strs[i] = *r.Value
	}
	return strs
}

// FormatValue returns the formatted value of the given record. It will return the unmodified value
// for regular records and the value prefixed with "ALIAS" if it's an alias record
func (r *Route53Record) FormatValue() string {
	if r.IsAlias {
		return "ALIAS " + r.Value[0]
	}
	return strings.Join(r.Value, "\n")
}
