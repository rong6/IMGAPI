package providers

import "sync"

var (
	registry = make(map[string]Provider)
	regMutex sync.RWMutex
)

// init 初始化所有提供商
func init() {
	RegisterProvider(&Provider111666{})
	RegisterProvider(&ProviderMeituan{})
	RegisterProvider(&Provider360Tu{})
	RegisterProvider(&ProviderCloudinary{})
	RegisterProvider(&ProviderIPFS{})
	RegisterProvider(&ProviderNodeSeek{})
	RegisterProvider(&ProviderEroLabs{})
	// RegisterProvider(&ProviderExample{})
}

// RegisterProvider 注册提供商
func RegisterProvider(provider Provider) {
	regMutex.Lock()
	defer regMutex.Unlock()
	registry[provider.GetName()] = provider
}

// GetProvider 获取提供商
func GetProvider(name string) (Provider, bool) {
	regMutex.RLock()
	defer regMutex.RUnlock()
	provider, exists := registry[name]
	return provider, exists
}

// GetAllProviders 获取所有提供商
func GetAllProviders() []Provider {
	regMutex.RLock()
	defer regMutex.RUnlock()

	var providers []Provider
	for _, provider := range registry {
		providers = append(providers, provider)
	}
	return providers
}

// GetEnabledProviders 获取所有启用的提供商
func GetEnabledProviders() []Provider {
	regMutex.RLock()
	defer regMutex.RUnlock()

	var providers []Provider
	for _, provider := range registry {
		if provider.IsEnabled() {
			providers = append(providers, provider)
		}
	}
	return providers
}
