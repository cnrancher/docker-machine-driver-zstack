package instance

import "github.com/cnrancher/go-zstack/common"

const (
	createInstanceURI  = "/zstack/v1/vm-instances"
	deleteInstanceURI  = "/zstack/v1/vm-instances/{uuid}"
	operateInstanceURI = "/zstack/v1/vm-instances/{uuid}/actions"
	queryInstanceURI   = "/zstack/v1/vm-instances/{uuid}"
	queryInstancesURI  = "/zstack/v1/vm-instances"
	//StopInstanceTypeGrace stop instance gracefully
	StopInstanceTypeGrace StopInstanceType = "grace"
	//StopInstanceTypeCold stop instance immediately, equal to power off.
	StopInstanceTypeCold StopInstanceType = "cold"
)

const (
	InstanceDefaultTimeout = 120
	DefaultWaitForInterval = 5
	timeout                = 300
)

type StopInstanceType string

type VmInstanceStatus string

type CreateRequest struct {
	Params struct {
		Name                            string   `json:"name,omitempty"`
		InstanceOfferingUUID            string   `json:"instanceOfferingUuid,omitempty"`
		ImageUUID                       string   `json:"imageUuid,omitempty"`
		L3NetworkUuids                  []string `json:"l3NetworkUuids,omitempty"`
		Type                            string   `json:"type,omitempty"`
		RootDiskOfferingUUID            string   `json:"rootDiskOfferingUuid,omitempty"`
		DataDiskOfferingUUIDs           []string `json:"dataDiskOfferingUuids,omitempty"`
		ZoneUUID                        string   `json:"zoneUuid,omitempty"`
		ClusterUUID                     string   `json:"clusterUuid,omitempty"`
		HostUUID                        string   `json:"hostUuid,omitempty"`
		PrimaryStorageUUIDForRootVolume string   `json:"primaryStorageUuidForRootVolume,omitempty"`
		Description                     string   `json:"description,omitempty"`
		DefaultL3NetworkUUID            string   `json:"DefaultL3NetworkUuid,omitempty"`
		ResourceUUID                    string   `json:"resourceUuid,omitempty"`
		Strategy                        string   `json:"strategy,omitempty"`
	} `json:"params,omitempty"`
	common.Tags `json:",inline"`
}

type CreateOfferingRequest struct {
	Params struct {
		CpuNum     int   `json:"cpuNum,omitempty"`
		MemorySize int64 `json:"memorySize,omitempty"`
	}
}

type Response struct {
	Error     *common.Error        `json:"error,omitempty"`
	Inventory *VMInstanceInventory `json:"inventory,omitempty"`
}

type VMInstanceInventory struct {
	common.ResourceBase `json:",inline"`

	Name                 string    `json:"name,omitempty"`
	Description          string    `json:"description,omitempty"`
	ZoneUUID             string    `json:"zoneUuid,omitempty"`
	ClusterUUID          string    `json:"clusterUuid,omitempty"`
	ImageUUID            string    `json:"imageUuid,omitempty"`
	HostUUID             string    `json:"hostUuid,omitempty"`
	LastHostUUID         string    `json:"lastHostUuid,omitempty"`
	InstanceOfferingUUID string    `json:"instanceOfferingUuid,omitempty"`
	RootVolumeUUID       string    `json:"rootVolumeUuid,omitempty"`
	Platform             string    `json:"platform,omitempty"`
	DefaultL3NetworkUUID string    `json:"defaultL3NetworkUuid,omitempty"`
	Type                 string    `json:"type,omitempty"`
	HypervisorType       string    `json:"hypervisorType,omitempty"`
	MemorySize           int64     `json:"memorySize,omitempty"`
	CPUNum               int       `json:"cpuNum,omitempty"`
	CPUSpeed             int64     `json:"cpuSpeed,omitempty"`
	AllocatorStrategy    string    `json:"allocatorStrategy,omitempty"`
	State                string    `json:"state,omitempty"`
	VMNics               []*VMNic  `json:"vmNics,omitempty"`
	AllVolumes           []*Volume `json:"allVolumes,omitempty"`
}

type VMNic struct {
	common.ResourceBase `json:",inline"`

	VMInstanceUUID string `json:"vmInstanceUuid,omitempty"`
	L3NetworkUUID  string `json:"l3NetworkUuid,omitempty"`
	IP             string `json:"ip,omitempty"`
	MAC            string `json:"mac,omitempty"`
	Netmask        string `json:"netmask,omitempty"`
	Gateway        string `json:"gateway,omitempty"`
	MetaData       string `json:"metaData,omitempty"`
	DeviceID       int    `json:"deviceId,omitempty"`
}

type Volume struct {
	common.ResourceBase `json:",inline"`

	Name               string `json:"name,omitempty"`
	Description        string `json:"description,omitempty"`
	PrimaryStorageUUID string `json:"primaryStorageUuid,omitempty"`
	VMInstanceUUID     string `json:"vmInstanceUuid,omitempty"`
	DiskOfferingUUID   string `json:"diskOfferingUuid,omitempty"`
	RootImageUUID      string `json:"rootImageUuid,omitempty"`
	InstallPath        string `json:"installPath,omitempty"`
	Type               string `json:"type,omitempty"`
	Format             string `json:"format,omitempty"`
	Size               int64  `json:"size,omitempty"`
	ActualSize         int64  `json:"actualSize,omitempty"`
	DeviceID           int    `json:"deviceId,omitempty"`
	State              string `json:"state,omitempty"`
	Status             string `json:"status,omitempty"`
	IsShareable        bool   `json:"isShareable,omitempty"`
}

type DeleteInstanceResponse struct {
	Error *common.Error `json:"error"`
}

type ExpungeInstanceRequest struct {
	ExpungeVMInstance map[string]string `json:"expungeVmInstance"`
	common.Tags       `json:",inline"`
}

type QueryInstanceResponse struct {
	Error       *common.Error          `json:"error,omitempty"`
	Inventories []*VMInstanceInventory `json:"inventories,omitempty"`
}

type StartInstanceRequest struct {
	StartVMInstance map[string]string `json:"startVmInstance"`
	common.Tags     `json:",inline"`
}

type StopInstanceRequest struct {
	StopVMInstance struct {
		Type StopInstanceType `json:"type,omitempty"`
	} `json:"stopVmInstance,omitempty"`
	common.Tags `json:",inline"`
}

type RestartInstanceRequest struct {
	RebootVMInstance map[string]string `json:"rebootVmInstance,omitempty"`
	common.Tags      `json:",inline"`
}
