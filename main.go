package main

import (
	"fmt"
	"os"

	"github.com/urfave/cli"
)

func main() {

	app := cli.NewApp()
	app.Name = "aws tools"
	app.Usage = "various high-level tools to work with aws api"

	zoneExportCommand := cli.Command{
		Name:   "export",
		Usage:  "export one or multiple zones to csv",
		Action: actionExportZone,
		Flags: []cli.Flag{
			cli.StringSliceFlag{
				Name:  "zone-id, z",
				Usage: "zone ID to export",
			},
			cli.StringFlag{
				Name:  "output, o",
				Usage: "output file, using stdout if omitted",
			},
		},
	}

	zoneListCommand := cli.Command{
		Name:   "list",
		Usage:  "list all hosted zones visible to the current user",
		Action: actionListZones,
	}

	zoneCommand := cli.Command{
		Name:  "zone",
		Usage: "work with Route53 zones",
		Subcommands: []cli.Command{
			zoneExportCommand,
			zoneListCommand,
		},
	}

	app.Commands = []cli.Command{
		zoneCommand,
	}

	if err := app.Run(os.Args); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func actionExportZone(c *cli.Context) error {
	zoneIDs := c.StringSlice("zone-id")
	f, err := getOutput(c.String("output"))

	if err != nil {
		return err
	}

	records, err := ExportZone(zoneIDs...)

	if err != nil {
		return err
	}

	return WriteToCSV(f, records)
}

func actionListZones(c *cli.Context) error {
	zones, err := ListZones()

	if err != nil {
		return err
	}

	fmt.Println("ZoneID\tName")
	for _, zone := range zones {
		fmt.Printf("%s\t%s\n", zone.ZoneID, zone.Name)
	}

	return nil
}

// getOutput returns the output stream to write to depending on the value of outParam. If out is
// empty it will return stdout, otherwise it will create a file at the path matching outParam
func getOutput(outParam string) (*os.File, error) {
	if outParam == "" {
		return os.Stdout, nil
	}

	f, err := os.Create(outParam)
	if err != nil {
		return nil, err
	}
	// defer f.Close()

	return f, err
}

// func matchRecords(left []*Route53Record, right []*Route53Record) RowSet {

// 	lookup := mapByName(right)
// 	rows := [][]string{}

// 	headers := []string{
// 		"Name",
// 		"ZoneID",
// 		"TypeA",
// 		"ValueA",
// 		"TypeB",
// 		"ValueB",
// 		"IsEqual",
// 		"RecordA",
// 		"RecordB",
// 	}

// 	for _, leftRecord := range left {
// 		switch leftRecord.Type {
// 		case "A", "CNAME":
// 			name := leftRecord.Name
// 			rightRecord := lookup[name]

// 			var rightType string
// 			var rightValue string
// 			var isEqual string

// 			if rightRecord != nil {
// 				rightType = rightRecord.Type
// 				rightValue = rightRecord.FormatValue()
// 				isEqual = fmt.Sprintf("%t", leftRecord.FormatValue() == rightRecord.FormatValue())
// 			}

// 			rows = append(rows, []string{
// 				name,
// 				leftRecord.ZoneID,
// 				leftRecord.Type,
// 				leftRecord.FormatValue(),
// 				rightType,
// 				rightValue,
// 				isEqual,
// 				fmt.Sprintf("%#v", leftRecord),
// 				fmt.Sprintf("%#v", rightRecord),
// 			})

// 		default:
// 			continue
// 		}
// 	}

// 	reverseLookup := mapByName(left)

// 	// append records from right side that don't exist on the left
// 	for _, rightRecord := range right {
// 		switch rightRecord.Type {
// 		case "A", "CNAME":
// 			// skip if record exists on the other side
// 			if reverseLookup[rightRecord.Name] != nil {
// 				continue
// 			}
// 			rows = append(rows, []string{
// 				rightRecord.Name,
// 				rightRecord.ZoneID,
// 				"",
// 				"",
// 				rightRecord.Type,
// 				rightRecord.FormatValue(),
// 				"false",
// 				fmt.Sprintf("%#v", nil),
// 				fmt.Sprintf("%#v", rightRecord),
// 			})
// 		default:
// 			continue
// 		}
// 	}

// 	exportData := RowSet{
// 		Headers: headers,
// 		SortCol: 0, // sort by name
// 		Rows:    rows,
// 	}
// 	sort.Sort(exportData)

// 	return exportData
// }

// func mapByName(records []*Route53Record) map[string]*Route53Record {
// 	// map by name
// 	zoneMap := make(map[string]*Route53Record)
// 	for _, record := range records {
// 		zoneMap[record.Name] = record
// 	}
// 	return zoneMap
// }

// func exportVolumes() {
// 	ec2Client := NewEC2()

// 	volumes, err := describeVolumes(ec2Client)
// 	if err != nil {
// 		log.Fatal(err)
// 	}

// 	rows := make([][]string, 0, len(volumes))
// 	header := []string{
// 		"volumeId",
// 		"size",
// 		"type",
// 		"instanceId",
// 		"instanceName",
// 		"kubernetesCluster",
// 		"ebEnv",
// 	}

// 	for _, vol := range volumes {
// 		var id string
// 		var name string
// 		var cluster string
// 		var ebEnv string

// 		if vol.attachment != nil {
// 			id = vol.attachment.instanceID
// 			if vol.attachment.instance != nil {
// 				name = vol.attachment.instance.name
// 				cluster = vol.attachment.instance.tags.valueOf("KubernetesCluster")
// 				ebEnv = vol.attachment.instance.tags.valueOf("elasticbeanstalk:environment-name")
// 			}
// 		}
// 		rows = append(rows, []string{
// 			vol.id,
// 			fmt.Sprintf("%d", vol.size),
// 			vol.class,
// 			id,
// 			name,
// 			cluster,
// 			ebEnv,
// 		})
// 	}

// 	exportData := RowSet{
// 		Headers: header,
// 		SortCol: 0, // sort by name
// 		Rows:    rows,
// 	}
// 	sort.Sort(exportData)

// 	WriteToCSV(outfile, exportData)
// 	log.Println("Results written to", outfile)
// }
