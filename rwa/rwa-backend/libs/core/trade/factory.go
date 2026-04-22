package trade

import (
	"fmt"
)

type TradeServiceManager struct {
	service TradeService
}

func NewTradeServiceManager() *TradeServiceManager {
	return &TradeServiceManager{}
}

func (m *TradeServiceManager) InitializeService(alpacaConfig *AlpacaConfig, provider string) error {
	if provider == "" {
		provider = "alpaca" // default to alpaca
	}

	// currently only alpaca is supported
	if provider != "alpaca" {
		return fmt.Errorf("unsupported trading provider: %s, only 'alpaca' is supported", provider)
	}

	if alpacaConfig == nil || alpacaConfig.APIKey == "" || alpacaConfig.APISecret == "" {
		return fmt.Errorf("alpaca credentials not configured")
	}

	m.service = NewAlpacaService(alpacaConfig)
	return nil
}

func (m *TradeServiceManager) GetService() TradeService {
	return m.service
}

func (m *TradeServiceManager) IsInitialized() bool {
	return m.service != nil
}

func (m *TradeServiceManager) GetSupportedProviders() []string {
	return []string{"alpaca"}
}

func (m *TradeServiceManager) SwitchProvider(alpacaConfig *AlpacaConfig, provider string) error {
	return m.InitializeService(alpacaConfig, provider)
}

