package udisks

import (
	"github.com/godbus/dbus/v5"
)

func prop[T any](path string, obj dbus.BusObject, p *T) error {
	v, err := obj.GetProperty(path)
	if err != nil {
		return err
	}
	t, ok := v.Value().(T)
	if !ok {
		return ErrInvalidPropertyFormat
	}
	*p = t
	return nil
}
func stringArrayPropertyFromByte(path string, obj dbus.BusObject, p *[]string) error {
	v, err := obj.GetProperty(path)
	if err != nil {
		return err
	}
	t, ok := v.Value().([][]uint8)
	if !ok {
		return ErrInvalidPropertyFormat
	}
	*p = make([]string, len(t))
	for i, v := range t {
		acc := *p
		acc[i] = string(v)
	}
	return nil
}
func stringProperty(path string, obj dbus.BusObject, p *string) error {
	v, err := obj.GetProperty(path)
	if err != nil {
		return err
	}

	var ok bool
	*p, ok = v.Value().(string)
	if !ok {
		return ErrInvalidPropertyFormat
	}

	return nil
}

func boolProperty(path string, obj dbus.BusObject, p *bool) error {
	v, err := obj.GetProperty(path)
	if err != nil {
		return err
	}

	var ok bool
	*p, ok = v.Value().(bool)
	if !ok {
		return ErrInvalidPropertyFormat
	}

	return nil
}

func uint64Property(path string, obj dbus.BusObject, p *uint64) error {
	v, err := obj.GetProperty(path)
	if err != nil {
		return err
	}

	var ok bool
	*p, ok = v.Value().(uint64)
	if !ok {
		return ErrInvalidPropertyFormat
	}

	return nil
}
func objGet(conn *dbus.Conn, property string, obj dbus.BusObject) (dbus.BusObject, error) {
	v, err := obj.GetProperty(property)
	if err != nil {
		return nil, err
	}

	var ok bool
	path, ok := v.Value().(dbus.ObjectPath)
	if !ok {
		return nil, ErrInvalidPropertyFormat
	}
	driveObj := conn.Object("org.freedesktop.UDisks2", path)

	return driveObj, nil
}

func (c *Client) getDrive(blkobj dbus.BusObject) (*Drive, error) {
	objDrv, err := objGet(c.conn, "org.freedesktop.UDisks2.Block.Drive", blkobj)
	if err != nil || objDrv.Path() == dbus.ObjectPath("/") {
		return nil, ErrInvalidDrive
	}

	return c.buildDrive(objDrv)
}

func (c *Client) buildDrive(objDrv dbus.BusObject) (*Drive, error) {
	drv := &Drive{
		Vendor:         "",
		Model:          "",
		Serial:         "",
		Id:             "",
		MediaRemovable: false,
		Ejectable:      false,
		MediaAvailable: false,
	}
	stringProperty("org.freedesktop.UDisks2.Drive.Vendor", objDrv, &drv.Vendor)
	stringProperty("org.freedesktop.UDisks2.Drive.Serial", objDrv, &drv.Serial)
	stringProperty("org.freedesktop.UDisks2.Drive.Model", objDrv, &drv.Model)
	stringProperty("org.freedesktop.UDisks2.Drive.Id", objDrv, &drv.Id)
	stringProperty("org.freedesktop.UDisks2.Drive.ConnectionBus", objDrv, &drv.ConnectionBus)
	stringProperty("org.freedesktop.UDisks2.Drive.Seat", objDrv, &drv.Seat)
	stringProperty("org.freedesktop.UDisks2.Drive.SiblingId", objDrv, &drv.SiblingId)
	boolProperty("org.freedesktop.UDisks2.Drive.MediaRemovable", objDrv, &drv.MediaRemovable)
	boolProperty("org.freedesktop.UDisks2.Drive.MediaAvailable", objDrv, &drv.MediaAvailable)
	boolProperty("org.freedesktop.UDisks2.Drive.Ejectable", objDrv, &drv.Ejectable)
	boolProperty("org.freedesktop.UDisks2.Drive.Removable", objDrv, &drv.Removable)
	uint64Property("org.freedesktop.UDisks2.Drive.Size", objDrv, &drv.Size)
	boolProperty("org.freedesktop.UDisks2.Drive.CanPowerOff", objDrv, &drv.CanPowerOff)
	isAtaSmartSupported := false
	boolProperty("org.freedesktop.UDisks2.Drive.Ata.SmartSupported", objDrv, &isAtaSmartSupported)
	if isAtaSmartSupported {
		ata := &Ata{}
		buildAtaSmart(objDrv, ata)
		drv.Ata = ata
	} else {
		nvmeController := &NVMeController{}
		if err := buildNVMeSmart(objDrv, nvmeController); err == nil {
			drv.NVMeController = nvmeController
		}
	}

	return drv, nil
}
func buildAtaSmart(objDrv dbus.BusObject, ata *Ata) error {
	if err := prop("org.freedesktop.UDisks2.Drive.Ata.SecurityFrozen", objDrv, &ata.SecurityFrozen); err != nil {
		return err
	}
	if err := prop("org.freedesktop.UDisks2.Drive.Ata.SmartSupported", objDrv, &ata.SmartSupported); err != nil {
		return err
	}
	if err := prop("org.freedesktop.UDisks2.Drive.Ata.SmartEnabled", objDrv, &ata.SmartEnabled); err != nil {
		return err
	}
	if err := prop("org.freedesktop.UDisks2.Drive.Ata.SmartFailing", objDrv, &ata.SmartFailing); err != nil {
		return err
	}
	if err := prop("org.freedesktop.UDisks2.Drive.Ata.PmSupported", objDrv, &ata.PmSupported); err != nil {
		return err
	}
	if err := prop("org.freedesktop.UDisks2.Drive.Ata.PmEnabled", objDrv, &ata.PmEnabled); err != nil {
		return err
	}
	if err := prop("org.freedesktop.UDisks2.Drive.Ata.ApmSupported", objDrv, &ata.ApmSupported); err != nil {
		return err
	}
	if err := prop("org.freedesktop.UDisks2.Drive.Ata.ApmEnabled", objDrv, &ata.ApmEnabled); err != nil {
		return err
	}
	if err := prop("org.freedesktop.UDisks2.Drive.Ata.WriteCacheSupported", objDrv, &ata.WriteCacheSupported); err != nil {
		return err
	}
	if err := prop("org.freedesktop.UDisks2.Drive.Ata.WriteCacheEnabled", objDrv, &ata.WriteCacheEnabled); err != nil {
		return err
	}
	if err := prop("org.freedesktop.UDisks2.Drive.Ata.ReadLookaheadSupported", objDrv, &ata.ReadLookaheadSupported); err != nil {
		return err
	}
	if err := prop("org.freedesktop.UDisks2.Drive.Ata.ReadLookaheadEnabled", objDrv, &ata.ReadLookaheadEnabled); err != nil {
		return err
	}
	if err := prop("org.freedesktop.UDisks2.Drive.Ata.SmartUpdated", objDrv, &ata.SmartUpdated); err != nil {
		return err
	}
	if err := prop("org.freedesktop.UDisks2.Drive.Ata.SmartPowerOnSeconds", objDrv, &ata.SmartPowerOnSeconds); err != nil {
		return err
	}
	if err := prop("org.freedesktop.UDisks2.Drive.Ata.SmartNumAttributesFailedInThePast", objDrv, &ata.SmartNumAttributesFailedInThePast); err != nil {
		return err
	}
	if err := prop("org.freedesktop.UDisks2.Drive.Ata.SmartSelftestPercentRemaining", objDrv, &ata.SmartSelftestPercentRemaining); err != nil {
		return err
	}
	if err := prop("org.freedesktop.UDisks2.Drive.Ata.AamVendorRecommendedValue", objDrv, &ata.AamVendorRecommendedValue); err != nil {
		return err
	}
	if err := prop("org.freedesktop.UDisks2.Drive.Ata.SecurityEraseUnitMinutes", objDrv, &ata.SecurityEraseUnitMinutes); err != nil {
		return err
	}
	if err := prop("org.freedesktop.UDisks2.Drive.Ata.SecurityEnhancedEraseUnitMinutes", objDrv, &ata.SecurityEnhancedEraseUnitMinutes); err != nil {
		return err
	}
	if err := prop("org.freedesktop.UDisks2.Drive.Ata.SmartTemperature", objDrv, &ata.SmartTemperature); err != nil {
		return err
	}
	if err := prop("org.freedesktop.UDisks2.Drive.Ata.SmartNumAttributesFailing", objDrv, &ata.SmartNumAttributesFailing); err != nil {
		return err
	}
	if err := prop("org.freedesktop.UDisks2.Drive.Ata.SmartSelftestStatus", objDrv, &ata.SmartSelftestStatus); err != nil {
		return err
	}
	if err := prop("org.freedesktop.UDisks2.Drive.Ata.SmartNumBadSectors", objDrv, &ata.SmartNumBadSectors); err != nil {
		return err
	}
	return nil
}
func buildNVMeSmart(objDrv dbus.BusObject, nvme *NVMeController) error {
	if err := prop("org.freedesktop.UDisks2.NVMe.Controller.SmartSelftestStatus", objDrv, &nvme.SmartSelftestStatus); err != nil {
		return err
	}
	if err := prop("org.freedesktop.UDisks2.NVMe.Controller.SanitizeStatus", objDrv, &nvme.SanitizeStatus); err != nil {
		return err
	}
	if err := prop("org.freedesktop.UDisks2.NVMe.Controller.FGUID", objDrv, &nvme.FGUID); err != nil {
		return err
	}
	if err := prop("org.freedesktop.UDisks2.NVMe.Controller.NVMeRevision", objDrv, &nvme.NVMeRevision); err != nil {
		return err
	}
	if err := prop("org.freedesktop.UDisks2.NVMe.Controller.State", objDrv, &nvme.State); err != nil {
		return err
	}
	if err := prop("org.freedesktop.UDisks2.NVMe.Controller.SmartPowerOnHours", objDrv, &nvme.SmartPowerOnHours); err != nil {
		return err
	}
	if err := prop("org.freedesktop.UDisks2.NVMe.Controller.UnallocatedCapacity", objDrv, &nvme.UnallocatedCapacity); err != nil {
		return err
	}
	if err := prop("org.freedesktop.UDisks2.NVMe.Controller.SmartUpdated", objDrv, &nvme.SmartUpdated); err != nil {
		return err
	}
	if err := prop("org.freedesktop.UDisks2.NVMe.Controller.SmartTemperature", objDrv, &nvme.SmartTemperature); err != nil {
		return err
	}
	if err := prop("org.freedesktop.UDisks2.NVMe.Controller.ControllerID", objDrv, &nvme.ControllerID); err != nil {
		return err
	}
	if err := prop("org.freedesktop.UDisks2.NVMe.Controller.SmartSelftestPercentRemaining", objDrv, &nvme.SmartSelftestPercentRemaining); err != nil {
		return err
	}
	if err := prop("org.freedesktop.UDisks2.NVMe.Controller.SanitizePercentRemaining", objDrv, &nvme.SanitizePercentRemaining); err != nil {
		return err
	}
	if err := prop("org.freedesktop.UDisks2.NVMe.Controller.SanitizePercentRemaining", objDrv, &nvme.SanitizePercentRemaining); err != nil {
		return err
	}
	if err := prop("org.freedesktop.UDisks2.NVMe.Controller.SubsystemNQN", objDrv, &nvme.SubsystemNQN); err != nil {
		return err
	}
	if err := prop("org.freedesktop.UDisks2.NVMe.Controller.SmartCriticalWarning", objDrv, &nvme.SmartCriticalWarning); err != nil {
		return err
	}

	return nil
}
