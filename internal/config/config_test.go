package config

import (
	"os"
	"path/filepath"
	"sync"
	"testing"
	"time"

	"github.com/kennyg/tome/internal/artifact"
)

func TestLoadState_NewFile(t *testing.T) {
	tmpDir := t.TempDir()
	statePath := filepath.Join(tmpDir, "state.json")

	state, err := LoadState(statePath)
	if err != nil {
		t.Fatalf("LoadState() error = %v", err)
	}

	if state.Version != "1" {
		t.Errorf("Version = %v, want %v", state.Version, "1")
	}
	if len(state.Installed) != 0 {
		t.Errorf("Installed = %v, want empty", state.Installed)
	}
}

func TestSaveAndLoadState(t *testing.T) {
	tmpDir := t.TempDir()
	statePath := filepath.Join(tmpDir, "state.json")

	// Create state with installed artifacts
	state := &State{
		Version: "1",
		Installed: []artifact.InstalledArtifact{
			{
				Artifact: artifact.Artifact{
					Name:        "test-skill",
					Type:        artifact.TypeSkill,
					Description: "A test skill",
					Source:      "kennyg/test",
					InstalledAt: time.Now(),
				},
				LocalPath: "/home/user/.claude/skills/test-skill/SKILL.md",
			},
		},
	}

	// Save
	if err := SaveState(statePath, state); err != nil {
		t.Fatalf("SaveState() error = %v", err)
	}

	// Verify file exists
	if _, err := os.Stat(statePath); os.IsNotExist(err) {
		t.Fatal("state file was not created")
	}

	// Load
	loaded, err := LoadState(statePath)
	if err != nil {
		t.Fatalf("LoadState() error = %v", err)
	}

	if loaded.Version != state.Version {
		t.Errorf("Version = %v, want %v", loaded.Version, state.Version)
	}
	if len(loaded.Installed) != 1 {
		t.Fatalf("Installed len = %d, want 1", len(loaded.Installed))
	}
	if loaded.Installed[0].Name != "test-skill" {
		t.Errorf("Name = %v, want test-skill", loaded.Installed[0].Name)
	}
}

func TestState_AddInstalled(t *testing.T) {
	state := &State{Version: "1"}

	// Add first artifact
	art1 := artifact.InstalledArtifact{
		Artifact: artifact.Artifact{
			Name: "skill-1",
			Type: artifact.TypeSkill,
		},
	}
	state.AddInstalled(art1)

	if len(state.Installed) != 1 {
		t.Errorf("Installed len = %d, want 1", len(state.Installed))
	}

	// Add second artifact
	art2 := artifact.InstalledArtifact{
		Artifact: artifact.Artifact{
			Name: "skill-2",
			Type: artifact.TypeSkill,
		},
	}
	state.AddInstalled(art2)

	if len(state.Installed) != 2 {
		t.Errorf("Installed len = %d, want 2", len(state.Installed))
	}

	// Replace first artifact (same name and type)
	art1Updated := artifact.InstalledArtifact{
		Artifact: artifact.Artifact{
			Name:        "skill-1",
			Type:        artifact.TypeSkill,
			Description: "Updated",
		},
	}
	state.AddInstalled(art1Updated)

	if len(state.Installed) != 2 {
		t.Errorf("Installed len = %d, want 2 (should replace, not add)", len(state.Installed))
	}
}

func TestState_RemoveInstalled(t *testing.T) {
	state := &State{
		Version: "1",
		Installed: []artifact.InstalledArtifact{
			{Artifact: artifact.Artifact{Name: "skill-1", Type: artifact.TypeSkill}},
			{Artifact: artifact.Artifact{Name: "skill-2", Type: artifact.TypeSkill}},
			{Artifact: artifact.Artifact{Name: "cmd-1", Type: artifact.TypeCommand}},
		},
	}

	state.RemoveInstalled("skill-1", artifact.TypeSkill)

	if len(state.Installed) != 2 {
		t.Errorf("Installed len = %d, want 2", len(state.Installed))
	}

	// Verify skill-1 is gone
	for _, a := range state.Installed {
		if a.Name == "skill-1" && a.Type == artifact.TypeSkill {
			t.Error("skill-1 should have been removed")
		}
	}

	// Remove non-existent (should be no-op)
	state.RemoveInstalled("nonexistent", artifact.TypeSkill)
	if len(state.Installed) != 2 {
		t.Errorf("Installed len = %d, want 2", len(state.Installed))
	}
}

func TestState_FindInstalled(t *testing.T) {
	state := &State{
		Version: "1",
		Installed: []artifact.InstalledArtifact{
			{Artifact: artifact.Artifact{Name: "skill-1", Type: artifact.TypeSkill}},
			{Artifact: artifact.Artifact{Name: "cmd-1", Type: artifact.TypeCommand}},
		},
	}

	// Find existing
	found := state.FindInstalled("skill-1")
	if found == nil {
		t.Fatal("FindInstalled() returned nil, want skill-1")
	}
	if found.Name != "skill-1" {
		t.Errorf("Name = %v, want skill-1", found.Name)
	}

	// Find non-existent
	notFound := state.FindInstalled("nonexistent")
	if notFound != nil {
		t.Errorf("FindInstalled() = %v, want nil", notFound)
	}
}

func TestSaveState_AtomicWrite(t *testing.T) {
	tmpDir := t.TempDir()
	statePath := filepath.Join(tmpDir, "state.json")

	state := &State{
		Version: "1",
		Installed: []artifact.InstalledArtifact{
			{Artifact: artifact.Artifact{Name: "test", Type: artifact.TypeSkill}},
		},
	}

	// Save multiple times to verify atomic writes don't leave temp files
	for i := 0; i < 5; i++ {
		if err := SaveState(statePath, state); err != nil {
			t.Fatalf("SaveState() iteration %d error = %v", i, err)
		}
	}

	// Check no temp files remain
	entries, err := os.ReadDir(tmpDir)
	if err != nil {
		t.Fatalf("ReadDir() error = %v", err)
	}

	for _, entry := range entries {
		if filepath.Ext(entry.Name()) == ".tmp" {
			t.Errorf("Found leftover temp file: %s", entry.Name())
		}
	}
}

func TestSaveState_ConcurrentWrites(t *testing.T) {
	tmpDir := t.TempDir()
	statePath := filepath.Join(tmpDir, "state.json")

	// Run concurrent writes
	var wg sync.WaitGroup
	errors := make(chan error, 10)

	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func(n int) {
			defer wg.Done()
			state := &State{
				Version: "1",
				Installed: []artifact.InstalledArtifact{
					{Artifact: artifact.Artifact{
						Name: "skill-from-goroutine",
						Type: artifact.TypeSkill,
					}},
				},
			}
			if err := SaveState(statePath, state); err != nil {
				errors <- err
			}
		}(i)
	}

	wg.Wait()
	close(errors)

	// Check for errors
	for err := range errors {
		t.Errorf("Concurrent SaveState() error = %v", err)
	}

	// Verify file is valid
	state, err := LoadState(statePath)
	if err != nil {
		t.Fatalf("LoadState() after concurrent writes error = %v", err)
	}
	if state.Version != "1" {
		t.Errorf("Version = %v, want 1", state.Version)
	}

	// Clean up lock file if exists
	lockFile := statePath + ".lock"
	if _, err := os.Stat(lockFile); err == nil {
		t.Error("Lock file should be cleaned up after writes")
	}
}

func TestAcquireLock_BlocksUntilAvailable(t *testing.T) {
	tmpDir := t.TempDir()
	statePath := filepath.Join(tmpDir, "state.json")

	// First goroutine acquires lock
	unlock1, err := acquireLock(statePath)
	if err != nil {
		t.Fatalf("First acquireLock() error = %v", err)
	}

	// Second goroutine tries to acquire, should block
	done := make(chan bool)
	go func() {
		unlock2, err := acquireLock(statePath)
		if err != nil {
			t.Errorf("Second acquireLock() error = %v", err)
			done <- false
			return
		}
		unlock2()
		done <- true
	}()

	// Give second goroutine time to start waiting
	time.Sleep(100 * time.Millisecond)

	// Release first lock
	unlock1()

	// Second goroutine should now succeed
	select {
	case success := <-done:
		if !success {
			t.Error("Second lock acquisition failed")
		}
	case <-time.After(2 * time.Second):
		t.Error("Second lock acquisition timed out")
	}
}

func TestAcquireLock_StaleLock(t *testing.T) {
	tmpDir := t.TempDir()
	statePath := filepath.Join(tmpDir, "state.json")
	lockFile := lockPath(statePath)

	// Create a stale lock file
	if err := os.WriteFile(lockFile, []byte("12345"), 0644); err != nil {
		t.Fatalf("Failed to create lock file: %v", err)
	}

	// Set modification time to old (stale)
	oldTime := time.Now().Add(-10 * time.Second)
	if err := os.Chtimes(lockFile, oldTime, oldTime); err != nil {
		t.Fatalf("Failed to set lock file time: %v", err)
	}

	// Should acquire lock by removing stale lock
	unlock, err := acquireLock(statePath)
	if err != nil {
		t.Fatalf("acquireLock() error = %v", err)
	}
	defer unlock()

	// Verify we have the lock
	if _, err := os.Stat(lockFile); os.IsNotExist(err) {
		t.Error("Lock file should exist after acquiring lock")
	}
}
