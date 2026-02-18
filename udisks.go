package udisks

import (
	"fmt"
	"strings"

	"github.com/godbus/dbus/v5"

	"github.com/godbus/dbus/v5/introspect"
)

type Client struct {
	conn *dbus.Conn
}

type Drive struct {
	Vendor         string
	Model          string
	Serial         string
	Id             string
	MediaRemovable bool
	Ejectable      bool
	MediaAvailable bool
	ConnectionBus  string
	SiblingId      string
	Seat           string
	Removable      bool
	Size           uint64
	CanPowerOff    bool
	NVMeController *NVMeController
	Ata            *Ata
}
type BlockDevices []*BlockDevice

func (b BlockDevices) ByDevice(device string) *BlockDevice {
	for _, v := range b {
		if v.Device == device {
			return v
		}
	}
	return nil
}

type BlockDevice struct {
	UUID                string
	Device              string
	Id                  string
	IdUsage             string
	IdLabel             string
	IdType              string
	Drive               *Drive
	Filesystems         []Filesystem
	Symlinks            []string
	CryptoBackingDevice *CryptoBackingDevice
}

func (b *BlockDevice) IsMounted() bool {
	for _, v := range b.Filesystems {
		if v.IsMounted() {
			return true
		}
	}
	return false
}

type CryptoBackingDevice struct {
	Path                string
	CleartextDevicePath string
	HintEncryptionType  string
	MetadataSize        uint64
}

type Filesystem struct {
	MountPoints []string
	Size        uint64
}

func (f Filesystem) IsMounted() bool {
	return len(f.MountPoints) > 0
}
func NewClient() (*Client, error) {
	c := &Client{}
	conn, err := dbus.ConnectSystemBus()
	if err != nil {
		return nil, err
	}
	c.conn = conn

	return c, nil
}

// PowerOff unmounts all blockdevices on the device, lock any unlocked encrypted containers and then powers off the device
func (c Client) PowerOff(d *Drive) error {
	if !d.CanPowerOff {
		return ErrPowerOffNotSupported
	}
	blocks, err := c.BlockDevices()
	if err != nil {
		return err
	}

	for _, b := range blocks {
		if b.CryptoBackingDevice != nil {
			if b.CryptoBackingDevice.CleartextDevicePath != "" {
				cryptoDrive := blocks.ByDevice(b.CryptoBackingDevice.Path)
				if cryptoDrive.Drive != nil && cryptoDrive.Drive.Id == d.Id {
					if b.IsMounted() {
						if err := c.UnmountBlockDevice(b.Device); err != nil {
							return fmt.Errorf("%w:  %w", ErrUnmountFailed, err)
						}
					}
					if err := c.LockCryptoDevice(b.CryptoBackingDevice.Path); err != nil {
						return fmt.Errorf("%w:  %w", ErrUnmountFailed, err)
					}
				}
			}
		} else {
			if b.Drive != nil && b.Drive.Id == d.Id {
				if len(b.Filesystems) > 0 {
					if err := c.UnmountBlockDevice(b.Id); err != nil {
						return fmt.Errorf("%w: %w", ErrUnmountFailed, err)
					}
				}
			}
		}
	}
	path := "/org/freedesktop/UDisks2/drives/" + strings.ReplaceAll(d.Id, "-", "_")
	powerOffObj := c.conn.Object("org.freedesktop.UDisks2", dbus.ObjectPath(path))
	opt := map[string]interface{}{
		"auth.no_user_interaction": true,
	}
	return powerOffObj.Call("org.freedesktop.UDisks2.Drive.PowerOff", 0, &opt).Err
}
func (c *Client) LockCryptoDevice(path string) error {
	opt := map[string]interface{}{
		"auth.no_user_interaction": true,
	}
	obj := c.conn.Object("org.freedesktop.UDisks2", dbus.ObjectPath(path))
	result := obj.Call("org.freedesktop.UDisks2.Encrypted.Lock", 0, opt)
	return result.Err
}
func (c *Client) UnmountBlockDevice(path string) error {
	opt := map[string]interface{}{
		"auth.no_user_interaction": true,
	}
	obj := c.conn.Object("org.freedesktop.UDisks2", dbus.ObjectPath(path))
	result := obj.Call("org.freedesktop.UDisks2.Filesystem.Unmount", 0, opt)
	return result.Err
}

// BlockDevices returns the list of all block devices known to UDisks
func (c *Client) BlockDevices() (BlockDevices, error) {
	conn := c.conn
	var list []string
	var filter map[string]interface{}
	obj := conn.Object("org.freedesktop.UDisks2", "/org/freedesktop/UDisks2/Manager")
	err := obj.Call("org.freedesktop.UDisks2.Manager.GetBlockDevices", 0, &filter).Store(&list)
	if err != nil {
		return BlockDevices{}, err
	}

	bdevs := []*BlockDevice{}
	for _, bd := range list {
		dev := &BlockDevice{}
		bdevs = append(bdevs, dev)
		obj = conn.Object("org.freedesktop.UDisks2", dbus.ObjectPath(bd))
		dev.Device = bd
		stringProperty("org.freedesktop.UDisks2.Block.IdUUID", obj, &dev.UUID)
		stringProperty("org.freedesktop.UDisks2.Block.Id", obj, &dev.Id)
		stringProperty("org.freedesktop.UDisks2.Block.IdUsage", obj, &dev.IdUsage)
		stringProperty("org.freedesktop.UDisks2.Block.IdLabel", obj, &dev.IdLabel)
		stringProperty("org.freedesktop.UDisks2.Block.IdType", obj, &dev.IdType)
		stringArrayPropertyFromByte("org.freedesktop.UDisks2.Block.Symlinks", obj, &dev.Symlinks)

		var props map[string]dbus.Variant
		cbd, err := objGet(conn, "org.freedesktop.UDisks2.Block.CryptoBackingDevice", obj)
		if err == nil {
			cbd.Call("org.freedesktop.DBus.Properties.GetAll", 0, "org.freedesktop.UDisks2.Encrypted").Store(&props)
			if len(props) != 0 {
				dev.CryptoBackingDevice = &CryptoBackingDevice{
					Path:               string(cbd.Path()),
					HintEncryptionType: props["HintEncryptionType"].Value().(string),
					MetadataSize:       props["MetadataSize"].Value().(uint64),
				}

				clearPath := props["CleartextDevice"].Value()
				if val, ok := clearPath.(dbus.ObjectPath); ok && val.IsValid() {
					dev.CryptoBackingDevice.CleartextDevicePath = string(val)
				}
			}
		}

		dev.Drive, err = c.getDrive(obj)

		obj.Call("org.freedesktop.DBus.Properties.GetAll", 0, "org.freedesktop.UDisks2.Filesystem").Store(&props)

		if len(props) != 0 {

			fs := Filesystem{}
			va := props["MountPoints"].Value()
			if va != nil {
				arr := va.([][]byte)
				for i := 0; i < len(arr); i++ {
					mpsa := arr[i]
					mpa := string(mpsa[0 : len(mpsa)-1])

					fs.MountPoints = append(fs.MountPoints, mpa)
				}
			}
			va = props["Size"].Value()
			if va != nil {
				fs.Size = va.(uint64)
			}

			dev.Filesystems = append(dev.Filesystems, fs)
		}
	}

	return bdevs, nil
}

// Drives returns the list of all block devices known to UDisks
func (c *Client) Drives() ([]*Drive, error) {
	drives := []*Drive{}
	node, err := introspect.Call(c.conn.Object("org.freedesktop.UDisks2", "/org/freedesktop/UDisks2/drives"))
	if err != nil {
		return drives, err
	}

	for _, ch := range node.Children {
		path := "/org/freedesktop/UDisks2/drives/" + ch.Name
		obj := c.conn.Object("org.freedesktop.UDisks2", dbus.ObjectPath(path))
		drv, err := c.buildDrive(obj)
		if err != nil {
			return drives, err
		}
		drives = append(drives, drv)
	}

	return drives, nil
}

func (c *Client) DriveById(name string) (*Drive, error) {
	//we need to replace - with _ to give the correct id
	path := "/org/freedesktop/UDisks2/drives/" + strings.ReplaceAll(name, "-", "_")
	obj := c.conn.Object("org.freedesktop.UDisks2", dbus.ObjectPath(path))
	drv, err := c.buildDrive(obj)
	if err != nil {
		return drv, err
	}
	if drv.Id == "" {
		return drv, ErrDriveNotFound
	}
	return drv, err
}
func (c *Client) BlockDevicesOnDrive(id string) ([]*BlockDevice, error) {
	blocks, err := c.BlockDevices()
	if err != nil {
		return []*BlockDevice{}, err
	}
	blockDevices := make([]*BlockDevice, 0)
	for _, b := range blocks {
		if b.CryptoBackingDevice != nil {
			if b.CryptoBackingDevice.CleartextDevicePath != "" {
				cryptoDrive := blocks.ByDevice(b.CryptoBackingDevice.Path)
				if cryptoDrive.Drive != nil && cryptoDrive.Drive.Id == id {
					blockDevices = append(blockDevices, b)
				}
			}
		} else {
			if b.Drive != nil && b.Drive.Id == id {
				blockDevices = append(blockDevices, b)
			}
		}
	}
	return blockDevices, nil
}
