package zstack

import (
	"fmt"
	"time"

	"github.com/docker/machine/libmachine/drivers"
	"github.com/docker/machine/libmachine/log"
	"github.com/docker/machine/libmachine/mcnflag"
	"github.com/docker/machine/libmachine/state"

	"io/ioutil"
	"strings"

	"github.com/cnrancher/go-zstack/common"
	"github.com/cnrancher/go-zstack/infrastructure"
	"github.com/cnrancher/go-zstack/instance"
	"github.com/cnrancher/go-zstack/network/l3"
	"github.com/cnrancher/go-zstack/volume"
	"github.com/docker/machine/libmachine/ssh"
	"github.com/pkg/errors"
)

const (
	driverName  = "zstack"
	dockerPort  = 2376
	sshUser     = "docker"
	sshPassword = "tcuser"
)

//func NewDriver(hostName, storePath string) *Driver {
//	return &Driver{}
//}

func NewDriver(hostName, storePath string) drivers.Driver {
	return &Driver{
		BaseDriver: &drivers.BaseDriver{
			SSHUser:     sshUser,
			MachineName: hostName,
			StorePath:   storePath,
		}}
}

type Driver struct {
	*drivers.BaseDriver
	AccountName    string
	Password       string
	ZstackEndpoint string

	Name        string
	Description string

	ZoneName       string
	ClusterName    string
	PrimaryStorage string

	ImageName        string
	InstanceOffering string

	PublicKey []byte

	L3NetworkNames string

	SystemDiskOffering string

	DataDiskOffering string

	PhysicalHost string

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

func (d *Driver) getInstanceClient() *instance.Client {
	if d.instanceClient == nil {
		client := instance.NewInstanceClient(d.AccountName, d.Password, d.ZstackEndpoint)
		d.instanceClient = client
	}

	return d.instanceClient
}

// Create a host using the driver's config
func (d *Driver) Create() error {

	var (
		err error
	)

	if err := d.createKeyPair(); err != nil {
		return errors.Wrap(err, "Failed to create key pair.")
	}
	request := instance.CreateRequest{}
	request.Params.Name = d.MachineName
	//Following is for testing
	request.Params.ZoneUUID = d.ZoneName
	request.Params.ClusterUUID = d.ClusterName
	request.Params.ImageUUID = d.ImageName
	request.Params.L3NetworkUuids = d.getNetworks()
	request.Params.InstanceOfferingUUID = d.InstanceOffering
	request.Params.RootDiskOfferingUUID = d.SystemDiskOffering
	request.Params.DataDiskOfferingUUIDs = d.getDataDisks()
	request.Params.PrimaryStorageUUIDForRootVolume = d.PrimaryStorage
	request.Params.HostUUID = d.PhysicalHost
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

	inventory, err := d.instanceClient.QueryInstance(d.InstanceUUID)
	if err != nil {
		return err
	}
	d.IPAddress = d.getIP(inventory)
	if d.SSHUser == "" {
		d.SSHUser = sshUser
	}
	if d.SSHPassword == "" {
		d.SSHPassword = sshPassword
	}
	ssh.SetDefaultClient(ssh.Native)
	err = d.configInstance()
	if err != nil {
		return err
	}

	return nil
}

func (d *Driver) getNetworks() []string {
	var networkList []string

	for _, t := range strings.Split(d.L3NetworkNames, ",") {
		t = strings.TrimSpace(t)
		if t != "" {
			networkList = append(networkList, t)
		}
	}

	return networkList
}

func (d *Driver) getDataDisks() []string {
	var dataDiskList []string

	for _, t := range strings.Split(d.DataDiskOffering, ",") {
		t = strings.TrimSpace(t)
		if t != "" {
			dataDiskList = append(dataDiskList, t)
		}
	}

	return dataDiskList
}

func (d *Driver) createKeyPair() error {

	log.Debug("SSH key path: %s", d.GetSSHKeyPath())
	if err := ssh.GenerateSSHKey(d.GetSSHKeyPath()); err != nil {
		return err
	}

	publicKey, err := ioutil.ReadFile(d.GetSSHKeyPath() + ".pub")
	if err != nil {
		return err
	}
	d.PublicKey = publicKey

	return nil
}

func (d *Driver) uploadKeyPair(sshClient ssh.Client) error {
	command := fmt.Sprintf("mkdir -p ~/.ssh; echo '%s' >> ~/.ssh/authorized_keys", string(d.PublicKey))

	log.Debugf("Upload the public key with command: %s", command)

	output, err := sshClient.Output(command)

	log.Debugf("Upload command err, output: %v: %s", err, output)

	return err
}

// Mount the addtional disk
func (d *Driver) autoFdisk(sshClient ssh.Client) {
	script := fmt.Sprintf("cat > ~/machine_autofdisk.sh <<MACHINE_EOF\n%s\nMACHINE_EOF\n", instance.AutoFdiskScript)
	output, err := sshClient.Output(script)
	if d.SSHUser == "root" {
		output, err = sshClient.Output("sh ~/machine_autofdisk.sh")
	} else {
		output, err = sshClient.Output("sudo su root ~/machine_autofdisk.sh")
	}

	log.Infof("%s | Auto Fdisk command err, output: %v: %s", d.MachineName, err, output)
}

func (d *Driver) configInstance() error {
	ipAddr := d.IPAddress
	port, _ := d.GetSSHPort()
	tcpAddr := fmt.Sprintf("%s:%d", ipAddr, port)

	log.Infof("Waiting SSH service %s is ready to connect ...", tcpAddr)

	log.Infof("Uploading SSH keypair to %s ...", tcpAddr)

	auth := ssh.Auth{
		Passwords: []string{d.SSHPassword},
	}

	sshClient, err := ssh.NewClient(d.GetSSHUsername(), ipAddr, port, &auth)

	if err != nil {
		return err
	}

	err = d.uploadKeyPair(sshClient)
	if err != nil {
		return err
	}

	d.autoFdisk(sshClient)

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
			Name:   "zstack-account-name",
			Usage:  "The login zstack account",
			EnvVar: "ZSTACK_ACCOUNT_NAME",
			Value:  "",
		},
		mcnflag.StringFlag{
			Name:   "zstack-account-password",
			Usage:  "The login zstack password",
			EnvVar: "ZSTACK_ACCOUNT_PASSWORD",
			Value:  "",
		},
		mcnflag.StringFlag{
			Name:   "zstack-endpoint",
			Usage:  "The endpoint of zstack server",
			EnvVar: "ZSTACK_ENDPOINT",
			Value:  "",
		},
		mcnflag.StringFlag{
			Name:   "zstack-description",
			Usage:  "Optional. The detailed description of vm",
			EnvVar: "ZSTACK_DESCRIPTION",
			Value:  "",
		},
		mcnflag.StringFlag{
			Name:   "zstack-zone-name",
			Usage:  "Optional. Specify the zone name vm belongs to",
			EnvVar: "ZSTACK_ZONE_NAME",
			Value:  "",
		},
		mcnflag.StringFlag{
			Name:   "zstack-cluster-name",
			Usage:  "Optional. Specify the cluster name vm belongs to",
			EnvVar: "ZSTACK_CLUSTER_NAME",
			Value:  "",
		},
		mcnflag.StringFlag{
			Name:   "zstack-machine-name",
			Usage:  "Optional. Specify the machine where host will be create",
			EnvVar: "ZSTACK_MACHINE_NAME",
			Value:  "",
		},
		mcnflag.StringFlag{
			Name:   "zstack-image-name",
			Usage:  "The image to create the vm",
			EnvVar: "ZSTACK_IMAGE_NAME",
			Value:  "",
		},
		mcnflag.StringFlag{
			Name:   "zstack-instance-offering",
			Usage:  "Instance offering defined in zstack.",
			EnvVar: "ZSTACK_INSTANCE_OFFERING",
			Value:  "",
		},
		mcnflag.StringFlag{
			Name:   "zstack-network-name",
			Usage:  "L3 network name in zone.",
			EnvVar: "ZSTACK_NETWORK_NAME",
			Value:  "",
		},
		mcnflag.StringFlag{
			Name:   "zstack-system-disk-offering",
			Usage:  "Optional. Specify the root disk offering.",
			EnvVar: "ZSTACK_SYSTEM_DISK_OFFERING",
			Value:  "",
		},
		mcnflag.StringFlag{
			Name:   "zstack-data-disk-offering",
			Usage:  "Optional. Specify the data disk offering.",
			EnvVar: "ZSTACK_DATA_DISK_OFFERING",
			Value:  "",
		},
		mcnflag.StringFlag{
			Name:   "zstack-primary-storage",
			Usage:  "Optional. Specify the root volume.",
			EnvVar: "ZSTACK_PRIMARY_STORAGE",
			Value:  "",
		},
		mcnflag.StringFlag{
			Name:   "zstack-physical-host",
			Usage:  "Optional. Specify the root volume.",
			EnvVar: "ZSTACK_PHYSICAL_HOST",
			Value:  "",
		},
		mcnflag.StringFlag{
			Name:   "zstack-ssh-user",
			Usage:  "Optional. Specify the ssh password for root user.",
			EnvVar: "ZSTACK_SSH_USER",
			Value:  "",
		},
		mcnflag.StringFlag{
			Name:   "zstack-ssh-password",
			Usage:  "Optional. Specify the ssh password for root user.",
			EnvVar: "ZSTACK_SSH_PASSWORD",
			Value:  "",
		},
	}
}

// GetIP returns an IP or hostname that this host is available at
// e.g. 1.2.3.4 or docker-host-d60b70a14d3a.cloudapp.net
func (d *Driver) GetIP() (string, error) {
	inventory, err := d.getInstanceClient().QueryInstance(d.InstanceUUID)
	if err != nil {
		return "", errors.Wrap(err, "Error when getting instance.")
	}
	return d.getIP(inventory), nil
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
	i, err := d.getInstanceClient().QueryInstance(d.InstanceUUID)
	if err != nil {
		return 0, errors.Wrap(err, "Get error when get instance info from zstack.")
	}
	var rtnState state.State
	switch i.State {
	// "",
	case "":
		rtnState = state.None
		// "Running",
	case "Running":
		rtnState = state.Running
		// "Paused",
	case "Paused":
		rtnState = state.Paused
		// "Saved",
		// "Stopped",
	case "Stopped":
		rtnState = state.Stopped
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
	//init login in to get sessionID
	err := d.initClients()
	if err != nil {
		return err
	}

	return nil
}

// Remove a host
func (d *Driver) Remove() error {
	//First delete it
	async, err := d.getInstanceClient().DeleteInstance(d.InstanceUUID)
	if err != nil {
		return errors.Wrap(err, "Get error when sending delete instance request.")
	}
	responseStruct := instance.Response{}
	if err = async.QueryRealResponse(&responseStruct, 60*time.Second); err != nil {
		return errors.Wrap(err, "Get error when query response for zstack delete instance job.")
	}
	if responseStruct.Error != nil {
		return errors.Wrap(responseStruct.Error.WrapError(), "Get error when delete zstack instance.")
	}

	//Then expunge the instance
	async, err = d.getInstanceClient().ExpungeInstance(d.InstanceUUID)
	if err != nil {
		return errors.Wrap(err, "Get error when sending expunge instance request.")
	}
	responseStruct = instance.Response{}
	if err = async.QueryRealResponse(&responseStruct, 60*time.Second); err != nil {
		return errors.Wrap(err, "Get error when query response for zstack expunge instance job.")
	}
	if responseStruct.Error != nil {
		return errors.Wrap(responseStruct.Error.WrapError(), "Get error when expunge zstack instance.")
	}

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
	d.AccountName = opts.String("zstack-account-name")
	d.Password = opts.String("zstack-account-password")
	d.ZstackEndpoint = opts.String("zstack-endpoint")
	if d.AccountName == "" || d.Password == "" || d.ZstackEndpoint == "" {
		return errors.Errorf("AccountName, password and endpoint are required.")
	}
	d.Description = opts.String("zstack-description")

	//Following configuration is about where the host is
	d.ZoneName = opts.String("zstack-zone-name")
	d.ClusterName = opts.String("zstack-cluster-name")
	if d.ClusterName != "" && d.ZoneName != "" {
		log.Warn("The cluster name has been set so the zone name will be omitted.")
	}
	//d.MachineName = opts.String("machine-name")
	//if d.MachineName != "" && (d.ClusterName != "" || d.ZoneName != "") {
	//	log.Warn("The machine name has been set so the cluster name and zone name will be omitted.")
	//}

	//Following configuration is about what the host is like
	d.ImageName = opts.String("zstack-image-name")
	if d.ImageName == "" {
		return errors.Errorf("The image name is required.")
	}
	d.InstanceOffering = opts.String("zstack-instance-offering")
	if d.InstanceOffering == "" {
		return errors.Errorf("The instance offering is required.")
	}
	d.L3NetworkNames = opts.String("zstack-network-name")
	if d.L3NetworkNames == "" {
		return errors.Errorf("The network configuration is required.")
	}

	//if the image is the type of ISO, then this argument is required
	d.SystemDiskOffering = opts.String("zstack-system-disk-offering")
	if d.SystemDiskOffering == "" {
		return errors.Errorf("The Root/System disk size is required.")
	}

	d.PrimaryStorage = opts.String("zstack-primary-storage")
	d.DataDiskOffering = opts.String("zstack-data-disk-offering")
	d.PhysicalHost = opts.String("zstack-physical-host")

	d.SSHPassword = opts.String("zstack-ssh-password")
	d.SSHUser = opts.String("zstack-ssh-user")

	return nil
}

// Start a host
func (d *Driver) Start() error {
	async, err := d.getInstanceClient().StartInstance(d.InstanceUUID)
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
	async, err := d.getInstanceClient().StopInstance(d.InstanceUUID, instance.StopInstanceTypeGrace)
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

func (d *Driver) getIP(instance *instance.VMInstanceInventory) string {
	if len(instance.VMNics) > 0 {
		return instance.VMNics[0].IP
	}
	return ""
}
