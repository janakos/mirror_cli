package cmd

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/alexcode/mirror_cli/internal/client"
	pb "github.com/alexcode/mirror_cli/proto/gen"
)

// mirrorCmd represents the mirror command
var mirrorCmd = &cobra.Command{
	Use:   "mirror",
	Short: "Manage PeerDB mirrors",
	Long:  "Commands for creating, listing, monitoring, and managing PeerDB mirrors.",
}

// mirrorCreateCmd represents the mirror create command
var mirrorCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a new mirror",
	Long:  "Create a new CDC mirror between source and destination peers.",
	RunE: func(cmd *cobra.Command, args []string) error {
		return createMirror(cmd)
	},
}

// mirrorListCmd represents the mirror list command
var mirrorListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all mirrors",
	Long:  "List all configured mirrors with their status.",
	RunE: func(cmd *cobra.Command, args []string) error {
		return listMirrors(cmd)
	},
}

// mirrorStatusCmd represents the mirror status command
var mirrorStatusCmd = &cobra.Command{
	Use:   "status [mirror-name]",
	Short: "Get mirror status",
	Long:  "Get detailed status information for a specific mirror.",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		return getMirrorStatus(cmd, args[0])
	},
}

// mirrorPauseCmd represents the mirror pause command
var mirrorPauseCmd = &cobra.Command{
	Use:   "pause [mirror-name]",
	Short: "Pause a mirror",
	Long:  "Pause a running mirror to temporarily stop replication.",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		return pauseMirror(cmd, args[0])
	},
}

// mirrorResumeCmd represents the mirror resume command
var mirrorResumeCmd = &cobra.Command{
	Use:   "resume [mirror-name]",
	Short: "Resume a mirror",
	Long:  "Resume a paused mirror to restart replication.",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		return resumeMirror(cmd, args[0])
	},
}

// mirrorDropCmd represents the mirror drop command
var mirrorDropCmd = &cobra.Command{
	Use:   "drop [mirror-name]",
	Short: "Drop a mirror",
	Long:  "Terminate and drop a mirror permanently.",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		return dropMirror(cmd, args[0])
	},
}

// mirrorEditCmd represents the mirror edit command
var mirrorEditCmd = &cobra.Command{
	Use:   "edit [mirror-name]",
	Short: "Edit mirror configuration",
	Long:  "Update configuration for an existing mirror.",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		return editMirror(cmd, args[0])
	},
}

func init() {
	rootCmd.AddCommand(mirrorCmd)
	mirrorCmd.AddCommand(mirrorCreateCmd)
	mirrorCmd.AddCommand(mirrorListCmd)
	mirrorCmd.AddCommand(mirrorStatusCmd)
	mirrorCmd.AddCommand(mirrorPauseCmd)
	mirrorCmd.AddCommand(mirrorResumeCmd)
	mirrorCmd.AddCommand(mirrorDropCmd)
	mirrorCmd.AddCommand(mirrorEditCmd)

	// Create command flags
	mirrorCreateCmd.Flags().String("name", "", "Mirror name (required)")
	mirrorCreateCmd.Flags().String("source", "", "Source peer name (required)")
	mirrorCreateCmd.Flags().String("destination", "", "Destination peer name (required)")
	mirrorCreateCmd.Flags().StringSlice("tables", []string{}, "Table mappings in format 'source_table->dest_table'")
	mirrorCreateCmd.Flags().Uint32("batch-size", 1000, "Maximum batch size")
	mirrorCreateCmd.Flags().Uint64("idle-timeout", 60, "Idle timeout in seconds")
	mirrorCreateCmd.Flags().Bool("initial-snapshot", true, "Perform initial snapshot")
	mirrorCreateCmd.Flags().String("publication", "", "PostgreSQL publication name")
	mirrorCreateCmd.Flags().String("replication-slot", "", "PostgreSQL replication slot name")

	mirrorCreateCmd.MarkFlagRequired("name")
	mirrorCreateCmd.MarkFlagRequired("source")
	mirrorCreateCmd.MarkFlagRequired("destination")
	mirrorCreateCmd.MarkFlagRequired("tables")

	// Drop command flags
	mirrorDropCmd.Flags().Bool("skip-destination-drop", false, "Skip dropping tables in destination")
	mirrorDropCmd.Flags().Bool("force", false, "Force drop without confirmation")

	// Edit command flags
	mirrorEditCmd.Flags().StringSlice("add-tables", []string{}, "Add table mappings")
	mirrorEditCmd.Flags().StringSlice("remove-tables", []string{}, "Remove table mappings")
	mirrorEditCmd.Flags().Uint32("batch-size", 0, "Update batch size")
	mirrorEditCmd.Flags().Uint64("idle-timeout", 0, "Update idle timeout")
}

func createMirror(cmd *cobra.Command) error {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Get flags
	name, _ := cmd.Flags().GetString("name")
	source, _ := cmd.Flags().GetString("source")
	destination, _ := cmd.Flags().GetString("destination")
	tables, _ := cmd.Flags().GetStringSlice("tables")
	batchSize, _ := cmd.Flags().GetUint32("batch-size")
	idleTimeout, _ := cmd.Flags().GetUint64("idle-timeout")
	initialSnapshot, _ := cmd.Flags().GetBool("initial-snapshot")
	publication, _ := cmd.Flags().GetString("publication")
	replicationSlot, _ := cmd.Flags().GetString("replication-slot")

	// Parse table mappings
	tableMappings := make([]*pb.TableMapping, 0, len(tables))
	for _, table := range tables {
		parts := strings.Split(table, "->")
		if len(parts) != 2 {
			return fmt.Errorf("invalid table mapping format: %s (expected: source->destination)", table)
		}
		tableMappings = append(tableMappings, &pb.TableMapping{
			SourceTableIdentifier:      strings.TrimSpace(parts[0]),
			DestinationTableIdentifier: strings.TrimSpace(parts[1]),
		})
	}

	// Create client
	client, err := client.NewClient(GetConfig())
	if err != nil {
		return err
	}
	defer client.Close()

	// Create mirror request
	req := &pb.CreateCDCFlowRequest{
		ConnectionConfigs: &pb.FlowConnectionConfigs{
			FlowJobName:         name,
			SourceName:          source,
			DestinationName:     destination,
			TableMappings:       tableMappings,
			MaxBatchSize:        batchSize,
			IdleTimeoutSeconds:  idleTimeout,
			DoInitialSnapshot:   initialSnapshot,
			PublicationName:     publication,
			ReplicationSlotName: replicationSlot,
		},
	}

	// Create the mirror
	resp, err := client.CreateCDCMirror(ctx, req)
	if err != nil {
		return fmt.Errorf("failed to create mirror: %w", err)
	}

	fmt.Printf("✓ Mirror '%s' created successfully\n", name)
	fmt.Printf("  Workflow ID: %s\n", resp.WorkflowId)
	fmt.Printf("  Source: %s\n", source)
	fmt.Printf("  Destination: %s\n", destination)
	fmt.Printf("  Tables: %d\n", len(tableMappings))

	return nil
}

func listMirrors(cmd *cobra.Command) error {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Create client
	client, err := client.NewClient(GetConfig())
	if err != nil {
		return err
	}
	defer client.Close()

	// List mirrors
	resp, err := client.ListMirrors(ctx)
	if err != nil {
		return fmt.Errorf("failed to list mirrors: %w", err)
	}

	if len(resp.Mirrors) == 0 {
		fmt.Println("No mirrors found")
		return nil
	}

	// Print header
	fmt.Printf("%-20s %-15s %-15s %-10s %-12s\n", "NAME", "SOURCE", "DESTINATION", "TYPE", "CREATED")
	fmt.Println(strings.Repeat("-", 80))

	// Print mirrors
	for _, mirror := range resp.Mirrors {
		mirrorType := "QRep"
		if mirror.IsCdc {
			mirrorType = "CDC"
		}

		createdAt := time.Unix(int64(mirror.CreatedAt), 0).Format("2006-01-02")

		fmt.Printf("%-20s %-15s %-15s %-10s %-12s\n",
			mirror.Name,
			mirror.SourceName,
			mirror.DestinationName,
			mirrorType,
			createdAt,
		)
	}

	return nil
}

func getMirrorStatus(cmd *cobra.Command, mirrorName string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Create client
	client, err := client.NewClient(GetConfig())
	if err != nil {
		return err
	}
	defer client.Close()

	// Get mirror status
	resp, err := client.GetMirrorStatus(ctx, mirrorName)
	if err != nil {
		return fmt.Errorf("failed to get mirror status: %w", err)
	}

	// Print status
	fmt.Printf("Mirror: %s\n", resp.FlowJobName)
	fmt.Printf("Status: %s\n", resp.CurrentFlowState.String())

	if resp.CreatedAt != nil {
		fmt.Printf("Created: %s\n", resp.CreatedAt.AsTime().Format(time.RFC3339))
	}

	if resp.CdcStatus != nil {
		fmt.Printf("Rows Synced: %d\n", resp.CdcStatus.RowsSynced)
		fmt.Printf("Source Type: %s\n", resp.CdcStatus.SourceType.String())
		fmt.Printf("Destination Type: %s\n", resp.CdcStatus.DestinationType.String())

		if resp.CdcStatus.SnapshotStatus != nil {
			fmt.Printf("Snapshot Tables: %d\n", len(resp.CdcStatus.SnapshotStatus.Clones))
		}

		fmt.Printf("CDC Batches: %d\n", len(resp.CdcStatus.CdcBatches))
	}

	return nil
}

func pauseMirror(cmd *cobra.Command, mirrorName string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	client, err := client.NewClient(GetConfig())
	if err != nil {
		return err
	}
	defer client.Close()

	if err := client.PauseMirror(ctx, mirrorName); err != nil {
		return fmt.Errorf("failed to pause mirror: %w", err)
	}

	fmt.Printf("✓ Mirror '%s' paused successfully\n", mirrorName)
	return nil
}

func resumeMirror(cmd *cobra.Command, mirrorName string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	client, err := client.NewClient(GetConfig())
	if err != nil {
		return err
	}
	defer client.Close()

	if err := client.ResumeMirror(ctx, mirrorName); err != nil {
		return fmt.Errorf("failed to resume mirror: %w", err)
	}

	fmt.Printf("✓ Mirror '%s' resumed successfully\n", mirrorName)
	return nil
}

func dropMirror(cmd *cobra.Command, mirrorName string) error {
	skipDestinationDrop, _ := cmd.Flags().GetBool("skip-destination-drop")
	force, _ := cmd.Flags().GetBool("force")

	// Confirmation unless forced
	if !force {
		fmt.Printf("Are you sure you want to drop mirror '%s'? This action cannot be undone. (y/N): ", mirrorName)
		var response string
		fmt.Scanln(&response)
		if strings.ToLower(response) != "y" && strings.ToLower(response) != "yes" {
			fmt.Println("Operation cancelled")
			return nil
		}
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	client, err := client.NewClient(GetConfig())
	if err != nil {
		return err
	}
	defer client.Close()

	if err := client.DropMirror(ctx, mirrorName, skipDestinationDrop); err != nil {
		return fmt.Errorf("failed to drop mirror: %w", err)
	}

	fmt.Printf("✓ Mirror '%s' dropped successfully\n", mirrorName)
	return nil
}

func editMirror(cmd *cobra.Command, mirrorName string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	addTables, _ := cmd.Flags().GetStringSlice("add-tables")
	removeTables, _ := cmd.Flags().GetStringSlice("remove-tables")
	batchSize, _ := cmd.Flags().GetUint32("batch-size")
	idleTimeout, _ := cmd.Flags().GetUint64("idle-timeout")

	// Parse additional tables
	additionalTables := make([]*pb.TableMapping, 0, len(addTables))
	for _, table := range addTables {
		parts := strings.Split(table, "->")
		if len(parts) != 2 {
			return fmt.Errorf("invalid table mapping format: %s", table)
		}
		additionalTables = append(additionalTables, &pb.TableMapping{
			SourceTableIdentifier:      strings.TrimSpace(parts[0]),
			DestinationTableIdentifier: strings.TrimSpace(parts[1]),
		})
	}

	// Parse tables to remove
	removedTables := make([]*pb.TableMapping, 0, len(removeTables))
	for _, table := range removeTables {
		parts := strings.Split(table, "->")
		if len(parts) != 2 {
			return fmt.Errorf("invalid table mapping format: %s", table)
		}
		removedTables = append(removedTables, &pb.TableMapping{
			SourceTableIdentifier:      strings.TrimSpace(parts[0]),
			DestinationTableIdentifier: strings.TrimSpace(parts[1]),
		})
	}

	// Build update request
	cdcUpdate := &pb.CDCFlowConfigUpdate{
		AdditionalTables: additionalTables,
		RemovedTables:    removedTables,
	}

	if batchSize > 0 {
		cdcUpdate.BatchSize = batchSize
	}
	if idleTimeout > 0 {
		cdcUpdate.IdleTimeout = idleTimeout
	}

	update := &pb.FlowConfigUpdate{
		CdcFlowConfigUpdate: cdcUpdate,
	}

	client, err := client.NewClient(GetConfig())
	if err != nil {
		return err
	}
	defer client.Close()

	if err := client.UpdateMirror(ctx, mirrorName, update); err != nil {
		return fmt.Errorf("failed to update mirror: %w", err)
	}

	fmt.Printf("✓ Mirror '%s' updated successfully\n", mirrorName)
	return nil
}
