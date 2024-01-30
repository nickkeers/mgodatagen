package generators_test

import (
	"github.com/feliixx/mgodatagen/datagen/generators"
	"go.mongodb.org/mongo-driver/bson"
	"testing"
)

var (
	topLeft     = []float64{-4.5, 53} // Min longitude, Max latitude
	bottomRight = []float64{1.7, 50}  // Max longitude, Min latitude
)

func checkInBounds(longLat []float64) bool {
	if len(longLat) != 2 {
		return false // Incorrect input format
	}

	long := longLat[0]
	lat := longLat[1]

	// Ensure that the provided longitude is between the longitudes of topLeft and bottomRight
	// and the latitude is between the latitudes of bottomRight and topLeft,
	// reflecting the [longitude, latitude] order.
	return long >= topLeft[0] && long <= bottomRight[0] &&
		lat >= bottomRight[1] && lat <= topLeft[1]
}

func TestDocumentWithValidCoordinateRange(t *testing.T) {
	ci := generators.NewCollInfo(1, []int{3, 6, 4}, defaultSeed, nil, nil)

	docGenerator, err := ci.NewDocumentGenerator(map[string]generators.Config{
		"coords": {
			Type:               generators.TypeCoordinates,
			LatLongTopLeft:     topLeft,
			LatLongBottomRight: bottomRight,
		},
	})

	if err != nil {
		t.Error(err)
	}

	var d struct {
		Key []float64 `bson:"coords"`
	}

	for i := 0; i < 10; i++ {
		err := bson.Unmarshal(docGenerator.Generate(), &d)

		if err != nil {
			t.Error(err)
		}

		if !checkInBounds(d.Key) {
			t.Errorf("%.2f not in range", d.Key)
		}
	}
}
