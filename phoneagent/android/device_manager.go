package android

import (
	"bufio"
	"context"
	"fmt"
	"os/exec"
	"strconv"
	"strings"
	"time"

	"github.com/spance/autoglm-go/phoneagent/definitions"

	"github.com/rs/zerolog/log"
)

const (
	adbPath = "adb"
)

func (r *ADBDevice) Connect(ctx context.Context, address string) (string, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	// Prepare the command
	cmdArgs := []string{"connect", address}

	log.Debug().Str("cmd", fmt.Sprintf("[Connect] run cmd: %s %s", adbPath, strings.Join(cmdArgs, " "))).Msg("")

	cmd := exec.CommandContext(ctx, adbPath, cmdArgs...)
	rawOutput, err := cmd.CombinedOutput()
	if err != nil {
		log.Error().Err(err).Msg("[Connect] run cmd failed")
		return fmt.Sprintf("Connect error: %v", err), err
	}

	log.Debug().Str("output", fmt.Sprintf("%s", rawOutput)).Msg("[Connect] raw output")

	output := string(rawOutput)

	lowerOutput := strings.ToLower(output)

	if strings.Contains(lowerOutput, "already connected") {
		return fmt.Sprintf("Already connected to %s", address), nil
	}
	if strings.Contains(lowerOutput, " connected") {
		return fmt.Sprintf("Connected to %s", address), nil
	}

	return fmt.Sprintf("Connection error: %s", strings.TrimSpace(output)), nil
}

func (r *ADBDevice) Disconnect(ctx context.Context, address string) (string, error) {

	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	cmdArgs := []string{"disconnect"}
	if len(address) > 0 {
		cmdArgs = append(cmdArgs, address)
	}
	log.Debug().Str("cmd", fmt.Sprintf("[Disconnect] run cmd: %s %s", adbPath, strings.Join(cmdArgs, " "))).Msg("")

	cmd := exec.CommandContext(ctx, adbPath, cmdArgs...)
	rawOutput, err := cmd.CombinedOutput()
	if err != nil {
		log.Error().Err(err).Msg("[Disconnect] run cmd failed")
		return fmt.Sprintf("Disconnect error: %v", err), err
	}

	log.Debug().Str("output", fmt.Sprintf("%s", rawOutput)).Msg("[Disconnect] raw output")

	return string(rawOutput), nil
}

func (r *ADBDevice) ListDevices(ctx context.Context) ([]definitions.DeviceInfo, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	cmdArgs := []string{"devices", "-l"}

	log.Debug().Str("cmd", fmt.Sprintf("[ListDevices] run cmd: %s %s", adbPath, strings.Join(cmdArgs, " "))).Msg("")
	cmd := exec.CommandContext(ctx, adbPath, cmdArgs...)

	rawOutput, err := cmd.CombinedOutput()
	if err != nil {
		log.Error().Err(err).Msg("[ListDevices] run cmd failed")
		return nil, err
	}
	output := string(rawOutput)

	var devices []definitions.DeviceInfo
	scanner := bufio.NewScanner(strings.NewReader(output))

	// Skip the first line (header)
	scanner.Scan()

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}

		parts := strings.Fields(line)
		if len(parts) < 2 {
			continue
		}

		deviceID := parts[0]
		status := parts[1]

		// Determine connection type
		var connType definitions.ConnectionType
		if strings.Contains(deviceID, ":") {
			connType = definitions.Remote
		} else if strings.Contains(deviceID, "emulator") {
			connType = definitions.USB // Emulator via USB
		} else {
			connType = definitions.USB
		}

		// Parse additional info
		var model string
		for _, part := range parts[2:] {
			if strings.HasPrefix(part, "model:") {
				model = strings.SplitN(part, ":", 2)[1]
				break
			}
		}

		devices = append(devices, definitions.DeviceInfo{
			DeviceID:       deviceID,
			Status:         status,
			ConnectionType: connType,
			Model:          model,
		})
	}

	return devices, nil
}

func (r *ADBDevice) GetDeviceInfo(ctx context.Context, deviceID string) (*definitions.DeviceInfo, error) {
	// TODO implement me
	panic("implement me")
}

func (r *ADBDevice) IsConnected(ctx context.Context, deviceID string) bool {
	// TODO implement me
	panic("implement me")
}

func (r *ADBDevice) EnableTCPIP(ctx context.Context, port int, deviceID string) error {
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	// Prepare the command
	var cmdArgs []string
	if len(deviceID) > 0 {
		cmdArgs = append(cmdArgs, "-s", deviceID)
	}
	cmdArgs = append(cmdArgs, "tcpip", strconv.Itoa(port))

	log.Debug().Str("cmd", fmt.Sprintf("[EnableTCPIP] run cmd: %s %s", adbPath, strings.Join(cmdArgs, " "))).Msg("")

	cmd := exec.CommandContext(ctx, adbPath, cmdArgs...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		log.Error().Err(err).Msg("[EnableTCPIP] run cmd failed")
		return err
	}

	// Process results
	if strings.Contains(strings.ToLower(string(output)), "restarting") {
		time.Sleep(time.Second * 5)
		return nil
	}
	return fmt.Errorf("error enabling TCP/IP: %v", err)
}

func (r *ADBDevice) GetDeviceIP(ctx context.Context, deviceID string) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// ---------- 1. adb shell ip route ----------
	var cmdArgs []string
	if len(deviceID) > 0 {
		cmdArgs = append(cmdArgs, "-s", deviceID)
	}
	cmdArgs = append(cmdArgs, "shell", "ip", "route")

	cmd := exec.CommandContext(ctx, adbPath, cmdArgs...)
	log.Debug().Str("cmd", fmt.Sprintf("[GetDeviceIP] run cmd1: %s %s", adbPath, strings.Join(cmdArgs, " "))).Msg("")

	output, err := cmd.CombinedOutput()
	if err != nil {
		log.Error().Err(err).Msg("[GetDeviceIP] run cmd1 failed")
		return "", err
	}

	log.Debug().Str("output", fmt.Sprintf("%s", string(output))).Msg("[GetDeviceIP] output1")

	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		if strings.Contains(line, "src") {
			parts := strings.Fields(line)
			for i, part := range parts {
				if part == "src" && i+1 < len(parts) {
					ip := parts[i+1]
					return ip, nil
				}
			}
		}
	}

	// ---------- 2. adb shell ip addr show wlan0 ----------
	cmdArgs = []string{}
	if len(deviceID) > 0 {
		cmdArgs = append(cmdArgs, "-s", deviceID)
	}
	cmdArgs = append(cmdArgs, "shell", "ip", "addr", "show", "wlan0")
	log.Debug().Str("cmd", fmt.Sprintf("[GetDeviceIP] run cmd2: %s %s", adbPath, strings.Join(cmdArgs, " "))).Msg("")

	cmd = exec.CommandContext(ctx, adbPath, cmdArgs...)
	output, err = cmd.CombinedOutput()
	if err != nil {
		log.Error().Err(err).Msg("[GetDeviceIP] run cmd2 failed")
		return "", err
	}
	log.Debug().Str("output", fmt.Sprintf("%s", string(output))).Msg("[GetDeviceIP] output2")

	lines = strings.Split(string(output), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.Contains(line, "inet ") {
			parts := strings.Fields(line)
			if len(parts) >= 2 {
				ip := strings.Split(parts[1], "/")[0]
				return ip, nil
			}
		}
	}
	return "", nil
}

func (r *ADBDevice) RestartServer(ctx context.Context) (string, error) {
	// TODO implement me
	panic("implement me")
}
