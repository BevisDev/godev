package framework

import (
	"github.com/BevisDev/godev/interfaces"
)

type Provider interface {
	interfaces.Initializer
	interfaces.Starter
	interfaces.Stopper
}
