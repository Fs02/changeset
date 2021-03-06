package changeset

import (
	"testing"

	"github.com/Fs02/rel"
	"github.com/stretchr/testify/assert"
)

func TestForeignKeyConstraint(t *testing.T) {
	ch := &Changeset{}
	assert.Nil(t, ch.Constraints())

	ForeignKeyConstraint(ch, "field1")
	assert.Equal(t, 1, len(ch.Constraints()))
	assert.Equal(t, rel.ForeignKeyConstraint, ch.Constraints()[0].Type)
}
