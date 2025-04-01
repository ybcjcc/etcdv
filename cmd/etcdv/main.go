package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/spf13/cobra"
	"etcdv/pkg/etcdhistory"
	"etcdv/pkg/types"
)

var (
	endpoints []string
	key       string
	format    string
	limit     int
	order     string
	username  string
	password  string
)

var rootCmd = &cobra.Command{
	Use:   "etcdv",
	Short: "Get version history of a key in etcd",
	Run:   run,
}

func init() {
	rootCmd.Flags().StringSliceVarP(&endpoints, "endpoints", "e", []string{"localhost:2379"}, "etcd endpoints")
	rootCmd.Flags().StringVarP(&key, "key", "k", "", "key to get history for")
	rootCmd.Flags().StringVarP(&format, "format", "f", "table", "output format (json or table)")
	rootCmd.Flags().IntVarP(&limit, "limit", "l", 0, "limit the number of records returned (0 means no limit)")
	rootCmd.Flags().StringVarP(&order, "order", "o", "desc", "sort order (asc or desc)")
	rootCmd.Flags().StringVarP(&username, "username", "u", "", "etcd username for authentication")
	rootCmd.Flags().StringVarP(&password, "password", "p", "", "etcd password for authentication")
	rootCmd.MarkFlagRequired("key")
}

func run(cmd *cobra.Command, args []string) {
	opts := types.Options{
		Endpoints: endpoints,
		Key:       key,
		Limit:     limit,
		Order:     order,
		Username:  username,
		Password:  password,
	}

	records, err := etcdhistory.GetVersionHistory(context.Background(), opts)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	// 输出结果
	switch format {
	case "json":
		jsonData, err := json.MarshalIndent(records, "", "  ")
		if err != nil {
			fmt.Printf("Failed to marshal to JSON: %v\n", err)
			os.Exit(1)
		}
		fmt.Println(string(jsonData))

	case "table":
		w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
		fmt.Fprintln(w, "Revision\tVersion\tCreateRevision\tModRevision\tValue")
		for _, r := range records {
			fmt.Fprintf(w, "%d\t%d\t%d\t%d\t%s\n", r.Revision, r.Version, r.CreateRevision, r.ModRevision, r.Value)
		}
		w.Flush()

	default:
		fmt.Printf("Unsupported format: %s\n", format)
		os.Exit(1)
	}
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}