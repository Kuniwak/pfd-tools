package allcheckers

type Format string

const (
	FormatTSV  Format = "pfdtsv"
	FormatJSON Format = "json"
)

func (f Format) String() string {
	return string(f)
}
