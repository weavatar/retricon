package retricon

import (
	"crypto/sha512"
	"encoding/hex"
	"errors"
	"image"
	"image/color"
	"math"
	"strings"
)

// Style represents predefined configurations for retricon generation
type Style string

const (
	Default  Style = "default"
	Github   Style = "github"
	Gravatar Style = "gravatar"
	Mono     Style = "mono"
	Mosaic   Style = "mosaic"
	Mini     Style = "mini"
	Window   Style = "window"
)

// Options represents configuration options for retricon generation
type Options struct {
	Tiles         int
	TileSize      int
	TileColor     any // Can be int, string, or color.RGBA
	BgColor       any // Can be int, string, or color.RGBA
	TilePadding   int
	ImagePadding  int
	MinFill       float64
	MaxFill       float64
	VerticalSym   bool
	HorizontalSym bool
}

// ApplyStyle configures options based on predefined styles
func (o *Options) ApplyStyle(style Style) error {
	switch style {
	case Default:
		// Default settings already applied
	case Github:
		o.TileSize = 70
		o.BgColor = "F0F0F0"
		o.TilePadding = -1
		o.ImagePadding = 35
		o.Tiles = 5
		o.VerticalSym = true
		o.HorizontalSym = false
	case Gravatar:
		o.BgColor = 1
		o.Tiles = 8
		o.VerticalSym = true
		o.HorizontalSym = false
	case Mono:
		o.BgColor = "F0F0F0"
		o.TileColor = "000000"
		o.Tiles = 6
		o.TileSize = 12
		o.TilePadding = -1
		o.ImagePadding = 6
		o.VerticalSym = true
		o.HorizontalSym = false
	case Mosaic:
		o.ImagePadding = 2
		o.TilePadding = 1
		o.TileSize = 16
		o.BgColor = "F0F0F0"
		o.VerticalSym = true
		o.HorizontalSym = false
	case Mini:
		o.TileSize = 10
		o.TilePadding = 1
		o.Tiles = 3
		o.BgColor = 0
		o.TileColor = 1
		o.VerticalSym = false
		o.HorizontalSym = false
	case Window:
		o.TileColor = color.RGBA{R: 255, G: 255, B: 255, A: 255}
		o.BgColor = 0
		o.ImagePadding = 2
		o.TilePadding = 1
		o.TileSize = 16
		o.VerticalSym = true
		o.HorizontalSym = false
	default:
		return errors.New("invalid style parameter")
	}
	return nil
}

// New creates a new reticon image with the given name and optional style
func New(name string, style ...Style) (image.Image, error) {
	opts := Options{
		Tiles:         5,
		TileSize:      1,
		TileColor:     0,
		BgColor:       nil,
		TilePadding:   0,
		ImagePadding:  0,
		MinFill:       0.3,
		MaxFill:       0.9,
		VerticalSym:   true,
		HorizontalSym: false,
	}
	if len(style) > 0 {
		if err := opts.ApplyStyle(style[0]); err != nil {
			return nil, err
		}
	}
	return NewWithOptions(name, opts)
}

// NewWithOptions creates a new reticon image with custom options
func NewWithOptions(name string, opts Options) (image.Image, error) {
	if opts.Tiles < 1 {
		return nil, errors.New("tiles must be greater than 0")
	}
	if opts.TileSize < 1 {
		return nil, errors.New("tile size must be greater than 0")
	}
	if opts.MinFill <= 0 {
		opts.MinFill = 0.3
	}
	if opts.MaxFill <= 0 {
		opts.MaxFill = 0.9
	}

	dimension := opts.Tiles
	var useColor bool

	_, isTileColorInt := opts.TileColor.(int)
	_, isBgColorInt := opts.BgColor.(int)
	useColor = isTileColorInt || isBgColorInt

	var raw *rawData
	var err error
	var pic [][]int

	mid := int(math.Ceil(float64(dimension) / 2.0))

	if opts.VerticalSym && opts.HorizontalSym {
		raw, err = idHash(name, mid*mid, opts.MinFill, opts.MaxFill, useColor)
		if err != nil {
			return nil, err
		}
		pic = fillPixelsCentSym(raw, dimension)
	} else if opts.VerticalSym || opts.HorizontalSym {
		raw, err = idHash(name, mid*dimension, opts.MinFill, opts.MaxFill, useColor)
		if err != nil {
			return nil, err
		}
		if opts.VerticalSym {
			pic = fillPixelsVertSym(raw, dimension)
		} else {
			pic = fillPixelsHoriSym(raw, dimension)
		}
	} else {
		raw, err = idHash(name, dimension*dimension, opts.MinFill, opts.MaxFill, useColor)
		if err != nil {
			return nil, err
		}
		pic = fillPixels(raw, dimension)
	}

	// Default to transparent background if not specified
	if opts.BgColor == nil {
		opts.BgColor = color.RGBA{}
	}

	bgColor, err := parseColor(opts.BgColor, raw)
	if err != nil {
		return nil, err
	}
	tileColor, err := parseColor(opts.TileColor, raw)
	if err != nil {
		return nil, err
	}

	tileWidth := opts.TileSize + opts.TilePadding*2
	canvasSize := tileWidth*opts.Tiles + opts.ImagePadding*2

	// Create the base image
	im := image.NewRGBA(image.Rect(0, 0, canvasSize, canvasSize))

	// Fill the background
	for y := 0; y < im.Bounds().Dy(); y++ {
		for x := 0; x < im.Bounds().Dx(); x++ {
			im.Set(x, y, bgColor)
		}
	}

	// Draw the tiles
	for y := 0; y < dimension; y++ {
		for x := 0; x < dimension; x++ {
			if pic[y][x] == 1 {
				x0 := (x * tileWidth) + opts.TilePadding + opts.ImagePadding
				y0 := (y * tileWidth) + opts.TilePadding + opts.ImagePadding

				// Draw the rectangle tile
				for py := y0; py < y0+opts.TileSize; py++ {
					for px := x0; px < x0+opts.TileSize; px++ {
						im.Set(px, py, tileColor)
					}
				}
			}
		}
	}

	return im, nil
}

// MustNew creates a new retricon image or panics on error
func MustNew(name string, style ...Style) image.Image {
	img, err := New(name, style...)
	if err != nil {
		panic(err)
	}
	return img
}

// MustNewWithOptions creates a new retricon image with options or panics on error
func MustNewWithOptions(name string, opts Options) image.Image {
	img, err := NewWithOptions(name, opts)
	if err != nil {
		panic(err)
	}
	return img
}

// rawData represents the hash result with color and pixel data
type rawData struct {
	Colors []color.RGBA
	Pixels []int
}

// brightness calculates the perceived brightness of a color
// http://www.nbdtech.com/Blog/archive/2008/04/27/Calculating-the-Perceived-Brightness-of-a-Color.aspx
func brightness(r, g, b uint8) float64 {
	return math.Sqrt(0.241*float64(r)*float64(r) + 0.691*float64(g)*float64(g) + 0.068*float64(b)*float64(b))
}

// fixedLengthHash generates a hash of specified length
func fixedLengthHash(buf []byte, length int) ([]byte, error) {
	if length > 64 {
		return nil, errors.New("sha512 can only generate 64B of data")
	}

	hash := sha512.Sum512(buf)
	hexStr := hex.EncodeToString(hash[:])

	hexLength := length * 2
	if len(hexStr)%hexLength != 0 {
		hexStr += strings.Repeat("0", hexLength-len(hexStr)%hexLength)
	}

	ret := hexStr[:hexLength]
	for i := hexLength; i < len(hexStr); i += hexLength {
		segment := hexStr[i : i+hexLength]
		retVal, _ := hex.DecodeString(ret)
		segVal, _ := hex.DecodeString(segment)

		result := make([]byte, length)
		for j := 0; j < length; j++ {
			if j < len(retVal) && j < len(segVal) {
				result[j] = retVal[j] ^ segVal[j]
			}
		}
		ret = hex.EncodeToString(result)
	}

	if len(ret) < hexLength {
		ret = strings.Repeat("0", hexLength-len(ret)) + ret
	}

	result, err := hex.DecodeString(ret)
	if err != nil {
		return nil, err
	}
	return result, nil
}

// idHash generates a hash with specific fill characteristics
func idHash(name string, length int, minFill, maxFill float64, useColors bool) (*rawData, error) {
	buf := []byte(name + " ")
	neededBytes := int(math.Ceil(float64(length) / 8.0))
	if useColors {
		neededBytes += 6
	}

	for i := 0; i < 256; i++ {
		bufCopy := make([]byte, len(buf))
		copy(bufCopy, buf)
		bufCopy[len(bufCopy)-1] = byte(i)

		fp, err := fixedLengthHash(bufCopy, neededBytes)
		if err != nil {
			return nil, err
		}

		var pixels []int
		setPixels := 0
		var colors []color.RGBA

		if useColors {
			colors = []color.RGBA{
				{fp[0], fp[1], fp[2], 255},
				{fp[3], fp[4], fp[5], 255},
			}

			// Sort colors by brightness
			if brightness(colors[0].R, colors[0].G, colors[0].B) > brightness(colors[1].R, colors[1].G, colors[1].B) {
				colors[0], colors[1] = colors[1], colors[0]
			}

			fp = fp[6:]
		} else {
			colors = []color.RGBA{}
		}

		for _, b := range fp {
			for offset := 7; offset >= 0; offset-- {
				pixelVal := (b >> offset) & 1
				if pixelVal == 1 {
					pixels = append(pixels, 1)
					setPixels++
				} else {
					pixels = append(pixels, 0)
				}

				if len(pixels) == length {
					break
				}
			}
			if len(pixels) == length {
				break
			}
		}

		fillRatio := float64(setPixels) / float64(length)
		if minFill < fillRatio && fillRatio < maxFill {
			return &rawData{
				Colors: colors,
				Pixels: pixels,
			}, nil
		}
	}

	return nil, errors.New("string unhashable in single-byte search space")
}

// fillPixels arranges pixels in a square grid
func fillPixels(raw *rawData, dimension int) [][]int {
	pic := make([][]int, dimension)
	for row := 0; row < dimension; row++ {
		pic[row] = make([]int, dimension)
		for col := 0; col < dimension; col++ {
			i := row*dimension + col
			pic[row][col] = raw.Pixels[i]
		}
	}
	return pic
}

// fillPixelsVertSym arranges pixels with vertical symmetry
func fillPixelsVertSym(raw *rawData, dimension int) [][]int {
	mid := int(math.Ceil(float64(dimension) / 2.0))
	odd := dimension%2 != 0

	pic := make([][]int, dimension)
	for row := 0; row < dimension; row++ {
		pic[row] = make([]int, dimension)
		for col := 0; col < dimension; col++ {
			var i int
			if col < mid {
				i = row*mid + col
			} else {
				distMiddle := mid - col
				if odd {
					distMiddle -= 1
				}
				distMiddle = int(math.Abs(float64(distMiddle)))
				i = row*mid + mid - 1 - distMiddle
			}
			pic[row][col] = raw.Pixels[i]
		}
	}
	return pic
}

// fillPixelsCentSym arranges pixels with central symmetry
func fillPixelsCentSym(raw *rawData, dimension int) [][]int {
	mid := int(math.Ceil(float64(dimension) / 2.0))
	odd := dimension%2 != 0

	pic := make([][]int, dimension)
	for row := 0; row < dimension; row++ {
		pic[row] = make([]int, dimension)
		for col := 0; col < dimension; col++ {
			var distMiddle int
			if col >= mid {
				distMiddle = mid - col
				if odd {
					distMiddle -= 1
				}
				distMiddle = int(math.Abs(float64(distMiddle)))
			}

			var i int
			if row < mid {
				if col < mid {
					i = (row * mid) + col
				} else {
					i = (row * mid) + mid - 1 - distMiddle
				}
			} else {
				if col < mid {
					i = (dimension-1-row)*mid + col
				} else {
					i = (dimension-1-row)*mid + mid - 1 - distMiddle
				}
			}
			pic[row][col] = raw.Pixels[i]
		}
	}
	return pic
}

// fillPixelsHoriSym arranges pixels with horizontal symmetry
func fillPixelsHoriSym(raw *rawData, dimension int) [][]int {
	mid := int(math.Ceil(float64(dimension) / 2.0))

	pic := make([][]int, dimension)
	for row := 0; row < dimension; row++ {
		pic[row] = make([]int, dimension)
		for col := 0; col < dimension; col++ {
			var i int
			if row < mid {
				i = (row * dimension) + col
			} else {
				i = (dimension-1-row)*dimension + col
			}
			pic[row][col] = raw.Pixels[i]
		}
	}
	return pic
}

// parseColor converts various color formats to RGBA
func parseColor(c any, raw *rawData) (color.RGBA, error) {
	switch v := c.(type) {
	case int:
		if v < 0 || v >= len(raw.Colors) {
			return color.RGBA{}, errors.New("color index out of range")
		}
		return raw.Colors[v], nil
	case string:
		if len(v) < 6 {
			return color.RGBA{}, errors.New("hex color string must be at least 6 characters")
		}
		r, err := hex.DecodeString(v[0:2])
		if err != nil {
			return color.RGBA{}, err
		}
		g, err := hex.DecodeString(v[2:4])
		if err != nil {
			return color.RGBA{}, err
		}
		b, err := hex.DecodeString(v[4:6])
		if err != nil {
			return color.RGBA{}, err
		}
		return color.RGBA{R: r[0], G: g[0], B: b[0], A: 255}, nil
	case color.RGBA:
		return v, nil
	case []uint8:
		if len(v) >= 3 {
			alpha := uint8(255)
			if len(v) >= 4 {
				alpha = v[3]
			}
			return color.RGBA{R: v[0], G: v[1], B: v[2], A: alpha}, nil
		}
		return color.RGBA{}, errors.New("color slice must have at least 3 elements")
	default:
		return color.RGBA{}, nil
	}
}
