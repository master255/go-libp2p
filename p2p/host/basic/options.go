package basichost

type Option func(*natManager) error

// WithUserAgent is a natManager option that sets specific user agent for the NAT manager.
func WithUserAgent(userAgent string) Option {
	return func(nmgr *natManager) error {
		nmgr.userAgent = userAgent
		return nil
	}
}