package flavor

import (
	"context"
	"fmt"
	"log"
	"os"
	"strings"
	"text/tabwriter"

	vars "github.com/etesami/skycluster-cli/internal"
	utils "github.com/etesami/skycluster-cli/internal/utils"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var pNames []string

func init() {
	flavorCmd.AddCommand(flavorListCmd)
	flavorListCmd.PersistentFlags().StringSliceVar(&pNames, "provider-name", nil, "Provider Names, seperated by comma")
}

var flavorCmd = &cobra.Command{
	Use:   "flavor commands",
	Short: "Flavor commands",
	Run: func(cmd *cobra.Command, args []string) {
		cmd.Help()
	},
}

var flavorListCmd = &cobra.Command{
	Use:   "list",
	Short: "List avaialble flavors across providers",
	Run: func(cmd *cobra.Command, args []string) {
		listFlavors()
	},
}

func listFlavors() {
	kconfig := viper.GetStringMapString("kubeconfig")
	kubeconfig := kconfig["sky-manager"]
	clientset, err := utils.GetClientset(kubeconfig)
	if err != nil {
		log.Fatalf("Error getting clientset: %v", err)
		return
	}
	flavorList := make(map[string][]string, 0)
	baseFilters := "skycluster.io/managed-by=skycluster, skycluster.io/config-type=provider-mappings, "
	for _, n := range pNames {
		filters := baseFilters + "skycluster.io/provider-name=" + n
		confgis, err := clientset.CoreV1().ConfigMaps(vars.SkyClusterName).List(context.Background(), metav1.ListOptions{
			LabelSelector: filters,
		})
		if err != nil {
			log.Fatalf("Error listing configmaps: %v", err)
		}

		for _, cm := range confgis.Items {
			fList := make([]string, 0)
			pName := cm.Labels["skycluster.io/provider-name"]
			pRegion := cm.Labels["skycluster.io/provider-region"]
			pZone := cm.Labels["skycluster.io/provider-zone"]
			pID := pName + "_" + pRegion + "_" + pZone
			for d, _ := range cm.Data {
				if strings.Contains(d, "flavor") {
					fList = append(fList, d)
				}
			}
			if len(fList) > 0 {
				flavorList[pID] = fList
			}
		}
	}

	availableFlavors := utils.IntersectionOfMapValues(flavorList, utils.GetMapStringKeys(flavorList))
	writer := tabwriter.NewWriter(os.Stdout, 0, 8, 1, '\t', tabwriter.AlignRight)
	if len(availableFlavors) == 0 {
		fmt.Println("No flavors available")
	} else {
		fmt.Fprintln(writer, "NUM\tNAME\tOFFERED BY")
	}
	for _, f := range availableFlavors {
		fmt.Fprintf(writer, "%s\t%d\n", f, len(flavorList))
	}
	writer.Flush()

}

func GetFlavorCmd() *cobra.Command {
	return flavorCmd
}
