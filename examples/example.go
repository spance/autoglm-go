package examples

import (
	"fmt"

	"github.com/spance/autoglm-go/constants"
	"github.com/spance/autoglm-go/examples/android"
	"github.com/spance/autoglm-go/phoneagent"
)

func CreateDevice(deviceType string) (phoneagent.Device, error) {
	switch deviceType {
	case constants.ADB:
		return &android.ADBDevice{}, nil
	case constants.IOS:
		// return &ios.IOSDevice{}, nil
		return nil, nil // 暂时不实现 iOS 设备
	default:
		return nil, fmt.Errorf("unknown device type: %v", deviceType)
	}
}
