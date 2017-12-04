package snmp

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/runner-mei/snmpclient2"

	"github.com/three-plus-three/modules/ds"
)

var SysDescr snmpclient2.Oid = snmpclient2.MustParseOidFromString("1.3.6.1.2.1.1.1")
var IfDescr snmpclient2.Oid = snmpclient2.MustParseOidFromString("1.3.6.1.2.1.2.2.1.2")
var IfAdminStatus snmpclient2.Oid = snmpclient2.MustParseOidFromString("1.3.6.1.2.1.2.2.1.7")
var IfOperStatus snmpclient2.Oid = snmpclient2.MustParseOidFromString("1.3.6.1.2.1.2.2.1.8")
var IfName snmpclient2.Oid = snmpclient2.MustParseOidFromString("1.3.6.1.2.1.31.1.1.1.1")
var IfAlias snmpclient2.Oid = snmpclient2.MustParseOidFromString("1.3.6.1.2.1.31.1.1.1.18")

func Concat(oid snmpclient2.Oid, values ...int) snmpclient2.Oid {
	newValues := make([]int, len(oid.Value)+len(values))
	copy(newValues, oid.Value)
	copy(newValues[len(oid.Value):], values)
	return snmpclient2.Oid{Value: newValues}
}

//                 up(1),        -- ready to pass packets
//                 down(2),
//                 testing(3),   -- in some test mode
//                 unknown(4),   -- status can not be determined
//                               -- for some reason.
//                 dormant(5),
//                 notPresent(6),    -- some component is missing
//                 lowerLayerDown(7) -- down due to state of
//                                   -- lower-layer interface(s)
const (
	IF_STATUS_UP             = 1
	IF_STATUS_DOWN           = 2
	IF_STATUS_TESTING        = 3
	IF_STATUS_UNKNOW         = 4
	IF_STATUS_DORMANT        = 5
	IF_STATUS_NotPresent     = 6
	IF_STATUS_LOWERLAYERDOWN = 7
)

var interface_status_list = []string{"",
	"up",
	"down",
	"testing",
	"unknown",
	"dormant",
	"notPresent",
	"lowerLayerDown"}

func InterfaceStatusString(status int) string {
	if status > 0 && status < len(interface_status_list) {
		return interface_status_list[status]
	}
	return "unkown(" + strconv.FormatInt(int64(status), 10) + ")"
}

func NewSnmp(dev *ds.NetworkDevice, isWrite bool) (*snmpclient2.SNMP, error) {
	params, err := dev.SnmpParams()
	if err != nil {
		return nil, err
	}

	address := params.Address
	if address == "" {
		address = dev.Address
	}

	if strings.HasSuffix(address, "/32") {
		address = strings.TrimSuffix(address, "/32")
	}

	if params.Port <= 0 {
		address = address + ":161"
	} else {
		address = address + ":" + strconv.Itoa(params.Port)
	}

	version, err := snmpclient2.ParseVersion(params.Version)
	if err != nil {
		return nil, err
	}

	secLevel, err := snmpclient2.ParseSecurityLevel(params.SecLevel)
	if err != nil && (version != snmpclient2.V1 && version != snmpclient2.V2c) {
		return nil, err
	}
	authProto, err := snmpclient2.ParseAuthProtocol(params.AuthProto)
	if err != nil && (version != snmpclient2.V1 && version != snmpclient2.V2c) {
		return nil, err
	}
	privProto, err := snmpclient2.ParsePrivProtocol(params.PrivProto)
	if err != nil && (version != snmpclient2.V1 && version != snmpclient2.V2c) {
		return nil, err
	}

	community := params.ReadCommunity
	if isWrite {
		community = params.WriteCommunity
	}

	args := snmpclient2.Arguments{
		Version:          version, // SNMP version to use
		Timeout:          20 * time.Second,
		Retries:          3,
		MessageMaxSize:   params.MaxMsgSize,
		Community:        community,
		UserName:         params.SecName,
		SecurityLevel:    secLevel,           // Security level (V3 specific)
		AuthPassword:     params.AuthPass,    // Authentication protocol pass phrase (V3 specific)
		AuthProtocol:     authProto,          // Authentication protocol (V3 specific)
		PrivPassword:     params.AuthPass,    // Privacy protocol pass phrase (V3 specific)
		PrivProtocol:     privProto,          // Privacy protocol (V3 specific)
		SecurityEngineId: params.EngineID,    // Security engine ID (V3 specific)
		ContextEngineId:  params.EngineID,    // Context engine ID (V3 specific)
		ContextName:      params.ContextName, // Context name (V3 specific)
	}
	return snmpclient2.NewSNMP("udp", address, args)
}

type InterfaceStatus struct {
	Name  string `json:"if_name"`
	Descr string `json:"if_descr"`

	AdminStatus       int64  `json:"if_admin_status"`
	AdminStatusString string `json:"if_admin_status_label"`
	OpStatus          int64  `json:"if_oper_status"`
	OpStatusString    string `json:"if_oper_status_label"`
}

func ReadInterfaceStatus(dev *ds.NetworkDevice, ifIndex int) (*InterfaceStatus, error) {
	snmp, err := NewSnmp(dev, false)
	if err != nil {
		return nil, err
	}
	defer snmp.Close()

	ifDescr := Concat(IfDescr, ifIndex)
	ifAdminStatus := Concat(IfAdminStatus, ifIndex)
	ifOperStatus := Concat(IfOperStatus, ifIndex)
	ifName := Concat(IfName, ifIndex)

	pdu, err := snmp.GetRequest(snmpclient2.Oids{
		ifDescr,
		ifAdminStatus,
		ifOperStatus,
		ifName,
	})
	if err != nil {
		return nil, err
	}

	if pdu.ErrorStatus() != snmpclient2.NoError {
		return nil, fmt.Errorf(
			"failed to get system information - %s(%d)", pdu.ErrorStatus(), pdu.ErrorIndex())
	}

	vb := pdu.VariableBindings()
	descr := vb.MatchOid(ifDescr)
	admStatus := vb.MatchOid(ifAdminStatus)
	opStatus := vb.MatchOid(ifOperStatus)
	name := vb.MatchOid(ifName)

	var status = &InterfaceStatus{}
	if descr != nil {
		status.Descr, _ = snmpclient2.AsString(descr.Variable)
	}
	if name != nil {
		status.Name, _ = snmpclient2.AsString(name.Variable)
	}
	if admStatus != nil {
		status.AdminStatus = admStatus.Variable.Int()
		status.AdminStatusString = InterfaceStatusString(int(status.AdminStatus))
	}
	if opStatus != nil {
		status.OpStatus = opStatus.Variable.Int()
		status.OpStatusString = InterfaceStatusString(int(status.OpStatus))
	}
	return status, nil
}

func setInterfaceStatus(dev *ds.NetworkDevice, ifIndex, status int) error {
	snmp, err := NewSnmp(dev, true)
	if err != nil {
		return err
	}
	defer snmp.Close()

	ifAdminStatus := Concat(IfAdminStatus, ifIndex)

	pdu, err := snmp.SetRequest([]snmpclient2.VariableBinding{
		{Oid: ifAdminStatus, Variable: snmpclient2.NewInteger(int32(status))},
	})
	if err != nil {
		return err
	}

	if pdu.ErrorStatus() != snmpclient2.NoError {
		return fmt.Errorf(
			"failed to get system information - %s(%d)", pdu.ErrorStatus(), pdu.ErrorIndex())
	}

	//fmt.Println(pdu.VariableBindings())
	return nil
}

func CloseInterface(dev *ds.NetworkDevice, ifIndex int) error {
	return setInterfaceStatus(dev, ifIndex, IF_STATUS_DOWN)
}
func OpenInterface(dev *ds.NetworkDevice, ifIndex int) error {
	return setInterfaceStatus(dev, ifIndex, IF_STATUS_UP)
}
