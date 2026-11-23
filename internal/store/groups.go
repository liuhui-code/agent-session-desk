package store

import (
	"context"
	"database/sql"
	"encoding/hex"
	"errors"
	"fmt"
	"math"
	"sort"

	"github.com/liuhui-code/agent-session-desk/internal/session"
)

const groupSchema = `
CREATE TABLE groups(scope TEXT NOT NULL,id TEXT NOT NULL,parent_id TEXT,name TEXT NOT NULL CHECK(length(name) BETWEEN 1 AND 50 AND length(CAST(name AS BLOB))<=200),default_dir TEXT NOT NULL CHECK(length(CAST(default_dir AS BLOB))<=4096),revision INTEGER NOT NULL CHECK(revision>0),ordinal INTEGER NOT NULL CHECK(ordinal>=0),PRIMARY KEY(scope,id),FOREIGN KEY(scope,parent_id) REFERENCES groups(scope,id),CHECK(id='default' OR (length(id)=32 AND id NOT GLOB '*[^0-9a-f]*')),CHECK(id<>'default' OR (parent_id IS NULL AND ordinal=0)));
CREATE UNIQUE INDEX groups_sibling_name ON groups(scope,coalesce(parent_id,''),name);
CREATE TABLE session_groups(scope TEXT NOT NULL,session_id TEXT NOT NULL,group_id TEXT NOT NULL,revision INTEGER NOT NULL CHECK(revision>0),PRIMARY KEY(scope,session_id),FOREIGN KEY(scope,session_id) REFERENCES sessions(scope,id),FOREIGN KEY(scope,group_id) REFERENCES groups(scope,id));
INSERT INTO groups SELECT DISTINCT scope,'default',NULL,'我的会话','',1,0 FROM sessions;
`

var ErrGroupConflict = errors.New("group or membership changed; refresh before editing")

const groupColumns = `SELECT scope,id,parent_id,name,default_dir,revision,ordinal FROM groups LIMIT 0`
const membershipColumns = `SELECT scope,session_id,group_id,revision FROM session_groups LIMIT 0`

func validGroupID(id string) bool {
	if id == session.DefaultGroupID {
		return true
	}
	if len(id) != 32 {
		return false
	}
	_, err := hex.DecodeString(id)
	return err == nil
}

type rowsReader interface {
	rowReader
	QueryContext(context.Context, string, ...any) (*sql.Rows, error)
}

func (s *Store) groups(ctx context.Context, q rowsReader) ([]session.Group, error) {
	rows, err := q.QueryContext(ctx, `SELECT id,coalesce(parent_id,''),name,default_dir,revision,ordinal FROM groups WHERE scope=? ORDER BY ordinal,id LIMIT ?`, s.scope, session.MaxGroups+1)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var result []session.Group
	for rows.Next() {
		var g session.Group
		if err = rows.Scan(&g.ID, &g.ParentID, &g.Name, &g.DefaultDir, &g.Revision, &g.Order); err != nil {
			return nil, err
		}
		result = append(result, g)
	}
	if err = rows.Err(); err != nil {
		return nil, err
	}
	if err = validateGroups(result); err != nil {
		return nil, err
	}
	byParent := map[string][]session.Group{}
	for _, g := range result {
		byParent[g.ParentID] = append(byParent[g.ParentID], g)
	}
	result = nil
	var visit func(string)
	visit = func(parent string) {
		children := byParent[parent]
		sort.SliceStable(children, func(i, j int) bool {
			if children[i].ID == session.DefaultGroupID || children[j].ID == session.DefaultGroupID {
				return children[i].ID == session.DefaultGroupID && children[j].ID != session.DefaultGroupID
			}
			if children[i].Order == children[j].Order {
				return children[i].ID < children[j].ID
			}
			return children[i].Order < children[j].Order
		})
		for _, g := range children {
			result = append(result, g)
			visit(g.ID)
		}
	}
	visit("")
	return result, nil
}

func validateGroups(groups []session.Group) error {
	if len(groups) == 0 || len(groups) > session.MaxGroups {
		return fmt.Errorf("group count outside defensive bounds")
	}
	byID := map[string]session.Group{}
	for _, g := range groups {
		name, err := session.NormalizeGroupName(g.Name)
		if err != nil || name != g.Name || !validGroupID(g.ID) || g.Revision < 1 || g.Order < 0 {
			return fmt.Errorf("invalid stored group")
		}
		dir, err := session.NormalizeGroupDir(g.DefaultDir)
		if err != nil || dir != g.DefaultDir {
			return fmt.Errorf("invalid stored group directory")
		}
		if _, exists := byID[g.ID]; exists {
			return fmt.Errorf("duplicate group identity")
		}
		byID[g.ID] = g
	}
	def, ok := byID[session.DefaultGroupID]
	if !ok || def.ParentID != "" || def.Order != 0 {
		return fmt.Errorf("default group missing or moved")
	}
	for _, g := range groups {
		seen := map[string]bool{}
		id := g.ID
		for depth := 0; id != ""; depth++ {
			if depth >= session.MaxGroupDepth || seen[id] {
				return fmt.Errorf("invalid group ancestry")
			}
			seen[id] = true
			parent, ok := byID[id]
			if !ok {
				return fmt.Errorf("group parent missing")
			}
			id = parent.ParentID
		}
	}
	return nil
}

func (s *Store) Groups(ctx context.Context) ([]session.Group, error) { return s.groups(ctx, s.db) }

func (s *Store) CreateGroup(ctx context.Context, parentID, name, dir string) (session.Group, error) {
	name, err := session.NormalizeGroupName(name)
	if err != nil {
		return session.Group{}, err
	}
	dir, err = session.NormalizeGroupDir(dir)
	if err != nil {
		return session.Group{}, err
	}
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return session.Group{}, err
	}
	defer tx.Rollback()
	groups, err := s.groups(ctx, tx)
	if err != nil {
		return session.Group{}, err
	}
	var order int64
	for _, g := range groups {
		if g.ParentID == parentID {
			if g.Name == name {
				if g.DefaultDir != dir {
					return session.Group{}, ErrGroupConflict
				}
				return g, nil
			}
			if g.Order == math.MaxInt64 {
				return session.Group{}, fmt.Errorf("group order exhausted")
			}
			order = max(order, g.Order+1)
		}
	}
	if len(groups) >= session.MaxGroups {
		return session.Group{}, fmt.Errorf("group count exceeds defensive bound")
	}
	id, err := session.NewID()
	if err != nil {
		return session.Group{}, err
	}
	g := session.Group{ID: id, ParentID: parentID, Name: name, DefaultDir: dir, Revision: 1, Order: order}
	if err = validateGroups(append(groups, g)); err != nil {
		return session.Group{}, err
	}
	var parent any
	if parentID != "" {
		parent = parentID
	}
	if _, err = tx.ExecContext(ctx, `INSERT INTO groups VALUES(?,?,?,?,?,?,?)`, s.scope, g.ID, parent, g.Name, g.DefaultDir, g.Revision, g.Order); err != nil {
		return session.Group{}, err
	}
	if err = tx.Commit(); err != nil {
		return session.Group{}, fmt.Errorf("group commit not confirmed; refresh before another create: %w", err)
	}
	return g, nil
}
