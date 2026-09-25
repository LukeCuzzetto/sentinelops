package health

import (
	"github.com/LukeCuzzetto/sentinelops/internal/telemetry"
)

type Status string

const (
	StatusNominal  Status = "nominal"
	StatusWarning  Status = "warning"
	StatusCritical Status = "critical"
)

type Finding struct {
	Code    string `json:"code"`
	Status  Status `json:"status"`
	Message string `json:"message"`
}

type Assessment struct {
	Status   Status    `json:"status"`
	Findings []Finding `json:"findings"`
}

type Thresholds struct {
	BatterySOCWarning  float64
	BatterySOCCritical float64

	TemperatureHighWarning  float64
	TemperatureHighCritical float64

	TemperatureLowWarning  float64
	TemperatureLowCritical float64
}

func Evaluate(sample telemetry.Sample, thresholds Thresholds) Assessment {
	assessment := Assessment{
		Status:   StatusNominal,
		Findings: []Finding{},
	}

	// Battery SOC checks

	if sample.BatterySOCPercent < thresholds.BatterySOCCritical {

		assessment.Findings = append(assessment.Findings, Finding{
			Code:    "battery_soc_critical",
			Status:  StatusCritical,
			Message: "Battery state of charge is below critical threshold",
		})
		assessment.Status = StatusCritical
	} else if sample.BatterySOCPercent < thresholds.BatterySOCWarning {
		assessment.Findings = append(assessment.Findings, Finding{
			Code:    "battery_soc_warning",
			Status:  StatusWarning,
			Message: "Battery state of charge is below warning threshold",
		})
		if assessment.Status != StatusCritical {
			assessment.Status = StatusWarning
		}
	}

	// Temperature checks start here

	if sample.TemperatureC >= thresholds.TemperatureHighCritical {
		assessment.Findings = append(assessment.Findings, Finding{
			Code:    "temperature_high_critical",
			Status:  StatusCritical,
			Message: "Temperature is above critical threshold",
		})
		assessment.Status = StatusCritical
	} else if sample.TemperatureC >= thresholds.TemperatureHighWarning {
		assessment.Findings = append(assessment.Findings, Finding{
			Code:    "temperature_high_warning",
			Status:  StatusWarning,
			Message: "Temperature is above warning threshold",
		})
		if assessment.Status != StatusCritical {
			assessment.Status = StatusWarning
		}
	}

	if sample.TemperatureC <= thresholds.TemperatureLowCritical {
		assessment.Findings = append(assessment.Findings, Finding{
			Code:    "temperature_low_critical",
			Status:  StatusCritical,
			Message: "Temperature is below critical threshold",
		})
		assessment.Status = StatusCritical
	} else if sample.TemperatureC <= thresholds.TemperatureLowWarning {
		assessment.Findings = append(assessment.Findings, Finding{
			Code:    "temperature_low_warning",
			Status:  StatusWarning,
			Message: "Temperature is below warning threshold",
		})
		if assessment.Status != StatusCritical {
			assessment.Status = StatusWarning
		}
	}

	return assessment
}
