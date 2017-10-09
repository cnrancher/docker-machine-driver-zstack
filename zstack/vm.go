package zstack

import (
	"fmt"
	"time"

	"github.com/docker/machine/libmachine/drivers"
	"github.com/docker/machine/libmachine/log"
	"github.com/docker/machine/libmachine/mcnflag"
	"github.com/docker/machine/libmachine/state"

	"github.com/orangedeng/go-zstack/common"
	"github.com/orangedeng/go-zstack/infrastructure"
	"github.com/orangedeng/go-zstack/instance"
	"github.com/orangedeng/go-zstack/network/l3"
	"github.com/orangedeng/go-zstack/volume"
	"github.com/pkg/errors"
)

const (
	driverName = "zstack"
	dockerPort = 2376
)

func NewDriver() *Driver {
	return &Driver{}
}

type Driver struct {
	*drivers.BaseDriver
	AccountName    string
	Password       string
	ZstackEndpoint string

	Name        string
	Description string

	ZoneName    string
	ClusterName string
	MachineName string

	ImageName        string
	InstanceOffering string

	L3NetworkName string
	IP            string

	SystemDiskOffering string
	SystemDiskSize     int

	DataDiskOffering string
	DataDiskSize     int

	SSHPassword string

	InstanceUUID string

	instanceClient         *instance.Client
	hostClient             *infrastructure.Host
	imageClient            *instance.Image
	clusterClient          *infrastructure.Cluster
	zoneCLient             *infrastructure.Zone
	instanceOfferingClient *instance.Offering
	l3NetworkClient        *l3.Client
	volumeOfferingClient   *volume.Offering
}

func (d *Driver) cleanup() error {
	defer func() {
		d.hostClient = nil
		d.imageClient = nil
		d.clusterClient = nil
		d.zoneCLient = nil
		d.instanceOfferingClient = nil
		d.l3NetworkClient = nil
		d.volumeOfferingClient = nil
		d.instanceClient = nil
	}()
	return d.instanceClient.Cleanup()
}

func (d *Driver) initClients() error {
	if d.instanceClient != nil {
		return nil
	}
	commonClient := common.Client{}
	if err := commonClient.Init(d.AccountName, d.Password, d.ZstackEndpoint); err != nil {
		log.Error(err)
		return err
	}
	d.instanceClient = &instance.Client{
		Client: commonClient,
	}
	d.hostClient = &infrastructure.Host{
		Client: commonClient,
	}
	d.imageClient = &instance.Image{
		Client: commonClient,
	}
	d.clusterClient = &infrastructure.Cluster{
		Client: commonClient,
	}
	d.zoneCLient = &infrastructure.Zone{
		Client: commonClient,
	}
	d.instanceOfferingClient = &instance.Offering{
		Client: commonClient,
	}
	d.l3NetworkClient = &l3.Client{
		Client: commonClient,
	}
	d.volumeOfferingClient = &volume.Offering{
		Client: commonClient,
	}
	return nil
}

// Create a host using the driver's config
func (d *Driver) Create() error {
	request := instance.CreateRequest{}
	request.Params.Name = d.Name
	//Following is for testing
	request.Params.ZoneUUID = d.ZoneName
	request.Params.ClusterUUID = d.ClusterName
	async, err := d.instanceClient.CreateInstance(request)
	if err != nil {
		return errors.Wrap(err, "Get error when create vm instance in zstack.")
	}
	response := instance.Response{}
	if err = async.QueryRealResponse(&response, 60*time.Second); err != nil {
		return errors.Wrap(err, "Get error when create vm instance in zstack.")
	}
	if response.Error != nil {
		return errors.Wrap(response.Error.WrapError(), "Get error when create vm instance in zstack.")
	}
	d.InstanceUUID = response.Inventory.UUID
	return nil
}

// DriverName returns the name of the driver
func (d *Driver) DriverName() string {
	return driverName
}

// GetCreateFlags returns the mcnflag.Flag slice representing the flags
// that can be set, their descriptions and defaults.
func (d *Driver) GetCreateFlags() []mcnflag.Flag {
	//ToDo fulfill the usage for each flag
	return []mcnflag.Flag{
		mcnflag.StringFlag{
			Name:   "account-name",
			Usage:  "",
			EnvVar: "ZSTACK_ACCOUNT_NAME",
			Value:  "",
		},
		mcnflag.StringFlag{
			Name:   "account-passowrd",
			Usage:  "",
			EnvVar: "ZSTACK_ACCOUNT_PASSWORD",
			Value:  "",
		},
		mcnflag.StringFlag{
			Name:   "zstack-endpoint",
			Usage:  "",
			EnvVar: "ZSTACK_ENDPOINT",
			Value:  "",
		},
		mcnflag.StringFlag{
			Name:   "description",
			Usage:  "",
			EnvVar: "ZSTACK_DESCRIPTION",
			Value:  "",
		},
		mcnflag.StringFlag{
			Name:   "zone-name",
			Usage:  "",
			EnvVar: "ZSTACK_ZONE_NAME",
			Value:  "",
		},
		mcnflag.StringFlag{
			Name:   "cluster-name",
			Usage:  "",
			EnvVar: "ZSTACK_CLUSTER_NAME",
			Value:  "",
		},
		mcnflag.StringFlag{
			Name:   "machine-name",
			Usage:  "Optional. Specific the machine where host will be create",
			EnvVar: "ZSTACK_MACHINE_NAME",
			Value:  "",
		},
		mcnflag.StringFlag{
			Name:   "image-name",
			Usage:  "Image name in zstack cluster.",
			EnvVar: "ZSTACK_IMAGE_NAME",
			Value:  "",
		},
		mcnflag.StringFlag{
			Name:   "instance-offering",
			Usage:  "Instance offering defined in zstack.",
			EnvVar: "ZSTACK_INSTANCE_OFFERING",
			Value:  "",
		},
		mcnflag.StringFlag{
			Name:   "network-name",
			Usage:  "L3 network name in zone.",
			EnvVar: "ZSTACK_NETWORK_NAME",
			Value:  "",
		},
		mcnflag.StringFlag{
			Name:   "ip",
			Usage:  "Specific the IP of this host.",
			EnvVar: "ZSTACK_HOST_IP",
			Value:  "",
		},
		mcnflag.StringFlag{
			Name:   "system-disk-offering",
			Usage:  "",
			EnvVar: "ZSTACK_SYSTEM_DISK_OFFERING",
			Value:  "",
		},
		mcnflag.IntFlag{
			Name:   "system-disk-size",
			Usage:  "",
			EnvVar: "ZSTACK_SYSTEM_DISK_SIZE",
			Value:  0,
		},
		mcnflag.StringFlag{
			Name:   "data-disk-offering",
			Usage:  "",
			EnvVar: "ZSTACK_DATA_DISK_OFFERING",
			Value:  "",
		},
		mcnflag.IntFlag{
			Name:   "data-disk-size",
			Usage:  "",
			EnvVar: "ZSTACK_DATA_DISK_SIZE",
			Value:  0,
		},
	}
}

// GetIP returns an IP or hostname that this host is available at
// e.g. 1.2.3.4 or docker-host-d60b70a14d3a.cloudapp.net
func (d *Driver) GetIP() (string, error) {
	instance, err := d.instanceClient.QueryInstance(d.InstanceUUID)
	if err != nil {
		return "", errors.Wrap(err, "Error when getting instance.")
	}
	return d.getIP(instance)
}

// GetSSHHostname returns hostname for use with ssh
func (d *Driver) GetSSHHostname() (string, error) {
	return d.GetIP()
}

// GetURL returns a Docker compatible host URL for connecting to this host
// e.g. tcp://1.2.3.4:2376
func (d *Driver) GetURL() (string, error) {
	ip, err := d.GetIP()
	if err != nil {
		return "", err
	}
	if ip == "" {
		return "", nil
	}
	return fmt.Sprintf("tcp://%s:%d", ip, dockerPort), nil
}

// GetState returns the state that the host is in (running, stopped, etc)
func (d *Driver) GetState() (state.State, error) {
	i, err := d.instanceClient.QueryInstance(d.InstanceUUID)
	if err != nil {
		return 0, errors.Wrap(err, "Get error when get instance info from zstack.")
	}
	var rtnState state.State
	switch i.State {
	// "",
	case "":
		rtnState = 0
		// "Running",
	case "Running":
		rtnState = 1
		// "Paused",
	case "Paused":
		rtnState = 2
		// "Saved",
		// "Stopped",
	case "Stopped":
		rtnState = 4
		// "Stopping",
		// "Starting",
		// "Error",
		// "Timeout",
	}
	return rtnState, nil
}

// Kill stops a host forcefully
func (d *Driver) Kill() error {
	async, err := d.instanceClient.StopInstance(d.InstanceUUID, instance.StopInstanceTypeCold)
	if err != nil {
		return errors.Wrap(err, "Get error when sending kill instance request.")
	}
	responseStruct := instance.Response{}
	if err = async.QueryRealResponse(&responseStruct, 60*time.Second); err != nil {
		return errors.Wrap(err, "Get error when query response for zstack kill instance job.")
	}
	if responseStruct.Error != nil {
		return errors.Wrap(responseStruct.Error.WrapError(), "Get error when kill zstack instance.")
	}
	//ToDo change it to new type
	if responseStruct.Inventory.State != "Stopped" {
		return errors.New("the target Instance state is not as expect,'Stopped'")
	}
	return nil
}

// PreCreateCheck allows for pre-create operations to make sure a driver is ready for creation
func (d *Driver) PreCreateCheck() error {
	return d.initClients()
}

// Remove a host
func (d *Driver) Remove() error {
	//ToDo
	d.instanceClient.DeleteInstance(d.InstanceUUID)
	return nil
}

// Restart a host. This may just call Stop(); Start() if the provider does not
// have any special restart behaviour.
func (d *Driver) Restart() error {
	return nil
}

// SetConfigFromFlags configures the driver with the object that was returned
// by RegisterCreateFlags
func (d *Driver) SetConfigFromFlags(opts drivers.DriverOptions) error {
	d.AccountName = opts.String("account-name")
	d.Password = opts.String("account-passowrd")
	d.ZstackEndpoint = opts.String("zstack-endpoint")
	if d.AccountName == "" || d.Password == "" || d.ZstackEndpoint == "" {
		return errors.Errorf("AccountName, password and endpoint are required.")
	}
	d.Description = opts.String("description")

	//Following configuration is about where the host is
	d.ZoneName = opts.String("zone-name")
	d.ClusterName = opts.String("cluster-name")
	if d.ClusterName != "" && d.ZoneName != "" {
		log.Warn("The cluster name has been set so the zone name will be omitted.")
	}
	d.MachineName = opts.String("machine-name")
	if d.MachineName != "" && (d.ClusterName != "" || d.ZoneName != "") {
		log.Warn("The machine name has been set so the cluster name and zone name will be omitted.")
	}

	//Following configuration is about what the host is like
	d.ImageName = opts.String("image-name")
	if d.ImageName == "" {
		return errors.Errorf("The image name is required.")
	}
	d.InstanceOffering = opts.String("instance-offering")
	if d.InstanceOffering == "" {
		return errors.Errorf("The instance offering is required.")
	}
	d.L3NetworkName = opts.String("network-name")
	if d.L3NetworkName == "" {
		return errors.Errorf("The network configuration is required.")
	}
	d.IP = opts.String("id")
	d.SystemDiskOffering = opts.String("system-disk-offering")
	d.SystemDiskSize = opts.Int("system-disk-size")
	if d.SystemDiskOffering != "" && d.SystemDiskSize > 0 {
		log.Warn("The system disk size will be omitted because the system disk offering is set.")
	}
	d.DataDiskOffering = opts.String("data-disk-offering")
	d.DataDiskSize = opts.Int("data-disk-size")
	if d.DataDiskOffering != "" && d.DataDiskSize > 0 {
		log.Warn("The data disk size will be omitted because the data disk offering is set.")
	}

	return nil
}

// Start a host
func (d *Driver) Start() error {
	async, err := d.instanceClient.StartInstance(d.InstanceUUID)
	if err != nil {
		return errors.Wrap(err, "Get error when sending start instance request.")
	}
	responseStruct := instance.Response{}
	if err = async.QueryRealResponse(&responseStruct, 60*time.Second); err != nil {
		return errors.Wrap(err, "Get error when querying response for zstack start instance job.")
	}
	if responseStruct.Error != nil {
		return errors.Wrap(responseStruct.Error.WrapError(), "Get error when start zstack instance.")
	}
	//ToDo change it to new type
	if responseStruct.Inventory.State != "Running" {
		return errors.New("the target Instance state is not as expect,'Running'")
	}
	return nil
}

// Stop a host gracefully
func (d *Driver) Stop() error {
	async, err := d.instanceClient.StopInstance(d.InstanceUUID, instance.StopInstanceTypeGrace)
	if err != nil {
		return errors.Wrap(err, "Get error when sending stop instance request.")
	}
	responseStruct := instance.Response{}
	if err = async.QueryRealResponse(&responseStruct, 60*time.Second); err != nil {
		return errors.Wrap(err, "Get error when query response for zstack stop instance job.")
	}
	if responseStruct.Error != nil {
		return errors.Wrap(responseStruct.Error.WrapError(), "Get error when stop zstack instance.")
	}
	//ToDo change it to new type
	if responseStruct.Inventory.State != "Stopped" {
		return errors.New("the target Instance state is not as expect,'Stopped'")
	}
	return nil
}

func (d *Driver) getIP(instance *instance.VMInstanceInventory) (string, error) {
	if len(instance.VMNics) > 0 {
		return instance.VMNics[0].IP, nil
	}
	return "", nil
}
