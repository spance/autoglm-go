package ios

import (
	"autoglm-go/phoneagent/definitions"
	"context"
)

type IOSDevice struct{}

func (r *IOSDevice) GetScreenshot(ctx context.Context, deviceID string) (*definitions.Screenshot, error) {
	// TODO implement me
	panic("implement me")
}

func (r *IOSDevice) GetCurrentApp(ctx context.Context, deviceID string) (string, error) {
	// TODO implement me
	panic("implement me")
}

func (r *IOSDevice) Tap(ctx context.Context, x, y int, deviceID string) error {
	// TODO implement me
	panic("implement me")
}

func (r *IOSDevice) DoubleTap(ctx context.Context, x, y int, deviceID string) error {
	// TODO implement me
	panic("implement me")
}

func (r *IOSDevice) LongPress(ctx context.Context, x, y int, deviceID string) error {
	// TODO implement me
	panic("implement me")
}

func (r *IOSDevice) Swipe(ctx context.Context, startX, startY, endX, endY int, deviceID string) error {
	// TODO implement me
	panic("implement me")
}

func (r *IOSDevice) Back(ctx context.Context, deviceID string) error {
	// TODO implement me
	panic("implement me")
}

func (r *IOSDevice) Home(ctx context.Context, deviceID string) error {
	// TODO implement me
	panic("implement me")
}

func (r *IOSDevice) LaunchApp(ctx context.Context, appName, deviceID string) (bool, error) {
	// TODO implement me
	panic("implement me")
}

func (r *IOSDevice) TypeText(ctx context.Context, text, deviceID string) error {
	// TODO implement me
	panic("implement me")
}

func (r *IOSDevice) ClearText(ctx context.Context, deviceID string) error {
	// TODO implement me
	panic("implement me")
}

func (r *IOSDevice) DetectAndSetADBKeyboard(ctx context.Context, deviceID string) (string, error) {
	// TODO implement me
	panic("implement me")
}

func (r *IOSDevice) RestoreKeyboard(ctx context.Context, ime, deviceID string) error {
	// TODO implement me
	panic("implement me")
}
