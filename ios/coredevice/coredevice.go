package coredevice

import "github.com/google/uuid"

func BuildRequest(deviceId, feature string, input map[string]interface{}) map[string]interface{} {
	u := uuid.New()
	return map[string]interface{}{
		"CoreDevice.CoreDeviceDDIProtocolVersion": int64(0),
		"CoreDevice.action":                       map[string]interface{}{},
		"CoreDevice.coreDeviceVersion": map[string]interface{}{
			"components":              []interface{}{},
			"originalComponentsCount": int64(0),
			"stringValue":             "348.1",
		},
		"CoreDevice.deviceIdentifier":     deviceId,
		"CoreDevice.featureIdentifier":    feature,
		"CoreDevice.input":                input,
		"CoreDevice.invocationIdentifier": u.String(),
	}
}

func BuildRequest1(deviceId, u string, feature string, input map[string]interface{}) map[string]interface{} {

	return map[string]interface{}{
		"CoreDevice.CoreDeviceDDIProtocolVersion": int64(0),
		"CoreDevice.action":                       map[string]interface{}{},
		"CoreDevice.coreDeviceVersion": map[string]interface{}{
			"components":              []interface{}{},
			"originalComponentsCount": int64(0),
			"stringValue":             "348.1",
		},
		"CoreDevice.deviceIdentifier":     deviceId,
		"CoreDevice.featureIdentifier":    feature,
		"CoreDevice.input":                input,
		"CoreDevice.invocationIdentifier": u,
	}
}

func BuildRequest2() map[string]interface{} {

	return map[string]interface{}{
		// "CoreDevice.CoreDeviceDDIProtocolVersion": int64(0),
		// "CoreDevice.action":                       map[string]interface{}{},
		"CoreDevice.coreDeviceVersion": map[string]interface{}{
			"components": []interface{}{},
			// "originalComponentsCount": int64(0),
			// "stringValue":             "348.1",
		},
	}
}
