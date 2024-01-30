package generators

import (
	"strconv"

	"github.com/MichaelTJones/pcg"
	"go.mongodb.org/mongo-driver/bson"
)

// Generator for creating random GPS coordinates
type positionGenerator struct {
	base
	pcg64       *pcg.PCG64
	topLeft     []float64 // [latitude, longitude]
	bottomRight []float64 // [latitude, longitude]
}

// validateCoordinates validates that topLeft is northwest of bottomRight considering [longitude, latitude] order
func validateCoordinates(topLeft, bottomRight []float64) bool {
	// Validate longitude range: topLeft's longitude should be less than bottomRight's longitude
	if topLeft[0] >= bottomRight[0] {
		return false
	}
	// Validate latitude range: topLeft's latitude should be greater than bottomRight's latitude
	if topLeft[1] <= bottomRight[1] {
		return false
	}
	// Ensure longitude and latitude are within valid global ranges
	if topLeft[0] < -180 || bottomRight[0] > 180 || topLeft[1] > 90 || bottomRight[1] < -90 {
		return false
	}
	return true
}

func newPositionGenerator(base base, pcg64 *pcg.PCG64, topLeft, bottomRight []float64) (Generator, error) {
	var defaultTopLeft = []float64{-180, 90}     // Min longitude, Max latitude (northwest top)
	var defaultBottomRight = []float64{180, -90} // Max longitude, Min latitude (southeast bottom)

	// Proceed with validation, relying on correct value order [longitude, latitude].
	if topLeft == nil || bottomRight == nil || len(topLeft) != 2 || len(bottomRight) != 2 || !validateCoordinates(topLeft, bottomRight) {
		topLeft = defaultTopLeft
		bottomRight = defaultBottomRight
	}

	return &positionGenerator{
		base:        base,
		pcg64:       pcg64,
		topLeft:     topLeft,
		bottomRight: bottomRight,
	}, nil
}

func (g *positionGenerator) randomInRange(min, max float64) float64 {
	return min + (max-min)*(float64(g.pcg64.Random())/(1<<64))
}

func (g *positionGenerator) EncodeValue() {
	current := g.buffer.Len()
	g.buffer.Reserve()

	// Generate a random longitude within the bounding box
	longitude := g.randomInRange(g.topLeft[0], g.bottomRight[0])

	// Generate a random latitude within the bounding box
	latitude := g.randomInRange(g.bottomRight[1], g.topLeft[1])

	// longitude
	g.buffer.WriteSingleByte(byte(bson.TypeDouble))
	g.buffer.WriteSingleByte(indexesBytes[0])
	g.buffer.WriteSingleByte(byte(0))
	g.buffer.Write(float64Bytes(longitude))

	// latitude
	g.buffer.WriteSingleByte(byte(bson.TypeDouble))
	g.buffer.WriteSingleByte(indexesBytes[1])
	g.buffer.WriteSingleByte(byte(0))
	g.buffer.Write(float64Bytes(latitude))

	g.buffer.WriteSingleByte(byte(0))
	g.buffer.WriteAt(current, int32Bytes(int32(g.buffer.Len()-current)))
}

func (g *positionGenerator) EncodeValueAsString() {
	// Generate longitude in the range [topLeft[0], bottomRight[0]]
	longitude := g.randomInRange(g.topLeft[0], g.bottomRight[0])

	// Generate latitude in the range [bottomRight[1], topLeft[1]]
	latitude := g.randomInRange(g.bottomRight[1], g.topLeft[1])

	g.buffer.WriteSingleByte('[')
	// Since longitude comes first in the [longitude, latitude] order:
	g.buffer.WriteString(strconv.FormatFloat(longitude, 'f', 10, 64))
	g.buffer.WriteSingleByte(',')
	g.buffer.WriteString(strconv.FormatFloat(latitude, 'f', 10, 64))
	g.buffer.WriteSingleByte(']')
}
