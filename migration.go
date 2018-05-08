package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/byuoitav/configuration-database-microservice/structs"

	"github.com/byuoitav/av-api/dbo"
	newstructs "github.com/byuoitav/common/structs"
)

func main() {

	buildingList, err := dbo.GetBuildings()
	roomList, err := dbo.GetRooms()
	configList, err := dbo.GetRoomConfigurations()
	if err != nil {
		log.Printf("Failed to get info from old config db : %s", err.Error())
	}

	moveBuildings(buildingList)
	moveRooms(buildingList, roomList, configList)
	moveRoomConfigurations(buildingList, roomList, configList)
	moveDevicesAndTypes(buildingList, roomList)
}

func moveBuildings(buildingList []structs.Building) {
	log.Print("Starting moveBuildings...")

	for i := range buildingList {

		bldg := newstructs.Building{}

		bldg.ID = buildingList[i].Shortname
		bldg.Description = buildingList[i].Description

		url := fmt.Sprintf("http://localhost:5984/buildings/%s", bldg.ID)

		body, err := json.Marshal(bldg)
		if err != nil {
			log.Printf("Cannot marshal building : %s", err.Error())
			return
		}

		req, err := http.NewRequest("PUT", url, bytes.NewReader(body))
		if err != nil {
			log.Printf("Error making request : %s", err.Error())
			return
		}

		client := &http.Client{}

		resp, err := client.Do(req)
		if err != nil {
			log.Printf("Error doing request : %s", err.Error())
			return
		}

		resp.Body.Close()
	}
}

func moveRooms(buildingList []structs.Building, roomList []structs.Room, configList []structs.RoomConfiguration) {
	log.Print("Starting moveRooms...")

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

		url := fmt.Sprintf("http://localhost:5984/rooms/%s", room.ID)

		body, err := json.Marshal(room)
		if err != nil {
			log.Printf("Cannot marshal room : %s", err.Error())
			return
		}

		req, err := http.NewRequest("PUT", url, bytes.NewReader(body))
		if err != nil {
			log.Printf("Error making request : %s", err.Error())
			return
		}

		client := &http.Client{}

		resp, err := client.Do(req)
		if err != nil {
			log.Printf("Error doing request : %s", err.Error())
			return
		}

		resp.Body.Close()
	}
}

func moveRoomConfigurations(buildingList []structs.Building, roomList []structs.Room, configList []structs.RoomConfiguration) {
	log.Print("Starting moveRoomConfigurations...")

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
		config.Description = c.Description
		config.Evaluators = evals

		url := fmt.Sprintf("http://localhost:5984/room_configurations/%s", config.ID)

		body, err := json.Marshal(config)
		if err != nil {
			log.Printf("Cannot marshal room configuration : %s", err.Error())
			return
		}

		req, err := http.NewRequest("PUT", url, bytes.NewReader(body))
		if err != nil {
			log.Printf("Error making request : %s", err.Error())
			return
		}

		client := &http.Client{}

		resp, err := client.Do(req)
		if err != nil {
			log.Printf("Error doing request : %s", err.Error())
			return
		}

		resp.Body.Close()
	}
}

func moveDevicesAndTypes(buildingList []structs.Building, roomList []structs.Room) {

	totalPortList, err := dbo.GetPorts()
	deviceClassList, err := dbo.GetDeviceClasses()
	microserviceList, err := dbo.GetMicroservices()
	endpointList, err := dbo.GetEndpoints()

	if err != nil {
		log.Printf("Failed to get info from old config db : %s", err.Error())
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

				portList[j].SourceDevice = port.Source
				portList[j].DestinationDevice = port.Destination
			}

			device.Ports = portList

			// Creating/moving the DeviceTypes here as well...
			deviceType := newstructs.DeviceType{}

			for _, t := range deviceClassList {
				if d.Class == t.Name {
					deviceType.ID = t.Name
					deviceType.Description = t.Description

					typePortList, _ := dbo.GetPortsByClass(t.Name)

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

						for _, m := range microserviceList {
							if command.Name == m.Address {
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
			url := fmt.Sprintf("http://localhost:5984/devices/%s", device.ID)

			body, err := json.Marshal(device)
			if err != nil {
				log.Printf("Cannot marshal device : %s", err.Error())
				return
			}

			req, err := http.NewRequest("PUT", url, bytes.NewReader(body))
			if err != nil {
				log.Printf("Error making request : %s", err.Error())
				return
			}

			client := &http.Client{}

			resp, err := client.Do(req)
			if err != nil {
				log.Printf("Error doing request : %s", err.Error())
				return
			}

			resp.Body.Close()

			// Send the DeviceType to Couch
			url2 := fmt.Sprintf("http://localhost:5984/device_types/%s", deviceType.ID)

			body2, err := json.Marshal(deviceType)
			if err != nil {
				log.Printf("Cannot marshal device type : %s", err.Error())
				return
			}

			req2, err := http.NewRequest("PUT", url2, bytes.NewReader(body2))
			if err != nil {
				log.Printf("Error making request : %s", err.Error())
				return
			}

			resp2, err := client.Do(req2)
			if err != nil {
				log.Printf("Error doing request : %s", err.Error())
				return
			}

			resp2.Body.Close()
		}
	}
}
