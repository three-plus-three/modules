package models

import (
	"encoding/json"
	"errors"
	"time"
)

type Object struct {
	ID        int64     `json:"id,omitempty" xorm:"id pk notnull"`
	Table     string    `json:"table_name,omitempty" xorm:"table_name notnull"`
	Name      string    `json:"name,omitempty" xorm:"name notnull"`
	Type      string    `json:"type,omitempty" xorm:"type notnull"`
	CreatedAt time.Time `json:"created_at,omitempty" xorm:"created_at created"`
	UpdatedAt time.Time `json:"updated_at,omitempty" xorm:"updated_at updated"`
}

func (*Object) TableName() string {
	return "tpt_objects"
}

var Objects = &Object{}

type NetworkDevice struct {
	ID          int64  `json:"id,omitempty" xorm:"id pk notnull"`
	Name        string `json:"name,omitempty" xorm:"name notnull"`
	FullName    string `json:"full_name,omitempty" xorm:"full_name"`
	ZhName      string `json:"zh_name,omitempty" xorm:"zh_name"`
	DisplayName string `json:"display_name,omitempty" xorm:"display_name notnull"`
	Description string `json:"description,omitempty" xorm:"description"`

	Type         string    `json:"type,omitempty" xorm:"type notnull"`
	ManagedType  string    `json:"managed_type,omitempty" xorm:"managed_type notnull"`
	Address      string    `json:"address,omitempty" xorm:"address notnull"`
	OsType       string    `json:"os_type,omitempty" xorm:"os_type"`
	DomainID     int64     `json:"domain_id,omitempty" xorm:"domain_id"`
	Owner        string    `json:"owner,omitempty" xorm:"owner"`
	OwnerPhone   string    `json:"owner_phone,omitempty" xorm:"owner_phone"`
	Category     string    `json:"category,omitempty" xorm:"category"`
	Level        int       `json:"level,omitempty" xorm:"level"`
	Location     string    `json:"location,omitempty" xorm:"location"`
	ManageURL    string    `json:"manage_url,omitempty" xorm:"manage_url"`
	DeviceType   int       `json:"device_type,omitempty" xorm:"device_type notnull"`
	EngineID     int64     `json:"engine_id,omitempty" xorm:"engine_id"`
	Oid          string    `json:"oid,omitempty" xorm:"oid"`
	Manufacturer string    `json:"manufacturer,omitempty" xorm:"manufacturer"`
	CreatedAt    time.Time `json:"created_at,omitempty" xorm:"created_at created"`
	UpdatedAt    time.Time `json:"updated_at,omitempty" xorm:"updated_at updated"`
}

func (dev *NetworkDevice) TableName() string {
	return "tpt_network_devices"
}

var NetworkDevices = &NetworkDevice{}

type NetworkLink struct {
	ID              int64     `json:"id,omitempty" xorm:"id pk notnull"`
	Name            string    `json:"name,omitempty" xorm:"name notnull"`
	FullName        string    `json:"full_name,omitempty"  xorm:"full_name notnull"`
	DisplayName     string    `json:"display_name,omitempty" xorm:"display_name notnull"`
	Description     string    `json:"description,omitempty" xorm:"description notnull"`
	ManagedType     string    `json:"managed_type,omitempty" xorm:"managed_type notnull"`
	LinkType        int       `json:"link_type,omitempty" xorm:"link_type notnull"`
	FromBased       bool      `json:"from_based,omitempty" xorm:"from_based"`
	Forward         bool      `json:"forward,omitempty" xorm:"forward"`
	Category        string    `json:"category,omitempty" xorm:"category"`
	Level           int       `json:"level,omitempty" xorm:"level"`
	ToDevice        int64     `json:"to_device,omitempty" xorm:"to_device"`
	ToIfIndex       int64     `json:"to_if_index,omitempty" xorm:"to_if_index"`
	ToPortID        int64     `json:"to_port_id,omitempty" xorm:"to_port_id"`
	FromDevice      int64     `json:"from_device,omitempty" xorm:"from_device"`
	FromIfIndex     int64     `json:"from_if_index,omitempty" xorm:"from_if_index"`
	FromPortID      int64     `json:"from_port_id,omitempty" xorm:"from_port_id"`
	CustomSpeedDown int64     `json:"custom_speed_down,omitempty" xorm:"custom_speed_down"`
	CustomSpeedUp   int64     `json:"custom_speed_up,omitempty" xorm:"custom_speed_up"`
	CreatedAt       time.Time `json:"created_at,omitempty" xorm:"created_at created"`
	UpdatedAt       time.Time `json:"updated_at,omitempty" xorm:"updated_at updated"`
}

func (dev *NetworkLink) TableName() string {
	return "tpt_network_links"
}

var NetworkLinks = &NetworkLink{}

type AccessParams struct {
	ID              int64     `json:"id,omitempty" xorm:"id pk notnull"`
	ManagedObjectID int64     `json:"managed_object_id,omitempty" xorm:"managed_object_id notnull"`
	Type            string    `json:"type,omitempty" xorm:"type"`
	Attributes      string    `json:"attributes,omitempty" xorm:"attributes"`
	OutOfBand       bool      `json:"out_of_band,omitempty" xorm:"out_of_band"`
	CreatedAt       time.Time `json:"created_at,omitempty" xorm:"created_at created"`
	UpdatedAt       time.Time `json:"updated_at,omitempty" xorm:"updated_at updated"`
}

func (ap *AccessParams) ToSnmpParams() (*SnmpParams, error) {
	if ap.Type != "snmp_param" {
		return nil, errors.New("access params isn't snmp params - " + ap.Type)
	}

	var snmp = &SnmpParams{}
	if err := json.Unmarshal([]byte(ap.Attributes), snmp); err != nil {
		return nil, errors.New("convert access params to snmp params fail, " + err.Error())
	}
	snmp.ID = ap.ID
	snmp.ManagedObjectID = ap.ManagedObjectID
	snmp.CreatedAt = ap.CreatedAt
	snmp.UpdatedAt = ap.UpdatedAt
	return snmp, nil
}

func (ap *AccessParams) TableName() string {
	return "tpt_access_params"
}

type SnmpParams struct {
	ID              int64     `json:"id,omitempty" xorm:"id pk notnull"`
	ManagedObjectID int64     `json:"managed_object_id,omitempty" xorm:"managed_object_id notnull"`
	Address         string    `json:"address,omitempty" xorm:"address"`
	Port            int       `json:"port,omitempty" xorm:"port"`
	Version         string    `json:"version,omitempty" xorm:"version"`
	WriteCommunity  string    `json:"write_community,omitempty" xorm:"write_community"`
	ReadCommunity   string    `json:"read_community,omitempty" xorm:"read_community"`
	SecModel        string    `json:"sec_model,omitempty" xorm:"sec_model"`
	ContextName     string    `json:"context_name,omitempty" xorm:"context_name"`
	EngineID        string    `json:"engine_id,omitempty" xorm:"engine_id"`
	Identifier      string    `json:"identifier,omitempty" xorm:"identifier"`
	SecName         string    `json:"sec_name,omitempty" xorm:"sec_name"`
	PrivProto       string    `json:"priv_proto,omitempty" xorm:"priv_proto"`
	PrivPass        string    `json:"priv_pass,omitempty" xorm:"priv_pass"`
	AuthProto       string    `json:"auth_proto,omitempty" xorm:"auth_proto"`
	AuthPass        string    `json:"auth_pass,omitempty" xorm:"auth_pass"`
	SecLevel        string    `json:"sec_level,omitempty" xorm:"sec_level"`
	MaxMsgSize      int       `json:"max_msg_size,omitempty" xorm:"max_msg_size"`
	CreatedAt       time.Time `json:"created_at,omitempty" xorm:"created_at created"`
	UpdatedAt       time.Time `json:"updated_at,omitempty" xorm:"updated_at updated"`
}

func (dev *SnmpParams) TableName() string {
	return "tpt_snmp_params"
}
