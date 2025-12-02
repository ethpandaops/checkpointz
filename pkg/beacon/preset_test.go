package beacon_test

import (
	"testing"

	"github.com/attestantio/go-eth2-client/spec/phase0"
	"github.com/ethpandaops/beacon/pkg/beacon/state"
	"github.com/ethpandaops/checkpointz/pkg/beacon"
	"github.com/ethpandaops/checkpointz/pkg/beacon/node"
	dynssz "github.com/pk910/dynamic-ssz"
	"github.com/sirupsen/logrus"
	"gopkg.in/yaml.v2"
)

// createMainnetSpec creates a mock spec with mainnet preset values
func createMainnetSpec() *state.Spec {
	specData := map[string]interface{}{
		"PRESET_BASE":                  "mainnet",
		"CONFIG_NAME":                  "mainnet",
		"SLOTS_PER_EPOCH":              "32",
		"SECONDS_PER_SLOT":             "12",
		"MAX_VALIDATORS_PER_COMMITTEE": "2048",
		"TARGET_COMMITTEE_SIZE":        "128",
		"MAX_EFFECTIVE_BALANCE":        "32000000000",
		"MIN_DEPOSIT_AMOUNT":           "1000000000",
		"EFFECTIVE_BALANCE_INCREMENT":  "1000000000",
	}

	spec := state.NewSpec(specData)

	return &spec
}

// createMinimalSpec creates a mock spec with minimal preset values
func createMinimalSpec() *state.Spec {
	specData := map[string]interface{}{
		"PRESET_BASE":                  "minimal",
		"CONFIG_NAME":                  "minimal",
		"SLOTS_PER_EPOCH":              "8",
		"SECONDS_PER_SLOT":             "6",
		"MAX_VALIDATORS_PER_COMMITTEE": "2048",
		"TARGET_COMMITTEE_SIZE":        "4",
		"MAX_EFFECTIVE_BALANCE":        "32000000000",
		"MIN_DEPOSIT_AMOUNT":           "1000000000",
		"EFFECTIVE_BALANCE_INCREMENT":  "1000000000",
	}

	spec := state.NewSpec(specData)

	return &spec
}

// initializeDynSsz mimics the initialization logic from refreshSpec
func initializeDynSsz(spec *state.Spec) (*dynssz.DynSsz, error) {
	staticSpec := map[string]any{}
	specYaml, err := yaml.Marshal(spec)

	if err != nil {
		return nil, err
	}

	if err := yaml.Unmarshal(specYaml, &staticSpec); err != nil {
		return nil, err
	}

	return dynssz.NewDynSsz(staticSpec), nil
}

func TestMainnetPresetDetection(t *testing.T) {
	spec := createMainnetSpec()

	if spec.PresetBase != "mainnet" {
		t.Errorf("Expected PresetBase to be 'mainnet', got '%s'", spec.PresetBase)
	}

	if spec.SlotsPerEpoch != phase0.Slot(32) {
		t.Errorf("Expected SlotsPerEpoch to be 32, got %d", spec.SlotsPerEpoch)
	}
}

func TestMinimalPresetDetection(t *testing.T) {
	spec := createMinimalSpec()

	if spec.PresetBase != "minimal" {
		t.Errorf("Expected PresetBase to be 'minimal', got '%s'", spec.PresetBase)
	}

	if spec.SlotsPerEpoch != phase0.Slot(8) {
		t.Errorf("Expected SlotsPerEpoch to be 8, got %d", spec.SlotsPerEpoch)
	}
}

func TestDynSszInitializationMainnet(t *testing.T) {
	spec := createMainnetSpec()

	dynSsz, err := initializeDynSsz(spec)
	if err != nil {
		t.Fatalf("Failed to initialize DynSsz: %v", err)
	}

	if dynSsz == nil {
		t.Error("Expected DynSsz to be initialized, got nil")
	}
}

func TestDynSszInitializationMinimal(t *testing.T) {
	spec := createMinimalSpec()

	dynSsz, err := initializeDynSsz(spec)
	if err != nil {
		t.Fatalf("Failed to initialize DynSsz: %v", err)
	}

	if dynSsz == nil {
		t.Error("Expected DynSsz to be initialized, got nil")
	}
}

func TestDynSszBeforeInitialization(t *testing.T) {
	// Create a Default provider without initializing spec/dynSsz
	log := logrus.New()
	log.SetLevel(logrus.PanicLevel) // Suppress log output during tests

	config := &beacon.Config{
		Mode: beacon.OperatingModeFull,
	}

	provider := beacon.NewDefaultProvider("test", log, []node.Config{}, config)

	// Try to get DynSsz before it's initialized
	_, err := provider.DynSsz()
	if err == nil {
		t.Error("Expected error when accessing DynSsz before initialization, got nil")
	}

	expectedErrMsg := "dynamic SSZ encoder not yet available"
	if err.Error() != expectedErrMsg {
		t.Errorf("Expected error message '%s', got '%s'", expectedErrMsg, err.Error())
	}
}

func TestPresetSpecificValues(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name          string
		createSpec    func() *state.Spec
		expectedBase  string
		expectedSlots phase0.Slot
	}{
		{
			name:          "Mainnet",
			createSpec:    createMainnetSpec,
			expectedBase:  "mainnet",
			expectedSlots: 32,
		},
		{
			name:          "Minimal",
			createSpec:    createMinimalSpec,
			expectedBase:  "minimal",
			expectedSlots: 8,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			spec := tt.createSpec()

			if spec.PresetBase != tt.expectedBase {
				t.Errorf("Expected PresetBase '%s', got '%s'", tt.expectedBase, spec.PresetBase)
			}

			if spec.SlotsPerEpoch != tt.expectedSlots {
				t.Errorf("Expected SlotsPerEpoch %d, got %d", tt.expectedSlots, spec.SlotsPerEpoch)
			}

			// Test that DynSsz can be initialized with this spec
			dynSsz, err := initializeDynSsz(spec)
			if err != nil {
				t.Fatalf("Failed to initialize DynSsz for %s: %v", tt.name, err)
			}

			if dynSsz == nil {
				t.Errorf("Expected DynSsz to be initialized for %s, got nil", tt.name)
			}
		})
	}
}
