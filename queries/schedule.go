package queries

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/mesos/mesos-go/api/v1/lib"
	"github.com/mesos/mesos-go/api/v1/lib/maintenance"
	"github.com/mesos/mesos-go/api/v1/lib/master"
	"github.com/minyk/dcos-maintenance/client"
	"strings"
	"time"
)

type Schedule struct {
	PrefixCb func() string
}

func NewSchedule() *Schedule {
	return &Schedule{}
}

func (q *Schedule) GetSchedule(rawJSON bool) error {
	body := master.Call{
		Type: master.Call_GET_MAINTENANCE_SCHEDULE,
	}

	requestContent, err := json.Marshal(body)
	if err != nil {
		return err
	}

	responseBytes, err := client.HTTPServicePostData("", requestContent, "application/json")
	if err != nil {
		return err
	}
	if rawJSON {
		client.PrintJSONBytes(responseBytes)
	} else {
		tree, err := toScheduleTable(responseBytes)
		if err != nil {
			return err
		}
		client.PrintMessage(tree)
	}
	return nil
}

func (q *Schedule) AddSchedule(start time.Time, duration time.Duration, file string) error {
	unav := mesos.Unavailability{
		Start: mesos.TimeInfo{
			Nanoseconds: start.UnixNano(),
		},
		Duration: &mesos.DurationInfo{
			Nanoseconds: duration.Nanoseconds(),
		},
	}

	schedule := maintenance.Schedule{
		Windows: []maintenance.Window{
			{
				Unavailability: unav,
			},
		},
	}

	body := master.Call{
		Type: master.Call_UPDATE_MAINTENANCE_SCHEDULE,
		UpdateMaintenanceSchedule: &master.Call_UpdateMaintenanceSchedule{
			Schedule: schedule,
		},
	}

	requestContent, err := json.Marshal(body)
	if err != nil {
		return err
	}

	client.PrintJSONBytes(requestContent)
	//_, err = client.HTTPServicePostJSON("", requestContent)
	//if err != nil {
	//	return err
	//} else {
	//	client.PrintMessage("Schedules updated")
	//}

	return nil
}

func (q *Schedule) RemoveSchedule(file string) error {
	// TODO get and update ?

	unav := mesos.Unavailability{
		Start: mesos.TimeInfo{
			Nanoseconds: 0,
		},
		Duration: &mesos.DurationInfo{
			Nanoseconds: 0,
		},
	}

	schedule := maintenance.Schedule{
		Windows: []maintenance.Window{
			{
				Unavailability: unav,
			},
		},
	}

	body := master.Call{
		Type: master.Call_UPDATE_MAINTENANCE_SCHEDULE,
		UpdateMaintenanceSchedule: &master.Call_UpdateMaintenanceSchedule{
			Schedule: schedule,
		},
	}

	requestContent, err := json.Marshal(body)
	if err != nil {
		return err
	}

	client.PrintJSONBytes(requestContent)
	//_, err = client.HTTPServicePostJSON("", requestContent)
	//if err != nil {
	//	return err
	//} else {
	//	client.PrintMessage("Schedules updated")
	//}

	return nil
}

func toScheduleTable(planJSONBytes []byte) (string, error) {
	schedule := master.Response{}
	err := json.Unmarshal(planJSONBytes, &schedule)
	if err != nil {
		return "", err
	}

	var buf bytes.Buffer
	buf.WriteString(fmt.Sprintln("Window\tHostname\t\tIP\t\tStart\t\t\t\tDuration"))

	for window, element := range schedule.GetMaintenanceSchedule.Schedule.Windows {

		start := time.Unix(0, element.Unavailability.Start.GetNanoseconds())
		duration := time.Duration(element.Unavailability.GetDuration().GetNanoseconds())

		for _, host := range element.MachineIDs {
			hostname := *host.Hostname
			ip := *host.IP
			buf.WriteString(fmt.Sprintf("%v\t%v\t%v\t%v\t%v\n", window, hostname, ip, start, duration))
		}

	}

	return strings.TrimRight(buf.String(), "\n"), nil
}
