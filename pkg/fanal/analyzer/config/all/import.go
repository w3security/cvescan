package all

import (
	_ "github.com/w3security/cvescan/pkg/fanal/analyzer/config/dockerfile"
	_ "github.com/w3security/cvescan/pkg/fanal/analyzer/config/helm"
	_ "github.com/w3security/cvescan/pkg/fanal/analyzer/config/json"
	_ "github.com/w3security/cvescan/pkg/fanal/analyzer/config/terraform"
	_ "github.com/w3security/cvescan/pkg/fanal/analyzer/config/yaml"
)
