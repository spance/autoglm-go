package definitions

type ConnectionType string

const (
	USB    ConnectionType = "usb"
	WiFi   ConnectionType = "wifi"
	Remote ConnectionType = "remote"
)

type DeviceInfo struct {
	DeviceID       string         `json:"device_id"`
	Status         string         `json:"status"`
	ConnectionType ConnectionType `json:"connection_type"`
	Model          string         `json:"model,omitempty"`
	AndroidVersion string         `json:"android_version,omitempty"`
}

// Screenshot represents a captured screenshot.
type Screenshot struct {
	Base64Data  string `json:"base64_data"`
	Width       int    `json:"width"`
	Height      int    `json:"height"`
	IsSensitive bool   `json:"is_sensitive"`
}
