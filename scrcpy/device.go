package scrcpy

import "io"

const deviceNameLength = 64

func (svr *server) ReadDeviceInfo() (deviceName string, screenSize size, err error) {
	buf := make([]byte, deviceNameLength+4)
	if _, err = io.ReadFull(svr.deviceConn, buf); err != nil {
		return
	}

	deviceName = string(buf[:deviceNameLength])
	screenSize.width = uint16(buf[deviceNameLength])<<8 | uint16(buf[deviceNameLength+1])
	screenSize.height = uint16(buf[deviceNameLength+2])<<8 | uint16(buf[deviceNameLength+3])
	return
}
