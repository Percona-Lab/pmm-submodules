package dashboards

import (
	"fmt"
	"testing"

	"github.com/grafana/grafana/pkg/infra/log"
	"github.com/grafana/grafana/pkg/setting"

	"github.com/grafana/grafana/pkg/bus"
	"github.com/grafana/grafana/pkg/models"
	"github.com/grafana/grafana/pkg/services/guardian"
	. "github.com/smartystreets/goconvey/convey"
)

func TestDashboardService(t *testing.T) {
	Convey("Dashboard service tests", t, func() {
		bus.ClearBusHandlers()

		service := &dashboardServiceImpl{
			log: log.New("test.logger"),
		}

		origNewDashboardGuardian := guardian.New
		guardian.MockDashboardGuardian(&guardian.FakeDashboardGuardian{CanSaveValue: true})

		Convey("Save dashboard validation", func() {
			dto := &SaveDashboardDTO{}

			Convey("When saving a dashboard with empty title it should return error", func() {
				titles := []string{"", " ", "   \t   "}

				for _, title := range titles {
					dto.Dashboard = models.NewDashboard(title)
					_, err := service.SaveDashboard(dto, false)
					So(err, ShouldEqual, models.ErrDashboardTitleEmpty)
				}
			})

			Convey("Should return validation error if it's a folder and have a folder id", func() {
				dto.Dashboard = models.NewDashboardFolder("Folder")
				dto.Dashboard.FolderId = 1
				_, err := service.SaveDashboard(dto, false)
				So(err, ShouldEqual, models.ErrDashboardFolderCannotHaveParent)
			})

			Convey("Should return validation error if folder is named General", func() {
				dto.Dashboard = models.NewDashboardFolder("General")
				_, err := service.SaveDashboard(dto, false)
				So(err, ShouldEqual, models.ErrDashboardFolderNameExists)
			})

			Convey("When saving a dashboard should validate uid", func() {
				bus.AddHandler("test", func(cmd *models.ValidateDashboardAlertsCommand) error {
					return nil
				})

				bus.AddHandler("test", func(cmd *models.ValidateDashboardBeforeSaveCommand) error {
					cmd.Result = &models.ValidateDashboardBeforeSaveResult{}
					return nil
				})

				bus.AddHandler("test", func(cmd *models.GetProvisionedDashboardDataByIdQuery) error {
					cmd.Result = nil
					return nil
				})

				testCases := []struct {
					Uid   string
					Error error
				}{
					{Uid: "", Error: nil},
					{Uid: "   ", Error: nil},
					{Uid: "  \t  ", Error: nil},
					{Uid: "asdf90_-", Error: nil},
					{Uid: "asdf/90", Error: models.ErrDashboardInvalidUid},
					{Uid: "   asdfghjklqwertyuiopzxcvbnmasdfghjklqwer   ", Error: nil},
					{Uid: "asdfghjklqwertyuiopzxcvbnmasdfghjklqwertyuiopzxcvbnmasdfghjklqwertyuiopzxcvbnm", Error: models.ErrDashboardUidTooLong},
				}

				for _, tc := range testCases {
					dto.Dashboard = models.NewDashboard("title")
					dto.Dashboard.SetUid(tc.Uid)
					dto.User = &models.SignedInUser{}

					_, err := service.buildSaveDashboardCommand(dto, true, false)
					So(err, ShouldEqual, tc.Error)
				}
			})

			Convey("Should return validation error if dashboard is provisioned", func() {
				provisioningValidated := false
				bus.AddHandler("test", func(cmd *models.GetProvisionedDashboardDataByIdQuery) error {
					provisioningValidated = true
					cmd.Result = &models.DashboardProvisioning{}
					return nil
				})

				bus.AddHandler("test", func(cmd *models.ValidateDashboardAlertsCommand) error {
					return nil
				})

				bus.AddHandler("test", func(cmd *models.ValidateDashboardBeforeSaveCommand) error {
					cmd.Result = &models.ValidateDashboardBeforeSaveResult{}
					return nil
				})

				dto.Dashboard = models.NewDashboard("Dash")
				dto.Dashboard.SetId(3)
				dto.User = &models.SignedInUser{UserId: 1}
				_, err := service.SaveDashboard(dto, false)
				So(provisioningValidated, ShouldBeTrue)
				So(err, ShouldEqual, models.ErrDashboardCannotSaveProvisionedDashboard)
			})

			Convey("Should not return validation error if dashboard is provisioned but UI updates allowed", func() {
				provisioningValidated := false
				bus.AddHandler("test", func(cmd *models.GetProvisionedDashboardDataByIdQuery) error {
					provisioningValidated = true
					cmd.Result = &models.DashboardProvisioning{}
					return nil
				})

				bus.AddHandler("test", func(cmd *models.ValidateDashboardAlertsCommand) error {
					return nil
				})

				bus.AddHandler("test", func(cmd *models.ValidateDashboardBeforeSaveCommand) error {
					cmd.Result = &models.ValidateDashboardBeforeSaveResult{}
					return nil
				})

				dto.Dashboard = models.NewDashboard("Dash")
				dto.Dashboard.SetId(3)
				dto.User = &models.SignedInUser{UserId: 1}
				_, err := service.SaveDashboard(dto, true)
				So(provisioningValidated, ShouldBeFalse)
				So(err, ShouldNotBeNil)
			})

			Convey("Should return validation error if alert data is invalid", func() {
				bus.AddHandler("test", func(cmd *models.GetProvisionedDashboardDataByIdQuery) error {
					cmd.Result = nil
					return nil
				})

				bus.AddHandler("test", func(cmd *models.ValidateDashboardAlertsCommand) error {
					return fmt.Errorf("alert validation error")
				})

				dto.Dashboard = models.NewDashboard("Dash")
				_, err := service.SaveDashboard(dto, false)
				So(err.Error(), ShouldEqual, "alert validation error")
			})
		})

		Convey("Save provisioned dashboard validation", func() {
			dto := &SaveDashboardDTO{}

			Convey("Should not return validation error if dashboard is provisioned", func() {
				provisioningValidated := false
				bus.AddHandler("test", func(cmd *models.GetProvisionedDashboardDataByIdQuery) error {
					provisioningValidated = true
					cmd.Result = &models.DashboardProvisioning{}
					return nil
				})

				bus.AddHandler("test", func(cmd *models.ValidateDashboardAlertsCommand) error {
					return nil
				})

				bus.AddHandler("test", func(cmd *models.ValidateDashboardBeforeSaveCommand) error {
					cmd.Result = &models.ValidateDashboardBeforeSaveResult{}
					return nil
				})

				bus.AddHandler("test", func(cmd *models.SaveProvisionedDashboardCommand) error {
					return nil
				})

				bus.AddHandler("test", func(cmd *models.UpdateDashboardAlertsCommand) error {
					return nil
				})

				dto.Dashboard = models.NewDashboard("Dash")
				dto.Dashboard.SetId(3)
				dto.User = &models.SignedInUser{UserId: 1}
				_, err := service.SaveProvisionedDashboard(dto, nil)
				So(err, ShouldBeNil)
				So(provisioningValidated, ShouldBeFalse)
			})

			Convey("Should override invalid refresh interval if dashboard is provisioned", func() {
				oldRefreshInterval := setting.MinRefreshInterval
				setting.MinRefreshInterval = "5m"
				defer func() { setting.MinRefreshInterval = oldRefreshInterval }()

				bus.AddHandler("test", func(cmd *models.GetProvisionedDashboardDataByIdQuery) error {
					cmd.Result = &models.DashboardProvisioning{}
					return nil
				})

				bus.AddHandler("test", func(cmd *models.ValidateDashboardAlertsCommand) error {
					return nil
				})

				bus.AddHandler("test", func(cmd *models.ValidateDashboardBeforeSaveCommand) error {
					cmd.Result = &models.ValidateDashboardBeforeSaveResult{}
					return nil
				})

				bus.AddHandler("test", func(cmd *models.SaveProvisionedDashboardCommand) error {
					return nil
				})

				bus.AddHandler("test", func(cmd *models.UpdateDashboardAlertsCommand) error {
					return nil
				})

				dto.Dashboard = models.NewDashboard("Dash")
				dto.Dashboard.SetId(3)
				dto.User = &models.SignedInUser{UserId: 1}
				dto.Dashboard.Data.Set("refresh", "1s")
				_, err := service.SaveProvisionedDashboard(dto, nil)
				So(err, ShouldBeNil)
				So(dto.Dashboard.Data.Get("refresh").MustString(), ShouldEqual, "5m")
			})
		})

		Convey("Import dashboard validation", func() {
			dto := &SaveDashboardDTO{}

			Convey("Should return validation error if dashboard is provisioned", func() {
				provisioningValidated := false
				bus.AddHandler("test", func(cmd *models.GetProvisionedDashboardDataByIdQuery) error {
					provisioningValidated = true
					cmd.Result = &models.DashboardProvisioning{}
					return nil
				})

				bus.AddHandler("test", func(cmd *models.ValidateDashboardAlertsCommand) error {
					return nil
				})

				bus.AddHandler("test", func(cmd *models.ValidateDashboardBeforeSaveCommand) error {
					cmd.Result = &models.ValidateDashboardBeforeSaveResult{}
					return nil
				})

				bus.AddHandler("test", func(cmd *models.SaveProvisionedDashboardCommand) error {
					return nil
				})

				bus.AddHandler("test", func(cmd *models.UpdateDashboardAlertsCommand) error {
					return nil
				})

				dto.Dashboard = models.NewDashboard("Dash")
				dto.Dashboard.SetId(3)
				dto.User = &models.SignedInUser{UserId: 1}
				_, err := service.ImportDashboard(dto)
				So(provisioningValidated, ShouldBeTrue)
				So(err, ShouldEqual, models.ErrDashboardCannotSaveProvisionedDashboard)
			})
		})

		Convey("Given provisioned dashboard", func() {
			result := setupDeleteHandlers(true)

			Convey("DeleteProvisionedDashboard should delete it", func() {
				err := service.DeleteProvisionedDashboard(1, 1)
				So(err, ShouldBeNil)
				So(result.deleteWasCalled, ShouldBeTrue)
			})

			Convey("DeleteDashboard should fail to delete it", func() {
				err := service.DeleteDashboard(1, 1)
				So(err, ShouldEqual, models.ErrDashboardCannotDeleteProvisionedDashboard)
				So(result.deleteWasCalled, ShouldBeFalse)
			})
		})

		Convey("Given non provisioned dashboard", func() {
			result := setupDeleteHandlers(false)

			Convey("DeleteProvisionedDashboard should delete it", func() {
				err := service.DeleteProvisionedDashboard(1, 1)
				So(err, ShouldBeNil)
				So(result.deleteWasCalled, ShouldBeTrue)
			})

			Convey("DeleteDashboard should delete it", func() {
				err := service.DeleteDashboard(1, 1)
				So(err, ShouldBeNil)
				So(result.deleteWasCalled, ShouldBeTrue)
			})
		})

		Reset(func() {
			guardian.New = origNewDashboardGuardian
		})
	})
}

type Result struct {
	deleteWasCalled bool
}

func setupDeleteHandlers(provisioned bool) *Result {
	bus.AddHandler("test", func(cmd *models.GetProvisionedDashboardDataByIdQuery) error {
		if provisioned {
			cmd.Result = &models.DashboardProvisioning{}
		} else {
			cmd.Result = nil
		}
		return nil
	})

	result := &Result{}
	bus.AddHandler("test", func(cmd *models.DeleteDashboardCommand) error {
		So(cmd.Id, ShouldEqual, 1)
		So(cmd.OrgId, ShouldEqual, 1)
		result.deleteWasCalled = true
		return nil
	})

	return result
}
