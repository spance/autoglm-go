package ios

import (
	"autoglm-go/phoneagent/definitions"
	"context"
)

func (r *IOSDevice) Connect(ctx context.Context, address string) (string, error) {
	// TODO implement me
	panic("implement me")
}

func (r *IOSDevice) Disconnect(ctx context.Context, address string) (string, error) {
	// TODO implement me
	panic("implement me")
}

func (r *IOSDevice) ListDevices(ctx context.Context) ([]definitions.DeviceInfo, error) {
	// TODO implement me
	panic("implement me")
}

func (r *IOSDevice) GetDeviceInfo(ctx context.Context, deviceID string) (*definitions.DeviceInfo, error) {
	// TODO implement me
	panic("implement me")
}

func (r *IOSDevice) IsConnected(ctx context.Context, deviceID string) bool {
	// TODO implement me
	panic("implement me")
}

func (r *IOSDevice) EnableTCPIP(ctx context.Context, port int, deviceID string) error {
	// TODO implement me
	panic("implement me")
}

func (r *IOSDevice) GetDeviceIP(ctx context.Context, deviceID string) (string, error) {
	// TODO implement me
	panic("implement me")
}

func (r *IOSDevice) RestartServer(ctx context.Context) (string, error) {
	// TODO implement me
	panic("implement me")
}
