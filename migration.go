package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"

	"github.com/byuoitav/configuration-database-microservice/structs"

	"github.com/byuoitav/av-api/dbo"
	"github.com/byuoitav/common/log"
	newstructs "github.com/byuoitav/common/structs"
)

var buildingList []structs.Building
var roomList []structs.Room
var configList []structs.RoomConfiguration
var deviceClassList []structs.DeviceClass

var typePortMap map[string][]structs.DeviceTypePort
var commandNameMap map[string]structs.RawCommand

var COUCH_ADDRESS string
var COUCH_USERNAME string
var COUCH_PASSWORD string

func main() {

	var err error

	buildingList, err = dbo.GetBuildings()
	if err != nil {
		log.L.Errorf("Failed to get info from old config db : %v", err)
	}
	roomList, err = dbo.GetRooms()
	if err != nil {
		log.L.Errorf("Failed to get info from old config db : %v", err)
	}
	configList, err = dbo.GetRoomConfigurations()
	if err != nil {
		log.L.Errorf("Failed to get info from old config db : %v", err)
	}
	deviceClassList, err = dbo.GetDeviceClasses()
	if err != nil {
		log.L.Errorf("Failed to get info from old config db : %v", err)
	}
	allCommands, err := dbo.GetAllRawCommands()
	if err != nil {
		log.L.Errorf("Failed to get info from old config db : %v", err)
	}

	COUCH_ADDRESS = os.Getenv("DB_ADDRESS")
	COUCH_USERNAME = os.Getenv("DB_USERNAME")
	COUCH_PASSWORD = os.Getenv("DB_PASSWORD")

	typePortMap = make(map[string][]structs.DeviceTypePort)

	for _, t := range deviceClassList {
		typePortMap[t.Name], err = dbo.GetPortsByClass(t.Name)
		if err != nil {
			log.L.Errorf("Failed to get info from old config db : %v", err)
		}
	}

	// log.L.Infof("%s", typePortMap)

	commandNameMap = make(map[string]structs.RawCommand)

	for _, c := range allCommands {
		commandNameMap[c.Name] = c
	}

	moveBuildings()
	moveRooms()
	moveRoomConfigurations()
	moveDevicesAndTypes()
}

func moveBuildings() {
	log.L.Info("Starting moveBuildings...")

	for i := range buildingList {

		bldg := newstructs.Building{}

		bldg.ID = buildingList[i].Shortname
		bldg.Name = buildingList[i].Name
		bldg.Description = buildingList[i].Description

		url := fmt.Sprintf("%v/buildings/%v", COUCH_ADDRESS, bldg.ID)

		body, err := json.Marshal(bldg)
		if err != nil {
			log.L.Errorf("Cannot marshal building : %v", err)
			return
		}

		req, err := http.NewRequest("PUT", url, bytes.NewReader(body))
		if err != nil {
			log.L.Errorf("Error making request : %v", err)
			return
		}

		// add auth
		if len(COUCH_USERNAME) > 0 && len(COUCH_PASSWORD) > 0 {
			req.SetBasicAuth(COUCH_USERNAME, COUCH_PASSWORD)
		}

		client := &http.Client{}

		resp, err := client.Do(req)
		if err != nil {
			log.L.Errorf("Error doing request : %v", err)
			return
		}

		resp.Body.Close()
	}
}

func moveRooms() {
	log.L.Info("Starting moveRooms...")

	for _, r := range roomList {

		room := newstructs.Room{}
		config := newstructs.RoomConfiguration{}

		bldgName := ""

		for a := 0; a < len(buildingList); a++ {
			if r.Building.ID == buildingList[a].ID {
				bldgName = buildingList[a].Shortname
			}
		}

		configName := ""

		for b := 0; b < len(configList); b++ {
			if r.ConfigurationID == configList[b].ID {
				configName = configList[b].Name
			}
		}

		room.ID = fmt.Sprintf("%s-%s", bldgName, r.Name)
		room.Description = r.Description
		config.ID = configName
		room.Configuration = config
		room.Designation = r.RoomDesignation

		url := fmt.Sprintf("%v/rooms/%v", COUCH_ADDRESS, room.ID)

		body, err := json.Marshal(room)
		if err != nil {
			log.L.Errorf("Cannot marshal room : %v", err)
			return
		}

		req, err := http.NewRequest("PUT", url, bytes.NewReader(body))
		if err != nil {
			log.L.Errorf("Error making request : %v", err)
			return
		}

		// add auth
		if len(COUCH_USERNAME) > 0 && len(COUCH_PASSWORD) > 0 {
			req.SetBasicAuth(COUCH_USERNAME, COUCH_PASSWORD)
		}

		client := &http.Client{}

		resp, err := client.Do(req)
		if err != nil {
			log.L.Errorf("Error doing request : %v", err)
			return
		}

		resp.Body.Close()
	}
}

func moveRoomConfigurations() {
	log.L.Info("Starting moveRoomConfigurations...")

	for _, c := range configList {

		config := newstructs.RoomConfiguration{}

		var evals []newstructs.Evaluator

		for _, r := range roomList {
			if r.ConfigurationID == c.ID {
				bName := ""
				for _, b := range buildingList {
					if r.Building.ID == b.ID {
						bName = b.Shortname
						break
					}
				}

				fullRoom, _ := dbo.GetRoomByInfo(bName, r.Name)

				evals = make([]newstructs.Evaluator, len(fullRoom.Configuration.Evaluators))

				for i, e := range fullRoom.Configuration.Evaluators {
					evals[i].ID = e.EvaluatorKey
					evals[i].CodeKey = e.EvaluatorKey
					evals[i].Priority = e.Priority
					evals[i].Description = e.EvaluatorKey
				}

				break
			}
		}

		config.ID = c.Name
		config.Description = c.RoomInitKey
		config.Evaluators = evals

		log.L.Info(config)

		url := fmt.Sprintf("%v/room_configurations/%v", COUCH_ADDRESS, config.ID)

		body, err := json.Marshal(config)
		if err != nil {
			log.L.Errorf("Cannot marshal room configuration : %v", err)
			return
		}

		req, err := http.NewRequest("PUT", url, bytes.NewReader(body))
		if err != nil {
			log.L.Errorf("Error making request : %v", err)
			return
		}

		// add auth
		if len(COUCH_USERNAME) > 0 && len(COUCH_PASSWORD) > 0 {
			req.SetBasicAuth(COUCH_USERNAME, COUCH_PASSWORD)
		}

		client := &http.Client{}

		resp, err := client.Do(req)
		if err != nil {
			log.L.Errorf("Error doing request : %v", err)
			return
		}

		resp.Body.Close()
	}
}

func moveDevicesAndTypes() {
	log.L.Infof("Building list size: %v", len(buildingList))
	log.L.Infof("Room list size: %v", len(roomList))
	log.L.Infof("Config list size: %v", len(configList))
	totalPortList, err := dbo.GetPorts()
	microserviceList, err := dbo.GetMicroservices()
	endpointList, err := dbo.GetEndpoints()

	if err != nil {
		log.L.Errorf("Failed to get info from old config db : %v", err)
	}

	for _, r := range roomList {
		bName := ""
		for _, b := range buildingList {
			if r.Building.ID == b.ID {
				bName = b.Shortname
				break
			}
		}

		fullRoom, _ := dbo.GetRoomByInfo(bName, r.Name)

		for _, d := range fullRoom.Devices {
			device := newstructs.Device{}

			device.ID = fmt.Sprintf("%v-%v-%v", fullRoom.Building.Shortname, fullRoom.Name, d.Name)
			device.Address = d.Address
			device.Name = d.Name
			device.Description = d.DisplayName
			device.DisplayName = d.DisplayName

			dType := newstructs.DeviceType{}
			dType.ID = d.Class
			device.Type = dType

			roleList := make([]newstructs.Role, len(d.Roles))

			for i, role := range d.Roles {
				roleList[i].ID = role
				roleList[i].Description = role
			}

			device.Roles = roleList

			portList := make([]newstructs.Port, len(d.Ports))

			for j, port := range d.Ports {
				for _, p := range totalPortList {
					if port.Name == p.Name {
						portList[j].ID = p.Name
						portList[j].FriendlyName = p.Description
						portList[j].Description = p.Description
						break
					}
				}

				portList[j].SourceDevice = fmt.Sprintf("%s-%s-%s", bName, r.Name, port.Source)
				portList[j].DestinationDevice = fmt.Sprintf("%s-%s-%s", bName, r.Name, port.Destination)
			}

			device.Ports = portList

			// Creating/moving the DeviceTypes here as well...
			deviceType := newstructs.DeviceType{}

			for _, t := range deviceClassList {
				if d.Class == t.Name {
					deviceType.ID = t.Name
					deviceType.Description = t.Description
					deviceType.Input = d.Input
					deviceType.Output = d.Output

					typePortList := typePortMap[t.Name]

					ports := make([]newstructs.Port, len(typePortList))

					for i, p := range typePortList {
						ports[i].ID = p.Port.Name
						ports[i].FriendlyName = p.Port.Description
						ports[i].Description = p.Port.Description
					}

					deviceType.Ports = ports

					commandList := make([]newstructs.Command, len(d.Commands))

					for k, command := range d.Commands {
						commandList[k].ID = command.Name
						commandList[k].Description = command.Name
						commandList[k].Priority = commandNameMap[command.Name].Priority

						for _, m := range microserviceList {
							if command.Microservice == m.Address {
								micro := newstructs.Microservice{}

								micro.ID = m.Name
								micro.Address = m.Address
								micro.Description = m.Description

								commandList[k].Microservice = micro
								break
							}
						}

						for _, e := range endpointList {
							if command.Endpoint.Path == e.Path {
								end := newstructs.Endpoint{}

								end.ID = e.Name
								end.Path = e.Path
								end.Description = e.Description

								commandList[k].Endpoint = end
								break
							}
						}
					}

					deviceType.Commands = commandList
				}
			}

			// // Send the Device to Couch
			url := fmt.Sprintf("%s/devices/%s", COUCH_ADDRESS, device.ID)

			body, err := json.Marshal(device)
			if err != nil {
				log.L.Errorf("Cannot marshal device : %v", err)
				return
			}

			req, err := http.NewRequest("PUT", url, bytes.NewReader(body))
			if err != nil {
				log.L.Errorf("Error making request : %v", err)
				return
			}

			// add auth
			if len(COUCH_USERNAME) > 0 && len(COUCH_PASSWORD) > 0 {
				req.SetBasicAuth(COUCH_USERNAME, COUCH_PASSWORD)
			}

			client := &http.Client{}

			resp, err := client.Do(req)
			if err != nil {
				log.L.Errorf("Error doing request : %v", err)
				return
			}

			resp.Body.Close()

			// Send the DeviceType to Couch
			url2 := fmt.Sprintf("%s/device_types/%s", COUCH_ADDRESS, deviceType.ID)

			body2, err := json.Marshal(deviceType)
			if err != nil {
				log.L.Errorf("Cannot marshal device type : %v", err)
				return
			}

			req2, err := http.NewRequest("PUT", url2, bytes.NewReader(body2))
			if err != nil {
				log.L.Errorf("Error making request : %v", err)
				return
			}

			// add auth
			if len(COUCH_USERNAME) > 0 && len(COUCH_PASSWORD) > 0 {
				req2.SetBasicAuth(COUCH_USERNAME, COUCH_PASSWORD)
			}

			resp2, err := client.Do(req2)
			if err != nil {
				log.L.Errorf("Error doing request : %v", err)
				return
			}

			resp2.Body.Close()
		}
	}
}
