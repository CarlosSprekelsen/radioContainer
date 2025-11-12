package fixtures

// WiFi24GHzChannels returns standard 2.4GHz WiFi channels for testing
func WiFi24GHzChannels() []ChannelProfile {
	return []ChannelProfile{
		{Index: 1, Frequency: 2412.0, Band: "2.4GHz", Type: "WiFi"},
		{Index: 2, Frequency: 2417.0, Band: "2.4GHz", Type: "WiFi"},
		{Index: 3, Frequency: 2422.0, Band: "2.4GHz", Type: "WiFi"},
		{Index: 4, Frequency: 2427.0, Band: "2.4GHz", Type: "WiFi"},
		{Index: 5, Frequency: 2432.0, Band: "2.4GHz", Type: "WiFi"},
		{Index: 6, Frequency: 2437.0, Band: "2.4GHz", Type: "WiFi"},
		{Index: 7, Frequency: 2442.0, Band: "2.4GHz", Type: "WiFi"},
		{Index: 8, Frequency: 2447.0, Band: "2.4GHz", Type: "WiFi"},
		{Index: 9, Frequency: 2452.0, Band: "2.4GHz", Type: "WiFi"},
		{Index: 10, Frequency: 2457.0, Band: "2.4GHz", Type: "WiFi"},
		{Index: 11, Frequency: 2462.0, Band: "2.4GHz", Type: "WiFi"},
	}
}

// WiFi5GHzChannels returns standard 5GHz WiFi channels for testing
func WiFi5GHzChannels() []ChannelProfile {
	return []ChannelProfile{
		{Index: 36, Frequency: 5180.0, Band: "5GHz", Type: "WiFi"},
		{Index: 40, Frequency: 5200.0, Band: "5GHz", Type: "WiFi"},
		{Index: 44, Frequency: 5220.0, Band: "5GHz", Type: "WiFi"},
		{Index: 48, Frequency: 5240.0, Band: "5GHz", Type: "WiFi"},
		{Index: 52, Frequency: 5260.0, Band: "5GHz", Type: "WiFi"},
		{Index: 56, Frequency: 5280.0, Band: "5GHz", Type: "WiFi"},
		{Index: 60, Frequency: 5300.0, Band: "5GHz", Type: "WiFi"},
		{Index: 64, Frequency: 5320.0, Band: "5GHz", Type: "WiFi"},
		{Index: 100, Frequency: 5500.0, Band: "5GHz", Type: "WiFi"},
		{Index: 104, Frequency: 5520.0, Band: "5GHz", Type: "WiFi"},
		{Index: 108, Frequency: 5540.0, Band: "5GHz", Type: "WiFi"},
		{Index: 112, Frequency: 5560.0, Band: "5GHz", Type: "WiFi"},
		{Index: 116, Frequency: 5580.0, Band: "5GHz", Type: "WiFi"},
		{Index: 120, Frequency: 5600.0, Band: "5GHz", Type: "WiFi"},
		{Index: 124, Frequency: 5620.0, Band: "5GHz", Type: "WiFi"},
		{Index: 128, Frequency: 5640.0, Band: "5GHz", Type: "WiFi"},
		{Index: 132, Frequency: 5660.0, Band: "5GHz", Type: "WiFi"},
		{Index: 136, Frequency: 5680.0, Band: "5GHz", Type: "WiFi"},
		{Index: 140, Frequency: 5700.0, Band: "5GHz", Type: "WiFi"},
	}
}

// UHFChannels returns UHF channels for testing
func UHFChannels() []ChannelProfile {
	return []ChannelProfile{
		{Index: 1, Frequency: 400.0, Band: "UHF", Type: "Radio"},
		{Index: 2, Frequency: 401.0, Band: "UHF", Type: "Radio"},
		{Index: 3, Frequency: 402.0, Band: "UHF", Type: "Radio"},
		{Index: 4, Frequency: 403.0, Band: "UHF", Type: "Radio"},
		{Index: 5, Frequency: 404.0, Band: "UHF", Type: "Radio"},
	}
}

// ValidChannelIndex returns a valid channel index for testing
func ValidChannelIndex() int {
	return 6 // Standard 2.4GHz WiFi channel
}

// InvalidChannelIndex returns an invalid channel index for testing
func InvalidChannelIndex() int {
	return 999 // Out of range
}

// ChannelIndexRange returns the valid range of channel indices
func ChannelIndexRange() (min, max int) {
	return 1, 255
}
