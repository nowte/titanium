package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"time"
)

const titaniumDir = ".titanium"
const configFile = "config.json"

// TitaniumConfig — .titanium/config.json yapısı
type TitaniumConfig struct {
	Version     string    `json:"version"`
	InitiatedAt time.Time `json:"initiated_at"`
	Rules       Rules     `json:"rules"`
}

// Rules — Branch ve commit kuralları (Faz 3'e zemin hazırlar)
type Rules struct {
	ProtectedBranches []string `json:"protected_branches"`
	RequireCommitMsg  bool     `json:"require_commit_msg"`
}

// ── State işlemleri ────────────────────────────────────────────────────────────

// IsTitaniumRepo — mevcut dizinde .titanium/ var mı kontrol eder.
func IsTitaniumRepo() bool {
	_, err := os.Stat(filepath.Join(titaniumDir, configFile))
	return err == nil
}

// InitTitanium — .titanium/ klasörünü oluşturur ve config.json yazar.
func InitTitanium() error {
	if err := os.MkdirAll(titaniumDir, 0755); err != nil {
		return err
	}

	cfg := TitaniumConfig{
		Version:     "0.1.0",
		InitiatedAt: time.Now().UTC(),
		Rules: Rules{
			ProtectedBranches: []string{"main", "master", "release"},
			RequireCommitMsg:  true,
		},
	}

	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(filepath.Join(titaniumDir, configFile), data, 0644)
}

// LoadConfig — mevcut config'i okur.
func LoadConfig() (*TitaniumConfig, error) {
	data, err := os.ReadFile(filepath.Join(titaniumDir, configFile))
	if err != nil {
		return nil, err
	}
	var cfg TitaniumConfig
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}
	return &cfg, nil
}

// IsProtectedBranch — branch adının korumalı olup olmadığını kontrol eder.
func IsProtectedBranch(branch string) bool {
	cfg, err := LoadConfig()
	if err != nil {
		// Config yoksa varsayılan kurallar
		for _, b := range []string{"main", "master", "release"} {
			if b == branch {
				return true
			}
		}
		return false
	}
	for _, b := range cfg.Rules.ProtectedBranches {
		if b == branch {
			return true
		}
	}
	return false
}
