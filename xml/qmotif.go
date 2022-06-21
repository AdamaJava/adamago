// To see what a typical qmotif XML file looks like, see the examples
// subdirectory.

package xml

import (
	"encoding/xml"
)

type Qmotif struct {
	XMLName xml.Name `xml:"qmotif"`
	Version string   `xml:"version,attr"`
	Ini     Ini      `xml:"ini"`
}

type Ini struct {
	XMLName  xml.Name    `xml:"ini"`
	File     string      `xml:"file,attr"`
	S1motif  Stage_motif `xml:"stage1_motif"`
	S2motif  Stage_motif `xml:"stage2_motif"`
	Includes []Region    `xml:"includes"`
}

type Stage_motif struct {
	XMLName xml.Name
	Strings []String `xml:"string"`
	Regexs  []Regex  `xml:"regex"`
}

type String struct {
	XMLName xml.Name `xml:"string"`
	Value   string   `xml:"value,attr"`
}

type Regex struct {
	XMLName xml.Name `xml:"regex"`
	Value   string   `xml:"value,attr"`
}

type Window_size struct {
	XMLName xml.Name `xml:"window_size"`
	Value   string   `xml:"value,attr"`
}

type Includes_only struct {
	XMLName xml.Name `xml:"includes_only"`
	Value   string   `xml:"value,attr"`
}

type Region struct {
	XMLName xml.Name `xml:"region"`
	ChrPos  string   `xml:"chrPos,attr"`
	Name    string   `xml:"name,attr"`
}
