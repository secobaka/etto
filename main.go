package main

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"github.com/secobaka/etto/task"
)

var store *task.Store

func main() {
	var err error
	store, err = task.NewStore()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	cobra.AddTemplateFunc("join", strings.Join)

	rootCmd := &cobra.Command{
		Use:   "etto",
		Short: "etto - what's next?",
		Run: func(cmd *cobra.Command, args []string) {
			runList(cmd, args)
		},
	}

	rootCmd.SetUsageTemplate(`Usage:{{if .Runnable}}
  {{.UseLine}}{{end}}{{if .HasAvailableSubCommands}}
  {{.CommandPath}} [command]{{end}}{{if gt (len .Aliases) 0}}

Aliases:
  {{.NameAndAliases}}{{end}}{{if .HasAvailableSubCommands}}

Available Commands:{{range .Commands}}{{if (or .IsAvailableCommand (eq .Name "help"))}}
  {{if .Aliases}}{{rpad (printf "%s, %s" .Name (join .Aliases ", ")) .NamePadding}} {{.Short}}{{else}}{{rpad .Name .NamePadding}} {{.Short}}{{end}}{{end}}{{end}}{{end}}{{if .HasAvailableLocalFlags}}

Flags:
{{.LocalFlags.FlagUsages | trimTrailingWhitespaces}}{{end}}{{if .HasAvailableInheritedFlags}}

Global Flags:
{{.InheritedFlags.FlagUsages | trimTrailingWhitespaces}}{{end}}{{if .HasAvailableSubCommands}}

Use "{{.CommandPath}} [command] --help" for more information about a command.{{end}}
`)

	// add
	addCmd := &cobra.Command{
		Use:     "add <title>",
		Aliases: []string{"a"},
		Short:   "Add a new task",
		Args:    cobra.ExactArgs(1),
		Run:     runAdd,
	}
	addCmd.Flags().StringP("due", "d", "", "Due date (2006-01-02 15:04)")
	addCmd.Flags().StringP("priority", "p", "low", "Priority (high|medium|low)")
	rootCmd.AddCommand(addCmd)

	// list
	listCmd := &cobra.Command{
		Use:     "list",
		Aliases: []string{"l", "ls"},
		Short:   "List tasks",
		Run:     runList,
	}
	listCmd.Flags().StringP("sort", "s", "due", "Sort order (priority|due)")
	listCmd.Flags().Bool("all", false, "Show all tasks including done")
	rootCmd.AddCommand(listCmd)

	// done
	doneCmd := &cobra.Command{
		Use:     "done <id>",
		Aliases: []string{"d"},
		Short:   "Toggle done/undone",
		Args:    cobra.ExactArgs(1),
		Run:     runDone,
	}
	rootCmd.AddCommand(doneCmd)

	// edit
	editCmd := &cobra.Command{
		Use:     "edit <id>",
		Aliases: []string{"e"},
		Short:   "Edit a task",
		Args:    cobra.MinimumNArgs(1),
		Run:     runEdit,
	}
	editCmd.Flags().StringP("title", "t", "", "Update title")
	editCmd.Flags().StringP("due", "d", "", "Update due date (2006-01-02 15:04)")
	editCmd.Flags().StringP("priority", "p", "", "Update priority (high|medium|low)")
	rootCmd.AddCommand(editCmd)

	// remove
	removeCmd := &cobra.Command{
		Use:     "remove <id>",
		Aliases: []string{"r", "rm"},
		Short:   "Remove a task",
		Args:    cobra.ExactArgs(1),
		Run:     runRemove,
	}
	rootCmd.AddCommand(removeCmd)

	// yabai
	yabaiCmd := &cobra.Command{
		Use:     "yabai",
		Aliases: []string{"yb"},
		Short:   "Show overdue & urgent tasks",
		Run:     runYabai,
	}
	yabaiCmd.Flags().IntP("hours", "h", 24, "Hours threshold")
	rootCmd.AddCommand(yabaiCmd)

	// momuri
	// merge
	mergeCmd := &cobra.Command{
		Use:     "merge <id> <id>",
		Aliases: []string{"m", "mg"},
		Short:   "Merge two tasks into one",
		Long:    "Merge two tasks into one. Keeps the first task's title, higher priority, and earlier due date.",
		Args:    cobra.ExactArgs(2),
		Run:     runMerge,
	}
	mergeCmd.Flags().StringP("title", "t", "", "Override merged task title")
	rootCmd.AddCommand(mergeCmd)

	momuriCmd := &cobra.Command{
		Use:   "momuri",
		Short: "Remove all active tasks",
		Run:   runMomuri,
	}
	rootCmd.AddCommand(momuriCmd)

	rootCmd.Execute()
}

func runAdd(cmd *cobra.Command, args []string) {
	title := args[0]
	var due *time.Time

	dueStr, _ := cmd.Flags().GetString("due")
	if dueStr != "" {
		t, err := time.ParseInLocation("2006-01-02 15:04", dueStr, time.Local)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Invalid date format: %s (expected: 2006-01-02 15:04)\n", dueStr)
			os.Exit(1)
		}
		due = &t
	}

	priStr, _ := cmd.Flags().GetString("priority")
	priority := task.ParsePriority(priStr)

	t := store.Add(title, due, priority)
	if err := store.Save(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("Added: #%d %s\n", t.ID, t.Title)
}

func runList(cmd *cobra.Command, args []string) {
	sortStr := "due"
	showAll := false
	if cmd.Flags().Lookup("sort") != nil {
		sortStr, _ = cmd.Flags().GetString("sort")
	}
	if cmd.Flags().Lookup("all") != nil {
		showAll, _ = cmd.Flags().GetBool("all")
	}

	var sortKey task.SortKey
	switch sortStr {
	case "due", "d":
		sortKey = task.SortByDue
	case "priority", "p":
		sortKey = task.SortByPriority
	default:
		fmt.Fprintf(os.Stderr, "Unknown sort key: %s (use priority|due)\n", sortStr)
		os.Exit(1)
	}

	store.Sort(sortKey)

	if len(store.Tasks) == 0 {
		fmt.Println("No tasks.")
		return
	}

	for _, t := range store.Tasks {
		if !showAll && t.Done {
			continue
		}
		check := "[ ]"
		if t.Done {
			check = "[x]"
		}

		pri := strings.ToUpper(t.Priority.String()[:1])

		dueStr := ""
		if t.Due != nil {
			dueStr = t.Due.Format("01/02 15:04")
		}

		fmt.Printf("#%-3d %s %s  (%s)  %s\n", t.ID, check, t.Title, pri, dueStr)
	}
}

func runDone(cmd *cobra.Command, args []string) {
	id, err := strconv.Atoi(args[0])
	if err != nil {
		fmt.Fprintf(os.Stderr, "Invalid ID: %s\n", args[0])
		os.Exit(1)
	}

	if !store.ToggleDone(id) {
		fmt.Fprintf(os.Stderr, "Task #%d not found\n", id)
		os.Exit(1)
	}

	if err := store.Save(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	for _, t := range store.Tasks {
		if t.ID == id {
			status := "done"
			if !t.Done {
				status = "undone"
			}
			fmt.Printf("#%d %s -> %s\n", t.ID, t.Title, status)
			break
		}
	}
}

func runEdit(cmd *cobra.Command, args []string) {
	id, err := strconv.Atoi(args[0])
	if err != nil {
		fmt.Fprintf(os.Stderr, "Invalid ID: %s\n", args[0])
		os.Exit(1)
	}

	var found *task.Task
	for i := range store.Tasks {
		if store.Tasks[i].ID == id {
			found = &store.Tasks[i]
			break
		}
	}
	if found == nil {
		fmt.Fprintf(os.Stderr, "Task #%d not found\n", id)
		os.Exit(1)
	}

	title := found.Title
	due := found.Due
	priority := found.Priority

	if t, _ := cmd.Flags().GetString("title"); t != "" {
		title = t
	}
	if d, _ := cmd.Flags().GetString("due"); d != "" {
		t, err := time.ParseInLocation("2006-01-02 15:04", d, time.Local)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Invalid date format: %s (expected: 2006-01-02 15:04)\n", d)
			os.Exit(1)
		}
		due = &t
	}
	if p, _ := cmd.Flags().GetString("priority"); p != "" {
		priority = task.ParsePriority(p)
	}

	store.Update(id, title, due, priority)
	if err := store.Save(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("Updated: #%d %s\n", id, title)
}

func runRemove(cmd *cobra.Command, args []string) {
	id, err := strconv.Atoi(args[0])
	if err != nil {
		fmt.Fprintf(os.Stderr, "Invalid ID: %s\n", args[0])
		os.Exit(1)
	}

	var title string
	for _, t := range store.Tasks {
		if t.ID == id {
			title = t.Title
			break
		}
	}

	if !store.Delete(id) {
		fmt.Fprintf(os.Stderr, "Task #%d not found\n", id)
		os.Exit(1)
	}

	if err := store.Save(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Removed: #%d %s\n", id, title)
}

func runMerge(cmd *cobra.Command, args []string) {
	id1, err := strconv.Atoi(args[0])
	if err != nil {
		fmt.Fprintf(os.Stderr, "Invalid ID: %s\n", args[0])
		os.Exit(1)
	}
	id2, err := strconv.Atoi(args[1])
	if err != nil {
		fmt.Fprintf(os.Stderr, "Invalid ID: %s\n", args[1])
		os.Exit(1)
	}

	t1 := store.Find(id1)
	if t1 == nil {
		fmt.Fprintf(os.Stderr, "Task #%d not found\n", id1)
		os.Exit(1)
	}
	t2 := store.Find(id2)
	if t2 == nil {
		fmt.Fprintf(os.Stderr, "Task #%d not found\n", id2)
		os.Exit(1)
	}

	// タイトル: 先のIDを採用、--titleで上書き可
	title := t1.Title
	if t, _ := cmd.Flags().GetString("title"); t != "" {
		title = t
	}

	// 優先度: 高い方
	priority := t1.Priority
	if t2.Priority > priority {
		priority = t2.Priority
	}

	// 期限: 早い方
	due := t1.Due
	if t1.Due == nil {
		due = t2.Due
	} else if t2.Due != nil && t2.Due.Before(*t1.Due) {
		due = t2.Due
	}

	store.Update(id1, title, due, priority)
	store.Delete(id2)

	if err := store.Save(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Merged: #%d + #%d -> #%d %s\n", id1, id2, id1, title)
}

func runYabai(cmd *cobra.Command, args []string) {
	hours, _ := cmd.Flags().GetInt("hours")

	now := time.Now()
	deadline := now.Add(time.Duration(hours) * time.Hour)

	store.Sort(task.SortByDue)

	var overdue, upcoming []task.Task
	for _, t := range store.Tasks {
		if t.Done || t.Due == nil {
			continue
		}
		if t.Due.Before(now) {
			overdue = append(overdue, t)
		} else if t.Due.Before(deadline) {
			upcoming = append(upcoming, t)
		}
	}

	if len(overdue) == 0 && len(upcoming) == 0 {
		fmt.Println("Nothing yabai. You're safe!")
		return
	}

	total := len(overdue) + len(upcoming)
	fmt.Printf("!!! YABAI !!! %d task(s) need your attention !!!\n\n", total)

	if len(overdue) > 0 {
		fmt.Println("OVERDUE:")
		for _, t := range overdue {
			pri := strings.ToUpper(t.Priority.String()[:1])
			fmt.Printf("  #%-3d %s  (%s)  %s\n", t.ID, t.Title, pri, t.Due.Format("01/02 15:04"))
		}
	}

	if len(upcoming) > 0 {
		fmt.Printf("\nDue within %dh:\n", hours)
		for _, t := range upcoming {
			pri := strings.ToUpper(t.Priority.String()[:1])
			remaining := time.Until(*t.Due)
			h := int(remaining.Hours())
			m := int(remaining.Minutes()) % 60
			fmt.Printf("  #%-3d %s  (%s)  %s  (remaining: %dh%dm)\n", t.ID, t.Title, pri, t.Due.Format("01/02 15:04"), h, m)
		}
	}
}

func runMomuri(cmd *cobra.Command, args []string) {
	undone := 0
	for _, t := range store.Tasks {
		if !t.Done {
			undone++
		}
	}

	if undone == 0 {
		fmt.Println("No active tasks to remove.")
		return
	}

	fmt.Printf("%d active task(s) will be gone forever. (done tasks will be kept)\n", undone)
	fmt.Print("Really momuri? (y/N): ")
	reader := bufio.NewReader(os.Stdin)
	input, _ := reader.ReadString('\n')
	input = strings.TrimSpace(strings.ToLower(input))

	if input != "y" && input != "yes" {
		fmt.Println("Not momuri yet. Keep going.")
		return
	}

	var kept []task.Task
	for _, t := range store.Tasks {
		if t.Done {
			kept = append(kept, t)
		}
	}
	store.Tasks = kept
	if err := store.Save(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("Removed %d task(s). You are free now!\n", undone)
}
