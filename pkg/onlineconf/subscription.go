package onlineconf

func NewSubscription(params []string, callback func() error) *SubscriptionCallback {
	return &SubscriptionCallback{path: params, callback: callback}
}

func (s *SubscriptionCallback) GetPaths() []string {
	return s.path
}

func (s *SubscriptionCallback) InvokeCallback() error {
	return s.callback()
}
