package cmd

import (
	"ajgo/gff3"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"regexp"
)

// submode genemodel > ensembl-gff3 > panel
var genemodelEnsemblGff3PanelCmd = &cobra.Command{
	Use:   "panel",
	Short: "select features from gene model by gene",
	Long: `Read a GFF3 file and only keep features related to a list of
gene names or ids read from a file. Write out the derived GFF3 file.`,
	Run: func(cmd *cobra.Command, args []string) {
		startLogging()
		genemodelEnsemblGff3PanelCmdRun(cmd, args)
		finishLogging()
	},
}

func init() {
	genemodelEnsemblGff3Cmd.AddCommand(genemodelEnsemblGff3PanelCmd)

	genemodelEnsemblGff3PanelCmd.Flags().StringVar(&flagInfileGeneModel, "in-gff3", "",
		"gene model as basis of panel - GFF3 format")
	genemodelEnsemblGff3PanelCmd.MarkFlagRequired("in-gff3")
	genemodelEnsemblGff3PanelCmd.Flags().StringVar(&flagOutfileGeneModel, "out-gff3", "",
		"gene model for selected genes - GFF3 format")
	genemodelEnsemblGff3PanelCmd.MarkFlagRequired("out-gff3")

	genemodelEnsemblGff3PanelCmd.Flags().StringVar(&flagInfile, "genes", "",
		"plain text file of gene names or ENSG IDs, - one per line")
	genemodelEnsemblGff3PanelCmd.MarkFlagRequired("genes")
}

func genemodelEnsemblGff3PanelCmdRun(cmd *cobra.Command, args []string) {
	// Get genes and build lookups
	log.Info("reading genes from: ", flagInfile)
	rex := regexp.MustCompile(`^ENSG`)
	genes, err := LinesFromFile(flagInfile)
	if err != nil {
		log.Fatal(err)
	}
	geneNames := make(map[string]int)
	geneIds := make(map[string]int)
	for _, gene := range genes {
		if rex.Match([]byte(gene)) {
			// Make key look like an Ensembl ID - gene:ENSG....
			geneIds["gene:"+gene]++
			if geneIds[gene] > 1 {
				log.Fatalf("gene ID %s was specified more than once", gene)
			}
		} else {
			geneNames[gene]++
			if geneNames[gene] > 1 {
				log.Fatalf("gene name %s was specified more than once", gene)
			}
		}
	}
	log.Info("  Genes found: ", len(genes))

	// Read source GFF3
	log.Info("reading GFF3: ", flagInfileGeneModel)
	gIn, err := gff3.NewFromFile(flagInfileGeneModel)
	if err != nil {
		log.Fatal(err)
	}
	log.Info("  Number of Features: ", gIn.FeatureCount())

	// Create tree
	log.Info("creating Gff3Tree")
	t := gIn.NewTree()
	log.Info("  Number of Nodes: ", len(t.Nodes))
	log.Info("  Number of Orphans: ", len(t.Orphans))

	// For each node, try to match firstly by IdString (ENSG...) and
	// secondly by Name= Attribute.
	log.Info("filtering by gene")
	nfs := gff3.NewFeatures()
	found := make(map[string]int)
	for _, n := range t.Nodes {
		// Try to match IdString as an ENSG
		if _, ok := geneIds[n.IdString]; ok {
			found[n.IdString]++
			feats := n.Features()
			nfs.Features = append(nfs.Features, feats...)
		} else {
			// See if the first Self Feature has a Name= attribute and
			// if so, try to match on gene Name
			if _, ok := n.Self[0].Attributes[`Name`]; ok {
				if _, ok := geneNames[n.Self[0].Attributes[`Name`]]; ok {
					found[n.Self[0].Attributes[`Name`]]++
					feats := n.Features()
					nfs.Features = append(nfs.Features, feats...)
				}
			}
		}
	}

	gOut := gff3.NewGff3()
	gOut.Features = nfs
	log.Info("  Number of Features: ", gOut.FeatureCount())

	// Check that all of the genes requested were found
	for _, gene := range genes {
		if _, ok := found[gene]; !ok {
			log.Warnf("gene %s was not found", gene)
		}
	}

	// Write out the new post-selection Gff3
	err = gOut.Write(flagOutfileGeneModel)
	if err != nil {
		log.Fatal(err)
	}
	log.Infof("writing complete: %s", flagOutfileGeneModel)
}
