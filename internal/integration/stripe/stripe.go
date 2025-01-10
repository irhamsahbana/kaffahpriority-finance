package stripe

type StripeIntegrationContract interface {
}

type stripeIntegration struct {
}

func NewStripeIntegration() StripeIntegrationContract {
	return &stripeIntegration{}
}
