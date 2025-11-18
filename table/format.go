package table

type Format string

const (
	FormatTSV  Format = "tsv"
	FormatHTML Format = "html"
)

func (f Format) String() string {
	return string(f)
}
