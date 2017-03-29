package configstring

import (
	"testing"
	"github.com/cosminrentea/gobbler/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/golang/mock/gomock"
)

func TestList_IsEmpty(t *testing.T) {
	ctrl, finish := testutil.NewMockCtrl(t)
	defer finish()

	m := NewMockSettings(ctrl)
	m.EXPECT().SetValue(gomock.Any())

	cs := NewFromKingpin(m)

	assert.True(t, cs.IsEmpty())
}