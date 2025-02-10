package xpc

import (
	"bytes"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path"
	"strings"
	"testing"

	"github.com/danielpaulus/go-ios/ios/coredevice"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"howett.net/plist"
)

func TestEmptyDictionary(t *testing.T) {
	b, _ := os.ReadFile(path.Join("xpc_empty_dict.bin"))

	res, err := DecodeMessage(bytes.NewReader(b))
	assert.NoError(t, err)
	assert.Equal(t, Message{
		Flags: AlwaysSetFlag,
		Body:  map[string]interface{}{},
	}, res)
}

func TestDictionary(t *testing.T) {
	b, _ := os.ReadFile(path.Join("xpc_dict.bin"))

	res, err := DecodeMessage(bytes.NewReader(b))
	assert.NoError(t, err)
	assert.Equal(t, Message{
		Flags: AlwaysSetFlag | DataFlag | HeartbeatRequestFlag,
		Body: map[string]interface{}{
			"CoreDevice.CoreDeviceDDIProtocolVersion": int64(0),
			"CoreDevice.action":                       map[string]interface{}{},
			"CoreDevice.coreDeviceVersion": map[string]interface{}{
				"components":              []interface{}{uint64(0x15c), uint64(0x1), uint64(0x0), uint64(0x0), uint64(0x0)},
				"originalComponentsCount": int64(2),
				"stringValue":             "348.1",
			},
			"CoreDevice.deviceIdentifier":  "A7DD28AC-2911-4549-811D-85917B9AC72F",
			"CoreDevice.featureIdentifier": "com.apple.coredevice.feature.launchapplication",
			"CoreDevice.input": map[string]interface{}{
				"applicationSpecifier": map[string]interface{}{
					"bundleIdentifier": map[string]interface{}{
						"_0": "xxx.xxxxxxxxx.xxxxxxxx",
					},
				},
				"options": map[string]interface{}{
					"arguments": []interface{}{},
					"environmentVariables": map[string]interface{}{
						"TERM": "xterm-256color",
					},
					"platformSpecificOptions":       base64Decode("YnBsaXN0MDDQCAAAAAAAAAEBAAAAAAAAAAEAAAAAAAAAAAAAAAAAAAAJ"),
					"standardIOUsesPseudoterminals": true,
					"startStopped":                  false,
					"terminateExisting":             false,
					"user": map[string]interface{}{
						"active": true,
					},
					"workingDirectory": nil,
				},
				"standardIOIdentifiers": map[string]interface{}{},
			},
			"CoreDevice.invocationIdentifier": "62419FC1-5ABF-4D96-BCA8-7A5F6F9A69EE",
		},
	}, res)
}

func base64Decode(s string) []byte {
	dst := make([]byte, base64.StdEncoding.DecodedLen(len(s)))
	_, err := base64.StdEncoding.Decode(dst, []byte(s))
	if err != nil {
		panic(err)
	}
	return dst
}

func TestEncodeDecode(t *testing.T) {
	tests := []struct {
		name          string
		input         map[string]interface{}
		expectedFlags uint32
	}{
		{
			name:          "empty dict",
			input:         map[string]interface{}{},
			expectedFlags: AlwaysSetFlag | DataFlag,
		},
		{
			name:          "no xpc body",
			input:         nil,
			expectedFlags: AlwaysSetFlag | DataFlag,
		},
		{
			name: "keys without padding",
			input: map[string]interface{}{
				"key":     "value",
				"key-key": "value",
			},
			expectedFlags: AlwaysSetFlag | DataFlag,
		},
		{
			name: "nested values",
			input: map[string]interface{}{
				"key1": "string-val",
				"nested-dict": map[string]interface{}{
					"bool":   true,
					"int64":  int64(123),
					"uint64": uint64(321),
					"data":   []byte{0x1},
					"double": float64(1.2),
				},
			},
			expectedFlags: AlwaysSetFlag | DataFlag,
		},
		{
			name: "null entry",
			input: map[string]interface{}{
				"null": nil,
			},
			expectedFlags: AlwaysSetFlag | DataFlag,
		},
		{
			name: "dictionary with array",
			input: map[string]interface{}{
				"array": []interface{}{uint64(1), uint64(2), uint64(3)},
			},
			expectedFlags: AlwaysSetFlag | DataFlag,
		},
		{
			name: "encode uuid",
			input: map[string]interface{}{
				"uuidvalue": func() uuid.UUID {
					u, _ := uuid.FromBytes(base64Decode("RYjS2yNAbEG+Y0WWxq5/4w=="))
					return u
				}(),
			},
			expectedFlags: AlwaysSetFlag | DataFlag,
		},
		{
			name: "encode uuid",
			input: map[string]interface{}{
				"uuidvalue": func() uuid.UUID {
					u, _ := uuid.FromBytes(base64Decode("RYjS2yNAbEG+Y0WWxq5/4w=="))
					return u
				}(),
			},
			expectedFlags: AlwaysSetFlag | DataFlag,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			buf := bytes.NewBuffer(nil)
			err := EncodeMessage(buf, Message{
				Flags: AlwaysSetFlag | DataFlag,
				Body:  tt.input,
				Id:    0,
			})
			assert.NoError(t, err)
			res, err := DecodeMessage(buf)
			assert.NoError(t, err)
			assert.Equal(t, tt.input, res.Body)
			assert.Equal(t, tt.expectedFlags, res.Flags)
		})
	}
}

func TestEncodeLaunch(t *testing.T) {

	deviceId := "3f9e370f-27cd-4e1e-b0d7-7952fc31afd1"
	bundleId := "com.apple.test.WebDriverAgentRunner-Runner"
	sessionIdentifier := "792F6A0A-8A28-4523-A810-6E15C89B5D79"
	testBundlePath := "/private/var/containers/Bundle/Application/3FDA3A90-C236-4322-872E-4F591E1CEDB6/WebDriverAgentRunner-Runner.app/PlugIns/WebDriverAgentRunner.xctest"
	args := []interface{}{}
	libraries := "/Developer/usr/lib/libMainThreadChecker.dylib"

	env := map[string]interface{}{
		"CA_ASSERT_MAIN_THREAD_TRANSACTIONS": "0",
		"CA_DEBUG_TRANSACTIONS":              "0",
		"DYLD_INSERT_LIBRARIES":              libraries,
		"DYLD_FRAMEWORK_PATH":                "/System/Developer/Library/Frameworks",
		"DYLD_LIBRARY_PATH":                  "/System/Developer/usr/lib",

		"MTC_CRASH_ON_REPORT":             "1",
		"NSUnbufferedIO":                  "YES",
		"OS_ACTIVITY_DT_MODE":             "YES",
		"SQLITE_ENABLE_THREAD_ASSERTIONS": "1",
		"XCTestBundlePath":                testBundlePath,
		"XCTestConfigurationFilePath":     "",
		"XCTestManagerVariant":            "DDI",
		"XCTestSessionIdentifier":         strings.ToUpper(sessionIdentifier),
	}

	options := map[string]interface{}{
		"ActivateSuspended":   uint64(1),
		"StartSuspendedKey":   uint64(0),
		"__ActivateSuspended": uint64(1),
	}
	terminateExisting := true
	byteArray := []byte{
		0x1f, 0xc2, 0x86, 0x84, 0x67, 0xd9, 0x4c, 0xff, 0x9f,
		0x06, 0x10, 0xf7, 0x02, 0xd9, 0x5b, 0x13,
	}
	uuidObj, _ := uuid.FromBytes(byteArray)
	stdIo := map[string]any{
		"standardInput":  uuidObj,
		"standardOutput": uuidObj,
		"standardError":  uuidObj,
	}

	// input := appservice.BuildAppLaunchPayload(deviceId, bundleId, args, env, options, terminateExisting, stdIoConfig)
	platformSpecificOptions := bytes.NewBuffer(nil)
	plistEncoder := plist.NewBinaryEncoder(platformSpecificOptions)
	err := plistEncoder.Encode(options)
	if err != nil {
		panic(err)
	}

	input := coredevice.BuildRequest1(deviceId, deviceId, "com.apple.coredevice.feature.launchapplication", map[string]interface{}{
		"applicationSpecifier": map[string]interface{}{
			"bundleIdentifier": map[string]interface{}{
				"_0": bundleId,
			},
		},
		"options": map[string]interface{}{
			"arguments":                     args,
			"environmentVariables":          env,
			"platformSpecificOptions":       platformSpecificOptions.Bytes(),
			"standardIOUsesPseudoterminals": true,
			"startStopped":                  false,
			"terminateExisting":             terminateExisting,
			"user": map[string]interface{}{
				"active": true,
			},
		},
		"standardIOIdentifiers": stdIo,
	})

	buf := bytes.NewBuffer(nil)
	msg := Message{
		Flags: HeartbeatRequestFlag | AlwaysSetFlag | DataFlag,
		Body:  input,
		Id:    0,
	}
	err1 := EncodeMessage(buf, msg)
	assert.NoError(t, err1)

	output1 := buf.Bytes()
	oo := hex.Dump(output1)
	println(oo)

	errf := os.WriteFile("/Users/xiao/Desktop/go-launch.bin", output1, 0644)
	if errf != nil {
		log.Fatalf("Error writing to file: %v", err)
	}

	fmt.Println("Data successfully written to output.txt")

}

// simple one
func TestEncodeLaunch_simple(t *testing.T) {

	input := coredevice.BuildRequest2()

	buf := bytes.NewBuffer(nil)
	msg := Message{
		Flags: HeartbeatRequestFlag | AlwaysSetFlag | DataFlag,
		Body:  input,
		Id:    0,
	}
	err1 := EncodeMessage(buf, msg)
	assert.NoError(t, err1)

	output1 := buf.Bytes()
	oo := hex.Dump(output1)
	println(oo)

	errf := os.WriteFile("/Users/xiao/Desktop/go-launch.bin", output1, 0644)
	if errf != nil {
		log.Fatalf("Error writing to file: %v", errf)
	}

	fmt.Println("Data successfully written to go-launch.bin")

}

func TestEncodeLaunch_result(t *testing.T) {

	input := map[string]interface{}{
		"auditToken": []interface{}{
			uint64(0xffffffff), //4294967295
			uint64(501),
			uint64(501),
			uint64(501),
			uint64(501),
			uint64(1584),
			uint64(0),
			uint64(4122)},
	}

	buf := bytes.NewBuffer(nil)
	msg := Message{
		Flags: HeartbeatRequestFlag | AlwaysSetFlag | DataFlag,
		Body:  input,
		Id:    0,
	}
	err1 := EncodeMessage(buf, msg)
	assert.NoError(t, err1)

	output1 := buf.Bytes()
	oo := hex.Dump(output1)
	println(oo)

	errf := os.WriteFile("/Users/xiao/Desktop/go-launch.bin", output1, 0644)
	if errf != nil {
		log.Fatalf("Error writing to file: %v", errf)
	}

	fmt.Println("Data successfully written to go-launch.bin")

}

func TestDictionary_decode_launch_result(t *testing.T) {
	b, _ := os.ReadFile("/Users/xiao/Desktop/14.bin")

	res, _ := DecodeMessage(bytes.NewReader(b))
	prettyJSON, _ := json.MarshalIndent(res, "", "  ")
	println(string(prettyJSON))
}

func TestDictionary2(t *testing.T) {
	b, _ := os.ReadFile("/Users/xiao/packets/utun7-go1/13.bin")
	// b, _ := os.ReadFile("/Users/xiao/packets/utun5-java/13.bin")

	// byteArray := []byte{
	// 	0x1f, 0xc2, 0x86, 0x84, 0x67, 0xd9, 0x4c, 0xff, 0x9f,
	// 	0x06, 0x10, 0xf7, 0x02, 0xd9, 0x5b, 0x13,
	// }
	// uuidObj, err1 := uuid.FromBytes(byteArray)

	// hexString := "bplist00\xd3\x01\x02\x03\x04\x05\x04_\x10\x11ActivateSuspended_\x10\x11StartSuspendedKey_\x10\x13__ActivateSuspended\x10\x01\x10\x00\b\x0f#7MO\x00\x00\x00\x00\x00\x00\x01\x01\x00\x00\x00\x00\x00\x00\x00\x06\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00Q"
	// optionsByte, err2 := hex.DecodeString(hexString)

	// if err1 != nil {
	// 	fmt.Println("Error creating UUID:", err1)
	// }
	// if err2 != nil {
	// 	fmt.Println("Error creating UUID:", err2)
	// }

	res, _ := DecodeMessage(bytes.NewReader(b))
	prettyJSON, _ := json.MarshalIndent(res, "", "  ")
	println(string(prettyJSON))

	/*
		assert.NoError(t, err)
		assert.Equal(t, Message{
			Flags: AlwaysSetFlag | DataFlag | HeartbeatRequestFlag,
			Body: map[string]interface{}{
				"CoreDevice.CoreDeviceDDIProtocolVersion": int64(0),
				"CoreDevice.action":                       map[string]interface{}{},
				"CoreDevice.coreDeviceVersion": map[string]interface{}{
					"components":              []interface{}{uint64(0x15c), uint64(0x1), uint64(0x0), uint64(0x0), uint64(0x0)},
					"originalComponentsCount": int64(2),
					"stringValue":             "348.1",
				},
				"CoreDevice.deviceIdentifier":  "54e0066e-0512-48af-8a58-21c18058eb1d",
				"CoreDevice.featureIdentifier": "com.apple.coredevice.feature.launchapplication",
				"CoreDevice.input": map[string]interface{}{
					"applicationSpecifier": map[string]interface{}{
						"bundleIdentifier": map[string]interface{}{
							"_0": "com.apple.test.WebDriverAgentRunner-Runner",
						},
					},
					"options": map[string]interface{}{
						"arguments": []interface{}{},
						"environmentVariables": map[string]interface{}{
							"CA_ASSERT_MAIN_THREAD_TRANSACTIONS": "0",
							"CA_DEBUG_TRANSACTIONS":              "0",
							"DYLD_INSERT_LIBRARIES":              "/Developer/usr/lib/libMainThreadChecker.dylib",
							"DYLD_FRAMEWORK_PATH":                "/System/Developer/Library/Frameworks",
							"DYLD_LIBRARY_PATH":                  "/System/Developer/usr/lib",
							"MTC_CRASH_ON_REPORT":                "1",
							"NSUnbufferedIO":                     "YES",
							"OS_ACTIVITY_DT_MODE":                "YES",
							"SQLITE_ENABLE_THREAD_ASSERTIONS":    "1",
							"XCTestBundlePath":                   "/private/var/containers/Bundle/Application/77EF4F4D-D2C1-4B6A-AAB6-22305FA8AE8E/WebDriverAgentRunner-Runner.app/PlugIns/WebDriverAgentRunner.xctest",
							"XCTestConfigurationFilePath":        "",
							"XCTestManagerVariant":               "DDI",
							"XCTestSessionIdentifier":            "C7792F5B-B00F-4C47-B721-C9815F02EE29",
						},
						"platformSpecificOptions":       optionsByte,
						"standardIOUsesPseudoterminals": true,
						"startStopped":                  false,
						"terminateExisting":             true,
						"user": map[string]interface{}{
							"active": true,
						},
						"workingDirectory": nil,
					},
					"standardIOIdentifiers": map[string]interface{}{
						"standardInput":  uuidObj,
						"standardOutput": uuidObj,
						"standardError":  uuidObj,
					},
				},
				"CoreDevice.invocationIdentifier": "80aea1e4-ccf8-4344-8452-ed3b68d5425d",
			},
		}, res)
	*/
}
