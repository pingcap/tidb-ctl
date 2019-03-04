package cmd

import (
	. "github.com/pingcap/check"
)

var _ = Suite(&Global{})

type Global struct{}

func (s *Global) TestBase64Encode(c *C) {
	c.Parallel()
	expected := "SGVsbG8sV29ybGQ="
	obtained := base64Encode("Hello,World")
	c.Assert(obtained, Equals, expected)
}

func (s *Global) TestBase64Decode(c *C) {
	c.Parallel()
	expected := "Hello,World"
	obtained, err := base64Decode("SGVsbG8sV29ybGQ=")
	c.Assert(err, IsNil)
	c.Assert(obtained, Equals, expected)
	obtained, err = base64Decode("ThisIsNotBase64")
	c.Assert(obtained, Equals, "")
	c.Assert(err, ErrorMatches, "*illegal.*")

}
