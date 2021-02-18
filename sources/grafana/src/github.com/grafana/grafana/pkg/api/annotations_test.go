package api

import (
	"testing"

	"github.com/grafana/grafana/pkg/api/dtos"
	"github.com/grafana/grafana/pkg/bus"
	"github.com/grafana/grafana/pkg/models"
	"github.com/grafana/grafana/pkg/services/annotations"

	. "github.com/smartystreets/goconvey/convey"
)

func TestAnnotationsApiEndpoint(t *testing.T) {
	Convey("Given an annotation without a dashboard id", t, func() {
		cmd := dtos.PostAnnotationsCmd{
			Time: 1000,
			Text: "annotation text",
			Tags: []string{"tag1", "tag2"},
		}

		updateCmd := dtos.UpdateAnnotationsCmd{
			Time: 1000,
			Text: "annotation text",
			Tags: []string{"tag1", "tag2"},
		}

		patchCmd := dtos.PatchAnnotationsCmd{
			Time: 1000,
			Text: "annotation text",
			Tags: []string{"tag1", "tag2"},
		}

		Convey("When user is an Org Viewer", func() {
			role := models.ROLE_VIEWER
			Convey("Should not be allowed to save an annotation", func() {
				postAnnotationScenario("When calling POST on", "/api/annotations", "/api/annotations", role, cmd, func(sc *scenarioContext) {
					sc.fakeReqWithParams("POST", sc.url, map[string]string{}).exec()
					So(sc.resp.Code, ShouldEqual, 403)
				})

				putAnnotationScenario("When calling PUT on", "/api/annotations/1", "/api/annotations/:annotationId", role, updateCmd, func(sc *scenarioContext) {
					sc.fakeReqWithParams("PUT", sc.url, map[string]string{}).exec()
					So(sc.resp.Code, ShouldEqual, 403)
				})

				patchAnnotationScenario("When calling PATCH on", "/api/annotations/1", "/api/annotations/:annotationId", role, patchCmd, func(sc *scenarioContext) {
					sc.fakeReqWithParams("PATCH", sc.url, map[string]string{}).exec()
					So(sc.resp.Code, ShouldEqual, 403)
				})

				loggedInUserScenarioWithRole("When calling DELETE on", "DELETE", "/api/annotations/1", "/api/annotations/:annotationId", role, func(sc *scenarioContext) {
					sc.handlerFunc = DeleteAnnotationByID
					sc.fakeReqWithParams("DELETE", sc.url, map[string]string{}).exec()
					So(sc.resp.Code, ShouldEqual, 403)
				})
			})
		})

		Convey("When user is an Org Editor", func() {
			role := models.ROLE_EDITOR
			Convey("Should be able to save an annotation", func() {
				postAnnotationScenario("When calling POST on", "/api/annotations", "/api/annotations", role, cmd, func(sc *scenarioContext) {
					sc.fakeReqWithParams("POST", sc.url, map[string]string{}).exec()
					So(sc.resp.Code, ShouldEqual, 200)
				})

				putAnnotationScenario("When calling PUT on", "/api/annotations/1", "/api/annotations/:annotationId", role, updateCmd, func(sc *scenarioContext) {
					sc.fakeReqWithParams("PUT", sc.url, map[string]string{}).exec()
					So(sc.resp.Code, ShouldEqual, 200)
				})

				patchAnnotationScenario("When calling PATCH on", "/api/annotations/1", "/api/annotations/:annotationId", role, patchCmd, func(sc *scenarioContext) {
					sc.fakeReqWithParams("PATCH", sc.url, map[string]string{}).exec()
					So(sc.resp.Code, ShouldEqual, 200)
				})

				loggedInUserScenarioWithRole("When calling DELETE on", "DELETE", "/api/annotations/1", "/api/annotations/:annotationId", role, func(sc *scenarioContext) {
					sc.handlerFunc = DeleteAnnotationByID
					sc.fakeReqWithParams("DELETE", sc.url, map[string]string{}).exec()
					So(sc.resp.Code, ShouldEqual, 200)
				})
			})
		})
	})

	Convey("Given an annotation with a dashboard id and the dashboard does not have an acl", t, func() {
		cmd := dtos.PostAnnotationsCmd{
			Time:        1000,
			Text:        "annotation text",
			Tags:        []string{"tag1", "tag2"},
			DashboardId: 1,
			PanelId:     1,
		}

		updateCmd := dtos.UpdateAnnotationsCmd{
			Time: 1000,
			Text: "annotation text",
			Tags: []string{"tag1", "tag2"},
			Id:   1,
		}

		patchCmd := dtos.PatchAnnotationsCmd{
			Time: 8000,
			Text: "annotation text 50",
			Tags: []string{"foo", "bar"},
			Id:   1,
		}

		deleteCmd := dtos.DeleteAnnotationsCmd{
			DashboardId: 1,
			PanelId:     1,
		}

		viewerRole := models.ROLE_VIEWER
		editorRole := models.ROLE_EDITOR

		aclMockResp := []*models.DashboardAclInfoDTO{
			{Role: &viewerRole, Permission: models.PERMISSION_VIEW},
			{Role: &editorRole, Permission: models.PERMISSION_EDIT},
		}

		bus.AddHandler("test", func(query *models.GetDashboardAclInfoListQuery) error {
			query.Result = aclMockResp
			return nil
		})

		bus.AddHandler("test", func(query *models.GetTeamsByUserQuery) error {
			query.Result = []*models.TeamDTO{}
			return nil
		})

		Convey("When user is an Org Viewer", func() {
			role := models.ROLE_VIEWER
			Convey("Should not be allowed to save an annotation", func() {
				postAnnotationScenario("When calling POST on", "/api/annotations", "/api/annotations", role, cmd, func(sc *scenarioContext) {
					sc.fakeReqWithParams("POST", sc.url, map[string]string{}).exec()
					So(sc.resp.Code, ShouldEqual, 403)
				})

				putAnnotationScenario("When calling PUT on", "/api/annotations/1", "/api/annotations/:annotationId", role, updateCmd, func(sc *scenarioContext) {
					sc.fakeReqWithParams("PUT", sc.url, map[string]string{}).exec()
					So(sc.resp.Code, ShouldEqual, 403)
				})

				patchAnnotationScenario("When calling PATCH on", "/api/annotations/1", "/api/annotations/:annotationId", role, patchCmd, func(sc *scenarioContext) {
					sc.fakeReqWithParams("PATCH", sc.url, map[string]string{}).exec()
					So(sc.resp.Code, ShouldEqual, 403)
				})

				loggedInUserScenarioWithRole("When calling DELETE on", "DELETE", "/api/annotations/1", "/api/annotations/:annotationId", role, func(sc *scenarioContext) {
					sc.handlerFunc = DeleteAnnotationByID
					sc.fakeReqWithParams("DELETE", sc.url, map[string]string{}).exec()
					So(sc.resp.Code, ShouldEqual, 403)
				})
			})
		})

		Convey("When user is an Org Editor", func() {
			role := models.ROLE_EDITOR
			Convey("Should be able to save an annotation", func() {
				postAnnotationScenario("When calling POST on", "/api/annotations", "/api/annotations", role, cmd, func(sc *scenarioContext) {
					sc.fakeReqWithParams("POST", sc.url, map[string]string{}).exec()
					So(sc.resp.Code, ShouldEqual, 200)
				})

				putAnnotationScenario("When calling PUT on", "/api/annotations/1", "/api/annotations/:annotationId", role, updateCmd, func(sc *scenarioContext) {
					sc.fakeReqWithParams("PUT", sc.url, map[string]string{}).exec()
					So(sc.resp.Code, ShouldEqual, 200)
				})

				patchAnnotationScenario("When calling PATCH on", "/api/annotations/1", "/api/annotations/:annotationId", role, patchCmd, func(sc *scenarioContext) {
					sc.fakeReqWithParams("PATCH", sc.url, map[string]string{}).exec()
					So(sc.resp.Code, ShouldEqual, 200)
				})

				loggedInUserScenarioWithRole("When calling DELETE on", "DELETE", "/api/annotations/1", "/api/annotations/:annotationId", role, func(sc *scenarioContext) {
					sc.handlerFunc = DeleteAnnotationByID
					sc.fakeReqWithParams("DELETE", sc.url, map[string]string{}).exec()
					So(sc.resp.Code, ShouldEqual, 200)
				})
			})
		})

		Convey("When user is an Admin", func() {
			role := models.ROLE_ADMIN
			Convey("Should be able to do anything", func() {
				postAnnotationScenario("When calling POST on", "/api/annotations", "/api/annotations", role, cmd, func(sc *scenarioContext) {
					sc.fakeReqWithParams("POST", sc.url, map[string]string{}).exec()
					So(sc.resp.Code, ShouldEqual, 200)
				})

				putAnnotationScenario("When calling PUT on", "/api/annotations/1", "/api/annotations/:annotationId", role, updateCmd, func(sc *scenarioContext) {
					sc.fakeReqWithParams("PUT", sc.url, map[string]string{}).exec()
					So(sc.resp.Code, ShouldEqual, 200)
				})

				patchAnnotationScenario("When calling PATCH on", "/api/annotations/1", "/api/annotations/:annotationId", role, patchCmd, func(sc *scenarioContext) {
					sc.fakeReqWithParams("PATCH", sc.url, map[string]string{}).exec()
					So(sc.resp.Code, ShouldEqual, 200)
				})

				deleteAnnotationsScenario("When calling POST on", "/api/annotations/mass-delete", "/api/annotations/mass-delete", role, deleteCmd, func(sc *scenarioContext) {
					sc.fakeReqWithParams("POST", sc.url, map[string]string{}).exec()
					So(sc.resp.Code, ShouldEqual, 200)
				})
			})
		})
	})
}

type fakeAnnotationsRepo struct {
}

func (repo *fakeAnnotationsRepo) Delete(params *annotations.DeleteParams) error {
	return nil
}
func (repo *fakeAnnotationsRepo) Save(item *annotations.Item) error {
	item.Id = 1
	return nil
}
func (repo *fakeAnnotationsRepo) Update(item *annotations.Item) error {
	return nil
}
func (repo *fakeAnnotationsRepo) Find(query *annotations.ItemQuery) ([]*annotations.ItemDTO, error) {
	annotations := []*annotations.ItemDTO{{Id: 1}}
	return annotations, nil
}

var fakeAnnoRepo *fakeAnnotationsRepo

func postAnnotationScenario(desc string, url string, routePattern string, role models.RoleType, cmd dtos.PostAnnotationsCmd, fn scenarioFunc) {
	Convey(desc+" "+url, func() {
		defer bus.ClearBusHandlers()

		sc := setupScenarioContext(url)
		sc.defaultHandler = Wrap(func(c *models.ReqContext) Response {
			sc.context = c
			sc.context.UserId = TestUserID
			sc.context.OrgId = TestOrgID
			sc.context.OrgRole = role

			return PostAnnotation(c, cmd)
		})

		fakeAnnoRepo = &fakeAnnotationsRepo{}
		annotations.SetRepository(fakeAnnoRepo)

		sc.m.Post(routePattern, sc.defaultHandler)

		fn(sc)
	})
}

func putAnnotationScenario(desc string, url string, routePattern string, role models.RoleType, cmd dtos.UpdateAnnotationsCmd, fn scenarioFunc) {
	Convey(desc+" "+url, func() {
		defer bus.ClearBusHandlers()

		sc := setupScenarioContext(url)
		sc.defaultHandler = Wrap(func(c *models.ReqContext) Response {
			sc.context = c
			sc.context.UserId = TestUserID
			sc.context.OrgId = TestOrgID
			sc.context.OrgRole = role

			return UpdateAnnotation(c, cmd)
		})

		fakeAnnoRepo = &fakeAnnotationsRepo{}
		annotations.SetRepository(fakeAnnoRepo)

		sc.m.Put(routePattern, sc.defaultHandler)

		fn(sc)
	})
}

func patchAnnotationScenario(desc string, url string, routePattern string, role models.RoleType, cmd dtos.PatchAnnotationsCmd, fn scenarioFunc) {
	Convey(desc+" "+url, func() {
		defer bus.ClearBusHandlers()

		sc := setupScenarioContext(url)
		sc.defaultHandler = Wrap(func(c *models.ReqContext) Response {
			sc.context = c
			sc.context.UserId = TestUserID
			sc.context.OrgId = TestOrgID
			sc.context.OrgRole = role

			return PatchAnnotation(c, cmd)
		})

		fakeAnnoRepo = &fakeAnnotationsRepo{}
		annotations.SetRepository(fakeAnnoRepo)

		sc.m.Patch(routePattern, sc.defaultHandler)

		fn(sc)
	})
}

func deleteAnnotationsScenario(desc string, url string, routePattern string, role models.RoleType, cmd dtos.DeleteAnnotationsCmd, fn scenarioFunc) {
	Convey(desc+" "+url, func() {
		defer bus.ClearBusHandlers()

		sc := setupScenarioContext(url)
		sc.defaultHandler = Wrap(func(c *models.ReqContext) Response {
			sc.context = c
			sc.context.UserId = TestUserID
			sc.context.OrgId = TestOrgID
			sc.context.OrgRole = role

			return DeleteAnnotations(c, cmd)
		})

		fakeAnnoRepo = &fakeAnnotationsRepo{}
		annotations.SetRepository(fakeAnnoRepo)

		sc.m.Post(routePattern, sc.defaultHandler)

		fn(sc)
	})
}
