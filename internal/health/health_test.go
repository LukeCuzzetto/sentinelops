package health

import (
	"testing"

	"github.com/LukeCuzzetto/sentinelops/internal/telemetry"
)

func TestEvaluateBatterySOC(t *testing.T) {
	thresholds := Thresholds{
		BatterySOCWarning:       30.0,
		BatterySOCCritical:      15.0,
		TemperatureHighWarning:  40.0,
		TemperatureHighCritical: 50.0,
		TemperatureLowWarning:   0.0,
		TemperatureLowCritical:  -10.0,
	}

	test := []struct {
		name           string
		batterySOC     float64
		expectedStatus Status
	}{
		{
			name:           "nominal battery",
			batterySOC:     75.0,
			expectedStatus: StatusNominal,
		},

		{
			name:           "warning battery",
			batterySOC:     25.0,
			expectedStatus: StatusWarning,
		},
		{
			name:           "critical battery",
			batterySOC:     10.0,
			expectedStatus: StatusCritical,
		},
	}

	for _, test := range test {
		t.Run(test.name, func(t *testing.T) {
			sample := telemetry.Sample{
				BatterySOCPercent: test.batterySOC,
				TemperatureC:      20.0, // Nominal temperature for battery SOC tests
			}

			assessment := Evaluate(sample, thresholds)

			if assessment.Status != test.expectedStatus {
				t.Errorf("expected status %q, got %q", test.expectedStatus, assessment.Status)
			}

		})
	}
}

func TestEvaluateTemperature(t *testing.T) {
	thresholds := Thresholds{
		BatterySOCWarning:       30.0,
		BatterySOCCritical:      15.0,
		TemperatureHighWarning:  40.0,
		TemperatureHighCritical: 50.0,
		TemperatureLowWarning:   0.0,
		TemperatureLowCritical:  -10.0,
	}

	test := []struct {
		name           string
		temperature    float64
		expectedStatus Status
	}{
		{
			name:           "nominal temperature",
			temperature:    25.0,
			expectedStatus: StatusNominal,
		},
		{
			name:           "warning high temperature",
			temperature:    45.0,
			expectedStatus: StatusWarning,
		},
		{
			name:           "critical high temperature",
			temperature:    55.0,
			expectedStatus: StatusCritical,
		},
		{
			name:           "warning low temperature",
			temperature:    -5.0,
			expectedStatus: StatusWarning,
		},
		{
			name:           "critical low temperature",
			temperature:    -15.0,
			expectedStatus: StatusCritical,
		},
	}

	for _, test := range test {
		t.Run(test.name, func(t *testing.T) {
			sample := telemetry.Sample{
				BatterySOCPercent: 75.0, // Nominal battery SOC for temperature tests
				TemperatureC:      test.temperature,
			}

			assessment := Evaluate(sample, thresholds)

			if assessment.Status != test.expectedStatus {
				t.Errorf("expected status %q, got %q", test.expectedStatus, assessment.Status)
			}
		})
	}
}

func TestMultipleFindings(t *testing.T) {
	thresholds := Thresholds{
		BatterySOCWarning:       30.0,
		BatterySOCCritical:      15.0,
		TemperatureHighWarning:  40.0,
		TemperatureHighCritical: 50.0,
		TemperatureLowWarning:   0.0,
		TemperatureLowCritical:  -10.0,
	}

	sample := telemetry.Sample{
		BatterySOCPercent: 10.0, // Critical battery SOC
		TemperatureC:      45.0, // Warning high temperature
	}

	assessment := Evaluate(sample, thresholds)

	if assessment.Status != StatusCritical {
		t.Errorf("expected status %q, got %q", StatusCritical, assessment.Status)
	}

	if len(assessment.Findings) != 2 {
		t.Errorf("expected 2 findings, got %d", len(assessment.Findings))
	}

}
