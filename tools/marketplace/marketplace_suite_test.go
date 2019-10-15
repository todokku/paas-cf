package main_test

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestMarketplace(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Marketplace Suite")
}
