package onlineconf

import "github.com/Nikolo/go-onlineconf/pkg/onlineconfInterface"

func NewSubscription(params []string, callback func() error) onlineconfInterface.SubscriptionCallback {
	return &SubscriptionCallback{path: params, callback: callback}
}

func (s *SubscriptionCallback) GetPaths() []string {
	return s.path
}

func (s *SubscriptionCallback) InvokeCallback() error {
	return s.callback()
}
