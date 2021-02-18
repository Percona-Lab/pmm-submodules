// +build integration

package sqlstore

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/grafana/grafana/pkg/components/simplejson"
	"github.com/grafana/grafana/pkg/models"
	"github.com/grafana/grafana/pkg/setting"
)

func TestDashboardSnapshotDBAccess(t *testing.T) {
	InitTestDB(t)

	origSecret := setting.SecretKey
	setting.SecretKey = "dashboard_snapshot_testing"
	t.Cleanup(func() {
		setting.SecretKey = origSecret
	})

	t.Run("Given saved snapshot", func(t *testing.T) {
		cmd := models.CreateDashboardSnapshotCommand{
			Key: "hej",
			Dashboard: simplejson.NewFromAny(map[string]interface{}{
				"hello": "mupp",
			}),
			UserId: 1000,
			OrgId:  1,
		}
		err := CreateDashboardSnapshot(&cmd)
		require.NoError(t, err)

		t.Run("Should be able to get snapshot by key", func(t *testing.T) {
			query := models.GetDashboardSnapshotQuery{Key: "hej"}
			err := GetDashboardSnapshot(&query)
			require.NoError(t, err)

			assert.NotNil(t, query.Result)

			dashboard, err := query.Result.DashboardJSON()
			require.NoError(t, err)
			assert.Equal(t, "mupp", dashboard.Get("hello").MustString())
		})

		t.Run("And the user has the admin role", func(t *testing.T) {
			query := models.GetDashboardSnapshotsQuery{
				OrgId:        1,
				SignedInUser: &models.SignedInUser{OrgRole: models.ROLE_ADMIN},
			}
			err := SearchDashboardSnapshots(&query)
			require.NoError(t, err)

			t.Run("Should return all the snapshots", func(t *testing.T) {
				assert.NotNil(t, query.Result)
				assert.Len(t, query.Result, 1)
			})
		})

		t.Run("And the user has the editor role and has created a snapshot", func(t *testing.T) {
			query := models.GetDashboardSnapshotsQuery{
				OrgId:        1,
				SignedInUser: &models.SignedInUser{OrgRole: models.ROLE_EDITOR, UserId: 1000},
			}
			err := SearchDashboardSnapshots(&query)
			require.NoError(t, err)

			t.Run("Should return all the snapshots", func(t *testing.T) {
				require.NotNil(t, query.Result)
				assert.Len(t, query.Result, 1)
			})
		})

		t.Run("And the user has the editor role and has not created any snapshot", func(t *testing.T) {
			query := models.GetDashboardSnapshotsQuery{
				OrgId:        1,
				SignedInUser: &models.SignedInUser{OrgRole: models.ROLE_EDITOR, UserId: 2},
			}
			err := SearchDashboardSnapshots(&query)
			require.NoError(t, err)

			t.Run("Should not return any snapshots", func(t *testing.T) {
				require.NotNil(t, query.Result)
				assert.Empty(t, query.Result)
			})
		})

		t.Run("And the user is anonymous", func(t *testing.T) {
			cmd := models.CreateDashboardSnapshotCommand{
				Key:       "strangesnapshotwithuserid0",
				DeleteKey: "adeletekey",
				Dashboard: simplejson.NewFromAny(map[string]interface{}{
					"hello": "mupp",
				}),
				UserId: 0,
				OrgId:  1,
			}
			err := CreateDashboardSnapshot(&cmd)
			require.NoError(t, err)

			t.Run("Should not return any snapshots", func(t *testing.T) {
				query := models.GetDashboardSnapshotsQuery{
					OrgId:        1,
					SignedInUser: &models.SignedInUser{OrgRole: models.ROLE_EDITOR, IsAnonymous: true, UserId: 0},
				}
				err := SearchDashboardSnapshots(&query)
				require.NoError(t, err)

				require.NotNil(t, query.Result)
				assert.Empty(t, query.Result)
			})
		})

		t.Run("Should have encrypted dashboard data", func(t *testing.T) {
			original, err := cmd.Dashboard.Encode()
			require.NoError(t, err)

			decrypted, err := cmd.Result.DashboardEncrypted.Decrypt()
			require.NoError(t, err)

			require.Equal(t, decrypted, original)
		})
	})
}

func TestDeleteExpiredSnapshots(t *testing.T) {
	sqlstore := InitTestDB(t)

	t.Run("Testing dashboard snapshots clean up", func(t *testing.T) {
		setting.SnapShotRemoveExpired = true

		nonExpiredSnapshot := createTestSnapshot(t, sqlstore, "key1", 48000)
		createTestSnapshot(t, sqlstore, "key2", -1200)
		createTestSnapshot(t, sqlstore, "key3", -1200)

		err := DeleteExpiredSnapshots(&models.DeleteExpiredSnapshotsCommand{})
		require.NoError(t, err)

		query := models.GetDashboardSnapshotsQuery{
			OrgId:        1,
			SignedInUser: &models.SignedInUser{OrgRole: models.ROLE_ADMIN},
		}
		err = SearchDashboardSnapshots(&query)
		require.NoError(t, err)

		assert.Len(t, query.Result, 1)
		assert.Equal(t, nonExpiredSnapshot.Key, query.Result[0].Key)

		err = DeleteExpiredSnapshots(&models.DeleteExpiredSnapshotsCommand{})
		require.NoError(t, err)

		query = models.GetDashboardSnapshotsQuery{
			OrgId:        1,
			SignedInUser: &models.SignedInUser{OrgRole: models.ROLE_ADMIN},
		}
		err = SearchDashboardSnapshots(&query)
		require.NoError(t, err)

		require.Len(t, query.Result, 1)
		require.Equal(t, nonExpiredSnapshot.Key, query.Result[0].Key)
	})
}

func createTestSnapshot(t *testing.T, sqlstore *SqlStore, key string, expires int64) *models.DashboardSnapshot {
	cmd := models.CreateDashboardSnapshotCommand{
		Key:       key,
		DeleteKey: "delete" + key,
		Dashboard: simplejson.NewFromAny(map[string]interface{}{
			"hello": "mupp",
		}),
		UserId:  1000,
		OrgId:   1,
		Expires: expires,
	}
	err := CreateDashboardSnapshot(&cmd)
	require.NoError(t, err)

	// Set expiry date manually - to be able to create expired snapshots
	if expires < 0 {
		expireDate := time.Now().Add(time.Second * time.Duration(expires))
		_, err = sqlstore.engine.Exec("UPDATE dashboard_snapshot SET expires = ? WHERE id = ?", expireDate, cmd.Result.Id)
		require.NoError(t, err)
	}

	return cmd.Result
}
