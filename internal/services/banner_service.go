package services

import (
	"errors"
	"littleeinsteinchildcare/backend/internal/models"
	"log"
	"sync"
	"time"
)

// BannerService manages the current banner and its expiration
type BannerService struct {
	currentBanner  *models.Banner
	mutex          sync.RWMutex
	timerChan      chan bool
	isTimerRunning bool
}

// NewBannerService creates a new banner service
func NewBannerService() *BannerService {
	return &BannerService{
		currentBanner:  nil,
		mutex:          sync.RWMutex{},
		timerChan:      make(chan bool, 1),
		isTimerRunning: false,
	}
}

// GetCurrentBanner returns the current active banner
func (s *BannerService) GetCurrentBanner() (*models.Banner, error) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	if s.currentBanner == nil {
		return nil, errors.New("no active banner")
	}

	// Check if banner has expired
	if s.currentBanner.IsExpired() {
		// Use a goroutine to avoid deadlock (we already hold a read lock)
		go s.DeleteBanner()
		return nil, errors.New("banner has expired")
	}

	return s.currentBanner, nil
}

// CreateOrUpdateBanner sets a new banner and starts the expiration timer
func (s *BannerService) CreateOrUpdateBanner(banner *models.Banner) error {
	if banner == nil {
		return errors.New("banner cannot be nil")
	}

	// Stop any existing timer
	s.stopTimer()

	s.mutex.Lock()
	s.currentBanner = banner
	s.mutex.Unlock()

	// Start expiration timer
	s.startTimer(banner.ExpiresAt)

	return nil
}

// DeleteBanner removes the current banner and stops the timer
func (s *BannerService) DeleteBanner() error {
	s.stopTimer()

	s.mutex.Lock()
	s.currentBanner = nil
	s.mutex.Unlock()

	return nil
}

// startTimer starts a goroutine to remove the banner when it expires
func (s *BannerService) startTimer(expirationTime time.Time) {
	duration := time.Until(expirationTime)
	if duration <= 0 {
		// If already expired, delete immediately
		s.DeleteBanner()
		return
	}

	s.mutex.Lock()
	// Reset channel if needed
	if s.isTimerRunning {
		s.timerChan <- true // Signal the existing timer to stop
	}

	s.isTimerRunning = true
	timerChan := s.timerChan // Create local reference to avoid race conditions
	s.mutex.Unlock()

	// Start the timer goroutine
	go func(stopChan chan bool, expiry time.Time) {
		log.Printf("Banner timer started, expires at: %v", expiry.Format(time.RFC3339))

		select {
		case <-stopChan:
			// Timer was cancelled
			log.Println("Banner timer cancelled")
			return
		case <-time.After(duration):
			// Timer expired naturally
			log.Println("Banner expired, removing automatically")
			s.DeleteBanner()

			s.mutex.Lock()
			s.isTimerRunning = false
			s.mutex.Unlock()
		}
	}(timerChan, expirationTime)
}

// stopTimer cancels any running expiration timer
func (s *BannerService) stopTimer() {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	if s.isTimerRunning {
		s.timerChan <- true // Signal timer to stop
		s.isTimerRunning = false
	}
}

// IsTimerRunning returns whether an expiration timer is active
func (s *BannerService) IsTimerRunning() bool {
	s.mutex.RLock()
	defer s.mutex.RUnlock()
	return s.isTimerRunning
}
