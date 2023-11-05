package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"strconv"
	"time"

	"dapp/rollups"
)

var (
	infolog = log.New(os.Stderr, "[ info ]  ", log.Lshortfile)
	errlog  = log.New(os.Stderr, "[ error ] ", log.Lshortfile)
)

func decodePayload(payload string) (map[string]interface{}, error) {
	st, e := rollups.Hex2Str(payload)

	if e != nil {
		return nil, e
	}

	jsonInfo := make(map[string]interface{})

	e = json.Unmarshal([]byte(st), &jsonInfo)

	if e != nil {
		return nil, e
	}

	return jsonInfo, nil
}

func sendJsonNotice(data map[string]interface{}) error {
	st, e := json.Marshal(data)

	if e != nil {
		return nil
	}

	hex := rollups.Str2Hex(string(st))

	notice := &rollups.NoticeRequest{
		Payload: hex,
	}
	_, err := rollups.SendNotice(notice)

	return err
}

func sendJsonReport(data any) error {
	st, e := json.Marshal(data)

	if e != nil {
		return nil
	}

	hex := rollups.Str2Hex(string(st))

	report := &rollups.ReportRequest{
		Payload: hex,
	}
	_, err := rollups.SendReport(report)

	return err
}

func HandleAction(input map[string]interface{}, address Address, timestamp int) error {
	method, exists := input["method"].(string)

	if !exists {
		return fmt.Errorf("method invalid")
	}

	switch method {
	case "CreateGroup":
		membersInterface, e := input["members"]

		if !e {
			return fmt.Errorf("invalid input")
		}

		membersList, e := membersInterface.([]interface{})

		members := make([]Address, 0)

		for _, v := range membersList {
			members = append(members, Address(v.(string)))
		}

		infolog.Println(members)

		group, id := CreateGroup(members)
		infolog.Println(group, id)

		res := map[string]interface{}{
			"group": group,
			"id":    id,
		}

		infolog.Println(res)

		groups[id] = group
		stateTransitions[id] = make([]StateTransition, 0)

		return sendJsonNotice(res)
	case "SubmitR1":
		r1Value, e := input["r1Value"]

		if !e {
			return fmt.Errorf("invalid input")
		}

		id, e := input["id"]

		if !e {
			return fmt.Errorf("invalid input")
		}

		return SubmitR1(id.(string), B64ToBigInt(r1Value.(string)), address, "")

	case "SubmitR2":
		r2Value, e := input["r2Value"]

		if !e {
			return fmt.Errorf("invalid input")
		}

		id, e := input["id"]

		if !e {
			return fmt.Errorf("invalid input")
		}

		return SubmitR2(id.(string), B64ToBigInt(r2Value.(string)), address, "")
	case "SubmitTransition":
		id, e := input["id"]

		if !e {
			return fmt.Errorf("invalid input")
		}

		action, e := input["action"]

		if !e {
			return fmt.Errorf("invalid input")
		}

		return SubmitTransition(id.(string), action.(string), address, int64(timestamp))
	}

	return nil
}

func HandleRead(input map[string]interface{}) error {
	method, exists := input["method"].(string)

	if !exists {
		return fmt.Errorf("method invalid")
	}

	switch method {
	case "groups":
		return sendJsonReport(groups)
	case "transitions":
		id, e := input["id"]

		if !e {
			return fmt.Errorf("invalid input")
		}

		groupsTransitions, e := stateTransitions[id.(string)]

		if !e {
			return fmt.Errorf("group does not exist")
		}

		infolog.Println(groupsTransitions)

		return sendJsonReport(groupsTransitions)
	}

	return nil
}

func HandleAdvance(data *rollups.AdvanceResponse) error {
	dataMarshal, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("HandleAdvance: failed marshaling json: %w", err)
	}

	decodedPayload, err := decodePayload(data.Payload)

	if err != nil {
		return fmt.Errorf("error decoding payload: %w", err)
	}

	infolog.Println("Received advance request data", string(dataMarshal))
	infolog.Println("Address", data.Metadata.MsgSender)
	infolog.Println("Payload", decodedPayload)

	return HandleAction(decodedPayload, Address(data.Metadata.MsgSender), int(data.Metadata.Timestamp))
}

func HandleInspect(data *rollups.InspectResponse) error {
	dataMarshal, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("HandleInspect: failed marshaling json: %w", err)
	}
	infolog.Println("Received inspect request data", string(dataMarshal))
	infolog.Println(rollups.Hex2Str(data.Payload))

	decodedPayload, err := decodePayload(data.Payload)

	if err != nil {
		return fmt.Errorf("error decoding payload: %w", err)
	}

	return HandleRead(decodedPayload)

}

func Handler(response *rollups.FinishResponse) error {
	var err error

	switch response.Type {
	case "advance_state":
		data := new(rollups.AdvanceResponse)
		if err = json.Unmarshal(response.Data, data); err != nil {
			return fmt.Errorf("Handler: Error unmarshaling advance:", err)
		}
		err = HandleAdvance(data)
	case "inspect_state":
		data := new(rollups.InspectResponse)
		if err = json.Unmarshal(response.Data, data); err != nil {
			return fmt.Errorf("Handler: Error unmarshaling inspect:", err)
		}
		err = HandleInspect(data)
	}
	return err
}

func main() {
	finish := rollups.FinishRequest{"accept"}

	for true {
		infolog.Println("Sending finish")
		res, err := rollups.SendFinish(&finish)
		if err != nil {
			errlog.Println("Error: error making http request: ", err)
      time.Sleep(3 * time.Second)
      continue
		}
		infolog.Println("Received finish status ", strconv.Itoa(res.StatusCode))

		if res.StatusCode == 202 {
			infolog.Println("No pending rollup request, trying again")
		} else {

			resBody, err := ioutil.ReadAll(res.Body)
			if err != nil {
				errlog.Println("Error: could not read response body: ", err)
			}

			var response rollups.FinishResponse
			err = json.Unmarshal(resBody, &response)
			if err != nil {
				errlog.Println("Error: unmarshaling body:", err)
			}

			finish.Status = "accept"
			err = Handler(&response)
			if err != nil {
				errlog.Println(err)
				finish.Status = "reject"
			}
		}
	}
}
