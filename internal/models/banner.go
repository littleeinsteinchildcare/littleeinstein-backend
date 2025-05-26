package models

import (
	"errors"
	"time"
)

// Valid banner types of
const ( //untyped constants
	BannerTypeWeather = "weather"
	BannerTypeClosure = "closure"
	BannerTypeCustom  = "custom"
)

const maxDuration = 72 * time.Hour

// Banner model for displaying site-wide notifications
type Banner struct {
	Type      string    // weather, closure, or custom
	Message   string    // Required for custom type
	ExpiresAt time.Time // Auto expire time of type of time.Time
}

// NewBanner creates a new Banner instance with validation
func NewBanner(bannerType string, message string, expiresAt time.Time) (*Banner, error) {
	// Validate banner type
	if bannerType != BannerTypeWeather && bannerType != BannerTypeClosure && bannerType != BannerTypeCustom {
		return nil, errors.New("invalid banner type: must be weather, closure, or custom")
	}

	// Validate message is provided for custom type
	if bannerType == BannerTypeCustom && message == "" {
		return nil, errors.New("message is required for custom banner type")
	}

	now := time.Now()

	// Validate expiresAt is in the future
	if expiresAt.Before(now) {
		return nil, errors.New("expiration time must be in the future")
	}

	//Validate expiration is not more than 72 hours in the future
	maxAllowedTime := now.Add(maxDuration)
	if expiresAt.After(maxAllowedTime) {
		return nil, errors.New("expiration time cannot be more than 72 hours in the future")
	}

	return &Banner{
		Type:      bannerType,
		Message:   message,
		ExpiresAt: expiresAt,
	}, nil
}

// IsExpired checks if the banner has expired
func (b *Banner) IsExpired() bool {
	return time.Now().After(b.ExpiresAt)
}
