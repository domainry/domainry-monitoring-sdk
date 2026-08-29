package monitoringsdk

import "testing"

func TestApplicationAndDescriptorValidation(t *testing.T) {
	if err := (ApplicationRef{}).Validate(); err == nil {
		t.Fatal("empty runtime identity accepted")
	}
	if err := (ApplicationRef{RuntimeID: "runtime-1"}).Validate(); err != nil {
		t.Fatal(err)
	}
	if err := (Descriptor{ProtocolVersion: ProtocolVersionV1, Mode: DeploymentModeModule}).Validate(); err != nil {
		t.Fatal(err)
	}
	if err := (Descriptor{ProtocolVersion: "old", Mode: DeploymentModeModule}).Validate(); err == nil {
		t.Fatal("unsupported protocol accepted")
	}
}
