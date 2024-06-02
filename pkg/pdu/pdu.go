package pdu

import (
	"encoding/json"
	"fmt"
)

const (
	TYPE_DATA            = 0
	TYPE_ACK             = 1
	TYPE_HELLO           = 2
	TYPE_CONFIG_UPDATE   = 3
	TYPE_CONFIG_ACK      = 4
	TYPE_HEALTH_DATA     = 5
	TYPE_HEALTH_REQUEST  = 6
	TYPE_HEALTH_RESPONSE = 7
	TYPE_ERROR           = 8
	TYPE_TERMINATE       = 9
	TYPE_TERMINATE_ACK   = 10

	MAX_PDU_SIZE = 1024
)

type PDU struct {
	Mtype  uint8  `json:"mtype"`
	Length uint16 `json:"length"`
	Data   []byte `json:"data"`
}

func MakePduBuffer() []byte {
	return make([]byte, MAX_PDU_SIZE)
}

func NewPDU(mtype uint8, data []byte) *PDU {
	return &PDU{
		Mtype:  mtype,
		Length: uint16(len(data)),
		Data:   data,
	}
}

func (pdu *PDU) GetTypeAsString() string {
	switch pdu.Mtype {
	case TYPE_DATA:
		return "***DATA"
	case TYPE_ACK:
		return "****ACK"
	case TYPE_HELLO:
		return "**HELLO"
	case TYPE_CONFIG_UPDATE:
		return "CONFIG_UPDATE"
	case TYPE_CONFIG_ACK:
		return "CONFIG_ACK"
	case TYPE_HEALTH_DATA:
		return "HEALTH_DATA"
	case TYPE_HEALTH_REQUEST:
		return "HEALTH_REQUEST"
	case TYPE_HEALTH_RESPONSE:
		return "HEALTH_RESPONSE"
	case TYPE_ERROR:
		return "**ERROR"
	case TYPE_TERMINATE:
		return "TERMINATE"
	case TYPE_TERMINATE_ACK:
		return "TERMINATE_ACK"
	default:
		return "UNKNOWN"
	}
}

func (pdu *PDU) ToJsonString() string {
	jsonData, err := json.MarshalIndent(pdu, "", "    ")
	if err != nil {
		fmt.Println("Error marshaling JSON:", err)
		return "{}"
	}

	return string(jsonData)
}

func PduFromBytes(raw []byte) (*PDU, error) {
	//log.Printf("[pdu] Received PDU bytes: %s", string(raw))
	pdu := &PDU{}
	err := json.Unmarshal(raw, pdu)
	if err != nil {
		return nil, err
	}
	return pdu, nil
}

func PduToBytes(pdu *PDU) ([]byte, error) {
	//log.Printf("[pdu] PDU to be marshaled: %+v", pdu)
	return json.Marshal(pdu)
}
