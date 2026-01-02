package android

import (
	"autoglm-go/constants"
	"autoglm-go/phoneagent/definitions"
	"bytes"
	"context"
	"encoding/base64"
	"fmt"
	"github.com/google/uuid"
	"image"
	"image/png"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

type ADBDevice struct {
}

// createFallbackScreenshot creates a black fallback image when screenshot fails.
func createFallbackScreenshot(isSensitive bool) *definitions.Screenshot {
	const (
		defaultWidth  = 1080
		defaultHeight = 2400
	)

	// 创建黑色 RGBA 图像
	img := image.NewRGBA(image.Rect(0, 0, defaultWidth, defaultHeight))

	// 直接编码 PNG 并转 Base64
	var buf bytes.Buffer
	encoder := base64.NewEncoder(base64.StdEncoding, &buf)
	if err := png.Encode(encoder, img); err != nil {
		fmt.Printf("Error encoding fallback image: %v\n", err)
		return &definitions.Screenshot{
			Base64Data:  "",
			Width:       defaultWidth,
			Height:      defaultHeight,
			IsSensitive: isSensitive,
		}
	}
	encoder.Close() // 确保 Base64 编码完整

	return &definitions.Screenshot{
		Base64Data:  buf.String(),
		Width:       defaultWidth,
		Height:      defaultHeight,
		IsSensitive: isSensitive,
	}
}

func (r *ADBDevice) GetScreenshot(ctx context.Context, deviceID string) (*definitions.Screenshot, error) {
	tempPath := filepath.Join(os.TempDir(), fmt.Sprintf("screenshot_%s.png", uuid.New().String()))
	defer func() {
		_ = os.Remove(tempPath) // 确保临时文件被清理
	}()

	// 设置 adb 参数
	var cmdArgs []string
	if len(deviceID) > 0 {
		cmdArgs = append(cmdArgs, "-s", deviceID)
	}

	// 执行截屏
	screenshotArgs := append(cmdArgs, "shell", "screencap", "-p", "/sdcard/tmp.png")
	fmt.Printf("[GetScreenshot] run cmd1: %s %s\n", adbPath, strings.Join(screenshotArgs, " "))

	output, err := exec.CommandContext(ctx, adbPath, screenshotArgs...).CombinedOutput()
	if err != nil {
		fmt.Printf("Screenshot command error: %v, output: %s\n", err, output)
		return createFallbackScreenshot(false), nil
	}

	outputStr := string(output)
	if strings.Contains(outputStr, "Status: -1") || strings.Contains(outputStr, "Failed") {
		return createFallbackScreenshot(true), nil
	}

	// 拉取文件
	pullArgs := append(cmdArgs, "pull", "/sdcard/tmp.png", tempPath)
	fmt.Printf("[GetScreenshot] run cmd2: %s %s\n", adbPath, strings.Join(pullArgs, " "))

	if out, err := exec.CommandContext(ctx, adbPath, pullArgs...).CombinedOutput(); err != nil {
		fmt.Printf("Pull command error: %v, output: %s\n", err, out)
		return createFallbackScreenshot(false), nil
	}

	// 读取文件并 Base64 编码
	data, err := os.ReadFile(tempPath)
	if err != nil {
		fmt.Printf("Error reading image file: %v\n", err)
		return createFallbackScreenshot(false), nil
	}

	img, _, err := image.Decode(bytes.NewReader(data))
	if err != nil {
		fmt.Printf("Error decoding image: %v\n", err)
		return createFallbackScreenshot(false), nil
	}
	bounds := img.Bounds()
	width, height := bounds.Dx(), bounds.Dy()

	base64Data := base64.StdEncoding.EncodeToString(data)

	return &definitions.Screenshot{
		Base64Data:  base64Data,
		Width:       width, // 如果对宽和高无要求，可以省略解图
		Height:      height,
		IsSensitive: false,
	}, nil
}

func (r *ADBDevice) GetCurrentApp(ctx context.Context, deviceID string) (string, error) {
	// adb 前缀
	var cmdArgs []string
	if deviceID != "" {
		cmdArgs = append(cmdArgs, "-s", deviceID)
	}

	// 执行 adb shell dumpsys window
	args := append(cmdArgs, "shell", "dumpsys", "window")
	out, err := exec.Command(adbPath, args...).Output()
	if err != nil {
		fmt.Printf("Error running dumpsys window: %v, output: %s\n", err, out)
		return "", fmt.Errorf("failed to run dumpsys window: %w", err)
	}

	output := string(out)
	if output == "" {
		fmt.Printf("dumpsys window output is empty\n")
		return "", fmt.Errorf("no output from dumpsys window")
	}

	// 遍历每行查找焦点窗口信息
	for _, line := range strings.Split(output, "\n") {
		line = strings.TrimSpace(line)
		if strings.Contains(line, "mCurrentFocus") || strings.Contains(line, "mFocusedApp") {
			for appName, packageName := range constants.APP_PACKAGES_ANDROID {
				if strings.Contains(line, packageName) {
					return appName, nil
				}
			}
		}
	}

	// 未匹配到任何 app
	return "System Home", nil
}

func (r *ADBDevice) Tap(ctx context.Context, x, y int, deviceID string) error {
	adbPrefix := r.GetADBPrefix(deviceID)

	args := append(adbPrefix,
		"shell", "input", "tap", strconv.Itoa(x), strconv.Itoa(y),
	)

	cmd := exec.Command(args[0], args[1:]...)
	_, err := cmd.CombinedOutput()
	time.Sleep(time.Second * 1)
	return err
}

func (r *ADBDevice) DoubleTap(ctx context.Context, x, y int, deviceID string) error {
	const doubleTapInterval = 100 * time.Millisecond

	if err := r.Tap(ctx, x, y, deviceID); err != nil {
		return err
	}

	time.Sleep(doubleTapInterval)

	if err := r.Tap(ctx, x, y, deviceID); err != nil {
		return err
	}

	time.Sleep(time.Second * 1)
	return nil
}

func (r *ADBDevice) LongPress(ctx context.Context, x, y int, deviceID string) error {
	adbPrefix := r.GetADBPrefix(deviceID)

	args := append(adbPrefix,
		"shell", "input", "swipe",
		strconv.Itoa(x), strconv.Itoa(y),
		strconv.Itoa(x), strconv.Itoa(y),
		strconv.Itoa(3000),
	)

	cmd := exec.Command(args[0], args[1:]...)
	_, err := cmd.CombinedOutput()
	time.Sleep(time.Second * 1)
	return err
}

func (r *ADBDevice) Swipe(ctx context.Context, startX, startY, endX, endY int, deviceID string) error {
	distSq := (startX-endX)*(startX-endX) + (startY-endY)*(startY-endY)
	durationMs := int(float64(distSq) / 1000)
	durationMs = max(1000, min(durationMs, 2000)) // Clamp between 1000-2000ms
	adbPrefix := r.GetADBPrefix(deviceID)

	args := append(adbPrefix,
		"shell", "input", "swipe",
		strconv.Itoa(startX), strconv.Itoa(startY),
		strconv.Itoa(endX), strconv.Itoa(endY),
		strconv.Itoa(durationMs),
	)

	cmd := exec.Command(args[0], args[1:]...)
	_, err := cmd.CombinedOutput()
	time.Sleep(time.Second * 1)
	return err
}

func (r *ADBDevice) Back(ctx context.Context, deviceID string) error {
	adbPrefix := r.GetADBPrefix(deviceID)
	args := append(adbPrefix,
		"shell", "input", "keyevent", "4",
	)
	cmd := exec.Command(args[0], args[1:]...)
	_, err := cmd.CombinedOutput()
	time.Sleep(time.Second * 1)
	return err
}

func (r *ADBDevice) Home(ctx context.Context, deviceID string) error {
	adbPrefix := r.GetADBPrefix(deviceID)
	args := append(adbPrefix,
		"shell", "input", "keyevent", "KEYCODE_HOME",
	)
	cmd := exec.Command(args[0], args[1:]...)
	_, err := cmd.CombinedOutput()
	time.Sleep(time.Second * 1)
	return err
}

func (r *ADBDevice) LaunchApp(ctx context.Context, appName, deviceID string) (bool, error) {
	if _, ok := constants.APP_PACKAGES_ANDROID[appName]; !ok {
		return false, fmt.Errorf("app name %s not found in APP_PACKAGES", appName)
	}
	adbPrefix := r.GetADBPrefix(deviceID)
	packageName := constants.APP_PACKAGES_ANDROID[appName]

	args := append(adbPrefix,
		"shell",
		"monkey",
		"-p",
		packageName,
		"-c", "android.intent.category.LAUNCHER",
		"1",
	)

	cmd := exec.Command(args[0], args[1:]...)
	_, err := cmd.CombinedOutput()
	time.Sleep(time.Second * 1)
	return true, err
}

func (r *ADBDevice) TypeText(ctx context.Context, text, deviceID string) error {
	adbPrefix := r.GetADBPrefix(deviceID)

	encoded := base64.StdEncoding.EncodeToString([]byte(text))

	args := append(adbPrefix,
		"shell", "am", "broadcast",
		"-a", "ADB_INPUT_B64",
		"--es", "msg", encoded,
	)

	cmd := exec.Command(args[0], args[1:]...)
	_, err := cmd.CombinedOutput()
	return err
}

func (r *ADBDevice) ClearText(ctx context.Context, deviceID string) error {
	adbPrefix := r.GetADBPrefix(deviceID)

	args := append(adbPrefix,
		"shell", "am", "broadcast",
		"-a", "ADB_CLEAR_TEXT",
	)

	cmd := exec.Command(args[0], args[1:]...)
	_, err := cmd.CombinedOutput()
	return err
}

func (r *ADBDevice) DetectAndSetADBKeyboard(ctx context.Context, deviceID string) (string, error) {
	adbPrefix := r.GetADBPrefix(deviceID)

	// 获取当前输入法
	getArgs := append(adbPrefix,
		"shell", "settings", "get", "secure", "default_input_method",
	)

	out, err := exec.Command(getArgs[0], getArgs[1:]...).CombinedOutput()
	if err != nil {
		return "", err
	}

	currentIME := strings.TrimSpace(string(out))

	// 如未启用 ADB Keyboard，则切换
	if !strings.Contains(currentIME, "com.android.adbkeyboard/.AdbIME") {
		setArgs := append(adbPrefix,
			"shell", "ime", "set", "com.android.adbkeyboard/.AdbIME",
		)

		if _, err := exec.Command(setArgs[0], setArgs[1:]...).CombinedOutput(); err != nil {
			return "", err
		}
	}

	// 预热键盘（与 Python 逻辑一致）
	_ = r.TypeText(ctx, "", deviceID)

	return currentIME, nil
}

func (r *ADBDevice) RestoreKeyboard(ctx context.Context, ime, deviceID string) error {
	if ime == "" {
		return fmt.Errorf("IME cannot be empty")
	}

	adbPrefix := r.GetADBPrefix(deviceID)

	args := append(adbPrefix,
		"shell", "ime", "set", ime,
	)

	cmd := exec.Command(args[0], args[1:]...)
	_, err := cmd.CombinedOutput()
	return err
}

func (r *ADBDevice) GetADBPrefix(deviceID string) []string {
	if deviceID != "" {
		return []string{"adb", "-s", deviceID}
	}
	return []string{"adb"}
}
