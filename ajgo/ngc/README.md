# ngc

This package provides data structures for working with NGScheck
output files, primarily in JSON format.

As a rule we will not provide getter or setter functions - you can
directly access the data structures including inner structures.

```
package main

import (
    "ajgo/ngc"
    "encoding/json"
    "io"
    "log"
    "os"
)

func main() {
    j, err := os.Open(file)
    if err != nil {
        log.Fatal(err)
    }
    byteValue, _ := io.ReadAll(j)

    var ngb ngc.NgscheckBasic
    err = json.Unmarshal(byteValue, &ngb)
    if err != nil {
        log.Fatal(err)
    }

    (&ngb).Finalise()

    log.Info("NGScheck UUID:    ",ngb.Ngscheck.ReportUuid)
    log.Info("NGScheck version: ",ngb.Ngscheck.Version)
    log.Info("BAM file:         ",ngb.Qprofiler2.Bam)
    log.Info("Predicted Genome: ",ngb.Scorecard.PredictedGenome)
    log.Info("Avg Read Depth:   ",ngb.Scorecard.AverageReadDepth)
}
```

