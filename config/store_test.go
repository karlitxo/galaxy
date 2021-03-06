package config

import (
	"errors"
	"testing"
)

func NewTestStore() (*Store, *MemoryBackend) {
	r := &Store{}
	b := NewMemoryBackend()
	r.Backend = b
	return r, b
}

func TestAppNotExists(t *testing.T) {
	r, _ := NewTestStore()

	if exists, err := r.AppExists("foo", "dev"); exists || err != nil {
		t.Errorf("AppExists(%q) = %t, %v, want %t, %v",
			"foo", exists, err, false, nil)
	}
}

func TestGetNonExistentApp(t *testing.T) {
	r, _ := NewTestStore()
	_, err := r.GetApp("bad", "bogus")
	if err == nil {
		t.Errorf("GetApp() should have returned an error.  got nil")
	}
}

func TestAppExists(t *testing.T) {
	r, _ := NewTestStore()
	assertAppCreated(t, r, "app")
	assertAppExists(t, r, "app")
}

func TestListAssignmentKeyFormat(t *testing.T) {
	r, b := NewTestStore()

	b.MembersFunc = func(key string) ([]string, error) {
		if key != "dev/pools/foo" {
			t.Errorf("ListAssignments(%q) wrong key, want %s", key, "dev/pools/foo")
		}
		return []string{}, nil
	}

	r.ListAssignments("dev", "foo")
}

func TestListAssignmentsEmpty(t *testing.T) {
	r, _ := NewTestStore()

	assignments, err := r.ListAssignments("dev", "foo")
	if err != nil {
		t.Error(err)
	}

	if len(assignments) != 0 {
		t.Errorf("ListAssignments(%q) = %d, want %d", "foo", len(assignments), 0)
	}
}

func TestListAssignmentsNotEmpty(t *testing.T) {
	r, _ := NewTestStore()

	assertPoolCreated(t, r, "web")
	for _, k := range []string{"one", "two"} {
		assertAppCreated(t, r, k)
		if assigned, err := r.AssignApp(k, "dev", "web"); !assigned || err != nil {
			t.Fatalf("AssignApp(%q) = %t, %v, want %t, %v", k, assigned, err, true, nil)
		}
	}

	var assignments []string
	var err error
	if assignments, err = r.ListAssignments("dev", "web"); len(assignments) != 2 || err != nil {
		t.Fatalf("ListAssignments(%q) = %d, %v, want %d, %v", "web", len(assignments), err, 2, nil)
	}

	if assignments[0] != "one" {
		t.Fatalf("assignments[0] = %v, want %v", assignments[0], "one")
	}

	if assignments[1] != "two" {
		t.Fatalf("assignments[1] = %v, want %v", assignments[0], "two")
	}
}

func TestAssignAppNotExists(t *testing.T) {
	r, _ := NewTestStore()

	assigned, err := r.AssignApp("foo", "dev", "web")
	if assigned {
		t.Errorf("AssignApp(%q) = %t, want %t", "foo", assigned, false)
	}

	if err != nil {
		t.Error(err)
	}
}

func TestAssignAppPoolExists(t *testing.T) {
	r, _ := NewTestStore()

	assertAppCreated(t, r, "app")

	if assigned, err := r.AssignApp("app", "dev", "web"); !assigned || err != nil {
		t.Errorf("AssignApp(%q) = %t, %v, want %t, %v", "app", assigned, err,
			true, nil)
	}
}

func TestAssignAppAddMemberFail(t *testing.T) {
	r, b := NewTestStore()

	assertAppCreated(t, r, "app")
	assertPoolCreated(t, r, "web")

	b.AssignAppFunc = func(app, env, pool string) (bool, error) {
		return false, errors.New("something failed")
	}

	if assigned, err := r.AssignApp("app", "dev", "web"); assigned || err == nil {
		t.Errorf("AssignApp(%q) = %t, %v, want %t, %v", "app", assigned, err, false,
			errors.New("something failed"))
	}
}

func TestAssignAppNotifyFail(t *testing.T) {
	r, b := NewTestStore()

	assertAppCreated(t, r, "app")
	assertPoolCreated(t, r, "web")

	b.NotifyFunc = func(key, value string) (int, error) {
		return 0, errors.New("something failed")
	}

	if assigned, err := r.AssignApp("app", "dev", "web"); !assigned || err == nil {
		t.Errorf("AssignApp(%q) = %t, %v, want %t, %v", "app", assigned, err, false,
			errors.New("something failed"))
	}
}

func TestUnassignAppNotExists(t *testing.T) {
	r, _ := NewTestStore()

	if unassigned, err := r.UnassignApp("foo", "dev", "web"); unassigned || err != nil {
		t.Errorf("UnAssignApp(%q) = %t, %v, want %t, %v", "foo", unassigned, err, false, nil)
	}
}

func TestUnassignAppRemoveMemberFail(t *testing.T) {
	r, b := NewTestStore()

	assertAppCreated(t, r, "app")
	assertPoolCreated(t, r, "web")

	if assigned, err := r.AssignApp("app", "dev", "web"); !assigned || err != nil {
		t.Errorf("AssignApp(%q) = %t, %v, want %t, %v", "app", assigned, err, true, nil)
	}

	b.UnassignAppFunc = func(app, env, pool string) (bool, error) {
		return false, errors.New("something failed")
	}

	if unassigned, err := r.UnassignApp("foo", "dev", "web"); unassigned || err == nil {
		t.Errorf("UnAssignApp(%q) = %t, %v, want %t, %v", "foo", unassigned, err,
			false, errors.New("something failed"))
	}
}

func TestUnassignAppAddMemberNotifyRestart(t *testing.T) {
	r, b := NewTestStore()

	assertAppCreated(t, r, "app")
	assertPoolCreated(t, r, "web")

	if assigned, err := r.AssignApp("app", "dev", "web"); !assigned || err != nil {
		t.Errorf("AssignApp() = %t, %v, want %t, %v", assigned, err, true, nil)
	}

	b.NotifyFunc = func(key, value string) (int, error) {
		if key != "galaxy-dev" {
			t.Errorf("UnassignApp(%q) wrong notify key, want %s. got %s", "app", key, "galaxy-dev")
		}

		if value != "restart app" {
			t.Errorf("UnassignApp(%q) wrong notify value, want %s. got %s", "app", value, "restart app")
		}
		return 1, nil
	}
	if unassigned, err := r.UnassignApp("app", "dev", "web"); !unassigned || err != nil {
		t.Errorf("UnAssignApp(%q) = %t, %v, want %t, %v", "app", unassigned, err, true, nil)
	}
}

func TestUnassignAppNotifyFailed(t *testing.T) {
	r, b := NewTestStore()

	assertAppCreated(t, r, "app")
	assertPoolCreated(t, r, "web")

	if assigned, err := r.AssignApp("app", "dev", "web"); !assigned || err != nil {
		t.Errorf("AssignApp() = %t, %v, want %t, %v", assigned, err, true, nil)
	}

	b.NotifyFunc = func(key, value string) (int, error) {
		return 0, errors.New("something failed")
	}

	if unassigned, err := r.UnassignApp("app", "dev", "web"); !unassigned || err == nil {
		t.Errorf("UnAssignApp(%q) = %t, %v, want %t, %v", "app", unassigned, err, true, nil)
	}

}

func TestCreatePool(t *testing.T) {
	r, _ := NewTestStore()
	assertPoolCreated(t, r, "web")
}

func TestCreatePoolAddMemberFailedl(t *testing.T) {
	r, b := NewTestStore()
	b.CreatePoolFunc = func(env, pool string) (bool, error) {
		return false, errors.New("something failed")
	}

	if created, err := r.CreatePool("web", "dev"); created || err == nil {
		t.Errorf("CreatePool(%q) = %t, %v, want %t, %v", "web", created, err, true, nil)
	}
}

func TestDeletePool(t *testing.T) {
	r, _ := NewTestStore()

	assertPoolCreated(t, r, "web")

	if exists, err := r.PoolExists("dev", "web"); !exists || err != nil {
		t.Errorf("PoolExists()) = %t, %v, want %t, %v", exists, err, true, nil)
	}

	if deleted, err := r.DeletePool("web", "dev"); !deleted || err != nil {
		t.Errorf("DeletePool(%q) = %t, %v, want %t, %v", "web", deleted, err, true, nil)
	}
}

func TestDeletePoolHasAssignments(t *testing.T) {
	r, _ := NewTestStore()

	assertAppCreated(t, r, "app")
	assertPoolCreated(t, r, "web")

	// This is weird. AssignApp should probably take app & pool as params.
	if assigned, err := r.AssignApp("app", "dev", "web"); !assigned || err != nil {
		t.Errorf("AssignApp() = %t, %v, want %t, %v", assigned, err, true, nil)
	}

	// Should fail.  Can't delete a pool if apps are assigned
	if deleted, err := r.DeletePool("web", "dev"); deleted || err != nil {
		t.Errorf("DeletePool(%q) = %t, %v, want %t, %v", "web", deleted, err, false, nil)
	}
}

func TestListPools(t *testing.T) {
	r, _ := NewTestStore()

	for _, pool := range []string{"one", "two"} {
		assertPoolCreated(t, r, pool)
	}

	if pools, err := r.ListPools("dev"); len(pools) == 0 || err != nil {
		t.Errorf("ListPools() = %d, %v, want %d, %v", len(pools), err, 2, nil)
	}
}

func TestCreateApp(t *testing.T) {
	r, _ := NewTestStore()

	assertAppCreated(t, r, "app")
}

func TestCreateAppAlreadyExists(t *testing.T) {
	r, _ := NewTestStore()

	assertAppCreated(t, r, "app")

	if created, err := r.CreateApp("app", "dev"); created || err != nil {
		t.Fatalf("CreateApp() = %t, %v, want %t, %v",
			created, err,
			false, nil)
	}
}

func TestCreateAppError(t *testing.T) {
	r, b := NewTestStore()

	b.CreateAppFunc = func(app, env string) (bool, error) {
		return false, errors.New("something failed")
	}

	if created, err := r.CreateApp("foo", "dev"); created || err == nil {
		t.Fatalf("CreateApp() = %t, %v, want %t, %v",
			created, err,
			false, errors.New("something failed"))
	}
}

func TestDeleteApp(t *testing.T) {
	r, _ := NewTestStore()

	assertAppCreated(t, r, "app")
	assertAppExists(t, r, "app")

	if deleted, err := r.DeleteApp("app", "dev"); !deleted || err != nil {
		t.Fatalf("DeleteApp(%q) = %t, %v, want %t, %v", "app", deleted, err,
			true, nil)
	}
}

func TestDeleteAppStillAssigned(t *testing.T) {
	r, _ := NewTestStore()

	assertAppCreated(t, r, "app")
	assertAppExists(t, r, "app")
	assertPoolCreated(t, r, "web")

	if assigned, err := r.AssignApp("app", "dev", "web"); !assigned || err != nil {
		t.Fatalf("AssignApp(%q) = %t, %v, want %t, %v", "app", assigned, err,
			true, nil)
	}

	if deleted, err := r.DeleteApp("app", "dev"); deleted || err == nil {
		t.Fatalf("DeleteApp(%q) = %t, %v, want %t, %v", "app", deleted, err,
			false, errors.New("app is assigned to pool web"))
	}
}

func TestListApps(t *testing.T) {
	r, _ := NewTestStore()

	if apps, err := r.ListApps("dev"); len(apps) > 0 || err != nil {
		t.Fatalf("ListApps() = %d, %v, want %d, %v", len(apps), err,
			0, nil)
	}

	for _, k := range []string{"one", "two"} {
		assertAppCreated(t, r, k)
	}

	if apps, err := r.ListApps("dev"); len(apps) != 2 || err != nil {
		t.Fatalf("ListApps() = %d, %v, want %d, %v", len(apps), err,
			2, nil)
	}
}

func TestListAppsIgnoreSpecialKeys(t *testing.T) {
	r, b := NewTestStore()

	b.maps["dev/hosts/environment"] = make(map[string]string)

	if apps, err := r.ListApps("dev"); len(apps) > 0 || err != nil {
		t.Fatalf("ListApps() = %d, %v, want %d, %v", len(apps), err,
			0, nil)
	}
}

func TestListEnvs(t *testing.T) {
	r, b := NewTestStore()

	b.assignments["dev/web"] = []string{}
	b.assignments["prod/web"] = []string{}
	b.assignments["prod/batch"] = []string{}

	if apps, err := r.ListEnvs(); len(apps) != 2 || err != nil {
		t.Fatalf("ListApps() = %d, %v, want %d, %v", len(apps), err,
			2, nil)
	}
}

func assertAppCreated(t *testing.T, r *Store, app string) {
	if created, err := r.CreateApp(app, "dev"); !created || err != nil {
		t.Fatalf("CreateApp(%q) = %t, %v, want %t, %v", app,
			created, err,
			true, nil)
	}
}

func assertAppExists(t *testing.T, r *Store, app string) {
	if exists, err := r.AppExists(app, "dev"); !exists || err != nil {
		t.Fatalf("AppExists(%q) = %t, %v, want %t, %v", app, exists, err,
			true, nil)
	}
}

func assertPoolCreated(t *testing.T, r *Store, pool string) {
	if created, err := r.CreatePool(pool, "dev"); !created || err != nil {
		t.Errorf("CreatePool(%q) = %t, %v, want %t, %v", pool, created, err, true, nil)
	}
}
