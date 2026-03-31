package main

import (
	"os"
	"testing"
)

func TestSelectRegistry(t *testing.T) {
	tests := []struct {
		name           string
		cmdName        string
		upstreamEnv    string
		expectedTarget string
		expectedEnvVar string
	}{
		{
			name:           "npm default",
			cmdName:        "npm",
			upstreamEnv:    "",
			expectedTarget: "https://registry.npmjs.org",
			expectedEnvVar: "NPM_CONFIG_REGISTRY",
		},
		{
			name:           "pip default",
			cmdName:        "pip",
			upstreamEnv:    "",
			expectedTarget: "https://pypi.org",
			expectedEnvVar: "PIP_INDEX_URL",
		},
		{
			name:           "npm override",
			cmdName:        "npm",
			upstreamEnv:    "http://localhost:8080",
			expectedTarget: "http://localhost:8080",
			expectedEnvVar: "NPM_CONFIG_REGISTRY",
		},
		{
			name:           "pip override",
			cmdName:        "pip",
			upstreamEnv:    "http://localhost:8081",
			expectedTarget: "http://localhost:8081",
			expectedEnvVar: "PIP_INDEX_URL",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			os.Setenv("BUFFER7_UPSTREAM_REGISTRY", tt.upstreamEnv)
			defer os.Unsetenv("BUFFER7_UPSTREAM_REGISTRY")

			target, envVar := selectRegistry(tt.cmdName)
			if target != tt.expectedTarget {
				t.Errorf("selectRegistry() target = %v, want %v", target, tt.expectedTarget)
			}
			if envVar != tt.expectedEnvVar {
				t.Errorf("selectRegistry() envVar = %v, want %v", envVar, tt.expectedEnvVar)
			}
		})
	}
}
