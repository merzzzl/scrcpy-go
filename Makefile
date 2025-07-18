ANDROID_SDK_ROOT := $(CURDIR)/android-sdk
CMDLINE_TOOLS_VERSION := 11076708_latest

BUILD_TOOLS_VERSION := 35.0.0
ANDROID_PLATFORM := android-35

CMDLINE_TOOLS_ARCHIVE_LINUX := commandlinetools-linux-$(CMDLINE_TOOLS_VERSION).zip
CMDLINE_TOOLS_ARCHIVE_MAC := commandlinetools-mac-$(CMDLINE_TOOLS_VERSION).zip

UNAME_S := $(shell uname -s)

ifeq ($(UNAME_S),Linux)
	CMDLINE_TOOLS_ARCHIVE := $(CMDLINE_TOOLS_ARCHIVE_LINUX)
	CMDLINE_TOOLS_URL := https://dl.google.com/android/repository/$(CMDLINE_TOOLS_ARCHIVE_LINUX)
endif

ifeq ($(UNAME_S),Darwin)
	CMDLINE_TOOLS_ARCHIVE := $(CMDLINE_TOOLS_ARCHIVE_MAC)
	CMDLINE_TOOLS_URL := https://dl.google.com/android/repository/$(CMDLINE_TOOLS_ARCHIVE_MAC)
endif

SDKMANAGER := $(ANDROID_SDK_ROOT)/cmdline-tools/latest/bin/sdkmanager
ADB := $(ANDROID_SDK_ROOT)/platform-tools/adb

.PHONY: sdk scrcpy-sync build run

cmdline-tools:
	@rm -rf $(ANDROID_SDK_ROOT)
	@mkdir -p $(ANDROID_SDK_ROOT)
	@wget $(CMDLINE_TOOLS_URL)
	@unzip -o $(CMDLINE_TOOLS_ARCHIVE) -d $(ANDROID_SDK_ROOT)/cmdline-tools-temp
	@mkdir -p $(ANDROID_SDK_ROOT)/cmdline-tools/latest
	@mv $(ANDROID_SDK_ROOT)/cmdline-tools-temp/cmdline-tools/* $(ANDROID_SDK_ROOT)/cmdline-tools/latest/
	@rm -rf $(ANDROID_SDK_ROOT)/cmdline-tools-temp
	@rm $(CMDLINE_TOOLS_ARCHIVE)

sdk: cmdline-tools
	@$(SDKMANAGER) --sdk_root=$(ANDROID_SDK_ROOT) --install \
		platform-tools \
		"platforms;$(ANDROID_PLATFORM)" \
		"build-tools;$(BUILD_TOOLS_VERSION)"

scrcpy-sync:
	@git submodule update --init scrcpy
	@cd scrcpy && \
		git fetch --tags --force && \
		git checkout --detach tags/v3.3.1
	@echo "scrcpy → v3.3.1"

build: scrcpy-sync
	@rm -rf build
	@mkdir -p build
	@BUILD_DIR=build ANDROID_HOME=$(ANDROID_SDK_ROOT) ./scrcpy/server/build_without_gradle.sh

run:
	@$(ADB) push build/scrcpy-server /data/local/tmp/scrcpy-server.jar
	@$(ADB) forward tcp:10000 localabstract:scrcpy_00010000
	@$(ADB) shell CLASSPATH=/data/local/tmp/scrcpy-server.jar app_process / com.genymobile.scrcpy.Server 3.3.1 scid=10000 tunnel_forward=true audio=false log_level=verbose