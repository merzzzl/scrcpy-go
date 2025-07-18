# scrcpy-go

`scrcpy-go` is a simple, lightweight Go client/wrapper for [scrcpy v3.3.1](https://github.com/Genymobile/scrcpy/tree/v3.3.1).  
It supports sending control commands and receiving H.264 video stream from Android devices.

## 📚 Features

- Written in pure Go
- Supports the following control commands:
  - **Inject Keycode** — send hardware key events (e.g. volume, back)
  - **Inject Text** — input text as if typed from a keyboard
  - **Inject Touch Event** — emulate tap, swipe, and multitouch gestures
  - **Inject Scroll Event** — emulate scroll gestures
  - **Back or Screen On** — trigger back button or turn on screen
  - **Expand Notification Panel** — open notification drawer
  - **Expand Settings Panel** — open quick settings drawer
  - **Collapse Panels** — close any open system panels
  - **Get Clipboard** — retrieve device clipboard content
  - **Set Clipboard** — copy text to device clipboard
  - **Set Screen Power Mode** — turn screen on/off
  - **Rotate Device** — request device to rotate screen
  - **Create UHID Device** — create virtual HID input device (e.g. mouse, keyboard)
  - **Send UHID Input** — send HID report to virtual device
  - **Destroy UHID Device** — remove previously created virtual HID device
  - **Open Hard Keyboard Settings** — open system hardware keyboard settings screen
  - **Start App** — start an Android application by package name

- Decodes and displays H.264 video stream
- Connects via TCP to the scrcpy server running on the Android device

## ✨ Example

A basic terminal UI demo is available in the [`cmd/`](./cmd) directory.  
It mirrors the screen and allows input using the keyboard and mouse.

> **Note:** FFmpeg is required for video decoding.  
> Make sure `ffmpeg` is installed and available in your `PATH`.

```bash
go run ./cmd
```

![screenshot](README.gif)

## 📦 Make Features

- Automatically downloads and installs:
  - Android SDK command-line tools
  - Platform tools (ADB)
  - Android platform and build-tools
- Clones and syncs `scrcpy` submodule to tag `v3.3.1`
- Builds `scrcpy-server.jar` using the official `build_without_gradle.sh`
- Runs the server on the connected device using `adb` and `app_process`

### Targets

| Target          | Description                                                |
|-----------------|------------------------------------------------------------|
| `make sdk`      | Download and install Android SDK + required tools          |
| `make scrcpy-sync` | Clone and checkout scrcpy tag `v3.3.1`                |
| `make build`    | Build `scrcpy-server.jar` into `./build` folder            |
| `make run`      | Push server to device and launch it via `app_process`      |

## 🛠 Requirements

- Git
- GNU Make
- Linux or macOS
- Go 1.22+ (if you plan to use the Go client)
- `wget`, `unzip`, `ffmpeg`

## 📄 License

MIT