// A typical qmotif XML file looks like this:
//
// <?xml version="1.0" encoding="UTF-8" standalone="no"?>
// <qmotif version="1.2 (a8ab31c1)">
//   <ini file="/working/genomeinfo/cromwell/cromwell-executions/qmotifWf/71654a06-7c32-48a5-898e-5135f3f07e12/call-runQmotif/shard-283/inputs/1752157915/qmotif.ini">
//     <stage1_motif>
//       <string value="TTAGGGTTAGGGTTAGGG"/>
//       <string value="CCCTAACCCTAACCCTAA"/>
//     </stage1_motif>
//     <stage2_motif>
//       <regex value="(...GGG){2,}|(CCC...){2,}"/>
//     </stage2_motif>
//     <window_size value="10000"/>
//     <includes_only value="true"/>
//     <includes>
//       <region chrPos="chr1:10001-12464" name="chr1p"/>
//       ...
//       <region chrPos="chrY:59360739-59363565" name="chrYq"/>
//     </includes>
//   </ini>
//   <summary bam="/working/genomeinfo/cromwell/cromwell-executions/qmotifWf/71654a06-7c32-48a5-898e-5135f3f07e12/call-runQmotif/shard-283/inputs/347273389/fb7fab57-5d85-4096-8d88-786499717fd1.bam">
//     <counts>
//       <totalReadsInThisAnalysis count="148191"/>
//       <noOfMotifs count="60620"/>
//       <rawUnmapped count="0"/>
//       <rawIncludes count="70811"/>
//       <rawGenomic count="0"/>
//       <scaledUnmapped count="-1"/>
//       <scaledIncludes count="-1"/>
//       <scaledGenomic count="-1"/>
//       <bases_containing_motifs count="9000618"/>
//     </counts>
//   </summary>
//   <motifs>
//     <motif id="1" motif="AAAGGGAAAGGG" noOfHits="1"/>
//     <motif id="2" motif="AAAGGGAGAGGG" noOfHits="1"/>
//     ...
//     <motif id="60619" motif="TTTGGGTTTGGGTTTGGGTTTGGGTTTGGGTTTGGGTGGGGGTGAGGGTGAGGGTGAGGGTGAGGGTTAGGGTGAGGGTTAGGGTTAGGGTTAGGGTTAGGG" noOfHits="1"/>
//     <motif id="60620" motif="TTTGGGTTTGGGTTTGGGTTTGGGTTTGGGTTTGGGTTAGGG" noOfHits="2"/>
//   </motifs>
//   <regions>
//     <region chrPos="chr1:10001-12464" name="chr1p" stage1Cov="3219" stage2Cov="3219" type="includes">
//       <motif motifRef="20269" number="1" strand="F"/>
//       <motif motifRef="12440" number="1" strand="F"/>
//       ...
//     </region>
//     <region chrPos="chr1:249237907-249240620" name="chr1q" stage1Cov="5454" stage2Cov="5454" type="includes">
//       <motif motifRef="33673" number="1" strand="F"/>
//       <motif motifRef="47922" number="1" strand="F"/>
//       ...
//     </region>
//     ...
//   </regions>
// </qmotif>

package xml

import (
    "encoding/xml"
)

type Qmotif struct {
    XMLName xml.Name     `xml:"qmotif"`
    Version string       `xml:"version,attr"`
    Ini     Ini          `xml:"ini"`
    Summary Summary      `xml:"summary"`
}

// <ini/>

type Ini struct {
    XMLName  xml.Name      `xml:"ini"`
    File     string        `xml:"file,attr"`
    S1motif  Stage1_motif  `xml:"stage1_motif"`
    S2motif  Stage2_motif  `xml:"stage2_motif"`
    WinSize  Window_size   `xml:"window_size"`
    InclOnly Includes_only `xml:"includes_only"`
    Includes Includes      `xml:"includes"`
}

type Stage1_motif struct {
    XMLName xml.Name `xml:"stage1_motif"`
    Strings []String `xml:"string"`
    Regexs  []Regex  `xml:"regex"`
}

type Stage2_motif struct {
    XMLName xml.Name `xml:"stage2_motif"`
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

type Includes struct {
    XMLName xml.Name `xml:"includes"`
    Regions []Region `xml:"region"`
}

type Region struct {
    XMLName xml.Name `xml:"region"`
    ChrPos  string   `xml:"chrPos,attr"`
    Name    string   `xml:"name,attr"`
}

// <summary/>

type Summary struct {
    XMLName  xml.Name     `xml:"summary"`
    Bam      string       `xml:"bam,attr"`
    Counts   Counts       `xml:"counts"`
}

type Counts struct {
    XMLName        xml.Name       `xml:"counts"`
    TotalReads     TotalReads     `xml:"totalReadsInThisAnalysis"`
    NoOfMotifs     NoOfMotifs     `xml:"noOfMotifs"`
    RawUnmapped    RawUnmapped    `xml:"rawUnmapped"`
    RawIncludes    RawIncludes    `xml:"rawIncludes"`
    RawGenomic     RawGenomic     `xml:"rawGenomic"`
    ScaledUnmapped ScaledUnmapped `xml:"scaledUnmapped"`
    ScaledIncludes ScaledIncludes `xml:"scaledIncludes"`
    ScaledGenomic  ScaledGenomic  `xml:"scaledGenomic"`
    BasesInMotifs  BasesInMotifs  `xml:"bases_containing_motifs"`
}

type TotalReads struct {
    XMLName xml.Name `xml:"totalReadsInThisAnalysis"`
    Count   string   `xml:"count,attr"`
}

type NoOfMotifs struct {
    XMLName xml.Name `xml:"noOfMotifs"`
    Count   string   `xml:"count,attr"`
}

type RawUnmapped struct {
    XMLName xml.Name `xml:"rawUnmapped"`
    Count   string   `xml:"count,attr"`
}

type RawIncludes struct {
    XMLName xml.Name `xml:"rawIncludes"`
    Count   string   `xml:"count,attr"`
}

type RawGenomic struct {
    XMLName xml.Name `xml:"rawGenomic"`
    Count   string   `xml:"count,attr"`
}

type ScaledUnmapped struct {
    XMLName xml.Name `xml:"scaledUnmapped"`
    Count   string   `xml:"count,attr"`
}

type ScaledIncludes struct {
    XMLName xml.Name `xml:"scaledIncludes"`
    Count   string   `xml:"count,attr"`
}

type ScaledGenomic struct {
    XMLName xml.Name `xml:"scaledGenomic"`
    Count   string   `xml:"count,attr"`
}

type BasesInMotifs struct {
    XMLName xml.Name `xml:"bases_containing_motifs"`
    Count   string   `xml:"count,attr"`
}
