package cmd

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFriendlyName(t *testing.T) {
	assert.Equal(t, "flag1-flag2-flag3", FriendlyName("flag1.flag2-flag3"))
}
