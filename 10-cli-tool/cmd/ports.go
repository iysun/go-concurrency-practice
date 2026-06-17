package cmd

import (
	"fmt"
	"net"
	"sync"
	"time"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var portsCmd = &cobra.Command{
	Use:   "ports <host>",
	Short: "Scan open TCP ports on a host",
	Long: `Scan the given host for open TCP ports in the specified range.
Uses a goroutine pool for concurrent scanning.

Example:
  devtool ports localhost --start 1 --end 1024 --workers 100`,
	Args: cobra.ExactArgs(1),
	RunE: runPortScan,
}

func init() {
	rootCmd.AddCommand(portsCmd)

	portsCmd.Flags().Int("start", 1, "first port to scan")
	portsCmd.Flags().Int("end", 1024, "last port to scan")
	portsCmd.Flags().Int("workers", 100, "number of concurrent workers")
	portsCmd.Flags().Duration("timeout", 500*time.Millisecond, "connection timeout per port")

	viper.BindPFlag("ports.timeout", portsCmd.Flags().Lookup("timeout"))
}

// scanResult holds the outcome for one port.
type scanResult struct {
	port int
	open bool
}

// runPortScan is the command handler.
// TODO: implement the goroutine pool pattern:
//  1. Fan out port numbers into a jobs channel
//  2. N workers each dial the port and send results
//  3. Collect and print open ports
func runPortScan(cmd *cobra.Command, args []string) error {
	host := args[0]
	start, _ := cmd.Flags().GetInt("start")
	end, _ := cmd.Flags().GetInt("end")
	workers, _ := cmd.Flags().GetInt("workers")
	timeout, _ := cmd.Flags().GetDuration("timeout")

	verbose := viper.GetBool("verbose")
	if verbose {
		fmt.Printf("Scanning %s ports %d-%d with %d workers (timeout %v)\n",
			host, start, end, workers, timeout)
	}

	jobs := make(chan int, workers)
	results := make(chan scanResult, workers)

	// Launch worker pool
	var wg sync.WaitGroup
	for i := 0; i < workers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for port := range jobs {
				addr := fmt.Sprintf("%s:%d", host, port)
				conn, err := net.DialTimeout("tcp", addr, timeout)
				open := err == nil
				if open {
					conn.Close()
				}
				results <- scanResult{port: port, open: open}
			}
		}()
	}

	// Feed ports
	go func() {
		for p := start; p <= end; p++ {
			jobs <- p
		}
		close(jobs)
	}()

	// Close results when all workers finish
	go func() {
		wg.Wait()
		close(results)
	}()

	// Collect open ports
	var open []int
	for r := range results {
		if r.open {
			open = append(open, r.port)
			// TODO: sort and print in order — currently prints as they arrive
			fmt.Printf("  port %d open\n", r.port)
		}
	}

	fmt.Printf("\nFound %d open ports on %s\n", len(open), host)
	return nil
}
