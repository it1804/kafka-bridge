package config

type (
	ExporterServiceList struct {
		Services []ExporterService `json:"services"`
		Metrics  MetricsService    `json:"metrics"`
	}

	ExporterService struct {
		Name string `json:"name"`
		Url  string `json:"url"`
	}

	MetricsService struct {
		Listen string `json:"listen,omitempty"`
		Path   string `json:"path,omitempty"`
	}
)
