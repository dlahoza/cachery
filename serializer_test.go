package cachery

import (
	"fmt"
	"math/rand"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

type A struct {
	Name     string
	BirthDay time.Time
	Phone    string
	Siblings int
	Spouse   bool
	Money    float64
}

func randString(l int) string {
	buf := make([]byte, l)
	for i := 0; i < (l+1)/2; i++ {
		buf[i] = byte(rand.Intn(256))
	}
	return fmt.Sprintf("%x", buf)[:l]
}

func TestSerializer(t *testing.T) {
	a := assert.New(t)

	tt := map[string]Serializer{
		"Gob":  new(GobSerializer),
		"JSON": new(JSONSerializer),
	}
	orig := make([]*A, 0, 1000)
	for i := 0; i < 1000; i++ {
		orig = append(orig, &A{
			Name:     randString(16),
			BirthDay: time.Now().Round(time.Nanosecond),
			Phone:    randString(10),
			Siblings: rand.Intn(5),
			Spouse:   rand.Intn(2) == 1,
			Money:    rand.Float64(),
		})
	}

	for k := range tt {
		t.Run(k, func(t *testing.T) {
			buf, err := tt[k].Serialize(orig)
			a.NoError(err)
			var dst []*A
			err = tt[k].Deserialize(buf, &dst)
			a.NoError(err)
			a.EqualValues(orig, dst)
		})
	}
}
