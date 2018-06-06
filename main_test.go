package response

import (
	"testing"

	"github.com/aphistic/sweet"
	"github.com/aphistic/sweet-junit"
	. "github.com/onsi/gomega"
)

func TestMain(m *testing.M) {
	RegisterFailHandler(sweet.GomegaFail)

	sweet.Run(m, func(s *sweet.S) {
		s.RegisterPlugin(junit.NewPlugin())

		s.AddSuite(&InterfaceSuite{})
		s.AddSuite(&ImplementationSuite{})
		s.AddSuite(&BaseSuite{})
		s.AddSuite(&StreamSuite{})
		s.AddSuite(&IOUtilSuite{})
	})
}
