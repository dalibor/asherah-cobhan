package main

import (
	"encoding/json"
	"testing"

	"github.com/godaddy/cobhan-go"
)

func setupAsherahForTesting(t *testing.T) {
	config := &Options{}

	config.KMS = "static"
	config.ServiceName = "TestService"
	config.ProductID = "TestProduct"
	config.Metastore = "memory"
	config.EnableSessionCaching = true
	config.SessionCacheDuration = 1000
	config.SessionCacheMaxSize = 2
	config.ExpireAfter = 1000
	config.CheckInterval = 1000
	config.Verbose = true
	config.RegionMap = RegionMap{}
	config.RegionMap.UnmarshalFlag("region1=arn1,region2=arn2")

	buf := testAllocateJsonBuffer(t, config)

	result := SetupJson(cobhan.Ptr(&buf))
	if result != ERR_NONE {
		t.Errorf("SetupJson returned %v", result)
	}
}

func testAllocateStringBuffer(t *testing.T, str string) []byte {
	buf, result := cobhan.AllocateStringBuffer(str)
	if result != ERR_NONE {
		t.Errorf("AllocateStringBuffer returned %v", result)
	}
	return buf
}

func testAllocateBytesBuffer(t *testing.T, bytes []byte) []byte {
	buf, result := cobhan.AllocateBytesBuffer(bytes)
	if result != ERR_NONE {
		t.Errorf("AllocateStringBuffer returned %v", result)
	}
	return buf
}

func testAllocateJsonBuffer(t *testing.T, obj interface{}) []byte {
	bytes, err := json.Marshal(obj)
	if err != nil {
		t.Errorf("json.Marshal returned %v", err)
	}
	return testAllocateBytesBuffer(t, bytes)
}

func TestSetupJson(t *testing.T) {
	setupAsherahForTesting(t)
	Shutdown()
}

func TestSetupJsonAlternateConfiguration(t *testing.T) {
	config := &Options{}

	config.KMS = "static"
	config.ServiceName = "TestService"
	config.ProductID = "TestProduct"
	config.Metastore = "memory"
	config.EnableSessionCaching = true
	config.Verbose = false

	buf := testAllocateJsonBuffer(t, config)

	result := SetupJson(cobhan.Ptr(&buf))
	if result != ERR_NONE {
		t.Errorf("SetupJson returned %v", result)
	}
	Shutdown()
}

func TestSetupJsonTwice(t *testing.T) {
	config := &Options{}

	config.KMS = "static"
	config.ServiceName = "TestService"
	config.ProductID = "TestProduct"
	config.Metastore = "memory"
	config.EnableSessionCaching = true
	config.Verbose = true

	buf := testAllocateJsonBuffer(t, config)

	result := SetupJson(cobhan.Ptr(&buf))
	if result != ERR_NONE {
		t.Errorf("SetupJson returned %v", result)
	}
	defer Shutdown()
	result = SetupJson(cobhan.Ptr(&buf))
	if result != ERR_ALREADY_INITIALIZED {
		t.Errorf("Expected SetupJson to return ERR_ALREADY_INITIALIZED got %v", result)
	}
}

func TestSetupInvalidJson(t *testing.T) {
	str := "}InvalidJson{"

	buf := testAllocateStringBuffer(t, str)

	result := SetupJson(cobhan.Ptr(&buf))
	if result != cobhan.ERR_JSON_DECODE_FAILED {
		t.Errorf("Expected SetupJson to return ERR_JSON_DECODE_FAILED got %v", result)
	}
	Shutdown()
}

func TestSetupNullJson(t *testing.T) {
	SetupJson(nil)
	Shutdown()
}

func TestEncryptDecryptRoundTrip(t *testing.T) {
	setupAsherahForTesting(t)
	defer Shutdown()

	input := "InputData"
	partitionId := testAllocateStringBuffer(t, "Partition")
	data := testAllocateStringBuffer(t, input)
	encryptedData := cobhan.AllocateBuffer(256)
	encryptedKey := cobhan.AllocateBuffer(256)
	createdBuf := cobhan.AllocateBuffer(8)
	parentKeyId := cobhan.AllocateBuffer(256)
	parentKeyCreatedBuf := cobhan.AllocateBuffer(8)

	result := Encrypt(cobhan.Ptr(&partitionId),
		cobhan.Ptr(&data),
		cobhan.Ptr(&encryptedData),
		cobhan.Ptr(&encryptedKey),
		cobhan.Ptr(&createdBuf),
		cobhan.Ptr(&parentKeyId),
		cobhan.Ptr(&parentKeyCreatedBuf),
	)
	if result != ERR_NONE {
		t.Errorf("Encrypt returned %v", result)
	}

	decryptedData := cobhan.AllocateBuffer(256)

	created, result := cobhan.BufferToInt64Safe(&createdBuf)
	if result != ERR_NONE {
		t.Errorf("BufferToInt64Safe returned %v", result)
	}

	parentKeyCreated, result := cobhan.BufferToInt64Safe(&parentKeyCreatedBuf)
	if result != ERR_NONE {
		t.Errorf("BufferToInt64Safe returned %v", result)
	}

	result = Decrypt(cobhan.Ptr(&partitionId),
		cobhan.Ptr(&encryptedData),
		cobhan.Ptr(&encryptedKey),
		created,
		cobhan.Ptr(&parentKeyId),
		parentKeyCreated,
		cobhan.Ptr(&decryptedData),
	)
	if result != ERR_NONE {
		t.Errorf("Decrypt returned %v", result)
	}

	output, result := cobhan.BufferToStringSafe(&decryptedData)
	if result != ERR_NONE {
		t.Errorf("BufferToStringSafe returned %v", result)
	}
	if output != input {
		t.Errorf("Expected %v Actual %v", input, output)
	}
}

func TestEncryptWithoutInit(t *testing.T) {
	partitionId := testAllocateStringBuffer(t, "Partition")
	data := testAllocateStringBuffer(t, "InputData")
	encryptedData := cobhan.AllocateBuffer(256)
	encryptedKey := cobhan.AllocateBuffer(256)
	createdBuf := cobhan.AllocateBuffer(8)
	parentKeyId := cobhan.AllocateBuffer(256)
	parentKeyCreatedBuf := cobhan.AllocateBuffer(8)

	result := Encrypt(cobhan.Ptr(&partitionId),
		cobhan.Ptr(&data),
		cobhan.Ptr(&encryptedData),
		cobhan.Ptr(&encryptedKey),
		cobhan.Ptr(&createdBuf),
		cobhan.Ptr(&parentKeyId),
		cobhan.Ptr(&parentKeyCreatedBuf),
	)
	if result != ERR_NOT_INITIALIZED {
		t.Error("Encrypt didn't return ERR_NOT_INITIALIZED")
	}
}

func TestDecryptWithoutInit(t *testing.T) {

	partitionId := testAllocateStringBuffer(t, "Partition")
	encryptedData := cobhan.AllocateBuffer(256)
	encryptedKey := cobhan.AllocateBuffer(256)
	parentKeyId := cobhan.AllocateBuffer(256)

	decryptedData := cobhan.AllocateBuffer(256)

	result := Decrypt(cobhan.Ptr(&partitionId),
		cobhan.Ptr(&encryptedData),
		cobhan.Ptr(&encryptedKey),
		1234,
		cobhan.Ptr(&parentKeyId),
		1234,
		cobhan.Ptr(&decryptedData),
	)
	if result != ERR_NOT_INITIALIZED {
		t.Error("Decrypt didn't return ERR_NOT_INITIALIZED")
	}
}

func TestEncryptNullPartitionId(t *testing.T) {
	setupAsherahForTesting(t)
	defer Shutdown()

	data := testAllocateStringBuffer(t, "InputData")

	encryptedData := cobhan.AllocateBuffer(256)
	encryptedKey := cobhan.AllocateBuffer(256)
	createdBuf := cobhan.AllocateBuffer(8)
	parentKeyId := cobhan.AllocateBuffer(256)
	parentKeyCreatedBuf := cobhan.AllocateBuffer(8)

	result := Encrypt(nil,
		cobhan.Ptr(&data),
		cobhan.Ptr(&encryptedData),
		cobhan.Ptr(&encryptedKey),
		cobhan.Ptr(&createdBuf),
		cobhan.Ptr(&parentKeyId),
		cobhan.Ptr(&parentKeyCreatedBuf),
	)
	if result != cobhan.ERR_NULL_PTR {
		t.Error("Encrypt didn't return ERR_NULL_PTR")
	}
}

func TestEncryptNullData(t *testing.T) {
	setupAsherahForTesting(t)
	defer Shutdown()

	partitionId := testAllocateStringBuffer(t, "Partition")
	encryptedData := cobhan.AllocateBuffer(256)
	encryptedKey := cobhan.AllocateBuffer(256)
	createdBuf := cobhan.AllocateBuffer(8)
	parentKeyId := cobhan.AllocateBuffer(256)
	parentKeyCreatedBuf := cobhan.AllocateBuffer(8)

	result := Encrypt(cobhan.Ptr(&partitionId),
		nil,
		cobhan.Ptr(&encryptedData),
		cobhan.Ptr(&encryptedKey),
		cobhan.Ptr(&createdBuf),
		cobhan.Ptr(&parentKeyId),
		cobhan.Ptr(&parentKeyCreatedBuf),
	)
	if result != cobhan.ERR_NULL_PTR {
		t.Error("Encrypt didn't return ERR_NULL_PTR")
	}
}

func TestEncryptNullEncryptedData(t *testing.T) {
	setupAsherahForTesting(t)
	defer Shutdown()

	partitionId := testAllocateStringBuffer(t, "Partition")
	data := testAllocateStringBuffer(t, "InputData")
	encryptedKey := cobhan.AllocateBuffer(256)
	createdBuf := cobhan.AllocateBuffer(8)
	parentKeyId := cobhan.AllocateBuffer(256)
	parentKeyCreatedBuf := cobhan.AllocateBuffer(8)

	result := Encrypt(cobhan.Ptr(&partitionId),
		cobhan.Ptr(&data),
		nil,
		cobhan.Ptr(&encryptedKey),
		cobhan.Ptr(&createdBuf),
		cobhan.Ptr(&parentKeyId),
		cobhan.Ptr(&parentKeyCreatedBuf),
	)
	if result != cobhan.ERR_NULL_PTR {
		t.Error("Encrypt didn't return ERR_NULL_PTR")
	}
}

func TestEncryptNullEncryptedKey(t *testing.T) {
	setupAsherahForTesting(t)
	defer Shutdown()

	partitionId := testAllocateStringBuffer(t, "Partition")
	data := testAllocateStringBuffer(t, "InputData")
	encryptedData := cobhan.AllocateBuffer(256)
	createdBuf := cobhan.AllocateBuffer(8)
	parentKeyId := cobhan.AllocateBuffer(256)
	parentKeyCreatedBuf := cobhan.AllocateBuffer(8)

	result := Encrypt(cobhan.Ptr(&partitionId),
		cobhan.Ptr(&data),
		cobhan.Ptr(&encryptedData),
		nil,
		cobhan.Ptr(&createdBuf),
		cobhan.Ptr(&parentKeyId),
		cobhan.Ptr(&parentKeyCreatedBuf),
	)
	if result != cobhan.ERR_NULL_PTR {
		t.Error("Encrypt didn't return ERR_NULL_PTR")
	}
}

func TestEncryptNullCreatedBuf(t *testing.T) {
	setupAsherahForTesting(t)
	defer Shutdown()

	partitionId := testAllocateStringBuffer(t, "Partition")
	data := testAllocateStringBuffer(t, "InputData")
	encryptedKey := cobhan.AllocateBuffer(256)
	encryptedData := cobhan.AllocateBuffer(256)
	parentKeyId := cobhan.AllocateBuffer(256)
	parentKeyCreatedBuf := cobhan.AllocateBuffer(8)

	result := Encrypt(cobhan.Ptr(&partitionId),
		cobhan.Ptr(&data),
		cobhan.Ptr(&encryptedData),
		cobhan.Ptr(&encryptedKey),
		nil,
		cobhan.Ptr(&parentKeyId),
		cobhan.Ptr(&parentKeyCreatedBuf),
	)
	if result != cobhan.ERR_NULL_PTR {
		t.Error("Encrypt didn't return ERR_NULL_PTR")
	}
}

func TestEncryptNullParentKeyId(t *testing.T) {
	setupAsherahForTesting(t)
	defer Shutdown()

	partitionId := testAllocateStringBuffer(t, "Partition")
	data := testAllocateStringBuffer(t, "InputData")
	encryptedKey := cobhan.AllocateBuffer(256)
	encryptedData := cobhan.AllocateBuffer(256)
	createdBuf := cobhan.AllocateBuffer(8)
	parentKeyCreatedBuf := cobhan.AllocateBuffer(8)

	result := Encrypt(cobhan.Ptr(&partitionId),
		cobhan.Ptr(&data),
		cobhan.Ptr(&encryptedData),
		cobhan.Ptr(&encryptedKey),
		cobhan.Ptr(&createdBuf),
		nil,
		cobhan.Ptr(&parentKeyCreatedBuf),
	)
	if result != cobhan.ERR_NULL_PTR {
		t.Error("Encrypt didn't return ERR_NULL_PTR")
	}
}

func TestEncryptNullParentKeyCreated(t *testing.T) {
	setupAsherahForTesting(t)
	defer Shutdown()

	partitionId := testAllocateStringBuffer(t, "Partition")
	data := testAllocateStringBuffer(t, "InputData")
	encryptedKey := cobhan.AllocateBuffer(256)
	encryptedData := cobhan.AllocateBuffer(256)
	createdBuf := cobhan.AllocateBuffer(8)
	parentKeyId := cobhan.AllocateBuffer(256)

	result := Encrypt(cobhan.Ptr(&partitionId),
		cobhan.Ptr(&data),
		cobhan.Ptr(&encryptedData),
		cobhan.Ptr(&encryptedKey),
		cobhan.Ptr(&createdBuf),
		cobhan.Ptr(&parentKeyId),
		nil,
	)
	if result != cobhan.ERR_NULL_PTR {
		t.Error("Encrypt didn't return ERR_NULL_PTR")
	}
}

func TestDecryptNullPartitionId(t *testing.T) {
	setupAsherahForTesting(t)
	defer Shutdown()

	input := "InputData"
	partitionId := testAllocateStringBuffer(t, "Partition")
	data := testAllocateStringBuffer(t, input)
	encryptedData := cobhan.AllocateBuffer(256)
	encryptedKey := cobhan.AllocateBuffer(256)
	createdBuf := cobhan.AllocateBuffer(8)
	parentKeyId := cobhan.AllocateBuffer(256)
	parentKeyCreatedBuf := cobhan.AllocateBuffer(8)

	result := Encrypt(cobhan.Ptr(&partitionId),
		cobhan.Ptr(&data),
		cobhan.Ptr(&encryptedData),
		cobhan.Ptr(&encryptedKey),
		cobhan.Ptr(&createdBuf),
		cobhan.Ptr(&parentKeyId),
		cobhan.Ptr(&parentKeyCreatedBuf),
	)
	if result != ERR_NONE {
		t.Errorf("Encrypt returned %v", result)
	}

	decryptedData := cobhan.AllocateBuffer(256)

	created, result := cobhan.BufferToInt64Safe(&createdBuf)
	if result != ERR_NONE {
		t.Errorf("BufferToInt64Safe returned %v", result)
	}

	parentKeyCreated, result := cobhan.BufferToInt64Safe(&parentKeyCreatedBuf)
	if result != ERR_NONE {
		t.Errorf("BufferToInt64Safe returned %v", result)
	}

	result = Decrypt(nil,
		cobhan.Ptr(&encryptedData),
		cobhan.Ptr(&encryptedKey),
		created,
		cobhan.Ptr(&parentKeyId),
		parentKeyCreated,
		cobhan.Ptr(&decryptedData),
	)
	if result != cobhan.ERR_NULL_PTR {
		t.Error("Decrypt didn't return ERR_NULL_PTR")
	}
}

func TestDecryptNullEncryptedData(t *testing.T) {
	setupAsherahForTesting(t)
	defer Shutdown()

	input := "InputData"
	partitionId := testAllocateStringBuffer(t, "Partition")
	data := testAllocateStringBuffer(t, input)
	encryptedData := cobhan.AllocateBuffer(256)
	encryptedKey := cobhan.AllocateBuffer(256)
	createdBuf := cobhan.AllocateBuffer(8)
	parentKeyId := cobhan.AllocateBuffer(256)
	parentKeyCreatedBuf := cobhan.AllocateBuffer(8)

	result := Encrypt(cobhan.Ptr(&partitionId),
		cobhan.Ptr(&data),
		cobhan.Ptr(&encryptedData),
		cobhan.Ptr(&encryptedKey),
		cobhan.Ptr(&createdBuf),
		cobhan.Ptr(&parentKeyId),
		cobhan.Ptr(&parentKeyCreatedBuf),
	)
	if result != ERR_NONE {
		t.Errorf("Encrypt returned %v", result)
	}

	decryptedData := cobhan.AllocateBuffer(256)

	created, result := cobhan.BufferToInt64Safe(&createdBuf)
	if result != ERR_NONE {
		t.Errorf("BufferToInt64Safe returned %v", result)
	}

	parentKeyCreated, result := cobhan.BufferToInt64Safe(&parentKeyCreatedBuf)
	if result != ERR_NONE {
		t.Errorf("BufferToInt64Safe returned %v", result)
	}

	result = Decrypt(cobhan.Ptr(&partitionId),
		nil,
		cobhan.Ptr(&encryptedKey),
		created,
		cobhan.Ptr(&parentKeyId),
		parentKeyCreated,
		cobhan.Ptr(&decryptedData),
	)
	if result != cobhan.ERR_NULL_PTR {
		t.Error("Decrypt didn't return ERR_NULL_PTR")
	}
}

func TestDecryptNullEncryptedKey(t *testing.T) {
	setupAsherahForTesting(t)
	defer Shutdown()

	input := "InputData"
	partitionId := testAllocateStringBuffer(t, "Partition")
	data := testAllocateStringBuffer(t, input)
	encryptedData := cobhan.AllocateBuffer(256)
	encryptedKey := cobhan.AllocateBuffer(256)
	createdBuf := cobhan.AllocateBuffer(8)
	parentKeyId := cobhan.AllocateBuffer(256)
	parentKeyCreatedBuf := cobhan.AllocateBuffer(8)

	result := Encrypt(cobhan.Ptr(&partitionId),
		cobhan.Ptr(&data),
		cobhan.Ptr(&encryptedData),
		cobhan.Ptr(&encryptedKey),
		cobhan.Ptr(&createdBuf),
		cobhan.Ptr(&parentKeyId),
		cobhan.Ptr(&parentKeyCreatedBuf),
	)
	if result != ERR_NONE {
		t.Errorf("Encrypt returned %v", result)
	}

	decryptedData := cobhan.AllocateBuffer(256)

	created, result := cobhan.BufferToInt64Safe(&createdBuf)
	if result != ERR_NONE {
		t.Errorf("BufferToInt64Safe returned %v", result)
	}

	parentKeyCreated, result := cobhan.BufferToInt64Safe(&parentKeyCreatedBuf)
	if result != ERR_NONE {
		t.Errorf("BufferToInt64Safe returned %v", result)
	}

	result = Decrypt(cobhan.Ptr(&partitionId),
		cobhan.Ptr(&encryptedData),
		nil,
		created,
		cobhan.Ptr(&parentKeyId),
		parentKeyCreated,
		cobhan.Ptr(&decryptedData),
	)
	if result != cobhan.ERR_NULL_PTR {
		t.Error("Decrypt didn't return ERR_NULL_PTR")
	}
}

func TestDecryptNullParentKeyId(t *testing.T) {
	setupAsherahForTesting(t)
	defer Shutdown()

	input := "InputData"
	partitionId := testAllocateStringBuffer(t, "Partition")
	data := testAllocateStringBuffer(t, input)
	encryptedData := cobhan.AllocateBuffer(256)
	encryptedKey := cobhan.AllocateBuffer(256)
	createdBuf := cobhan.AllocateBuffer(8)
	parentKeyId := cobhan.AllocateBuffer(256)
	parentKeyCreatedBuf := cobhan.AllocateBuffer(8)

	result := Encrypt(cobhan.Ptr(&partitionId),
		cobhan.Ptr(&data),
		cobhan.Ptr(&encryptedData),
		cobhan.Ptr(&encryptedKey),
		cobhan.Ptr(&createdBuf),
		cobhan.Ptr(&parentKeyId),
		cobhan.Ptr(&parentKeyCreatedBuf),
	)
	if result != ERR_NONE {
		t.Errorf("Encrypt returned %v", result)
	}

	decryptedData := cobhan.AllocateBuffer(256)

	created, result := cobhan.BufferToInt64Safe(&createdBuf)
	if result != ERR_NONE {
		t.Errorf("BufferToInt64Safe returned %v", result)
	}

	parentKeyCreated, result := cobhan.BufferToInt64Safe(&parentKeyCreatedBuf)
	if result != ERR_NONE {
		t.Errorf("BufferToInt64Safe returned %v", result)
	}

	result = Decrypt(cobhan.Ptr(&partitionId),
		cobhan.Ptr(&encryptedData),
		cobhan.Ptr(&encryptedKey),
		created,
		nil,
		parentKeyCreated,
		cobhan.Ptr(&decryptedData),
	)
	if result != cobhan.ERR_NULL_PTR {
		t.Error("Decrypt didn't return ERR_NULL_PTR")
	}
}

func TestDecryptNullDecryptedData(t *testing.T) {
	setupAsherahForTesting(t)
	defer Shutdown()

	input := "InputData"
	partitionId := testAllocateStringBuffer(t, "Partition")
	data := testAllocateStringBuffer(t, input)
	encryptedData := cobhan.AllocateBuffer(256)
	encryptedKey := cobhan.AllocateBuffer(256)
	createdBuf := cobhan.AllocateBuffer(8)
	parentKeyId := cobhan.AllocateBuffer(256)
	parentKeyCreatedBuf := cobhan.AllocateBuffer(8)

	result := Encrypt(cobhan.Ptr(&partitionId),
		cobhan.Ptr(&data),
		cobhan.Ptr(&encryptedData),
		cobhan.Ptr(&encryptedKey),
		cobhan.Ptr(&createdBuf),
		cobhan.Ptr(&parentKeyId),
		cobhan.Ptr(&parentKeyCreatedBuf),
	)
	if result != ERR_NONE {
		t.Errorf("Encrypt returned %v", result)
	}

	created, result := cobhan.BufferToInt64Safe(&createdBuf)
	if result != ERR_NONE {
		t.Errorf("BufferToInt64Safe returned %v", result)
	}

	parentKeyCreated, result := cobhan.BufferToInt64Safe(&parentKeyCreatedBuf)
	if result != ERR_NONE {
		t.Errorf("BufferToInt64Safe returned %v", result)
	}

	result = Decrypt(cobhan.Ptr(&partitionId),
		cobhan.Ptr(&encryptedData),
		cobhan.Ptr(&encryptedKey),
		created,
		cobhan.Ptr(&parentKeyId),
		parentKeyCreated,
		nil,
	)
	if result != cobhan.ERR_NULL_PTR {
		t.Error("Decrypt didn't return ERR_NULL_PTR")
	}
}

func TestDecryptBadData(t *testing.T) {
	setupAsherahForTesting(t)
	defer Shutdown()

	input := "InputData"
	partitionId := testAllocateStringBuffer(t, "Partition")
	data := testAllocateStringBuffer(t, input)
	encryptedData := cobhan.AllocateBuffer(256)
	encryptedKey := cobhan.AllocateBuffer(256)
	createdBuf := cobhan.AllocateBuffer(8)
	parentKeyId := cobhan.AllocateBuffer(256)
	parentKeyCreatedBuf := cobhan.AllocateBuffer(8)

	result := Encrypt(cobhan.Ptr(&partitionId),
		cobhan.Ptr(&data),
		cobhan.Ptr(&encryptedData),
		cobhan.Ptr(&encryptedKey),
		cobhan.Ptr(&createdBuf),
		cobhan.Ptr(&parentKeyId),
		cobhan.Ptr(&parentKeyCreatedBuf),
	)
	if result != ERR_NONE {
		t.Errorf("Encrypt returned %v", result)
	}

	// Intentionally corrupt the encrypted data
	encryptedData[cobhan.BUFFER_HEADER_SIZE+4] = 1
	encryptedData[cobhan.BUFFER_HEADER_SIZE+5] = 2
	encryptedData[cobhan.BUFFER_HEADER_SIZE+6] = 3
	encryptedData[cobhan.BUFFER_HEADER_SIZE+7] = 4

	decryptedData := cobhan.AllocateBuffer(256)

	created, result := cobhan.BufferToInt64Safe(&createdBuf)
	if result != ERR_NONE {
		t.Errorf("BufferToInt64Safe returned %v", result)
	}

	parentKeyCreated, result := cobhan.BufferToInt64Safe(&parentKeyCreatedBuf)
	if result != ERR_NONE {
		t.Errorf("BufferToInt64Safe returned %v", result)
	}

	result = Decrypt(cobhan.Ptr(&partitionId),
		cobhan.Ptr(&encryptedData),
		cobhan.Ptr(&encryptedKey),
		created,
		cobhan.Ptr(&parentKeyId),
		parentKeyCreated,
		cobhan.Ptr(&decryptedData),
	)
	if result != ERR_DECRYPT_FAILED {
		t.Error("Decrypt didn't return ERR_DECRYPT_FAILED")
	}
}
